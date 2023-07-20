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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "MinIO Service Created")
		} else {
			return err
		}
	}

	// compare any other change from what is specified on the tenant, since some of the state of the service is saved
	// on the service.spec we will compare individual parts
	expectedSvc := services.NewClusterIPForMinIO(tenant)

	// check the expose status of the MinIO ClusterIP service
	minioSvcMatchesSpec, err := minioSvcMatchesSpecification(svc, expectedSvc)

	// check the specification of the MinIO ClusterIP service
	if !minioSvcMatchesSpec {
		if err != nil {
			klog.Infof("MinIO Services don't match: %s", err)
		}

		svc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		svc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		svc.Spec.Ports = expectedSvc.Spec.Ports

		// Only when ExposeServices is set an explicit value we do modifications to the service type
		if tenant.Spec.ExposeServices != nil {
			if tenant.Spec.ExposeServices.MinIO {
				svc.Spec.Type = v1.ServiceTypeLoadBalancer
			} else {
				svc.Spec.Type = v1.ServiceTypeClusterIP
			}
		}

		// update the selector
		svc.Spec.Selector = expectedSvc.Spec.Selector

		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Updated", "MinIO Service Updated")
	}
	return err
}

func minioSvcMatchesSpecification(svc *v1.Service, expectedSvc *v1.Service) (bool, error) {
	// expected labels match
	for k, expVal := range expectedSvc.ObjectMeta.Labels {
		if value, ok := svc.ObjectMeta.Labels[k]; !ok || value != expVal {
			return false, errors.New("service labels don't match")
		}
	}
	// expected annotations match
	for k, expVal := range expectedSvc.ObjectMeta.Annotations {
		if value, ok := svc.ObjectMeta.Annotations[k]; !ok || value != expVal {
			return false, errors.New("service annotations don't match")
		}
	}
	// expected ports match
	if len(svc.Spec.Ports) != len(expectedSvc.Spec.Ports) {
		return false, errors.New("service ports don't match")
	}

	for i, expPort := range expectedSvc.Spec.Ports {
		if expPort.Name != svc.Spec.Ports[i].Name ||
			expPort.Port != svc.Spec.Ports[i].Port ||
			expPort.TargetPort != svc.Spec.Ports[i].TargetPort {
			return false, errors.New("service ports don't match")
		}
	}
	// compare selector
	if !equality.Semantic.DeepDerivative(expectedSvc.Spec.Selector, svc.Spec.Selector) {
		// some field set by the operator has changed
		return false, errors.New("selectors don't match")
	}
	if svc.Spec.Type != expectedSvc.Spec.Type {
		return false, errors.New("Service type doesn't match")
	}
	return true, nil
}
