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
	"os"
	"reflect"
	"strings"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
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

func (c *Controller) getPrometheuses(ctx context.Context) ([]*promv1.Prometheus, error) {
	ns := miniov2.GetPrometheusNamespace()
	promName := miniov2.GetPrometheusName()

	// If in namespace scope, only look in current namespace
	if os.Getenv("OPERATOR_SCOPE") == "namespace" {
		ns = miniov2.GetNSFromFile()
	}

	var pList []*promv1.Prometheus

	if promName != "" {
		p, err := c.promClient.MonitoringV1().Prometheuses(ns).Get(ctx, promName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		pList = append(pList, p)
	} else {
		promList, err := c.promClient.MonitoringV1().Prometheuses(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		pList = promList.Items
	}
	return pList, nil
}

func (c *Controller) getPrometheusAgents(ctx context.Context) ([]*promv1alpha1.PrometheusAgent, error) {
	ns := miniov2.GetPrometheusNamespace()
	promName := miniov2.GetPrometheusName()

	var pList []*promv1alpha1.PrometheusAgent

	if promName != "" {
		p, err := c.promClient.MonitoringV1alpha1().PrometheusAgents(ns).Get(ctx, promName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		pList = append(pList, p)
	} else {
		promAgentList, err := c.promClient.MonitoringV1alpha1().PrometheusAgents(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		pList = promAgentList.Items
	}
	return pList, nil
}

func (c *Controller) getPrometheus(ctx context.Context) (promv1.PrometheusInterface, error) {
	ns := miniov2.GetPrometheusNamespace()

	var instances []promv1.PrometheusInterface
	proms, err := c.getPrometheuses(ctx)
	if err != nil {
		return nil, err
	}
	for _, prom := range proms {
		instances = append(instances, prom) // Append Prometheus instances to the interface slice
	}

	promAgents, err := c.getPrometheusAgents(ctx)
	if err != nil {
		return nil, err
	}
	for _, promAgent := range promAgents {
		instances = append(instances, promAgent) // Append PrometheusAgent instances to the interface slice
	}

	if len(instances) == 0 {
		return nil, errors.New("no Prometheus or PrometheusAgent found in namespace " + ns)
	}
	if len(instances) > 1 {
		return nil, errors.New("more than one Prometheus or PrometheusAgent instance found in namespace " + ns + ".")
	}

	return instances[0], nil
}

func (c *Controller) checkAndCreatePrometheusAddlConfig(ctx context.Context, tenant *miniov2.Tenant, accessKey, secretKey string) error {
	ns := miniov2.GetPrometheusNamespace()

	p, err := c.getPrometheus(ctx)
	if err != nil {
		return err
	}

	// We use common fields of Prometheus & PrometheusAgents to make sure we only operate on fields available on both types.
	cpf := p.GetCommonPrometheusFields()
	// If the additional scrape config is set to something else, we will error out
	if cpf.AdditionalScrapeConfigs != nil && cpf.AdditionalScrapeConfigs.Name != miniov2.PrometheusAddlScrapeConfigSecret {
		return errors.New(cpf.AdditionalScrapeConfigs.Name + " is alreay set as additional scrape config in prometheus")
	}

	secret, err := c.kubeClientSet.CoreV1().Secrets(ns).Get(ctx, miniov2.PrometheusAddlScrapeConfigSecret, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	promCfg := configmaps.GetPrometheusConfig(tenant, accessKey, secretKey)

	// If the secret is not found, create the secret
	if k8serrors.IsNotFound(err) {
		klog.Infof("Adding MinIO tenant %s/%s Prometheus scrape config", ns, tenant.Name)
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
		var scrapeConfigs, expectedScrapeConfigs []configmaps.ScrapeConfig
		err := yaml.Unmarshal(secret.Data[miniov2.PrometheusAddlScrapeConfigKey], &scrapeConfigs)
		if err != nil {
			return err
		}
		// get other scrape configs
		for _, sc := range scrapeConfigs {
			if !strings.HasPrefix(sc.JobName, tenant.PrometheusOperatorAddlConfigJobName()) {
				expectedScrapeConfigs = append(expectedScrapeConfigs, sc)
			}
		}
		ignoreScrapeConfigsIndex := len(expectedScrapeConfigs)
		expectedScrapeConfigs = append(expectedScrapeConfigs, promCfg.ScrapeConfigs...)
		updateScrapeConfig := false
		if len(scrapeConfigs) != len(expectedScrapeConfigs) {
			updateScrapeConfig = true
		} else {
			for i := range scrapeConfigs {
				// can't compare that is generated by operator
				if i < ignoreScrapeConfigsIndex {
					continue
				}
				if scrapeConfigs[i].JobName != expectedScrapeConfigs[i].JobName ||
					scrapeConfigs[i].MetricsPath != expectedScrapeConfigs[i].MetricsPath ||
					scrapeConfigs[i].Scheme != expectedScrapeConfigs[i].Scheme ||
					!reflect.DeepEqual(scrapeConfigs[i].TLSConfig, expectedScrapeConfigs[i].TLSConfig) ||
					!reflect.DeepEqual(scrapeConfigs[i].StaticConfigs, expectedScrapeConfigs[i].StaticConfigs) {
					updateScrapeConfig = true
					break
				}
				accKey, _ := miniov2.GetAccessKeyFromBearerToken(scrapeConfigs[i].BearerToken, secretKey)
				if accKey != accessKey {
					updateScrapeConfig = true
					break
				}
			}
		}
		if updateScrapeConfig {
			klog.Infof("Updating MinIO tenant %s/%s Prometheus scrape config", ns, tenant.Name)
			scrapeCfgYaml, err := yaml.Marshal(expectedScrapeConfigs)
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
	if cpf.AdditionalScrapeConfigs == nil {
		cpf.AdditionalScrapeConfigs = &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: miniov2.PrometheusAddlScrapeConfigSecret},
			Key:                  miniov2.PrometheusAddlScrapeConfigKey,
		}
		p.SetCommonPrometheusFields(cpf)

		switch prom := p.(type) {
		case *promv1alpha1.PrometheusAgent:
			_, err = c.promClient.MonitoringV1alpha1().PrometheusAgents(ns).Update(ctx, prom, metav1.UpdateOptions{})
		case *promv1.Prometheus:
			_, err = c.promClient.MonitoringV1().Prometheuses(ns).Update(ctx, prom, metav1.UpdateOptions{})
		}

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

	var scrapeConfigs, exceptedScrapeConfigs []configmaps.ScrapeConfig
	err = yaml.Unmarshal(secret.Data[miniov2.PrometheusAddlScrapeConfigKey], &scrapeConfigs)
	if err != nil {
		return err
	}
	for _, sc := range scrapeConfigs {
		if !strings.HasPrefix(sc.JobName, tenant.PrometheusOperatorAddlConfigJobName()) {
			exceptedScrapeConfigs = append(exceptedScrapeConfigs, sc)
		}
	}
	if !reflect.DeepEqual(scrapeConfigs, exceptedScrapeConfigs) {
		klog.Infof("Deleting MinIO tenant Prometheus scrape config")
		// Update the secret
		scrapeCfgYaml, err := yaml.Marshal(exceptedScrapeConfigs)
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
