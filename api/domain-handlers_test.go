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

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
)

func (suite *TenantTestSuite) TestUpdateTenantDomainsHandlerWithError() {
	params, api := suite.initUpdateTenantDomainsRequest()
	response := api.OperatorAPIUpdateTenantDomainsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.UpdateTenantDomainsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initUpdateTenantDomainsRequest() (params operator_api.UpdateTenantDomainsParams, api operations.OperatorAPI) {
	registerDomainHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace-domain"
	params.Tenant = "mock-tenant-domain"
	params.Body = &models.UpdateDomainsRequest{
		Domains: &models.DomainsConfiguration{},
	}
	return params, api
}
