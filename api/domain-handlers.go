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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerDomainHandlers(api *operations.OperatorAPI) {
	// Update Tenant Domains
	api.OperatorAPIUpdateTenantDomainsHandler = operator_api.UpdateTenantDomainsHandlerFunc(func(params operator_api.UpdateTenantDomainsParams, principal *models.Principal) middleware.Responder {
		err := getUpdateDomainsResponse(principal, params)
		if err != nil {
			return operator_api.NewUpdateTenantDomainsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantDomainsNoContent()
	})
}

func getUpdateDomainsResponse(session *models.Principal, params operator_api.UpdateTenantDomainsParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	operatorCli, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: operatorCli,
	}

	err = updateTenantDomains(ctx, opClient, params.Namespace, params.Tenant, params.Body.Domains)

	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	return nil
}

func updateTenantDomains(ctx context.Context, operatorClient OperatorClientI, namespace string, tenantName string, domainConfig *models.DomainsConfiguration) error {
	minTenant, err := getTenant(ctx, operatorClient, namespace, tenantName)
	if err != nil {
		return err
	}

	var features miniov2.Features
	var domains miniov2.TenantDomains

	// We include current value for BucketDNS. Domains will be overwritten as we are passing all the values that must be saved.
	if minTenant.Spec.Features != nil {
		features = miniov2.Features{
			BucketDNS: minTenant.Spec.Features.BucketDNS,
		}
	}

	if domainConfig != nil {
		// tenant domains
		if domainConfig.Console != "" {
			domains.Console = domainConfig.Console
		}

		if domainConfig.Minio != nil {
			domains.Minio = domainConfig.Minio
		}

		features.Domains = &domains
	}

	minTenant.Spec.Features = &features

	_, err = operatorClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})

	return err
}
