// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/minio/madmin-go/v2"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorSubnetTestSuite struct {
	suite.Suite
	assert                  *assert.Assertions
	loginServer             *httptest.Server
	loginWithError          bool
	loginMFAServer          *httptest.Server
	loginMFAWithError       bool
	getAPIKeyServer         *httptest.Server
	getAPIKeyWithError      bool
	registerAPIKeyServer    *httptest.Server
	registerAPIKeyWithError bool
	k8sClient               k8sClientMock
	adminClient             AdminClientMock
}

func (suite *OperatorSubnetTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())
	suite.loginServer = httptest.NewServer(http.HandlerFunc(suite.loginHandler))
	suite.loginMFAServer = httptest.NewServer(http.HandlerFunc(suite.loginMFAHandler))
	suite.getAPIKeyServer = httptest.NewServer(http.HandlerFunc(suite.getAPIKeyHandler))
	suite.registerAPIKeyServer = httptest.NewServer(http.HandlerFunc(suite.registerAPIKeyHandler))
	suite.k8sClient = k8sClientMock{}
	suite.adminClient = AdminClientMock{}
	k8sClientCreateSecretMock = func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
		return &corev1.Secret{}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		data := make(map[string][]byte)
		data["secretkey"] = []byte("secret")
		data["accesskey"] = []byte("access")
		sec := &corev1.Secret{
			Data: data,
		}
		return sec, nil
	}
}

func (suite *OperatorSubnetTestSuite) loginHandler(
	w http.ResponseWriter, r *http.Request,
) {
	if suite.loginWithError {
		w.WriteHeader(400)
	} else {
		fmt.Fprintf(w, `{"mfa_required": true, "mfa_token": "mockToken"}`)
	}
}

func (suite *OperatorSubnetTestSuite) loginMFAHandler(
	w http.ResponseWriter, r *http.Request,
) {
	if suite.loginMFAWithError {
		w.WriteHeader(400)
	} else {
		fmt.Fprintf(w, `{"token_info": {"access_token": "mockToken"}}`)
	}
}

func (suite *OperatorSubnetTestSuite) getAPIKeyHandler(
	w http.ResponseWriter, r *http.Request,
) {
	if suite.getAPIKeyWithError {
		w.WriteHeader(400)
	} else {
		fmt.Fprintf(w, `{"api_key": "mockAPIKey"}`)
	}
}

func (suite *OperatorSubnetTestSuite) registerAPIKeyHandler(
	w http.ResponseWriter, r *http.Request,
) {
	if suite.registerAPIKeyWithError {
		w.WriteHeader(400)
	} else {
		fmt.Fprintf(w, `{"api_key": "mockAPIKey"}`)
	}
}

func (suite *OperatorSubnetTestSuite) TearDownSuite() {
}

func (suite *OperatorSubnetTestSuite) TestRegisterOperatorSubnetHanlders() {
	api := &operations.OperatorAPI{}
	suite.assert.Nil(api.OperatorAPIOperatorSubnetLoginHandler)
	suite.assert.Nil(api.OperatorAPIOperatorSubnetLoginMFAHandler)
	suite.assert.Nil(api.OperatorAPIOperatorSubnetAPIKeyHandler)
	suite.assert.Nil(api.OperatorAPIOperatorSubnetRegisterAPIKeyHandler)
	registerOperatorSubnetHandlers(api)
	suite.assert.NotNil(api.OperatorAPIOperatorSubnetLoginHandler)
	suite.assert.NotNil(api.OperatorAPIOperatorSubnetLoginMFAHandler)
	suite.assert.NotNil(api.OperatorAPIOperatorSubnetAPIKeyHandler)
	suite.assert.NotNil(api.OperatorAPIOperatorSubnetRegisterAPIKeyHandler)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetLoginHandlerWithEmptyCredentials() {
	params, api := suite.initSubnetLoginRequest("", "")
	response := api.OperatorAPIOperatorSubnetLoginHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetLoginDefault)
	suite.assert.True(ok)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetLoginHandlerWithServerError() {
	params, api := suite.initSubnetLoginRequest("mockusername", "mockpassword")
	suite.loginWithError = true
	os.Setenv(subnet.SubnetURL, suite.loginServer.URL)
	response := api.OperatorAPIOperatorSubnetLoginHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetLoginDefault)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetLoginHandlerWithoutError() {
	params, api := suite.initSubnetLoginRequest("mockusername", "mockpassword")
	suite.loginWithError = false
	os.Setenv(subnet.SubnetURL, suite.loginServer.URL)
	response := api.OperatorAPIOperatorSubnetLoginHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetLoginOK)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) initSubnetLoginRequest(username, password string) (params operator_api.OperatorSubnetLoginParams, api operations.OperatorAPI) {
	registerOperatorSubnetHandlers(&api)
	params.Body = &models.OperatorSubnetLoginRequest{Username: username, Password: password}
	params.HTTPRequest = &http.Request{}
	return params, api
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetLoginMFAHandlerWithServerError() {
	params, api := suite.initSubnetLoginMFARequest("mockusername", "mockMFA", "mockOTP")
	suite.loginMFAWithError = true
	os.Setenv(subnet.SubnetURL, suite.loginMFAServer.URL)
	response := api.OperatorAPIOperatorSubnetLoginMFAHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetLoginMFADefault)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetLoginMFAHandlerWithoutError() {
	params, api := suite.initSubnetLoginMFARequest("mockusername", "mockMFA", "mockOTP")
	suite.loginMFAWithError = false
	os.Setenv(subnet.SubnetURL, suite.loginMFAServer.URL)
	response := api.OperatorAPIOperatorSubnetLoginMFAHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetLoginMFAOK)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) initSubnetLoginMFARequest(username, mfa, otp string) (params operator_api.OperatorSubnetLoginMFAParams, api operations.OperatorAPI) {
	registerOperatorSubnetHandlers(&api)
	params.Body = &models.OperatorSubnetLoginMFARequest{Username: &username, MfaToken: &mfa, Otp: &otp}
	params.HTTPRequest = &http.Request{}
	return params, api
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetAPIKeyHandlerWithServerError() {
	params, api := suite.initSubnetAPIKeyRequest()
	suite.getAPIKeyWithError = true
	os.Setenv(subnet.SubnetURL, suite.getAPIKeyServer.URL)
	response := api.OperatorAPIOperatorSubnetAPIKeyHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetAPIKeyDefault)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetAPIKeyHandlerWithoutError() {
	params, api := suite.initSubnetAPIKeyRequest()
	suite.getAPIKeyWithError = false
	os.Setenv(subnet.SubnetURL, suite.getAPIKeyServer.URL)
	response := api.OperatorAPIOperatorSubnetAPIKeyHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetAPIKeyOK)
	suite.assert.True(ok)
	os.Unsetenv(subnet.SubnetURL)
}

func (suite *OperatorSubnetTestSuite) initSubnetAPIKeyRequest() (params operator_api.OperatorSubnetAPIKeyParams, api operations.OperatorAPI) {
	registerOperatorSubnetHandlers(&api)
	params.HTTPRequest = &http.Request{URL: &url.URL{}}
	return params, api
}

// TODO: Improve register tests (make code more testable)
// func (suite *OperatorSubnetTestSuite) TestOperatorSubnetRegisterAPIKeyHandlerWithServerError() {
// 	params, api := suite.initSubnetRegisterAPIKeyRequest()
// 	suite.registerAPIKeyWithError = true
// 	os.Setenv(subnet.SubnetURL, suite.registerAPIKeyServer.URL)
// 	response := api.OperatorAPIOperatorSubnetRegisterAPIKeyHandler.Handle(params, &models.Principal{})
// 	_, ok := response.(*operator_api.OperatorSubnetRegisterAPIKeyDefault)
// 	suite.assert.True(ok)
// 	os.Unsetenv(subnet.SubnetURL)
// }

// func (suite *OperatorSubnetTestSuite) TestOperatorSubnetRegisterAPIKeyHandlerWithError() {
// 	ctx := context.Background()
// 	suite.registerAPIKeyWithError = false
// 	os.Setenv(subnet.SubnetURL, suite.registerAPIKeyServer.URL)
// 	res, err := registerTenants([]v2.Tenant{{
// 		Spec: v2.TenantSpec{CredsSecret: &corev1.LocalObjectReference{Name: "secret-name"}},
// 	}}, "mockAPIKey", ctx, suite.k8sClient)
// 	suite.assert.Nil(res)
// 	suite.assert.NotNil(err)
// 	os.Unsetenv(subnet.SubnetURL)
// }

// func (suite *OperatorSubnetTestSuite) TestOperatorSubnetRegisterAPIKeyHandlerGetTenants() {
// 	ctx := context.Background()
// 	res, err := getTenantsToRegister(ctx, &models.Principal{})
// 	suite.assert.NotNil(res)
// 	suite.assert.Nil(err)
// }

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetRegisterAPIKeyHandlerWithUnreachableTenant() {
	ctx := context.Background()
	MinioServerInfoMock = func(ctx context.Context) (madmin.InfoMessage, error) {
		return madmin.InfoMessage{}, errors.New("error")
	}
	res, err := registerTenants(ctx, suite.k8sClient, []tenantInterface{
		{
			tenant: v2.Tenant{
				Spec: v2.TenantSpec{CredsSecret: &corev1.LocalObjectReference{Name: "secret-name"}},
			},
			mAdminClient: suite.adminClient,
		},
	}, "mockAPIKey")
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetRegisterAPIKeyHandlerZeroTenants() {
	ctx := context.Background()
	res, err := registerTenants(ctx, suite.k8sClient, []tenantInterface{}, "mockAPIKey")
	suite.assert.NotNil(res)
	suite.assert.Nil(err)
}

// func (suite *OperatorSubnetTestSuite) initSubnetRegisterAPIKeyRequest() (params operator_api.OperatorSubnetRegisterAPIKeyParams, api operations.OperatorAPI) {
// 	registerOperatorSubnetHandlers(&api)
// 	params.Body = &models.OperatorSubnetAPIKey{APIKey: "mockAPIKey"}
// 	params.HTTPRequest = &http.Request{}
// 	return params, api
// }

func (suite *OperatorSubnetTestSuite) TestOperatorSubnetAPIKeyInfoHandlerWithNoSecret() {
	params, api := suite.initSubnetAPIKeyInfoRequest()
	os.Setenv(apiKeySecretEnvVar, "mock-operator-subnet")
	response := api.OperatorAPIOperatorSubnetAPIKeyInfoHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.OperatorSubnetAPIKeyInfoDefault)
	suite.assert.True(ok)
	os.Unsetenv(apiKeySecretEnvVar)
}

// func (suite *OperatorSubnetTestSuite) TestOperatorSubnetAPIKeyInfoHandlerWithSecret() {
// 	params, api := suite.initSubnetAPIKeyInfoRequest()
// 	os.Setenv(apiKeySecretEnvVar, "mock-operator-subnet")
// 	session := &models.Principal{}
// 	clientSet, _ := cluster.K8sClient(session.STSSessionToken)
// 	k8sClient := &k8sClient{client: clientSet}
// 	ctx := context.Background()
// 	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: getAPIKeySecretName()}}
// 	k8sClient.createSecret(ctx, "default", secret, metav1.CreateOptions{})
// 	response := api.OperatorAPIOperatorSubnetAPIKeyInfoHandler.Handle(params, session)
// 	_, ok := response.(*operator_api.OperatorSubnetAPIKeyInfoOK)
// 	suite.assert.True(ok)
// 	k8sClient.deleteSecret(ctx, "default", getAPIKeySecretName(), metav1.DeleteOptions{})
// 	os.Unsetenv(apiKeySecretEnvVar)
// }

func (suite *OperatorSubnetTestSuite) initSubnetAPIKeyInfoRequest() (params operator_api.OperatorSubnetAPIKeyInfoParams, api operations.OperatorAPI) {
	registerOperatorSubnetHandlers(&api)
	params.HTTPRequest = &http.Request{}
	return params, api
}

func TestOperatorSubnet(t *testing.T) {
	suite.Run(t, new(OperatorSubnetTestSuite))
}
