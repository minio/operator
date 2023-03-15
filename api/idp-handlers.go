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

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
)

func registerIDPHandlers(api *operations.OperatorAPI) {
	// Tenant identity provider details
	api.OperatorAPITenantIdentityProviderHandler = operator_api.TenantIdentityProviderHandlerFunc(func(params operator_api.TenantIdentityProviderParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantIdentityProviderResponse(session, params)
		if err != nil {
			return operator_api.NewTenantIdentityProviderDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantIdentityProviderOK().WithPayload(resp)
	})

	// Update Tenant identity provider configuration
	api.OperatorAPIUpdateTenantIdentityProviderHandler = operator_api.UpdateTenantIdentityProviderHandlerFunc(func(params operator_api.UpdateTenantIdentityProviderParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantIdentityProviderResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantIdentityProviderDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantIdentityProviderNoContent()
	})
}

func getTenantIdentityProviderResponse(session *models.Principal, params operator_api.TenantIdentityProviderParams) (*models.IdpConfiguration, *models.Error) {
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
	k8sClient := k8sClient{
		client: clientSet,
	}
	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	info, err := getTenantIdentityProvider(ctx, &k8sClient, minTenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return info, nil
}

func getUpdateTenantIdentityProviderResponse(session *models.Principal, params operator_api.UpdateTenantIdentityProviderParams) *models.Error {
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
	if err := updateTenantIdentityProvider(ctx, opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant"))
	}
	return nil
}
