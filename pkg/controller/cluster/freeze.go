// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"context"
	"errors"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) checkFreezeState(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, bool, error) {
	// if the tenant is market for freezing, and it's not a fresh setup, freeze it.
	if tenant.Spec.Freeze != nil && *tenant.Spec.Freeze {
		var zero int32
		// if it's not already frozen, freeze tenant
		if tenant.Status.CurrentState != StatusFrozen {
			// Check if we need to create any of the pools. It's important not to update the statefulsets
			// in this loop because we need all the pools "as they are" for the hot-update below
			for i := range tenant.Spec.Pools {
				// Get the StatefulSet with the name specified in Tenant.status.pools[i].SSName

				// if this index is in the status of pools use it, else capture the desired name in the status and store it
				var ssName string
				if len(tenant.Status.Pools) > i {
					ssName = tenant.Status.Pools[i].SSName
				} else {
					return nil, true, errors.New("invalid number of pools in status")
				}
				ss, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(ssName)
				if err != nil {
					return nil, false, err
				}

				ss.Spec.Replicas = &zero
				_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ss, metav1.UpdateOptions{})
				if err != nil {
					return nil, false, err
				}

			}
			if tenant, err := c.updateTenantStatus(ctx, tenant, StatusFrozen, 0); err != nil {
				return tenant, false, err
			}

			// freeze prometheus if any
			if tenant.HasPrometheusEnabled() {
				promSs, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.PrometheusStatefulsetName())
				if err != nil && !k8serrors.IsNotFound(err) {
					return tenant, false, err
				}
				promSs.Spec.Replicas = &zero
				_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, promSs, metav1.UpdateOptions{})
				if err != nil {
					return tenant, false, err
				}
			}
			// freeze log search
			if tenant.HasLogEnabled() {
				// freeze statefulset
				logSs, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.LogStatefulsetName())
				if err != nil && !k8serrors.IsNotFound(err) {
					return tenant, false, err
				}
				logSs.Spec.Replicas = &zero
				_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, logSs, metav1.UpdateOptions{})
				if err != nil {
					return tenant, false, err
				}
				// freeze deployment
				logSearchDeployment, err := c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Get(ctx, tenant.LogSearchAPIDeploymentName(), metav1.GetOptions{})
				if err != nil {
					return tenant, false, err
				}
				logSearchDeployment.Spec.Replicas = &zero
				_, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Update(ctx, logSearchDeployment, metav1.UpdateOptions{})
				if err != nil {
					return tenant, false, err
				}

			}

		}

		return tenant, true, nil
	}
	// if status is frozen, but we reach here, we should remove this status
	if tenant.Spec.Freeze != nil && !*tenant.Spec.Freeze && tenant.Status.CurrentState == StatusFrozen {
		if tenant, err := c.updateTenantStatus(ctx, tenant, StatusInitialized, 0); err != nil {
			return tenant, false, err
		}
	}
	// return no error, and not complete, the pool sync loop will unfreeze the pools
	return tenant, false, nil
}
