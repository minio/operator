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
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/auth/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerCertificateHandlers(api *operations.OperatorAPI) {
	// Tenant Security details
	api.OperatorAPITenantSecurityHandler = operator_api.TenantSecurityHandlerFunc(func(params operator_api.TenantSecurityParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantSecurityResponse(session, params)
		if err != nil {
			return operator_api.NewTenantSecurityDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantSecurityOK().WithPayload(resp)
	})

	// Update Tenant Security configuration
	api.OperatorAPIUpdateTenantSecurityHandler = operator_api.UpdateTenantSecurityHandlerFunc(func(params operator_api.UpdateTenantSecurityParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantSecurityResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantSecurityDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantSecurityNoContent()
	})

	// Update Tenant Certificates
	api.OperatorAPITenantUpdateCertificateHandler = operator_api.TenantUpdateCertificateHandlerFunc(func(params operator_api.TenantUpdateCertificateParams, session *models.Principal) middleware.Responder {
		err := getTenantUpdateCertificatesResponse(session, params)
		if err != nil {
			return operator_api.NewTenantUpdateCertificateDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantUpdateCertificateCreated()
	})
}

func getTenantSecurityResponse(session *models.Principal, params operator_api.TenantSecurityParams) (*models.TenantSecurityResponse, *models.Error) {
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
	info, err := getTenantSecurity(ctx, &k8sClient, minTenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return info, nil
}

func getUpdateTenantSecurityResponse(session *models.Principal, params operator_api.UpdateTenantSecurityParams) *models.Error {
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
	if err := updateTenantSecurity(ctx, opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant"))
	}
	return nil
}

// updateTenantSecurity
func updateTenantSecurity(ctx context.Context, operatorClient OperatorClientI, client K8sClientI, namespace string, params operator_api.UpdateTenantSecurityParams) error {
	minInst, err := operatorClient.TenantGet(ctx, namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Update AutoCert
	minInst.Spec.RequestAutoCert = &params.Body.AutoCert
	var newExternalCertSecret []*miniov2.LocalCertificateReference
	var newExternalClientCertSecrets []*miniov2.LocalCertificateReference
	var newExternalCaCertSecret []*miniov2.LocalCertificateReference
	secretsToBeRemoved := map[string]bool{}

	if params.Body.CustomCertificates != nil {
		// Copy certificate secrets to be deleted into map
		for _, secret := range params.Body.CustomCertificates.SecretsToBeDeleted {
			secretsToBeRemoved[secret] = true
		}

		// Remove certificates from Tenant.Spec.ExternalCertSecret
		for _, certificate := range minInst.Spec.ExternalCertSecret {
			if _, ok := secretsToBeRemoved[certificate.Name]; !ok {
				newExternalCertSecret = append(newExternalCertSecret, certificate)
			}
		}
		// Remove certificates from Tenant.Spec.ExternalClientCertSecrets
		for _, certificate := range minInst.Spec.ExternalClientCertSecrets {
			if _, ok := secretsToBeRemoved[certificate.Name]; !ok {
				newExternalClientCertSecrets = append(newExternalClientCertSecrets, certificate)
			}
		}
		// Remove certificates from Tenant.Spec.ExternalCaCertSecret
		for _, certificate := range minInst.Spec.ExternalCaCertSecret {
			if _, ok := secretsToBeRemoved[certificate.Name]; !ok {
				newExternalCaCertSecret = append(newExternalCaCertSecret, certificate)
			}
		}

	}
	secretName := fmt.Sprintf("%s-%s", minInst.Name, strings.ToLower(utils.RandomCharString(5)))
	// Create new Server Certificate Secrets for MinIO
	externalServerCertSecretName := fmt.Sprintf("%s-external-server-certificate", secretName)
	externalServerCertSecrets, err := createOrReplaceExternalCertSecrets(ctx, client, minInst.Namespace, params.Body.CustomCertificates.MinioServerCertificates, externalServerCertSecretName, minInst.Name)
	if err != nil {
		return err
	}
	newExternalCertSecret = append(newExternalCertSecret, externalServerCertSecrets...)
	// Create new Client Certificate Secrets for MinIO
	externalClientCertSecretName := fmt.Sprintf("%s-external-client-certificate", secretName)
	externalClientCertSecrets, err := createOrReplaceExternalCertSecrets(ctx, client, minInst.Namespace, params.Body.CustomCertificates.MinioClientCertificates, externalClientCertSecretName, minInst.Name)
	if err != nil {
		return err
	}
	newExternalClientCertSecrets = append(newExternalClientCertSecrets, externalClientCertSecrets...)
	// Create new CAs Certificate Secrets for MinIO
	var caCertificates []tenantSecret
	for i, caCertificate := range params.Body.CustomCertificates.MinioCAsCertificates {
		certificateContent, err := base64.StdEncoding.DecodeString(caCertificate)
		if err != nil {
			return err
		}
		caCertificates = append(caCertificates, tenantSecret{
			Name: fmt.Sprintf("%s-ca-certificate-%d", secretName, i),
			Content: map[string][]byte{
				"public.crt": certificateContent,
			},
		})
	}
	if len(caCertificates) > 0 {
		certificateSecrets, err := createOrReplaceSecrets(ctx, client, minInst.Namespace, caCertificates, minInst.Name)
		if err != nil {
			return err
		}
		newExternalCaCertSecret = append(newExternalCaCertSecret, certificateSecrets...)
	}

	// set Security Context
	var newTenantSecurityContext *corev1.PodSecurityContext
	newTenantSecurityContext, err = convertModelSCToK8sSC(params.Body.SecurityContext)
	if err != nil {
		return err
	}
	for index := range minInst.Spec.Pools {
		minInst.Spec.Pools[index].SecurityContext = newTenantSecurityContext
	}

	// Update External Certificates
	minInst.Spec.ExternalCertSecret = newExternalCertSecret
	minInst.Spec.ExternalClientCertSecrets = newExternalClientCertSecrets
	minInst.Spec.ExternalCaCertSecret = newExternalCaCertSecret
	_, err = operatorClient.TenantUpdate(ctx, minInst, metav1.UpdateOptions{})
	return err
}
