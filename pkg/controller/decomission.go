// Copyright (C) 2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7/pkg/set"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// checkForPoolDecommission validates the spec of the tenant and it's status to detect a pool being removed
func (c *Controller) checkForPoolDecommission(ctx context.Context, key string, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (*miniov2.Tenant, error) {
	var err error
	// duplicate status.pools first
	haveDuplicateStatusPools := false
	distinctStatusPoolsMap := map[string]struct{}{}
	distinctStatusPools := []miniov2.PoolStatus{}
	for _, pool := range tenant.Status.Pools {
		if _, ok := distinctStatusPoolsMap[pool.SSName]; !ok {
			distinctStatusPoolsMap[pool.SSName] = struct{}{}
			distinctStatusPools = append(distinctStatusPools, *pool.DeepCopy())
		} else {
			haveDuplicateStatusPools = true
		}
	}
	tenant.Status.Pools = distinctStatusPools
	if haveDuplicateStatusPools {
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusNotOwned, 0); err != nil {
			return nil, err
		}
	}

	// if the number of pools in the spec is less that what we know in the status, a decomission is taking place
	if len(tenant.Status.Pools) > len(tenant.Spec.Pools) {
		// check for empty pool names
		var noDecom bool
		for _, pool := range tenant.Spec.Pools {
			if pool.Name != "" {
				continue
			} // pool.Name empty decommission is not allowed.
			noDecom = true
			break
		}
		if noDecom {
			klog.Warningf("%s Detected we are removing a pool but spec.Pool[].Name is empty - disallowing removal", key)
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusDecommissioningNotAllowed, 0); err != nil {
				return nil, err
			}
			return nil, errors.New("removing pool not allowed")
		}
		// Check for duplicate names
		var noDecomCommon bool
		commonNames := set.NewStringSet()
		for _, pool := range tenant.Spec.Pools {
			if commonNames.Contains(pool.Name) {
				noDecomCommon = true
				break
			}
			commonNames.Add(pool.Name)
		}
		if noDecomCommon {
			klog.Warningf("%s Detected we are removing a pool but spec.Pool[].Name's are duplicated - disallowing removal", key)
			return nil, errors.New("removing pool not allowed")
		}

		klog.Infof("%s Detected we are removing a pool", key)
		// This means we are attempting to remove a "pool", perhaps after a decommission event.
		var poolNamesRemoved []string
		var initializedPool miniov2.Pool
		for i, pstatus := range tenant.Status.Pools {
			var found bool
			for _, pool := range tenant.Spec.Pools {
				if pstatus.SSName == tenant.PoolStatefulsetName(&pool) {
					found = true
					if pstatus.State == miniov2.PoolInitialized {
						initializedPool = pool
					}
					continue
				}
			}
			if !found {
				poolNamesRemoved = append(poolNamesRemoved, pstatus.SSName)
				tenant.Status.Pools = append(tenant.Status.Pools[:i], tenant.Status.Pools[i+1:]...)
			}
		}

		var restarted bool
		// Only restart if there is an initialized pool to fetch the new args.
		if len(poolNamesRemoved) > 0 && initializedPool.Name != "" {
			// Restart services to get new args since we are shrinking the deployment here.
			if err := c.restartInitializedPool(ctx, tenant, initializedPool, tenantConfiguration); err != nil {
				return nil, err
			}
			metaNowTime := metav1.Now()
			tenant.Status.WaitingOnReady = &metaNowTime
			tenant.Status.CurrentState = StatusRestartingMinIO
			if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
				klog.Infof("'%s' Can't update tenant status: %v", key, err)
				return nil, err
			}
			klog.Infof("'%s' was restarted", key)
			restarted = true
		}

		for _, ssName := range poolNamesRemoved {
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "PoolRemoved", fmt.Sprintf("Tenant pool %s removed", ssName))
			if err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Delete(ctx, ssName, metav1.DeleteOptions{}); err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, err
			}
		}

		if restarted {
			return nil, ErrMinIORestarting
		}

		return tenant, nil
	}
	return tenant, err
}
