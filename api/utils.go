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

package api

import (
	"context"
	"errors"
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetTenantConfiguration returns the config for a tenant
func GetTenantConfiguration(ctx context.Context, clientSet K8sClientI, tenant *miniov2.Tenant) (map[string]string, error) {
	if tenant == nil {
		return nil, errors.New("tenant cannot be nil")
	}
	tenantConfiguration := map[string]string{}
	for _, config := range tenant.GetEnvVars() {
		tenantConfiguration[config.Name] = config.Value
	}
	// legacy support for tenants with tenant.spec.credsSecret
	if tenant.HasCredsSecret() {
		minioSecret, err := clientSet.getSecret(ctx, tenant.Namespace, tenant.Spec.CredsSecret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configFromCredsSecret := minioSecret.Data
		for key, val := range configFromCredsSecret {
			tenantConfiguration[key] = string(val)
		}
	}
	if tenant.HasConfigurationSecret() {
		minioConfigurationSecret, err := clientSet.getSecret(ctx, tenant.Namespace, tenant.Spec.Configuration.Name, metav1.GetOptions{})
		if err != nil {
			return tenantConfiguration, err
		}
		if minioConfigurationSecret == nil {
			return tenantConfiguration, errors.New("tenant configuration secret is empty")
		}
		configFromFile := miniov2.ParseRawConfiguration(minioConfigurationSecret.Data["config.env"])
		for key, val := range configFromFile {
			tenantConfiguration[key] = string(val)
		}
	}
	return tenantConfiguration, nil
}

// GenerateTenantConfigurationFile generate config for tenant
func GenerateTenantConfigurationFile(configuration map[string]string) string {
	var rawConfiguration string
	for key, val := range configuration {
		rawConfiguration += fmt.Sprintf("export %s=\"%s\"\n", key, val)
	}
	return rawConfiguration
}

// Create a copy of a string and return its pointer
func stringPtr(str string) *string {
	return &str
}
