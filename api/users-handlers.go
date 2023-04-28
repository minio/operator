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

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func registerUsersHandlers(api *operations.OperatorAPI) {
	// Set Tenant Administrators
	api.OperatorAPISetTenantAdministratorsHandler = operator_api.SetTenantAdministratorsHandlerFunc(func(params operator_api.SetTenantAdministratorsParams, session *models.Principal) middleware.Responder {
		err := getSetTenantAdministratorsResponse(session, params)
		if err != nil {
			return operator_api.NewSetTenantAdministratorsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewSetTenantAdministratorsNoContent()
	})
}

func getSetTenantAdministratorsResponse(session *models.Principal, params operator_api.SetTenantAdministratorsParams) *models.Error {
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
	k8sClient := &k8sClient{
		client: clientSet,
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	svcURL := GetTenantServiceURL(minTenant)
	// getTenantAdminClient will use all certificates under ~/.console/certs/CAs to trust the TLS connections with MinIO tenants
	mAdmin, err := getTenantAdminClient(
		ctx,
		k8sClient,
		minTenant,
		svcURL,
	)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	// create a minioClient interface implementation
	// defining the client to be used
	adminClient := AdminClient{Client: mAdmin}
	return setTenantAdministrators(ctx, minTenant, k8sClient, adminClient, params)
}

func setTenantAdministrators(ctx context.Context, minTenant *miniov2.Tenant, k8sClient K8sClientI, adminClient MinioAdmin, params operator_api.SetTenantAdministratorsParams) *models.Error {
	minTenant.EnsureDefaults()

	for _, user := range params.Body.UserDNS {
		if err := SetPolicy(ctx, adminClient, "consoleAdmin", user, "user"); err != nil {
			return ErrorWithContext(ctx, err)
		}
	}
	for _, group := range params.Body.GroupDNS {
		if err := SetPolicy(ctx, adminClient, "consoleAdmin", group, "group"); err != nil {
			return ErrorWithContext(ctx, err)
		}
	}
	return nil
}
