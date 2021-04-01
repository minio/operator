/*
 * Copyright (C) 2020, MinIO, Inc.
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

package secrets

import (
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PromServiceMonitorSecret creates a secret with Prometheus token to be added in the ServiceMonitor
func PromServiceMonitorSecret(t *miniov2.Tenant, accessKey, secretKey string) *corev1.Secret {
	return &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.PromServiceMonitorSecret(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Data: map[string][]byte{
			miniov2.PrometheusServiceMonitorSecretKey: []byte(t.GenBearerToken(accessKey, secretKey)),
		},
	}
}
