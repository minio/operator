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

package controller

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
	tenantCopy.Spec = miniov2.TenantSpec{}
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

func (c *Controller) updatePoolStatus(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	return c.updatePoolStatusWithRetry(ctx, tenant, true)
}

func (c *Controller) updatePoolStatusWithRetry(ctx context.Context, tenant *miniov2.Tenant, retry bool) (*miniov2.Tenant, error) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status = *tenant.Status.DeepCopy()
	tenantCopy.Status.Pools = tenant.Status.Pools
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
			return c.updatePoolStatusWithRetry(ctx, tenant, false)
		}
		return t, err
	}
	return t, nil
}

func (c *Controller) updateCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, autoCertEnabled bool) (*miniov2.Tenant, error) {
	return c.updateCertificatesWithRetry(ctx, tenant, autoCertEnabled, true)
}

func (c *Controller) updateCertificatesWithRetry(ctx context.Context, tenant *miniov2.Tenant, autoCertEnabled bool, retry bool) (*miniov2.Tenant, error) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status = *tenant.Status.DeepCopy()
	tenantCopy.Status.Certificates.AutoCertEnabled = &autoCertEnabled
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
			return c.updateCertificatesWithRetry(ctx, tenant, autoCertEnabled, false)
		}
		return t, err
	}
	return t, nil
}

func (c *Controller) updateCustomCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, customCertificates *miniov2.CustomCertificates) (*miniov2.Tenant, error) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status.Certificates.CustomCertificates = customCertificates

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Tenant resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	opts := metav1.UpdateOptions{}
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()

	if err != nil {
		return t, err
	}
	return t, nil
}

func (c *Controller) updateProvisionedUsersStatus(ctx context.Context, tenant *miniov2.Tenant, provisionedUsers bool) (*miniov2.Tenant, error) {
	return c.updateProvisionedUsersWithRetry(ctx, tenant, provisionedUsers, true)
}

func (c *Controller) updateProvisionedUsersWithRetry(ctx context.Context, tenant *miniov2.Tenant, provisionedUsers bool, retry bool) (*miniov2.Tenant, error) {
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status = *tenant.Status.DeepCopy()
	tenantCopy.Status.ProvisionedUsers = provisionedUsers
	opts := metav1.UpdateOptions{}
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()
	if err != nil {
		if k8serrors.IsConflict(err) && retry {
			klog.Info("Hit conflict issue, getting latest version of tenant")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.updateProvisionedUsersWithRetry(ctx, tenant, provisionedUsers, false)
		}
		return t, err
	}
	return t, nil
}

func (c *Controller) updateProvisionedBucketStatus(ctx context.Context, tenant *miniov2.Tenant, provisionedBuckets bool) (*miniov2.Tenant, error) {
	return c.updateProvisionedBucketsWithRetry(ctx, tenant, provisionedBuckets, true)
}

func (c *Controller) updateProvisionedBucketsWithRetry(ctx context.Context, tenant *miniov2.Tenant, provisionedBuckets bool, retry bool) (*miniov2.Tenant, error) {
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status = *tenant.Status.DeepCopy()
	tenantCopy.Status.ProvisionedBuckets = provisionedBuckets
	opts := metav1.UpdateOptions{}
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).UpdateStatus(ctx, tenantCopy, opts)
	t.EnsureDefaults()
	if err != nil {
		if k8serrors.IsConflict(err) && retry {
			klog.Info("Hit conflict issue, getting latest version of tenant")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.updateProvisionedBucketsWithRetry(ctx, tenant, provisionedBuckets, false)
		}
		return t, err
	}
	return t, nil
}

func (c *Controller) updateTenantSyncVersion(ctx context.Context, tenant *miniov2.Tenant, syncVersion string) (*miniov2.Tenant, error) {
	return c.updateTenantSyncVersionWithRetry(ctx, tenant, syncVersion, true)
}

func (c *Controller) updateTenantSyncVersionWithRetry(ctx context.Context, tenant *miniov2.Tenant, syncVersion string, retry bool) (*miniov2.Tenant, error) {
	// If we are updating the tenant with the same sync version as before we are going to skip it as to avoid a resource number
	// change and have the operator loop re-processing the tenant endlessly
	if tenant.Status.SyncVersion == syncVersion {
		return tenant, nil
	}
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	tenantCopy := tenant.DeepCopy()
	tenantCopy.Spec = miniov2.TenantSpec{}
	tenantCopy.Status.SyncVersion = syncVersion
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
			klog.Info("Hit conflict issue, getting latest version of tenant to update version")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.updateTenantSyncVersionWithRetry(ctx, tenant, syncVersion, false)
		}
		return t, err
	}
	return t, nil
}
