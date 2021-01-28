// This file is part of MinIO Operator
// Copyright (c) 2020 MinIO, Inc.
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
	"fmt"
	"reflect"
	"strings"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/statefulsets"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func (c *Controller) getSSForPool(tenant *miniov2.Tenant, pool *miniov2.Pool) (*appsv1.StatefulSet, error) {
	ss, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.PoolStatefulsetName(pool))
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}

		// check if there are legacy statefulsets
		ss, err = c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.LegacyStatefulsetName(pool))
		if err != nil {
			return nil, err
		}
		// Update the name of the pool
		pool.Name = strings.Replace(ss.Name, fmt.Sprintf("%s-", tenant.Name), "", 1)
	}
	return ss, nil
}

func (c *Controller) getAllSSForTenant(tenant *miniov2.Tenant) (map[int]*appsv1.StatefulSet, error) {
	poolDir := make(map[int]*appsv1.StatefulSet)
	// TODO: Load all statefulsets by using the tenant label in a single list call
	for i := range tenant.Spec.Pools {
		ss, err := c.getSSForPool(tenant, &tenant.Spec.Pools[i])
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}
		if ss != nil {
			poolDir[i] = ss
		}
	}
	return poolDir, nil
}

// poolSSMatchesSpec checks if the statefulset for the pool matches what is expected and described from the Tenant
func poolSSMatchesSpec(tenant *miniov2.Tenant, pool *miniov2.Pool, ss *appsv1.StatefulSet, opVersion string) (bool, error) {
	// Verify Resources
	poolMatchesSS := true
	if pool.Resources.String() != ss.Spec.Template.Spec.Containers[0].Resources.String() {
		klog.V(4).Infof("resource requirements updates for pool %s", pool.Name)
		poolMatchesSS = false
	}
	// Verify Affinity clauses
	if pool.Affinity.String() != ss.Spec.Template.Spec.Affinity.String() {
		klog.V(4).Infof("affinity update for pool %s", pool.Name)
		poolMatchesSS = false
	}
	// Verify all sidecars
	if tenant.Spec.SideCars != nil {
		if len(ss.Spec.Template.Spec.Containers) != len(tenant.Spec.SideCars.Containers)+1 {
			klog.V(4).Infof("Side cars for pool %s don't match", pool.Name)
			poolMatchesSS = false
		}
		// compare each container spec to the sidecars (shifted by one as container 0 is MinIO)
		for i := 1; i < len(ss.Spec.Template.Spec.Containers); i++ {
			if !equality.Semantic.DeepDerivative(ss.Spec.Template.Spec.Containers[i], tenant.Spec.SideCars.Containers[i-1]) {
				// container doesn't match
				poolMatchesSS = false
				break
			}
		}
	}
	if tenant.Spec.SideCars == nil && len(ss.Spec.Template.Spec.Containers) > 1 {
		klog.V(4).Infof("Side cars  removed for pool %s", pool.Name)
		poolMatchesSS = false
	}

	// Try to detect changes in the labels or annotations
	expectedMetadata := statefulsets.PodMetadata(tenant, pool, opVersion)
	if !reflect.DeepEqual(expectedMetadata.Labels, ss.ObjectMeta.Labels) {
		poolMatchesSS = false
	}
	if !reflect.DeepEqual(expectedMetadata.Annotations, ss.ObjectMeta.Annotations) {
		poolMatchesSS = false
	}

	return poolMatchesSS, nil
}
