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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// checkConsoleSvc validates the existence of the MinIO service and validate it's status against what the specification
// states
func (c *Controller) checkConsoleSvc(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {

	// Handle the Internal ClusterIP Service for Tenant
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.ConsoleCIServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningConsoleService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Console Cluster IP Service for cluster %q", nsName)
			// Create the clusterIP service for the Tenant
			svc = services.NewClusterIPForConsole(tenant)
			svc, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, svc, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// check the expose status of the Console service
	svcMatchesSpec := true
	// compare any other change from what is specified on the tenant
	expectedSvc := services.NewClusterIPForConsole(tenant)
	if !equality.Semantic.DeepDerivative(expectedSvc.Spec, svc.Spec) {
		// some field set by the operator has changed
		svcMatchesSpec = false
	}

	// check the specification of the MinIO ClusterIP service
	if !svcMatchesSpec {
		svc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		svc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		svc.Spec.Ports = expectedSvc.Spec.Ports
		// we can only expose the service, not un-expose it
		if tenant.Spec.ExposeServices != nil && tenant.Spec.ExposeServices.MinIO && svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			svc.Spec.Type = v1.ServiceTypeLoadBalancer
		}
		// update the selector
		svc.Spec.Selector = expectedSvc.Spec.Selector

		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return err
}
