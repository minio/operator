// This file is part of MinIO Operator
// Copyright (c) 2020 MinIO, Inc.
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
	"fmt"

	"github.com/minio/madmin-go"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/subnet"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) registerTenantInSubnet(ctx context.Context, tenant *miniov2.Tenant, adminClnt *madmin.AdminClient) error {
	apiKey, err := c.getTenantAPIKey(ctx, tenant, adminClnt)
	if err != nil {
		return err
	}
	license, err := c.getTenantLicense(ctx, tenant, adminClnt)
	if err != nil {
		return err
	}
	return c.registerTenant(ctx, tenant, adminClnt, apiKey, license)
}

func (c *Controller) getTenantAPIKey(ctx context.Context, tenant *miniov2.Tenant, adminClnt *madmin.AdminClient) (string, error) {
	s, err := c.kubeClientSet.CoreV1().Secrets("default").Get(ctx, "operator-subnet", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if apiKey, ok := s.Data["api-key"]; ok {
		return string(apiKey), nil
	}

	return "", errors.New("api-key not found in secret")
}

func (c *Controller) getTenantLicense(ctx context.Context, tenant *miniov2.Tenant, adminClnt *madmin.AdminClient) (*subnet.LicenseTokenConfig, error) {
	license, err := subnet.GetSubnetKeyFromMinIOConfig(ctx, adminClnt)
	if err != nil {
		return nil, err
	}
	if license.License != "" {
		return nil, errors.New("tenant is already registered")
	}
	return license, nil
}

func (c *Controller) registerTenant(
	ctx context.Context, tenant *miniov2.Tenant, adminClnt *madmin.AdminClient,
	apiKey string, license *subnet.LicenseTokenConfig,
) error {
	serverInfo, err := adminClnt.ServerInfo(ctx)
	if err != nil {
		return err
	}
	res, err := subnet.RegisterWithAPIKey(serverInfo, apiKey)
	if err != nil {
		return err
	}
	// Keep existing subnet proxy if exists
	configStr := fmt.Sprintf("subnet license=%s api_key=%s proxy=%s", res.License, res.APIKey, license.Proxy)
	_, err = adminClnt.SetConfigKV(ctx, configStr)
	return err
}
