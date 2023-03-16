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
	"fmt"
	"sort"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerEventHandlers(api *operations.OperatorAPI) {
	// Get Tenant Events
	api.OperatorAPIGetTenantEventsHandler = operator_api.GetTenantEventsHandlerFunc(func(params operator_api.GetTenantEventsParams, principal *models.Principal) middleware.Responder {
		payload, err := getTenantEventsResponse(principal, params)
		if err != nil {
			return operator_api.NewGetTenantEventsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantEventsOK().WithPayload(payload)
	})
}

func getTenantEventsResponse(session *models.Principal, params operator_api.GetTenantEventsParams) (models.EventListWrapper, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	client, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	tenant, err := client.MinioV2().Tenants(params.Namespace).Get(ctx, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	events, err := clientset.CoreV1().Events(params.Namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", tenant.UID)})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	retval := models.EventListWrapper{}
	for _, event := range events.Items {
		retval = append(retval, &models.EventListElement{
			Namespace: event.Namespace,
			LastSeen:  event.LastTimestamp.Unix(),
			Message:   event.Message,
			EventType: event.Type,
			Reason:    event.Reason,
		})
	}
	sort.SliceStable(retval, func(i int, j int) bool {
		return retval[i].LastSeen < retval[j].LastSeen
	})
	return retval, nil
}
