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

package cluster

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/configmaps"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
)

// MinIOPrometheusMetrics holds metrics pulled from prometheus
type MinIOPrometheusMetrics struct {
	UsableCapacity int64
	Usage          int64
}

func getPrometheusMetricsForTenant(tenant *miniov2.Tenant, bearer string) (*MinIOPrometheusMetrics, error) {
	return getPrometheusMetricsForTenantWithRetry(tenant, bearer, 5)
}

func getPrometheusMetricsForTenantWithRetry(tenant *miniov2.Tenant, bearer string, tryCount int) (*MinIOPrometheusMetrics, error) {
	// build the endpoint to contact the Tenant
	svcURL := tenant.GetTenantServiceURL()

	endpoint := fmt.Sprintf("%s%s", svcURL, "/minio/v2/metrics/cluster")

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		klog.Infof("error request pinging: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))

	httpClient := &http.Client{
		Transport: getHealthCheckTransport()(),
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// if we fail due to timeout, retry
		if err, ok := err.(net.Error); ok && err.Timeout() && tryCount > 0 {
			klog.Infof("health check failed, retrying %d, err: %s", tryCount, err)
			time.Sleep(10 * time.Second)
			return getPrometheusMetricsForTenantWithRetry(tenant, bearer, tryCount-1)
		}
		klog.Infof("error pinging: %v", err)
		return nil, err
	}
	defer drainBody(resp.Body)

	promMetrics := MinIOPrometheusMetrics{}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Usage
		if strings.HasPrefix(line, "minio_bucket_usage_total_bytes") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				usage, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					klog.Infof("%s/%s Could not parse usage for tenant: %s", tenant.Namespace, tenant.Name, line)
				}
				promMetrics.Usage = promMetrics.Usage + int64(usage)
			}
		}
		// Usable capacity
		if strings.HasPrefix(line, "minio_cluster_capacity_usable_total_bytes") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				usable, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					klog.Infof("%s/%s Could not parse usable capacity for tenant: %s", tenant.Namespace, tenant.Name, line)
				}
				promMetrics.UsableCapacity = int64(usable)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		klog.Error(err)
	}
	return &promMetrics, nil
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

func (c *Controller) deletePrometheusHeadless(ctx context.Context, tenant *miniov2.Tenant) error {
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.PrometheusHLServiceName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Deleting Prometheus Headless Service for %s", tenant.Namespace)
	err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Delete(ctx, tenant.PrometheusHLServiceName(), metav1.DeleteOptions{})
	return err
}

func (c *Controller) deletePrometheusStatefulSet(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.PrometheusStatefulsetName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Deleting Prometheus StatefulSet for %s", tenant.Namespace)
	err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Delete(ctx, tenant.PrometheusStatefulsetName(), metav1.DeleteOptions{})
	return err
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

func (c *Controller) checkPrometheusStatus(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte, totalReplicas int32, cOpts metav1.CreateOptions, uOpts metav1.UpdateOptions, nsName types.NamespacedName) error {
	if tenant.HasPrometheusEnabled() {
		_, err := c.checkAndCreatePrometheusConfigMap(ctx, tenant, string(tenantConfiguration["accesskey"]), string(tenantConfiguration["secretkey"]))
		if err != nil {
			return err
		}
		_, err = c.checkAndCreatePrometheusHeadless(ctx, tenant)
		if err != nil {
			return err
		}
		return c.checkAndCreatePrometheusStatefulSet(ctx, tenant, totalReplicas, cOpts, uOpts, nsName)
	}
	err := c.deletePrometheusHeadless(ctx, tenant)
	if err != nil {
		return err
	}
	return c.deletePrometheusStatefulSet(ctx, tenant)
}

// prometheusStatefulSetMatchesSpec checks if the StatefulSet for Prometheus matches what is expected and described from the Tenant
func prometheusStatefulSetMatchesSpec(tenant *miniov2.Tenant, prometheusStatefulSet *appsv1.StatefulSet) (bool, error) {
	if prometheusStatefulSet == nil {
		return false, errors.New("cannot process an empty prometheus StatefulSet")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}
	// compare image directly
	if !tenant.Spec.Prometheus.EqualImages(prometheusStatefulSet.Spec.Template.Spec.Containers) {
		klog.V(2).Infof("Tenant %s Prometheus version doesn't match", tenant.Name)
		return false, nil
	}
	// compare any other change from what is specified on the tenant
	expectedStatefulSet := statefulsets.NewForPrometheus(tenant, tenant.PrometheusHLServiceName())
	// compare containers environment variables
	if miniov2.IsContainersEnvUpdated(expectedStatefulSet.Spec.Template.Spec.Containers, prometheusStatefulSet.Spec.Template.Spec.Containers) {
		return false, nil
	}
	if !equality.Semantic.DeepDerivative(expectedStatefulSet.Spec, prometheusStatefulSet.Spec) {
		// some field set by the operator has changed
		return false, nil
	}
	return true, nil
}

func (c *Controller) checkAndCreatePrometheusStatefulSet(ctx context.Context, tenant *miniov2.Tenant, totalReplicas int32, cOpts metav1.CreateOptions, uOpts metav1.UpdateOptions, nsName types.NamespacedName) error {
	// Get the StatefulSet with the name specified in spec
	prometheusStatefulSet, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.PrometheusStatefulsetName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningPrometheusStatefulSet, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Prometheus StatefulSet for %s", tenant.Namespace)
			prometheusSS := statefulsets.NewForPrometheus(tenant, tenant.PrometheusHLServiceName())
			_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, prometheusSS, metav1.CreateOptions{})
			return err
		}
		return err
	}
	// Verify if this prometheus StatefulSet matches the spec on the tenant (resources, affinity, sidecars, etc)
	var ssSpecMatches bool
	if ssSpecMatches, err = prometheusStatefulSetMatchesSpec(tenant, prometheusStatefulSet); err != nil {
		return err
	}
	// if the Prometheus StatefulSet doesn't match the spec
	if !ssSpecMatches {
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingPrometheus, totalReplicas); err != nil {
			return err
		}
		newPrometheusSS := statefulsets.NewForPrometheus(tenant, tenant.PrometheusHLServiceName())
		// updating fields
		prometheusStatefulSet.Spec.Template = newPrometheusSS.Spec.Template
		prometheusStatefulSet.Spec.Replicas = newPrometheusSS.Spec.Replicas
		prometheusStatefulSet.Spec.UpdateStrategy = newPrometheusSS.Spec.UpdateStrategy
		if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, prometheusStatefulSet, uOpts); err != nil {
			c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "StsFailed", fmt.Sprintf("Prometheus Statefulset failed to update: %s", err))
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "StsUpdated", "Prometheus Statefulset Updated")
	}
	return nil
}

func (c *Controller) getPrometheus(ctx context.Context) (*promv1.Prometheus, error) {
	ns := miniov2.GetPrometheusNamespace()
	promName := miniov2.GetPrometheusName()
	var p *promv1.Prometheus
	var err error
	if promName != "" {
		p, err = c.promClient.MonitoringV1().Prometheuses(ns).Get(ctx, promName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		pList, err := c.promClient.MonitoringV1().Prometheuses(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		if len(pList.Items) == 0 {
			return nil, errors.New("No prometheus found on namespace " + ns)
		}
		if len(pList.Items) > 1 {
			return nil, errors.New("More than 1 prometheus found on namespace " + ns + ". PROMETHEUS_NAME not specified.")
		}
		p = pList.Items[0]
	}
	return p, nil
}

func (c *Controller) checkAndCreatePrometheusAddlConfig(ctx context.Context, tenant *miniov2.Tenant, accessKey, secretKey string) error {
	ns := miniov2.GetPrometheusNamespace()

	p, err := c.getPrometheus(ctx)
	if err != nil {
		return err
	}

	// If the additional scrape config is set to something else, we will error out
	if p.Spec.AdditionalScrapeConfigs != nil && p.Spec.AdditionalScrapeConfigs.Name != miniov2.PrometheusAddlScrapeConfigSecret {
		return errors.New(p.Spec.AdditionalScrapeConfigs.Name + " is alreay set as additional scrape config in prometheus")
	}

	secret, err := c.kubeClientSet.CoreV1().Secrets(ns).Get(ctx, miniov2.PrometheusAddlScrapeConfigSecret, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	promCfg := configmaps.GetPrometheusConfig(tenant, accessKey, secretKey)

	// If the secret is not found, create the secret
	if k8serrors.IsNotFound(err) {
		klog.Infof("Adding MinIO tenant Prometheus scrape config")
		scrapeCfgYaml, err := yaml.Marshal(&promCfg.ScrapeConfigs)
		if err != nil {
			return err
		}
		secret = &corev1.Secret{
			Type: "Opaque",
			ObjectMeta: metav1.ObjectMeta{
				Name:      miniov2.PrometheusAddlScrapeConfigSecret,
				Namespace: ns,
			},
			Data: map[string][]byte{
				miniov2.PrometheusAddlScrapeConfigKey: scrapeCfgYaml,
			},
		}
		_, err = c.kubeClientSet.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		var scrapeConfigs []configmaps.ScrapeConfig
		err := yaml.Unmarshal(secret.Data[miniov2.PrometheusAddlScrapeConfigKey], &scrapeConfigs)
		if err != nil {
			return err
		}
		// Check if the scrape config is already present
		hasScrapeConfig := false
		for _, sc := range scrapeConfigs {
			if sc.JobName == tenant.PrometheusOperatorAddlConfigJobName() {
				hasScrapeConfig = true
				break
			}
		}
		if !hasScrapeConfig {
			klog.Infof("Adding MinIO tenant Prometheus scrape config")
			scrapeConfigs = append(scrapeConfigs, promCfg.ScrapeConfigs...)
			scrapeCfgYaml, err := yaml.Marshal(scrapeConfigs)
			if err != nil {
				return err
			}
			secret.Data[miniov2.PrometheusAddlScrapeConfigKey] = scrapeCfgYaml
			_, err = c.kubeClientSet.CoreV1().Secrets(ns).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}

	// Update prometheus if its not done alreay
	if p.Spec.AdditionalScrapeConfigs == nil {
		p.Spec.AdditionalScrapeConfigs = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: miniov2.PrometheusAddlScrapeConfigSecret},
			Key:                  miniov2.PrometheusAddlScrapeConfigKey,
		}

		_, err = c.promClient.MonitoringV1().Prometheuses(ns).Update(ctx, p, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) deletePrometheusAddlConfig(ctx context.Context, tenant *miniov2.Tenant) error {
	ns := miniov2.GetPrometheusNamespace()

	secret, err := c.kubeClientSet.CoreV1().Secrets(ns).Get(ctx, miniov2.PrometheusAddlScrapeConfigSecret, metav1.GetOptions{})
	// If the secret is not found, return from here
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var scrapeConfigs []configmaps.ScrapeConfig
	err = yaml.Unmarshal(secret.Data[miniov2.PrometheusAddlScrapeConfigKey], &scrapeConfigs)
	if err != nil {
		return err
	}
	// Check if the scrape config is present
	hasScrapeConfig := false
	scIndex := -1
	for i, sc := range scrapeConfigs {
		if sc.JobName == tenant.PrometheusOperatorAddlConfigJobName() {
			hasScrapeConfig = true
			scIndex = i
			break
		}
	}
	if hasScrapeConfig {
		klog.Infof("Deleting MinIO tenant Prometheus scrape config")
		// Delete the config
		newScrapeConfigs := append(scrapeConfigs[:scIndex], scrapeConfigs[scIndex+1:]...)
		// Update the secret
		scrapeCfgYaml, err := yaml.Marshal(newScrapeConfigs)
		if err != nil {
			return err
		}
		secret.Data[miniov2.PrometheusAddlScrapeConfigKey] = scrapeCfgYaml
		_, err = c.kubeClientSet.CoreV1().Secrets(ns).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
