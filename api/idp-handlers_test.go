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

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *TenantTestSuite) TestUpdateTenantIdentityProviderHandlerWithError() {
	params, api := suite.initUpdateTenantIdentityProviderRequest()
	response := api.OperatorAPIUpdateTenantIdentityProviderHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.UpdateTenantIdentityProviderDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initUpdateTenantIdentityProviderRequest() (params operator_api.UpdateTenantIdentityProviderParams, api operations.OperatorAPI) {
	registerIDPHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.Body = &models.IdpConfiguration{}
	return params, api
}

func (suite *TenantTestSuite) TestUpdateTenantIdentityProviderWithTenantError() {
	ctx := context.Background()
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return nil, errors.New("mock")
	}
	params, _ := suite.initUpdateTenantIdentityProviderRequest()
	err := updateTenantIdentityProvider(ctx, suite.opClient, suite.k8sclient, "mock-namespace", params)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestUpdateTenantIdentityProviderWithTenantConfigurationError() {
	ctx := context.Background()
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				CredsSecret: &corev1.LocalObjectReference{
					Name: "mock",
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initUpdateTenantIdentityProviderRequest()
	err := updateTenantIdentityProvider(ctx, suite.opClient, suite.k8sclient, "mock-namespace", params)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestUpdateTenantIdentityProviderWithSecretCreationError() {
	ctx := context.Background()
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				Env: []corev1.EnvVar{
					{Name: "mock", Value: "mock"},
				},
			},
		}, nil
	}
	k8sClientCreateSecretMock = func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
		return nil, errors.New("mock-create-error")
	}
	params, _ := suite.initUpdateTenantIdentityProviderRequest()
	err := updateTenantIdentityProvider(ctx, suite.opClient, suite.k8sclient, "mock-namespace", params)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestUpdateTenantIdentityProviderWithoutError() {
	ctx := context.Background()
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{}, nil
	}
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{}, nil
	}
	k8sClientCreateSecretMock = func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
		return nil, nil
	}
	params, _ := suite.initUpdateTenantIdentityProviderRequest()
	params.Body.ActiveDirectory = &models.IdpConfigurationActiveDirectory{}
	configURL := "mock"
	clientID := "mock"
	clientSecret := "mock"
	claimName := "mock"
	params.Body.Oidc = &models.IdpConfigurationOidc{
		ConfigurationURL: &configURL,
		ClientID:         &clientID,
		SecretID:         &clientSecret,
		ClaimName:        &claimName,
	}
	params.Body.ActiveDirectory = &models.IdpConfigurationActiveDirectory{
		URL:                 &configURL,
		LookupBindDn:        &claimName,
		SkipTLSVerification: true,
		ServerInsecure:      true,
		ServerStartTLS:      true,
	}
	err := updateTenantIdentityProvider(ctx, suite.opClient, suite.k8sclient, "mock-namespace", params)
	suite.assert.Nil(err)
}
