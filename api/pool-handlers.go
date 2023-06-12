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
	"encoding/json"
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func registerPoolHandlers(api *operations.OperatorAPI) {
	// Add Tenant Pools
	api.OperatorAPITenantAddPoolHandler = operator_api.TenantAddPoolHandlerFunc(func(params operator_api.TenantAddPoolParams, session *models.Principal) middleware.Responder {
		// check the poolName if it exists
		resp, err := getTenantDetailsResponse(session, operator_api.TenantDetailsParams{Namespace: params.Namespace, Tenant: params.Tenant, HTTPRequest: params.HTTPRequest})
		if err != nil {
			return operator_api.NewTenantAddPoolDefault(int(err.Code)).WithPayload(err)
		}
		for _, p := range resp.Pools {
			if p.Name == params.Body.Name {
				err = ErrorWithContext(params.HTTPRequest.Context(), ErrPoolExists)
				return operator_api.NewTenantAddPoolDefault(int(err.Code)).WithPayload(err)
			}
		}
		err = getTenantAddPoolResponse(session, params)
		if err != nil {
			return operator_api.NewTenantAddPoolDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantAddPoolCreated()
	})
	// Update Tenant Pools
	api.OperatorAPITenantUpdatePoolsHandler = operator_api.TenantUpdatePoolsHandlerFunc(func(params operator_api.TenantUpdatePoolsParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantUpdatePoolResponse(session, params)
		if err != nil {
			return operator_api.NewTenantUpdatePoolsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantUpdatePoolsOK().WithPayload(resp)
	})
}

func getTenantAddPoolResponse(session *models.Principal, params operator_api.TenantAddPoolParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	if err := addTenantPool(ctx, opClient, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to add pool"))
	}
	return nil
}

func getTenantUpdatePoolResponse(session *models.Principal, params operator_api.TenantUpdatePoolsParams) (*models.Tenant, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	tenant, err := updateTenantPools(ctx, opClient, params.Namespace, params.Tenant, params.Body.Pools)
	if err != nil {
		LogError("error updating Tenant's pools: %v", err)
		return nil, ErrorWithContext(ctx, err)
	}
	return tenant, nil
}

// updateTenantPools Sets the Tenant's pools to the ones provided by the request
//
// It does the equivalent to a PUT request on Tenant's pools
func updateTenantPools(
	ctx context.Context,
	operatorClient OperatorClientI,
	namespace string,
	tenantName string,
	poolsReq []*models.Pool,
) (*models.Tenant, error) {
	minInst, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// set the pools if they are provided
	var newPoolArray []miniov2.Pool
	for _, pool := range poolsReq {
		pool, err := parseTenantPoolRequest(pool)
		if err != nil {
			return nil, err
		}
		newPoolArray = append(newPoolArray, *pool)
	}

	// replace pools array
	minInst.Spec.Pools = newPoolArray

	minInst = minInst.DeepCopy()
	minInst.EnsureDefaults()

	payloadBytes, err := json.Marshal(minInst)
	if err != nil {
		return nil, err
	}
	tenantUpdated, err := operatorClient.TenantPatch(ctx, namespace, minInst.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return getTenantInfo(tenantUpdated), nil
}
