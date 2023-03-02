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
	"net/http"
	"testing"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantLogsTestSuite struct {
	suite.Suite
	assert   *assert.Assertions
	opClient opClientMock
}

func (suite *TenantLogsTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())
	suite.opClient = opClientMock{}
}

func (suite *TenantLogsTestSuite) SetupTest() {
}

func (suite *TenantLogsTestSuite) TearDownSuite() {
}

func (suite *TenantLogsTestSuite) TearDownTest() {
}

func (suite *TenantLogsTestSuite) TestRegisterTenantLogsHandlers() {
	api := &operations.OperatorAPI{}
	suite.assertHandlersAreNil(api)
	registerTenantLogsHandlers(api)
	suite.assertHandlersAreNotNil(api)
}

func (suite *TenantLogsTestSuite) assertHandlersAreNil(api *operations.OperatorAPI) {
	suite.assert.Nil(api.OperatorAPIGetTenantLogsHandler)
	suite.assert.Nil(api.OperatorAPISetTenantLogsHandler)
	suite.assert.Nil(api.OperatorAPIEnableTenantLoggingHandler)
	suite.assert.Nil(api.OperatorAPIDisableTenantLoggingHandler)
}

func (suite *TenantLogsTestSuite) assertHandlersAreNotNil(api *operations.OperatorAPI) {
	suite.assert.NotNil(api.OperatorAPIGetTenantLogsHandler)
	suite.assert.NotNil(api.OperatorAPISetTenantLogsHandler)
	suite.assert.NotNil(api.OperatorAPIEnableTenantLoggingHandler)
	suite.assert.NotNil(api.OperatorAPIDisableTenantLoggingHandler)
}

func (suite *TenantLogsTestSuite) TestGetTenantLogsHandlerWithError() {
	params, api := suite.initGetTenantLogsRequest()
	response := api.OperatorAPIGetTenantLogsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetTenantLogsDefault)
	suite.assert.True(ok)
}

func (suite *TenantLogsTestSuite) initGetTenantLogsRequest() (params operator_api.GetTenantLogsParams, api operations.OperatorAPI) {
	registerTenantLogsHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-name"
	return params, api
}

func (suite *TenantLogsTestSuite) TestGetTenantLogsInfoWithoutLogAndNoError() {
	tenant := suite.createMockTenant(false, true)
	tenantLogs := getTenantLogsInfo(tenant)
	suite.assert.True(tenantLogs.Disabled)
}

func (suite *TenantLogsTestSuite) TestGetTenantLogsInfoWithLogAndNoError() {
	tenant := suite.createMockTenant(true, true)
	tenantLogs := getTenantLogsInfo(tenant)
	suite.assert.False(tenantLogs.Disabled)
	suite.assert.NotNil(tenantLogs.Labels)
	suite.assert.NotNil(tenantLogs.Annotations)
}

func (suite *TenantLogsTestSuite) TestSetTenantLogsWithoutError() {
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return tenant, nil
	}
	tenant := suite.createMockTenant(true, true)
	success, err := setTenantLogs(context.Background(), tenant, suite.opClient, suite.createMockParams())
	suite.assert.True(success)
	suite.assert.Nil(err)
}

func (suite *TenantLogsTestSuite) TestSetTenantLogsWithoutDBAndError() {
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return tenant, nil
	}
	tenant := suite.createMockTenant(true, false)
	params := suite.createMockParams()
	params.Data.LogDBCPURequest = ""
	params.Data.LogDBMemRequest = ""
	success, err := setTenantLogs(context.Background(), tenant, suite.opClient, params)
	suite.assert.True(success)
	suite.assert.Nil(err)
}

func (suite *TenantLogsTestSuite) TestSetTenantLogsWithError() {
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return nil, errors.New("mock-error")
	}
	tenant := suite.createMockTenant(true, true)
	success, err := setTenantLogs(context.Background(), tenant, suite.opClient, suite.createMockParams())
	suite.assert.False(success)
	suite.assert.NotNil(err)
}

func (suite *TenantLogsTestSuite) TestEnableTenantLoggingWithoutError() {
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return tenant, nil
	}
	tenant := suite.createMockTenant(true, true)
	success, err := enableTenantLogging(context.Background(), tenant, suite.opClient, "mock-tenant")
	suite.assert.True(success)
	suite.assert.Nil(err)
}

func (suite *TenantLogsTestSuite) TestEnableTenantLoggingWithError() {
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return nil, errors.New("mock-error")
	}
	tenant := suite.createMockTenant(true, true)
	success, err := enableTenantLogging(context.Background(), tenant, suite.opClient, "mock-tenant")
	suite.assert.False(success)
	suite.assert.NotNil(err)
}

func (suite *TenantLogsTestSuite) createMockParams() operator_api.SetTenantLogsParams {
	runAsUser := "1000"
	runAsGroup := "1000"
	fsGroup := "1000"
	return operator_api.SetTenantLogsParams{
		Data: &models.TenantLogs{
			Labels: []*models.Label{
				{
					Key:   "key",
					Value: "value",
				},
			},
			Annotations: []*models.Annotation{
				{
					Key:   "key",
					Value: "value",
				},
			},
			NodeSelector: []*models.NodeSelector{
				{
					Key:   "key",
					Value: "value",
				},
			},
			LogCPURequest: "5Gi",
			LogMemRequest: "5Gi",
			DbLabels: []*models.Label{
				{
					Key:   "key",
					Value: "value",
				},
			},
			DbAnnotations: []*models.Annotation{
				{
					Key:   "key",
					Value: "value",
				},
			},
			DbNodeSelector: []*models.NodeSelector{
				{
					Key:   "key",
					Value: "value",
				},
			},
			LogDBCPURequest: "5Gi",
			LogDBMemRequest: "5Gi",
			Image:           "mock-image",
			SecurityContext: &models.SecurityContext{
				RunAsUser:  &runAsUser,
				RunAsGroup: &runAsGroup,
				FsGroup:    fsGroup,
			},
			DiskCapacityGB:     "10",
			ServiceAccountName: "mock-service-account-name",
			DbImage:            "mock-db-image",
			DbSecurityContext: &models.SecurityContext{
				RunAsUser:  &runAsUser,
				RunAsGroup: &runAsGroup,
				FsGroup:    fsGroup,
			},
		},
	}
}

func (suite *TenantLogsTestSuite) createMockTenant(withSpecLog, withDB bool) *miniov2.Tenant {
	cap := 10
	runAsUser := int64(1000)
	runAsGroup := int64(1000)
	fsGroup := int64(1000)
	tenant := &miniov2.Tenant{}
	if withSpecLog {
		tenant.Spec.Log = &miniov2.LogConfig{
			Annotations: map[string]string{
				"key": "value",
			},
			Labels: map[string]string{
				"key": "value",
			},
			NodeSelector: map[string]string{
				"key": "value",
			},
			Db: &miniov2.LogDbConfig{
				Annotations: map[string]string{
					"key": "value",
				},
				Labels: map[string]string{
					"key": "value",
				},
				NodeSelector: map[string]string{
					"key": "value",
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser:  &runAsUser,
					RunAsGroup: &runAsGroup,
					FSGroup:    &fsGroup,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:  &runAsUser,
				RunAsGroup: &runAsGroup,
				FSGroup:    &fsGroup,
			},
			Audit: &miniov2.AuditConfig{
				DiskCapacityGB: &cap,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		}
		if !withDB {
			tenant.Spec.Log.Db = nil
		}
	}
	return tenant
}

func (suite *TenantLogsTestSuite) TestSetTenantLogsHandlerWithError() {
	params, api := suite.initSetTenantLogsRequest()
	response := api.OperatorAPISetTenantLogsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.SetTenantLogsDefault)
	suite.assert.True(ok)
}

func (suite *TenantLogsTestSuite) initSetTenantLogsRequest() (params operator_api.SetTenantLogsParams, api operations.OperatorAPI) {
	registerTenantLogsHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-name"
	return params, api
}

func (suite *TenantLogsTestSuite) TestEnableTenantLoggingHandlerWithError() {
	params, api := suite.initEnableTenantLoggingRequest()
	response := api.OperatorAPIEnableTenantLoggingHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.EnableTenantLoggingDefault)
	suite.assert.True(ok)
}

func (suite *TenantLogsTestSuite) initEnableTenantLoggingRequest() (params operator_api.EnableTenantLoggingParams, api operations.OperatorAPI) {
	registerTenantLogsHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-name"
	return params, api
}

func (suite *TenantLogsTestSuite) TestDisableTenantLoggingHandlerWithError() {
	params, api := suite.initDisableTenantLoggingRequest()
	response := api.OperatorAPIDisableTenantLoggingHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.DisableTenantLoggingDefault)
	suite.assert.True(ok)
}

func (suite *TenantLogsTestSuite) initDisableTenantLoggingRequest() (params operator_api.DisableTenantLoggingParams, api operations.OperatorAPI) {
	registerTenantLogsHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-name"
	return params, api
}

func TestTenantLogs(t *testing.T) {
	suite.Run(t, new(TenantLogsTestSuite))
}
