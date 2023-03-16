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

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/configmaps"
)

// MinIOPrometheusMetrics holds metrics pulled from prometheus
type MinIOPrometheusMetrics struct {
	UsableCapacity int64
	Usage          int64
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

	// Update prometheus if it's not done alreay
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
