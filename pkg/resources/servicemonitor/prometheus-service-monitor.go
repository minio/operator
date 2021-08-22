/*
 * Copyright (C) 2021, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package servicemonitor

import (
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewForPrometheus creates a new Prometheus ServiceMonitor for prometheus metrics
func NewForPrometheus(t *miniov2.Tenant) *promv1.ServiceMonitor {
	port := miniov2.MinIOServiceHTTPPortName
	scheme := "http"
	if t.TLS() {
		port = miniov2.MinIOServiceHTTPSPortName
		scheme = "https"
	}

	p := &promv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.PrometheusServiceMonitorName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
			Labels:          t.Spec.PrometheusOperator.Labels,
			Annotations:     t.Spec.PrometheusOperator.Annotations,
		},
		Spec: promv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: t.MinIOPodLabelsForSM(),
			},
			Endpoints: []promv1.Endpoint{
				{
					Port:          port,
					Path:          v2.MinIOPrometheusPathNode,
					Scheme:        scheme,
					Interval:      v2.MinIOPrometheusScrapeInterval.String(),
					ScrapeTimeout: v2.MinIOPrometheusScrapeTimeout.String(),
					BearerTokenSecret: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: t.PromServiceMonitorSecret()},
						Key:                  miniov2.PrometheusServiceMonitorSecretKey,
					},
					TLSConfig: &promv1.TLSConfig{
						SafeTLSConfig: promv1.SafeTLSConfig{
							InsecureSkipVerify: true,
						},
					},
				},
				{
					Port:          port,
					Path:          v2.MinIOPrometheusPathCluster,
					Scheme:        scheme,
					Interval:      v2.MinIOPrometheusScrapeInterval.String(),
					ScrapeTimeout: v2.MinIOPrometheusScrapeTimeout.String(),
					BearerTokenSecret: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: t.PromServiceMonitorSecret()},
						Key:                  miniov2.PrometheusServiceMonitorSecretKey,
					},
					TLSConfig: &promv1.TLSConfig{
						SafeTLSConfig: promv1.SafeTLSConfig{
							InsecureSkipVerify: true,
						},
					},
				},
			},
		},
	}
	return p
}
