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
	acv2 "github.com/minio/operator/pkg/client/applyconfiguration/minio.min.io/v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	ApplyConfigurationFieldManager = "minio-operator"
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

	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}
	tac := acv2.Tenant(tenant.Name, tenant.Namespace).
		WithStatus(ExtactTenantStatus(tenant).
			WithAvailableReplicas(availableReplicas).
			WithCurrentState(currentState))
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
	}
	return t, nil
}

func (c *Controller) updatePoolStatus(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	return c.updatePoolStatusWithRetry(ctx, tenant, true)
}

func (c *Controller) updatePoolStatusWithRetry(ctx context.Context, tenant *miniov2.Tenant, retry bool) (*miniov2.Tenant, error) {
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}

	tac := acv2.Tenant(tenant.Name, tenant.Namespace).WithStatus(ExtactTenantStatus(tenant))
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
	}
	return t, nil
}

func (c *Controller) updateCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, autoCertEnabled bool) (*miniov2.Tenant, error) {
	return c.updateCertificatesWithRetry(ctx, tenant, autoCertEnabled, true)
}

func (c *Controller) updateCertificatesWithRetry(ctx context.Context, tenant *miniov2.Tenant, autoCertEnabled bool, retry bool) (*miniov2.Tenant, error) {
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}
	tac := acv2.Tenant(tenant.Name, tenant.Namespace).
		WithStatus(ExtactTenantStatus(tenant))
	tac.Status.Certificates.WithAutoCertEnabled(autoCertEnabled)

	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
	}
	return t, nil
}

func (c *Controller) updateCustomCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, customCertificates *miniov2.CustomCertificates) (*miniov2.Tenant, error) {
	return c.updateCustomCertificatesStatusWithRetry(ctx, tenant, customCertificates, true)
}

func (c *Controller) updateCustomCertificatesStatusWithRetry(ctx context.Context, tenant *miniov2.Tenant, customCertificates *miniov2.CustomCertificates, retry bool) (*miniov2.Tenant, error) {
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}

	tac := acv2.Tenant(tenant.Name, tenant.Namespace).WithStatus(ExtactTenantStatus(tenant))

	if len(customCertificates.Client) > 0 {
		tac.Status.Certificates.CustomCertificates.WithClient(customCertificates.Client...)
	}

	if len(customCertificates.Minio) > 0 {
		tac.Status.Certificates.CustomCertificates.WithMinio(customCertificates.Minio...)
	}

	if len(customCertificates.MinioCAs) > 0 {
		tac.Status.Certificates.CustomCertificates.WithMinioCAs(customCertificates.MinioCAs...)
	}

	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
	t.EnsureDefaults()
	if err != nil {
		if k8serrors.IsConflict(err) && retry {
			klog.Info("Hit conflict issue, getting latest version of tenant")
			tenant, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Get(ctx, tenant.Name, metav1.GetOptions{})
			if err != nil {
				return tenant, err
			}
			return c.updateCustomCertificatesStatusWithRetry(ctx, tenant, customCertificates, false)
		}
		return tenant, err
	}
	return t, nil
}

func (c *Controller) updateProvisionedUsersStatus(ctx context.Context, tenant *miniov2.Tenant, provisionedUsers bool) (*miniov2.Tenant, error) {
	return c.updateProvisionedUsersWithRetry(ctx, tenant, provisionedUsers, true)
}

func (c *Controller) updateProvisionedUsersWithRetry(ctx context.Context, tenant *miniov2.Tenant, provisionedUsers bool, retry bool) (*miniov2.Tenant, error) {
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}
	tac := acv2.Tenant(tenant.Name, tenant.Namespace).
		WithStatus(ExtactTenantStatus(tenant).
			WithProvisionedUsers(provisionedUsers))
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
	}
	return t, nil
}

func (c *Controller) updateProvisionedBucketStatus(ctx context.Context, tenant *miniov2.Tenant, provisionedBuckets bool) (*miniov2.Tenant, error) {
	return c.updateProvisionedBucketsWithRetry(ctx, tenant, provisionedBuckets, true)
}

func (c *Controller) updateProvisionedBucketsWithRetry(ctx context.Context, tenant *miniov2.Tenant, provisionedBuckets bool, retry bool) (*miniov2.Tenant, error) {
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}
	tac := acv2.Tenant(tenant.Name, tenant.Namespace).
		WithStatus(ExtactTenantStatus(tenant).
			WithProvisionedBuckets(provisionedBuckets))
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
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
	applyOpts := metav1.ApplyOptions{FieldManager: ApplyConfigurationFieldManager, Force: true}
	tac := acv2.Tenant(tenant.Name, tenant.Namespace).
		WithStatus(ExtactTenantStatus(tenant).
			WithSyncVersion(syncVersion))
	t, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).ApplyStatus(ctx, tac, applyOpts)
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
		return tenant, err
	}
	return t, nil
}

func ExtactTenantStatus(tenant *miniov2.Tenant) *acv2.TenantStatusApplyConfiguration {
	cs := acv2.CertificateStatus().
		WithCustomCertificates(acv2.CustomCertificates())

	if tenant.Status.Certificates.AutoCertEnabled != nil {
		cs.WithAutoCertEnabled(*tenant.Status.Certificates.AutoCertEnabled)
	}

	if tenant.Status.Certificates.CustomCertificates != nil {
		if len(tenant.Status.Certificates.CustomCertificates.Client) > 0 {
			cs.CustomCertificates.WithClient(tenant.Status.Certificates.CustomCertificates.Client...)
		}
		if len(tenant.Status.Certificates.CustomCertificates.Minio) > 0 {
			cs.CustomCertificates.WithMinio(tenant.Status.Certificates.CustomCertificates.Minio...)
		}
		if len(tenant.Status.Certificates.CustomCertificates.MinioCAs) > 0 {
			cs.CustomCertificates.WithMinioCAs(tenant.Status.Certificates.CustomCertificates.MinioCAs...)
		}
	}
	pools := []*acv2.PoolStatusApplyConfiguration{}
	if len(tenant.Spec.Pools) > 0 {
		for _, po := range tenant.Status.Pools {
			pools = append(pools, acv2.PoolStatus().
				WithLegacySecurityContext(po.LegacySecurityContext).
				WithSSName(po.SSName).
				WithState(po.State))
		}
	}

	ts := acv2.TenantStatus().
		WithSyncVersion(tenant.Status.SyncVersion).
		WithAvailableReplicas(tenant.Status.AvailableReplicas).
		WithCurrentState(tenant.Status.CurrentState).
		WithProvisionedUsers(tenant.Status.ProvisionedUsers).
		WithProvisionedBuckets(tenant.Status.ProvisionedBuckets).
		WithCertificates(cs).
		WithPools(pools...)
	return ts
}
