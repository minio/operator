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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func registerYAMLHandlers(api *operations.OperatorAPI) {
	// Get Tenant YAML
	api.OperatorAPIGetTenantYAMLHandler = operator_api.GetTenantYAMLHandlerFunc(func(params operator_api.GetTenantYAMLParams, principal *models.Principal) middleware.Responder {
		payload, err := getTenantYAML(principal, params)
		if err != nil {
			return operator_api.NewGetTenantYAMLDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantYAMLOK().WithPayload(payload)
	})
	// Update Tenant YAML
	api.OperatorAPIPutTenantYAMLHandler = operator_api.PutTenantYAMLHandlerFunc(func(params operator_api.PutTenantYAMLParams, principal *models.Principal) middleware.Responder {
		err := getUpdateTenantYAML(principal, params)
		if err != nil {
			return operator_api.NewPutTenantYAMLDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewPutTenantYAMLCreated()
	})
}

func getTenantYAML(session *models.Principal, params operator_api.GetTenantYAMLParams) (*models.TenantYAML, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	opClient, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	tenant, err := opClient.MinioV2().Tenants(params.Namespace).Get(params.HTTPRequest.Context(), params.Tenant, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	// remove managed fields
	tenant.ManagedFields = []metav1.ManagedFieldsEntry{}
	// yb, err := yaml.Marshal(tenant)
	j8sJSONSerializer := k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	buf := new(bytes.Buffer)

	err = j8sJSONSerializer.Encode(tenant, buf)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	yb := buf.String()

	x, _ := json.Marshal(tenant)
	fmt.Println(string(x))

	return &models.TenantYAML{Yaml: yb}, nil
}

func getUpdateTenantYAML(session *models.Principal, params operator_api.PutTenantYAMLParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// https://godoc.org/k8s.io/apimachinery/pkg/runtime#Scheme
	scheme := runtime.NewScheme()

	// https://godoc.org/k8s.io/apimachinery/pkg/runtime/serializer#CodecFactory
	codecFactory := serializer.NewCodecFactory(scheme)

	// https://godoc.org/k8s.io/apimachinery/pkg/runtime#Decoder
	deserializer := codecFactory.UniversalDeserializer()

	tenantObject, _, err := deserializer.Decode([]byte(params.Body.Yaml), nil, &miniov2.Tenant{})
	if err != nil {
		return &models.Error{Code: 400, Message: swag.String(err.Error())}
	}
	inTenant := tenantObject.(*miniov2.Tenant)
	// check the inTenant if have the same poolName
	poolNameMapSet := make(map[string]struct{})
	for _, pool := range inTenant.Spec.Pools {
		if _, ok := poolNameMapSet[pool.Name]; ok {
			return &models.Error{Code: 400, Message: swag.String(fmt.Sprintf("Tenant %s have pool named '%s' already", inTenant.Name, pool.Name))}
		}
		poolNameMapSet[pool.Name] = struct{}{}
	}
	if len(poolNameMapSet) == 0 {
		return &models.Error{Code: 400, Message: swag.String(fmt.Sprintf("Tenant %s must have one pool at least!", inTenant.Name))}
	}
	// get Kubernetes Client
	opClient, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	tenant, err := opClient.MinioV2().Tenants(params.Namespace).Get(params.HTTPRequest.Context(), params.Tenant, metav1.GetOptions{})
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	upTenant := tenant.DeepCopy()
	// only update safe fields: spec, metadata.finalizers, metadata.labels and metadata.annotations
	upTenant.Labels = inTenant.Labels
	upTenant.Annotations = inTenant.Annotations
	upTenant.Finalizers = inTenant.Finalizers
	upTenant.Spec = inTenant.Spec

	_, err = opClient.MinioV2().Tenants(upTenant.Namespace).Update(params.HTTPRequest.Context(), upTenant, metav1.UpdateOptions{})
	if err != nil {
		return &models.Error{Code: 400, Message: swag.String(err.Error())}
	}

	return nil
}
