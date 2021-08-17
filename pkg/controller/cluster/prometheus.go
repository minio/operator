// This file is part of MinIO Prometheus
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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"k8s.io/klog/v2"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/configmaps"
	"github.com/minio/operator/pkg/resources/secrets"
	"github.com/minio/operator/pkg/resources/servicemonitor"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
)

func (c *Controller) syncPrometheusState(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) error {

	if tenant.HasPrometheusEnabled() {
		_, err := c.checkAndCreatePrometheusConfigMap(ctx, tenant, string(tenantConfiguration["accesskey"]), string(tenantConfiguration["secretkey"]))
		if err != nil {
			return err
		}

		_, err = c.checkAndCreatePrometheusHeadless(ctx, tenant)
		if err != nil {
			return err
		}

		err = c.checkAndCreatePrometheusStatefulSet(ctx, tenant)
		if err != nil {
			return err
		}
	}

	if tenant.HasPrometheusSMEnabled() {
		err := c.checkAndCreatePrometheusServiceMonitorSecret(ctx, tenant, string(tenantConfiguration["accesskey"]), string(tenantConfiguration["secretkey"]))
		if err != nil {
			return err
		}
		err = c.checkAndCreatePrometheusServiceMonitor(ctx, tenant)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Controller) checkAndCreatePrometheusConfigMap(ctx context.Context, tenant *miniov2.Tenant, accessKey, secretKey string) (*corev1.ConfigMap, error) {
	configMap, err := c.kubeClientSet.CoreV1().ConfigMaps(tenant.Namespace).Get(ctx, tenant.PrometheusConfigMapName(), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return configMap, err
	} else if err == nil {
		// check if configmap needs update.
		updatedConfigMap := configmaps.UpdatePrometheusConfigMap(tenant, accessKey, secretKey, configMap)
		if updatedConfigMap == nil {
			return configMap, nil
		}

		klog.V(2).Infof("Updating Prometheus config-map for %s", tenant.Name)
		configMap, err = c.kubeClientSet.CoreV1().ConfigMaps(tenant.Namespace).Update(ctx, updatedConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return configMap, err
		}

		return configMap, err
	}

	// otherwise create the config
	klog.V(2).Infof("Creating a new Prometheus config-map for %s", tenant.Name)
	return c.kubeClientSet.CoreV1().ConfigMaps(tenant.Namespace).Create(ctx, configmaps.PrometheusConfigMap(tenant, accessKey, secretKey), metav1.CreateOptions{})
}

func (c *Controller) checkAndCreatePrometheusHeadless(ctx context.Context, tenant *miniov2.Tenant) (*corev1.Service, error) {
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.PrometheusHLServiceName())
	if err == nil || !k8serrors.IsNotFound(err) {
		return svc, err
	}

	klog.V(2).Infof("Creating a new Prometheus Headless Service for %s", tenant.Namespace)
	svc = services.NewHeadlessForPrometheus(tenant)
	_, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	return svc, err
}

func (c *Controller) checkAndCreatePrometheusStatefulSet(ctx context.Context, tenant *miniov2.Tenant) error {
	promSs, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.PrometheusStatefulsetName())
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	if k8serrors.IsNotFound(err) {
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningPrometheusStatefulSet, 0); err != nil {
			return err
		}

		klog.V(2).Infof("Creating a new Prometheus StatefulSet for %s", tenant.Namespace)
		prometheusSS := statefulsets.NewForPrometheus(tenant, tenant.PrometheusHLServiceName())
		_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, prometheusSS, metav1.CreateOptions{})
		return err
	}
	// check the current state of the ss of prometheus
	ssMatchesSpec, err := prometheusStatefulsetMatchesSpec(tenant, promSs)
	if err != nil {
		return err
	}

	if !ssMatchesSpec {
		prometheusSS := statefulsets.NewForPrometheus(tenant, tenant.PrometheusHLServiceName())
		_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, prometheusSS, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func (c *Controller) checkAndCreatePrometheusServiceMonitorSecret(ctx context.Context, tenant *miniov2.Tenant, accessKey, secretKey string) error {
	_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.PromServiceMonitorSecret(), metav1.GetOptions{})
	if err == nil || !k8serrors.IsNotFound(err) {
		return err
	}

	if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningPrometheusServiceMonitor, 0); err != nil {
		return err
	}

	klog.V(2).Infof("Creating a new Prometheus Service Monitor secret for %s", tenant.Namespace)
	secret := secrets.PromServiceMonitorSecret(tenant, accessKey, secretKey)
	_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (c *Controller) checkAndCreatePrometheusServiceMonitor(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.serviceMonitorLister.ServiceMonitors(tenant.Namespace).Get(tenant.PrometheusServiceMonitorName())
	if err == nil || !k8serrors.IsNotFound(err) {
		return err
	}

	if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningPrometheusServiceMonitor, 0); err != nil {
		return err
	}

	klog.V(2).Infof("Creating a new Prometheus Service Monitor for %s", tenant.Namespace)
	prometheusSM := servicemonitor.NewForPrometheus(tenant)
	_, err = c.promClient.MonitoringV1().ServiceMonitors(tenant.Namespace).Create(ctx, prometheusSM, metav1.CreateOptions{})
	return err
}

// prometheusStatefulsetMatchesSpec checks if the Prometheus statefulset `actualSS`
// matches the desired spec provided by `tenant`
func prometheusStatefulsetMatchesSpec(tenant *miniov2.Tenant, actualSS *appsv1.StatefulSet) (bool, error) {
	if actualSS == nil {
		return false, errors.New("cannot process an empty prometheus statefulset")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}

	// Note: Since tenant's Headless service name for postgres server in log
	// search feature is internal to operator and is unaffectd by tenant
	// spec changes we use `actualSS`'s service name to create `expectedSS`.
	expectedSS := statefulsets.NewForPrometheus(tenant, actualSS.Spec.ServiceName)
	if !equality.Semantic.DeepDerivative(expectedSS.Spec, actualSS.Spec) {
		// some fields set by the operator have changed
		return false, nil
	}
	return true, nil
}
