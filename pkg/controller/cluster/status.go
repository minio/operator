// This file is part of MinIO Console Server
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
	"context"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (c *Controller) updateTenantStatus(ctx context.Context, tenant *miniov2.Tenant, currentState string, availableReplicas int32) (*miniov2.Tenant, error) {
	return c.updateTenantStatusWithRetry(ctx, tenant, currentState, availableReplicas, true)
}

func (c *Controller) updateTenantStatusWithRetry(ctx context.Context, tenant *miniov2.Tenant, currentState string, availableReplicas int32, retry bool) (*miniov2.Tenant, error) {
	// If we are updating the tenant with the same status as before we are going to skip it as to avoid a resource number
	// change and have the operator loop re-processing the tenant endlessly
	if tenant.Status.CurrentState == currentState && tenant.Status.AvailableReplicas == availableReplicas {
		return tenant, nil
	}
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Status.AvailableReplicas = availableReplicas
	tenantCopy.Status.CurrentState = currentState
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Tenant resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	opts := metav1.UpdateOptions{}
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()
	if err != nil {
		// if rejected due to conflict, get the latest tenant and retry once
		if k8serrors.IsConflict(err) && retry {
			klog.Info("Hit conflict issue, getting latest version of tenant")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.updateTenantStatusWithRetry(ctx, tenant, currentState, availableReplicas, false)
		}
		return t, err
	}
	return t, nil
}

func (c *Controller) increaseTenantRevision(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	return c.increaseTenantRevisionWithRetry(ctx, tenant, true)
}

func (c *Controller) increaseTenantRevisionWithRetry(ctx context.Context, tenant *miniov2.Tenant, retry bool) (*miniov2.Tenant, error) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Status.Revision = tenantCopy.Status.Revision + 1
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Tenant resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	opts := metav1.UpdateOptions{}
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()
	if err != nil {
		// if rejected due to conflict, get the latest tenant and retry once
		if k8serrors.IsConflict(err) && retry {
			klog.Info("Hit conflict issue, getting latest version of tenant")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.increaseTenantRevisionWithRetry(ctx, tenant, false)
		}
		return t, err
	}
	return t, nil
}
