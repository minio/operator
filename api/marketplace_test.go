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
	"os"
	"testing"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	testWithError       = false
	testServerWithError = false
	errMock             = errors.New("mock error")
)

type MarketplaceTestSuite struct {
	suite.Suite
	assert           *assert.Assertions
	kubernetesClient k8sClientMock
	namespace        string
	postServer       *httptest.Server
}

func (suite *MarketplaceTestSuite) SetupSuite() {
	suite.assert = assert.New(suite.T())
	suite.namespace = "default"
	k8sClientGetConfigMapMock = suite.getConfigMapMock
	k8sClientCreateConfigMapMock = suite.createConfigMapMock
	k8sClientUpdateConfigMapMock = suite.updateConfigMapMock
	k8sClientDeleteConfigMapMock = suite.deleteConfigMapMock
	os.Setenv(mpConfigMapKey, "mp-mock-config")
	suite.postServer = httptest.NewServer(http.HandlerFunc(suite.postHandler))
}

func (suite *MarketplaceTestSuite) postHandler(
	w http.ResponseWriter, r *http.Request,
) {
	if testServerWithError {
		w.WriteHeader(400)
	} else {
		fmt.Fprintf(w, `{"post": "Post response"}`)
	}
}

func (suite *MarketplaceTestSuite) TearDownSuite() {
	os.Unsetenv(mpConfigMapKey)
}

func (suite *MarketplaceTestSuite) getConfigMapMock(ctx context.Context, namespace, configMap string, opts metav1.GetOptions) (*corev1.ConfigMap, error) {
	if testWithError {
		return nil, errMock
	}
	return &corev1.ConfigMap{Data: map[string]string{isMPEmailSet: "true"}}, nil
}

func (suite *MarketplaceTestSuite) createConfigMapMock(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.CreateOptions) (*corev1.ConfigMap, error) {
	if testWithError {
		return nil, errMock
	}
	return &corev1.ConfigMap{}, nil
}

func (suite *MarketplaceTestSuite) updateConfigMapMock(ctx context.Context, namespace string, cm *corev1.ConfigMap, opts metav1.UpdateOptions) (*corev1.ConfigMap, error) {
	if testWithError {
		return nil, errMock
	}
	return &corev1.ConfigMap{}, nil
}

func (suite *MarketplaceTestSuite) deleteConfigMapMock(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) error {
	if testWithError {
		return errMock
	}
	return nil
}

func (suite *MarketplaceTestSuite) TestRegisterMarketplaceHandlers() {
	api := &operations.OperatorAPI{}
	suite.assert.Nil(api.OperatorAPIGetMPIntegrationHandler)
	suite.assert.Nil(api.OperatorAPIPostMPIntegrationHandler)
	registerMarketplaceHandlers(api)
	suite.assert.NotNil(api.OperatorAPIGetMPIntegrationHandler)
	suite.assert.NotNil(api.OperatorAPIPostMPIntegrationHandler)
}

func (suite *MarketplaceTestSuite) TestGetMPIntegrationHandlerWithError() {
	api := &operations.OperatorAPI{}
	registerMarketplaceHandlers(api)
	params := operator_api.NewGetMPIntegrationParams()
	params.HTTPRequest = &http.Request{}
	response := api.OperatorAPIGetMPIntegrationHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetMPIntegrationDefault)
	suite.assert.True(ok)
}

func (suite *MarketplaceTestSuite) TestPostMPIntegrationHandlerWithError() {
	api := &operations.OperatorAPI{}
	registerMarketplaceHandlers(api)
	params := operator_api.NewPostMPIntegrationParams()
	params.Body = &models.MpIntegration{Email: ""}
	params.HTTPRequest = &http.Request{}
	params.HTTPRequest.Header = map[string][]string{}
	params.HTTPRequest.AddCookie(&http.Cookie{Value: "token", Name: "token"})
	response := api.OperatorAPIPostMPIntegrationHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.PostMPIntegrationDefault)
	suite.assert.True(ok)
}

func (suite *MarketplaceTestSuite) TestGetMPEmailWithError() {
	testWithError = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	email, err := getMPEmail(ctx, &suite.kubernetesClient)
	suite.assert.NotNil(err)
	suite.assert.Empty(email)
}

func (suite *MarketplaceTestSuite) TestGetMPEmailNoError() {
	testWithError = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	isSet, err := getMPEmail(ctx, &suite.kubernetesClient)
	suite.assert.Nil(err)
	suite.assert.True(isSet)
}

func (suite *MarketplaceTestSuite) TestSetMPIntegrationNoEmail() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := setMPIntegration(ctx, "", false, &suite.kubernetesClient)
	suite.assert.NotNil(err)
}

func (suite *MarketplaceTestSuite) TestSetMPIntegrationWithError() {
	testWithError = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	os.Setenv(mpHostEnvVar, "  ")
	err := setMPIntegration(ctx, "mock@mock.com", false, &suite.kubernetesClient)
	suite.assert.NotNil(err)
	os.Unsetenv(mpHostEnvVar)
}

func (suite *MarketplaceTestSuite) TestSetMPIntegrationNoError() {
	testWithError = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	os.Setenv(mpHostEnvVar, suite.postServer.URL)
	err := setMPIntegration(ctx, "mock@mock.com", false, &suite.kubernetesClient)
	suite.assert.Nil(err)
	os.Unsetenv(mpHostEnvVar)
}

func (suite *MarketplaceTestSuite) TestSetMPIntegrationWithRequestError() {
	testWithError = false
	testServerWithError = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	os.Setenv(mpHostEnvVar, suite.postServer.URL)
	err := setMPIntegration(ctx, "mock@mock.com", false, &suite.kubernetesClient)
	suite.assert.NotNil(err)
	os.Unsetenv(mpHostEnvVar)
}

func TestMarketplace(t *testing.T) {
	suite.Run(t, new(MarketplaceTestSuite))
}
