// This file is part of MinIO Console Server
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
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/hashicorp/go-version"

	"github.com/minio/minio-go/v7/pkg/set"

	"k8s.io/klog/v2"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	version420 = "v4.2.0"
	version424 = "v4.2.4"
)

// checkForUpgrades verifies if the tenant definition needs any upgrades
func (c *Controller) checkForUpgrades(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {

	upgradesToDo := set.NewStringSet()

	// if the version is empty, do all upgrades
	if tenant.Status.SyncVersion == "" {
		upgradesToDo.Add(version420)
		upgradesToDo.Add(version424)
	} else {
		currentSyncVersion, err := version.NewVersion(tenant.Status.SyncVersion)
		if err != nil {
			return tenant, err
		}
		// check which upgrades we need to do
		versionsThatNeedUpgrades := []string{
			version420,
			version424,
		}
		for _, v := range versionsThatNeedUpgrades {
			vp, _ := version.NewVersion(v)
			if currentSyncVersion.LessThan(vp) {
				upgradesToDo.Add(v)
			}
		}
	}

	if upgradesToDo.Contains(version420) {
		if tenant, err := c.upgrade420(ctx, tenant); err != nil {
			klog.V(2).Infof("'%s/%s' Error upgrading tenant: %v", tenant.Namespace, tenant.Name, err.Error())
			return tenant, err
		}
	}
	if upgradesToDo.Contains(version424) {
		if tenant, err := c.upgrade424(ctx, tenant); err != nil {
			klog.V(2).Infof("'%s/%s' Error upgrading tenant: %v", tenant.Namespace, tenant.Name, err.Error())
			return tenant, err
		}
	}
	return tenant, nil
}

// Upgrades the sync version to v4.2.0
// in this version we renamed a bunch of environment variables and removed the
// stand-alone console deployment. I swear the name of the function is a coincidence.
func (c *Controller) upgrade420(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	logSearchSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.LogSecretName(), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if k8serrors.IsNotFound(err) {
		klog.Infof("%s has no log secret", tenant.Name)
	} else {
		secretChanged := false
		if _, ok := logSearchSecret.Data["LOGSEARCH_QUERY_AUTH_TOKEN"]; ok {
			logSearchSecret.Data["MINIO_QUERY_AUTH_TOKEN"] = logSearchSecret.Data["LOGSEARCH_QUERY_AUTH_TOKEN"]
			delete(logSearchSecret.Data, "LOGSEARCH_QUERY_AUTH_TOKEN")
			secretChanged = true
		}

		if _, ok := logSearchSecret.Data["CONSOLE_PROMETHEUS_URL"]; ok {
			logSearchSecret.Data["MINIO_PROMETHEUS_URL"] = logSearchSecret.Data["CONSOLE_PROMETHEUS_URL"]
			delete(logSearchSecret.Data, "CONSOLE_PROMETHEUS_URL")
			secretChanged = true
		}

		if secretChanged {
			_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, logSearchSecret, metav1.UpdateOptions{})
			if err != nil {
				return nil, err
			}
		}

	}
	// delete the previous operator secrets, they may be in a bad state
	err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Delete(ctx,
		miniov2.WebhookSecret, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Error deleting operator webhook secret, manual deletion is needed: %v", err)
	}

	if tenant, err = c.updateTenantSyncVersion(ctx, tenant, version420, nil); err != nil {
		return nil, err
	}

	return tenant, nil
}

// Upgrades the sync version to v4.2.4
// we started running all deployment with a default non-root `securityContext` which breaks previous tenants
// running without a security context, so to preserve the behavior, we will add the root securityContext to
// the tenant definition
func (c *Controller) upgrade424(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	// only do this update if the Tenant has at least 1 pool initialized which means it's not a fresh deployment
	atLeastOnePoolInitialized := false
	for _, pool := range tenant.Status.Pools {
		if pool.State == miniov2.PoolInitialized {
			atLeastOnePoolInitialized = true
			break
		}
	}

	var notes []string
	// if we found at least 1 pool initialized this is not a fresh tenant and needs upgrade
	if atLeastOnePoolInitialized {
		securityContextWasChanged := false
		var poolsToUpdate []int
		for i, pool := range tenant.Spec.Pools {
			if pool.SecurityContext == nil {
				securityContextWasChanged = true
				poolsToUpdate = append(poolsToUpdate, i)
				notes = append(notes, fmt.Sprintf("Pool %s had no security context so we added one for running as root (not recommended)", pool.Name))
			}
		}
		if securityContextWasChanged {

			// Use the implicit defaults from kubernetes
			var runAsNonRoot bool
			var runAsUser int64
			var runAsGroup int64
			var fsGroup int64
			var securityContext = corev1.PodSecurityContext{
				RunAsNonRoot: &runAsNonRoot,
				RunAsUser:    &runAsUser,
				RunAsGroup:   &runAsGroup,
				FSGroup:      &fsGroup,
			}
			for i := range poolsToUpdate {
				tenant.Spec.Pools[i].SecurityContext = &securityContext
			}

			klog.Infof("'%s/%s' Updating tenant's pools security context", tenant.Namespace, tenant.Name)
			if tenant, err := c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Update(ctx, tenant, metav1.UpdateOptions{}); err != nil {
				klog.Errorf("'%s/%s' Error upgrading implicit tenant security context, MinIO may not start: %v", tenant.Namespace, tenant.Name, err)
			}
		}
	}

	var err error
	if tenant, err = c.updateTenantSyncVersion(ctx, tenant, version424, notes); err != nil {
		return nil, err
	}

	return tenant, nil
}
