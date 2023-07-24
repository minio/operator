// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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
	"sort"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerConfigurationHandlers(api *operations.OperatorAPI) {
	// Tenant Configuration details
	// Tenant Security details
	api.OperatorAPITenantConfigurationHandler = operator_api.TenantConfigurationHandlerFunc(func(params operator_api.TenantConfigurationParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantConfigurationResponse(session, params)
		if err != nil {
			return operator_api.NewTenantConfigurationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantConfigurationOK().WithPayload(resp)
	})
	// Update Tenant Configuration
	api.OperatorAPIUpdateTenantConfigurationHandler = operator_api.UpdateTenantConfigurationHandlerFunc(func(params operator_api.UpdateTenantConfigurationParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantConfigurationResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantConfigurationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantConfigurationNoContent()
	})
}

func getTenantConfigurationResponse(session *models.Principal, params operator_api.TenantConfigurationParams) (*models.TenantConfigurationResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	k8sClient := &k8sClient{
		client: clientSet,
	}
	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return parseTenantConfiguration(ctx, k8sClient, minTenant)
}

func parseTenantConfiguration(ctx context.Context, k8sClient K8sClientI, minTenant *miniov2.Tenant) (*models.TenantConfigurationResponse, *models.Error) {
	tenantConfiguration, err := GetTenantConfiguration(ctx, k8sClient, minTenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	delete(tenantConfiguration, "accesskey")
	delete(tenantConfiguration, "secretkey")
	var envVars []*models.EnvironmentVariable
	for key, value := range tenantConfiguration {
		envVars = append(envVars, &models.EnvironmentVariable{
			Key:   key,
			Value: value,
		})
	}
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Key < envVars[j].Key
	})
	configurationInfo := &models.TenantConfigurationResponse{EnvironmentVariables: envVars}
	if minTenant.Spec.Features != nil && minTenant.Spec.Features.EnableSFTP != nil {
		configurationInfo.SftpExposed = *minTenant.Spec.Features.EnableSFTP
	}
	return configurationInfo, nil
}

func getUpdateTenantConfigurationResponse(session *models.Principal, params operator_api.UpdateTenantConfigurationParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	k8sClient := k8sClient{
		client: clientSet,
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	if err := updateTenantConfigurationFile(ctx, opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant configuration"))
	}
	return nil
}

func updateTenantConfigurationFile(ctx context.Context, operatorClient OperatorClientI, client K8sClientI, namespace string, params operator_api.UpdateTenantConfigurationParams) error {
	tenant, err := operatorClient.TenantGet(ctx, namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return err
	}
	tenantConfiguration, err := GetTenantConfiguration(ctx, client, tenant)
	if err != nil {
		return err
	}

	delete(tenantConfiguration, "accesskey")
	delete(tenantConfiguration, "secretkey")

	requestBody := params.Body
	if requestBody == nil {
		return errors.New("missing request body")
	}
	// Patch tenant configuration file with the new values provided by the user
	for _, envVar := range requestBody.EnvironmentVariables {
		if envVar.Key == "" {
			continue
		}
		tenantConfiguration[envVar.Key] = envVar.Value
	}
	// Remove existing values from configuration file
	for _, keyToBeDeleted := range requestBody.KeysToBeDeleted {
		delete(tenantConfiguration, keyToBeDeleted)
	}

	if !tenant.HasConfigurationSecret() {
		return errors.New("tenant configuration file not found")
	}
	tenantConfigurationSecret, err := client.getSecret(ctx, tenant.Namespace, tenant.Spec.Configuration.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	tenantConfigurationSecret.Data["config.env"] = []byte(GenerateTenantConfigurationFile(tenantConfiguration))
	_, err = client.updateSecret(ctx, namespace, tenantConfigurationSecret, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	// Update SFTP flag
	if tenant.Spec.Features != nil {
		tenant.Spec.Features.EnableSFTP = &requestBody.SftpExposed
		_, err = operatorClient.TenantUpdate(ctx, tenant, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// Restart all MinIO pods at the same time for they to take the new configuration
	err = client.deletePodCollection(ctx, namespace, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	})
	if err != nil {
		return err
	}

	return nil
}
