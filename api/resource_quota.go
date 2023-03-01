// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerResourceQuotaHandlers(api *operations.OperatorAPI) {
	// Get Resource Quota
	api.OperatorAPIGetResourceQuotaHandler = operator_api.GetResourceQuotaHandlerFunc(func(params operator_api.GetResourceQuotaParams, session *models.Principal) middleware.Responder {
		resp, err := getResourceQuotaResponse(session, params)
		if err != nil {
			return operator_api.NewGetResourceQuotaDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetResourceQuotaOK().WithPayload(resp)
	})
}

func getResourceQuota(ctx context.Context, client K8sClientI, namespace, resourcequota string) (*models.ResourceQuota, error) {
	resourceQuota, err := client.getResourceQuota(ctx, namespace, resourcequota, metav1.GetOptions{})
	if err != nil {
		// if there's no resource quotas
		if errors.IsNotFound(err) {
			// validate if at least the namespace is valid, if it is, return all storage classes with max capacity
			_, err := client.getNamespace(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			storageClasses, err := client.getStorageClasses(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, err
			}
			rq := models.ResourceQuota{Name: resourceQuota.Name}
			for _, sc := range storageClasses.Items {
				// Create Resource element with hard limit maxed out
				name := fmt.Sprintf("%s.storageclass.storage.k8s.io/requests.storage", sc.Name)
				element := models.ResourceQuotaElement{
					Name: name,
					Hard: 9223372036854775807,
				}
				rq.Elements = append(rq.Elements, &element)
			}
			return &rq, nil
		}

		return nil, err
	}
	rq := models.ResourceQuota{Name: resourceQuota.Name}
	resourceElementss := make(map[string]models.ResourceQuotaElement)
	for name, quantity := range resourceQuota.Status.Hard {
		// Create Resource element with hard limit
		element := models.ResourceQuotaElement{
			Name: string(name),
			Hard: quantity.Value(),
		}
		resourceElementss[string(name)] = element
	}
	for name, quantity := range resourceQuota.Status.Used {
		// Update resource element with Used quota
		if r, ok := resourceElementss[string(name)]; ok {
			r.Used = quantity.Value()
			// Element will only be returned if it has Hard and Used status
			rq.Elements = append(rq.Elements, &r)
		}
	}
	return &rq, nil
}

func getResourceQuotaResponse(session *models.Principal, params operator_api.GetResourceQuotaParams) (*models.ResourceQuota, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	client, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	k8sClient := &k8sClient{
		client: client,
	}
	resourceQuota, err := getResourceQuota(ctx, k8sClient, params.Namespace, params.ResourceQuotaName)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return resourceQuota, nil
}
