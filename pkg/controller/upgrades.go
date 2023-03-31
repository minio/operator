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

package controller

import (
	"context"
	"fmt"

	"github.com/minio/operator/pkg/controller/legacy"
	corev1 "k8s.io/api/core/v1"

	"github.com/blang/semver/v4"
	"github.com/hashicorp/go-version"

	"k8s.io/klog/v2"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	version420 = "v4.2.0"
	version424 = "v4.2.4"
	version429 = "v4.2.9"
	version430 = "v4.3.0"
	version45  = "v4.5"
	version500 = "v5.0.0"
)

// Legacy const
const (
	WebhookSecret = "operator-webhook-secret"
)

type upgradeFunction func(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error)

// checkForUpgrades verifies if the tenant definition needs any upgrades
func (c *Controller) checkForUpgrades(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	var upgradesToDo []string
	upgrades := map[string]upgradeFunction{
		version420: c.upgrade420,
		version424: c.upgrade424,
		version429: c.upgrade429,
		version430: c.upgrade430,
		version45:  c.upgrade45,
		version500: c.upgrade500,
	}

	// if the version is not empty, this is not a new tenant, upgrade accordingly
	if tenant.Status.SyncVersion != "" {
		currentSyncVersion, err := version.NewVersion(tenant.Status.SyncVersion)
		if err != nil {
			return tenant, err
		}
		// check which upgrades we need to do
		versionsThatNeedUpgrades := []string{
			version420,
			version424,
			version429,
			version430,
			version45,
			version500,
		}
		for _, v := range versionsThatNeedUpgrades {
			vp, _ := version.NewVersion(v)
			if currentSyncVersion.LessThan(vp) {
				upgradesToDo = append(upgradesToDo, v)
			}
		}
	}

	for _, u := range upgradesToDo {
		klog.Infof("Upgrading %s", u)
		if tenant, err := upgrades[u](ctx, tenant); err != nil {
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
	logSearchSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, legacy.LogSecretName(tenant), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}

	if k8serrors.IsNotFound(err) {
		klog.Infof("%s has no log secret", tenant.Name)
	} else {
		secretChanged := false
		if _, ok := logSearchSecret.Data["LOGSEARCH_QUERY_AUTH_TOKEN"]; ok {
			logSearchSecret.Data["MINIO_LOG_QUERY_AUTH_TOKEN"] = logSearchSecret.Data["LOGSEARCH_QUERY_AUTH_TOKEN"]
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
				return tenant, err
			}
		}

	}
	// delete the previous operator secrets, they may be in a bad state
	err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Delete(ctx,
		WebhookSecret, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Error deleting operator webhook secret, manual deletion is needed: %v", err)
	}

	return c.updateTenantSyncVersion(ctx, tenant, version420)
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

	var err error
	// if we found at least 1 pool initialized this is not a fresh tenant and needs upgrade
	if atLeastOnePoolInitialized {
		for i, pool := range tenant.Spec.Pools {
			if pool.SecurityContext == nil {
				// Report the pool is now Legacy Security Context
				tenant.Status.Pools[i].LegacySecurityContext = true
				// push updates to status
				if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
					klog.Errorf("'%s/%s' Error upgrading implicit tenant security context, MinIO may not start: %v", tenant.Namespace, tenant.Name, err)
				}
			}
		}
	}

	return c.updateTenantSyncVersion(ctx, tenant, version424)
}

// Method to compare two versions.
// Returns 1 if v2 is smaller, -1
// if v1 is smaller, 0 if equal
func versionCompare(version1 string, version2 string) int {
	vs1, err := semver.ParseTolerant(version1)
	if err != nil {
		klog.Errorf("Error parsing version %s: %v", version1, err)
		return -1
	}
	vs2, err := semver.ParseTolerant(version2)
	if err != nil {
		klog.Errorf("Error parsing version %s: %v", version2, err)
		return -1
	}
	return vs1.Compare(vs2)
}

// Upgrades the sync version to v4.2.9
// we need to mark any pool with a security context = root as a .status.pools[*].legacySC, this is due to do a
// reversal on the fix we previously did on v4.2.4
func (c *Controller) upgrade429(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	// only do this update if the Tenant has at least 1 pool initialized which means it's not a fresh deployment
	atLeastOnePoolInitialized := false
	for _, pool := range tenant.Status.Pools {
		if pool.State == miniov2.PoolInitialized {
			atLeastOnePoolInitialized = true
			break
		}
	}

	var err error
	// if we found at least 1 pool initialized this is not a fresh tenant and needs upgrade
	if atLeastOnePoolInitialized {
		for i, pool := range tenant.Spec.Pools {
			// if they have a security context, and is having them run as root
			scontext := pool.SecurityContext
			if scontext != nil && scontext.RunAsNonRoot != nil && scontext.RunAsUser != nil {
				if !*scontext.RunAsNonRoot && *scontext.RunAsUser == 0 {
					// Report the pool is now Legacy Security Context
					tenant.Status.Pools[i].LegacySecurityContext = true
					// push updates to status
					if tenant, err = c.updatePoolStatus(ctx, tenant); err != nil {
						klog.Errorf("'%s/%s' Error upgrading implicit tenant security context, MinIO may not start: %v", tenant.Namespace, tenant.Name, err)
					}
				}
			}
		}
	}

	return c.updateTenantSyncVersion(ctx, tenant, version429)
}

// Upgrades the sync version to v4.3.0
// in this version we renamed MINIO_QUERY_AUTH_TOKEN to MINIO_LOG_QUERY_AUTH_TOKEN.
func (c *Controller) upgrade430(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	logSearchSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, legacy.LogSecretName(tenant), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}

	if k8serrors.IsNotFound(err) {
		klog.Infof("%s has no log secret", tenant.Name)
	} else {
		secretChanged := false
		if _, ok := logSearchSecret.Data["MINIO_QUERY_AUTH_TOKEN"]; ok {
			logSearchSecret.Data["MINIO_LOG_QUERY_AUTH_TOKEN"] = logSearchSecret.Data["MINIO_QUERY_AUTH_TOKEN"]
			delete(logSearchSecret.Data, "MINIO_QUERY_AUTH_TOKEN")
			secretChanged = true
		}

		if secretChanged {
			_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, logSearchSecret, metav1.UpdateOptions{})
			if err != nil {
				return tenant, err
			}
		}

	}

	return c.updateTenantSyncVersion(ctx, tenant, version430)
}

// Upgrades the sync version to v4.5
// in this version we finally deprecated tenant.spec.credsSecret field.
func (c *Controller) upgrade45(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	if tenant.HasConfigurationSecret() {
		return c.updateTenantSyncVersion(ctx, tenant, version45)
	}
	if !tenant.HasCredsSecret() {
		return tenant, fmt.Errorf("'%s/%s' error migrating tenant credsSecret, credsSecret does not exist", tenant.Namespace, tenant.Name)
	}
	// Create new configuration secret based on the existing credsSecret
	credsSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.Spec.CredsSecret.Name, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	var accessKey string
	var secretKey string
	if _, ok := credsSecret.Data["accesskey"]; ok {
		accessKey = string(credsSecret.Data["accesskey"])
	}
	if _, ok := credsSecret.Data["secretkey"]; ok {
		secretKey = string(credsSecret.Data["secretkey"])
	}
	if accessKey == "" || secretKey == "" {
		return tenant, fmt.Errorf("accessKey/secretKey are empty - '%s/%s' error in migrating tenant credsSecret to newer configuration", tenant.Namespace, tenant.Name)
	}
	tenantConfiguration := map[string]string{}
	tenantConfiguration["MINIO_ROOT_USER"] = accessKey
	tenantConfiguration["MINIO_ROOT_PASSWORD"] = secretKey
	configurationSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.ConfigurationSecretName(),
			Namespace: tenant.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		Data: map[string][]byte{
			"config.env": []byte(miniov2.GenerateTenantConfigurationFile(tenantConfiguration)),
		},
	}
	_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, configurationSecret, metav1.CreateOptions{})
	if err != nil {
		return tenant, err
	}
	// Update tenant fields
	tenantCopy := tenant.DeepCopy()
	tenantCopy.EnsureDefaults()
	tenantCopy.Spec.Configuration = &corev1.LocalObjectReference{
		Name: tenantCopy.ConfigurationSecretName(),
	}
	tenantCopy.Spec.CredsSecret = nil
	_, err = c.minioClientSet.MinioV2().Tenants(tenant.Namespace).Update(ctx, tenantCopy, metav1.UpdateOptions{})
	if err != nil {
		return tenant, fmt.Errorf("error updating tenant '%s/%s', could not update tenant.spec.configuration field: %v", tenant.Namespace, tenant.Name, err)
	}
	return c.updateTenantSyncVersion(ctx, tenant, version45)
}

// Upgrades the sync version to v5.0.0
// in this version we finally deprecated tenant.spec.credsSecret field.
func (c *Controller) upgrade500(ctx context.Context, tenant *miniov2.Tenant) (*miniov2.Tenant, error) {
	// log search deployment
	logSearchDeployment, err := c.deploymentLister.Deployments(tenant.Namespace).Get(legacy.LogSearchAPIDeploymentName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	// we can still get nil due to not found
	if logSearchDeployment != nil {
		logSearchDeployment.ObjectMeta.OwnerReferences = nil
		if _, err := c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Update(ctx, logSearchDeployment, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// log search secret
	logSearchSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, legacy.LogSecretName(tenant), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && logSearchSecret != nil {
		logSearchSecret.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, logSearchSecret, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// log search api service
	logSearchAPISvc, err := c.serviceLister.Services(tenant.Namespace).Get(legacy.LogSearchAPIServiceName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && logSearchAPISvc != nil {
		logSearchAPISvc.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, logSearchAPISvc, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// log search sts
	logPgSS, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(legacy.LogStatefulsetName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && logPgSS != nil {
		logPgSS.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, logPgSS, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// log hl svc
	logHlSvc, err := c.serviceLister.Services(tenant.Namespace).Get(legacy.LogHLServiceName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && logHlSvc != nil {
		logHlSvc.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, logHlSvc, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// prometheus cm
	promCM, err := c.kubeClientSet.CoreV1().ConfigMaps(tenant.Namespace).Get(ctx, tenant.PrometheusConfigMapName(), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && promCM != nil {
		promCM.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.CoreV1().ConfigMaps(tenant.Namespace).Update(ctx, promCM, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// prometheus hl svc
	promHlSvc, err := c.serviceLister.Services(tenant.Namespace).Get(legacy.PrometheusHLServiceName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && promHlSvc != nil {
		promHlSvc.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, promHlSvc, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	// prometheus sts
	prometheusStatefulSet, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(legacy.PrometheusStatefulsetName(tenant))
	if err != nil && !k8serrors.IsNotFound(err) {
		return tenant, err
	}
	if err == nil && prometheusStatefulSet != nil {
		prometheusStatefulSet.ObjectMeta.OwnerReferences = nil
		if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, prometheusStatefulSet, metav1.UpdateOptions{}); err != nil {
			return tenant, err
		}
	}
	return c.updateTenantSyncVersion(ctx, tenant, version500)
}
