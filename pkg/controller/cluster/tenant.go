/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cluster

import (
	"context"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// checkMinIOSvc validates the existence of the MinIO service and validate it's status against what the specification
// states
func (c *Controller) checkMinIOSvc(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	// Handle the Internal ClusterIP Service for Tenant
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.MinIOCIServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningCIService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Cluster IP Service for cluster %q", nsName)
			// Create the clusterIP service for the Tenant
			svc = services.NewClusterIPForMinIO(tenant)
			svc, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, svc, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// check the expose status of the MinIO ClusterIP service
	minioSvcMatchesSpec := true
	// compare any other change from what is specified on the tenant
	expectedSvc := services.NewClusterIPForMinIO(tenant)
	if !equality.Semantic.DeepDerivative(expectedSvc.Spec, svc.Spec) {
		// some field set by the operator has changed
		minioSvcMatchesSpec = false
	}

	// check the specification of the MinIO ClusterIP service
	if !minioSvcMatchesSpec {
		svc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		svc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		// we can only expose the service, not un-expose it
		if tenant.Spec.ExposeServices != nil && tenant.Spec.ExposeServices.MinIO && svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			svc.Spec.Type = v1.ServiceTypeLoadBalancer
		}
		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return err
}
