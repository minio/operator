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
	"net/http"
	"time"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *TenantTestSuite) TestDeletePodHandlerWithoutError() {
	params, api := suite.initDeletePodRequest()
	response := api.OperatorAPIDeletePodHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.DeletePodDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initDeletePodRequest() (params operator_api.DeletePodParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-tenantmock-pod"
	return params, api
}

func (suite *TenantTestSuite) TestGetTenantPodsHandlerWithError() {
	params, api := suite.initGetTenantPodsRequest()
	response := api.OperatorAPIGetTenantPodsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetTenantPodsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetTenantPodsRequest() (params operator_api.GetTenantPodsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}

func (suite *TenantTestSuite) TestGetTenantPodsWithoutError() {
	pods := getTenantPods(&corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{{}},
				},
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
		},
	})
	suite.assert.Equal(1, len(pods))
}

func (suite *TenantTestSuite) TestGetPodLogsHandlerWithError() {
	params, api := suite.initGetPodLogsRequest()
	response := api.OperatorAPIGetPodLogsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetPodLogsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetPodLogsRequest() (params operator_api.GetPodLogsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mokc-pod"
	return params, api
}

func (suite *TenantTestSuite) TestGetPodEventsHandlerWithError() {
	params, api := suite.initGetPodEventsRequest()
	response := api.OperatorAPIGetPodEventsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetPodEventsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetPodEventsRequest() (params operator_api.GetPodEventsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-pod"
	return params, api
}

func (suite *TenantTestSuite) TestDescribePodHandlerWithError() {
	params, api := suite.initDescribePodRequest()
	response := api.OperatorAPIDescribePodHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.DescribePodDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initDescribePodRequest() (params operator_api.DescribePodParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-pod"
	return params, api
}
