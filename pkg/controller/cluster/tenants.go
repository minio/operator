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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func (c *Controller) getTenantCredentials(ctx context.Context, tenant *miniov2.Tenant) (map[string][]byte, error) {
	// Configuration for tenant can be passed using 3 different sources, tenant.spec.env, k8s credsSecret and config.env secret
	// If the user provides duplicated configuration the override order will be:
	// tenant.Spec.Env < credsSecret (k8s secret) < config.env file (k8s secret)
	tenantConfiguration := map[string][]byte{}

	for _, config := range tenant.GetEnvVars() {
		tenantConfiguration[config.Name] = []byte(config.Value)
	}

	if tenant.HasCredsSecret() {
		minioSecretName := tenant.Spec.CredsSecret.Name
		minioSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, minioSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configFromCredsSecret := minioSecret.Data
		for key, val := range configFromCredsSecret {
			tenantConfiguration[key] = val
		}
	}

	// Load tenant configuration from file
	if tenant.HasConfigurationSecret() {
		minioConfigurationSecretName := tenant.Spec.Configuration.Name
		minioConfigurationSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, minioConfigurationSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configFromFile := miniov2.ParseRawConfiguration(minioConfigurationSecret.Data["config.env"])
		for key, val := range configFromFile {
			tenantConfiguration[key] = val
		}
	}
	return tenantConfiguration, nil
}
