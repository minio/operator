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
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func registerNamespaceHandlers(api *operations.OperatorAPI) {
	// Add Namespace
	// api.OperatorAPICreateNamespaceHandler = operator_api.CreateNamespaceHandlerFunc(func(params operator_api.CreateNamespaceParams, session *models.Principal) middleware.Responder {
	api.OperatorAPICreateNamespaceHandler = operator_api.CreateNamespaceHandlerFunc(func(params operator_api.CreateNamespaceParams, session *models.Principal) middleware.Responder {
		err := getNamespaceCreatedResponse(session, params)
		if err != nil {
			return operator_api.NewCreateNamespaceDefault(int(err.Code)).WithPayload(err)
		}
		return nil
	})
}

func getNamespaceCreatedResponse(session *models.Principal, params operator_api.CreateNamespaceParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	namespace := *params.Body.Name

	errCreation := getNamespaceCreated(ctx, clientset.CoreV1(), namespace)

	if errCreation != nil {
		return ErrorWithContext(ctx, errCreation)
	}

	return nil
}

func getNamespaceCreated(ctx context.Context, clientset v1.CoreV1Interface, namespace string) error {
	if namespace == "" {
		errNS := errors.New("Namespace cannot be blank")

		return errNS
	}

	coreNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err := clientset.Namespaces().Create(ctx, &coreNamespace, metav1.CreateOptions{})

	return err
}
