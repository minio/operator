// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
	"fmt"
	"os"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/http"
	"github.com/minio/operator/pkg/subnet"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	apiKeySecretDefault = "operator-subnet"
	apiKeySecretEnvVar  = "API_KEY_SECRET_NAME"
)

type tenantInterface struct {
	tenant       miniov2.Tenant
	mAdminClient MinioAdmin
}

func registerOperatorSubnetHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIOperatorSubnetLoginHandler = operator_api.OperatorSubnetLoginHandlerFunc(func(params operator_api.OperatorSubnetLoginParams, session *models.Principal) middleware.Responder {
		res, err := getOperatorSubnetLoginResponse(session, params)
		if err != nil {
			return operator_api.NewOperatorSubnetLoginDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewOperatorSubnetLoginOK().WithPayload(res)
	})

	api.OperatorAPIOperatorSubnetLoginMFAHandler = operator_api.OperatorSubnetLoginMFAHandlerFunc(func(params operator_api.OperatorSubnetLoginMFAParams, session *models.Principal) middleware.Responder {
		res, err := getOperatorSubnetLoginMFAResponse(session, params)
		if err != nil {
			return operator_api.NewOperatorSubnetLoginMFADefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewOperatorSubnetLoginMFAOK().WithPayload(res)
	})

	api.OperatorAPIOperatorSubnetAPIKeyHandler = operator_api.OperatorSubnetAPIKeyHandlerFunc(func(params operator_api.OperatorSubnetAPIKeyParams, session *models.Principal) middleware.Responder {
		res, err := getOperatorSubnetAPIKeyResponse(session, params)
		if err != nil {
			return operator_api.NewOperatorSubnetAPIKeyDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewOperatorSubnetAPIKeyOK().WithPayload(res)
	})

	api.OperatorAPIOperatorSubnetRegisterAPIKeyHandler = operator_api.OperatorSubnetRegisterAPIKeyHandlerFunc(func(params operator_api.OperatorSubnetRegisterAPIKeyParams, session *models.Principal) middleware.Responder {
		res, err := getOperatorSubnetRegisterAPIKeyResponse(session, params)
		if err != nil {
			return operator_api.NewOperatorSubnetRegisterAPIKeyDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewOperatorSubnetRegisterAPIKeyOK().WithPayload(res)
	})
	api.OperatorAPIOperatorSubnetAPIKeyInfoHandler = operator_api.OperatorSubnetAPIKeyInfoHandlerFunc(func(params operator_api.OperatorSubnetAPIKeyInfoParams, session *models.Principal) middleware.Responder {
		res, err := getOperatorSubnetAPIKeyInfoResponse(session, params)
		if err != nil {
			return operator_api.NewOperatorSubnetAPIKeyInfoDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewOperatorSubnetAPIKeyInfoOK().WithPayload(res)
	})
}

func getOperatorSubnetLoginResponse(session *models.Principal, params operator_api.OperatorSubnetLoginParams) (*models.OperatorSubnetLoginResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	username := params.Body.Username
	password := params.Body.Password
	if username == "" || password == "" {
		return nil, ErrorWithContext(ctx, errors.New("empty credentials"))
	}
	subnetHTTPClient := &xhttp.Client{Client: GetConsoleHTTPClient("")}
	token, mfa, err := SubnetLogin(subnetHTTPClient, username, password)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return &models.OperatorSubnetLoginResponse{
		AccessToken: token,
		MfaToken:    mfa,
	}, nil
}

func getOperatorSubnetLoginMFAResponse(session *models.Principal, params operator_api.OperatorSubnetLoginMFAParams) (*models.OperatorSubnetLoginResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	subnetHTTPClient := &xhttp.Client{Client: GetConsoleHTTPClient("")}
	res, err := subnet.LoginWithMFA(subnetHTTPClient, *params.Body.Username, *params.Body.MfaToken, *params.Body.Otp)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return &models.OperatorSubnetLoginResponse{
		AccessToken: res.AccessToken,
	}, nil
}

func getOperatorSubnetAPIKeyResponse(session *models.Principal, params operator_api.OperatorSubnetAPIKeyParams) (*models.OperatorSubnetAPIKey, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	subnetHTTPClient := &xhttp.Client{Client: GetConsoleHTTPClient("")}
	token := params.HTTPRequest.URL.Query().Get("token")
	apiKey, err := subnet.GetAPIKey(subnetHTTPClient, token)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return &models.OperatorSubnetAPIKey{APIKey: apiKey}, nil
}

func getOperatorSubnetRegisterAPIKeyResponse(session *models.Principal, params operator_api.OperatorSubnetRegisterAPIKeyParams) (*models.OperatorSubnetRegisterAPIKeyResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	k8sClient := &k8sClient{client: clientSet}
	tenants, err := getTenantsToRegister(ctx, session, k8sClient)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return registerTenants(ctx, k8sClient, tenants, params.Body.APIKey)
}

func getTenantsToRegister(ctx context.Context, session *models.Principal, k8sClient K8sClientI) ([]tenantInterface, error) {
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, err
	}
	opClient := &operatorClient{client: opClientClientSet}
	tenantList, err := opClient.TenantList(ctx, "", metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	tenantStructs := make([]tenantInterface, len(tenantList.Items))
	for _, tenant := range tenantList.Items {
		svcURL := tenant.GetTenantServiceURL()
		mAdmin, err := getTenantAdminClient(ctx, k8sClient, &tenant, svcURL)
		if err != nil {
			return nil, err
		}
		tenantStructs = append(tenantStructs, tenantInterface{tenant: tenant, mAdminClient: AdminClient{Client: mAdmin}})
	}
	return tenantStructs, nil
}

func registerTenants(ctx context.Context, k8sClient K8sClientI, tenants []tenantInterface, apiKey string) (*models.OperatorSubnetRegisterAPIKeyResponse, *models.Error) {
	for _, tenant := range tenants {
		if err := registerTenant(ctx, k8sClient, tenant.mAdminClient, tenant.tenant, apiKey); err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
	}
	if err := createSubnetAPIKeySecret(ctx, apiKey, k8sClient); err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return &models.OperatorSubnetRegisterAPIKeyResponse{Registered: true}, nil
}

// SubnetRegisterWithAPIKey start registration with api key
func SubnetRegisterWithAPIKey(ctx context.Context, minioClient MinioAdmin, apiKey string) (bool, error) {
	serverInfo, err := minioClient.serverInfo(ctx)
	if err != nil {
		return false, err
	}
	registerResult, err := subnet.Register(GetConsoleHTTPClient(""), serverInfo, apiKey, "", "")
	if err != nil {
		return false, err
	}
	// Keep existing subnet proxy if exists
	subnetKey, err := GetSubnetKeyFromMinIOConfig(ctx, minioClient)
	if err != nil {
		return false, err
	}
	configStr := fmt.Sprintf("subnet license=%s api_key=%s proxy=%s", registerResult.License, registerResult.APIKey, subnetKey.Proxy)
	_, err = minioClient.setConfigKV(ctx, configStr)
	if err != nil {
		return false, err
	}
	// cluster registered correctly
	return true, nil
}

func registerTenant(ctx context.Context, k8sClient K8sClientI, adminClient MinioAdmin, tenant v2.Tenant, apiKey string) error {
	_, err := SubnetRegisterWithAPIKey(ctx, adminClient, apiKey)
	return err
}

func createSubnetAPIKeySecret(ctx context.Context, apiKey string, k8sClient K8sClientI) error {
	apiKeySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: getAPIKeySecretName()},
		Type:       corev1.SecretTypeOpaque,
		Data:       map[string][]byte{"api-key": []byte(apiKey)},
	}
	_, err := k8sClient.createSecret(ctx, "default", apiKeySecret, metav1.CreateOptions{})
	return err
}

func getOperatorSubnetAPIKeyInfoResponse(session *models.Principal, params operator_api.OperatorSubnetAPIKeyInfoParams) (*models.OperatorSubnetRegisterAPIKeyResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	k8sClient := &k8sClient{client: clientSet}
	if _, err := k8sClient.getSecret(ctx, "default", getAPIKeySecretName(), metav1.GetOptions{}); err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return &models.OperatorSubnetRegisterAPIKeyResponse{Registered: true}, nil
}

func getAPIKeySecretName() string {
	if s := os.Getenv(apiKeySecretEnvVar); s != "" {
		return s
	}
	return apiKeySecretDefault
}

// SubnetLogin start login
func SubnetLogin(client xhttp.ClientI, username, password string) (string, string, error) {
	tokens, err := subnet.Login(client, username, password)
	if err != nil {
		return "", "", err
	}
	if tokens.MfaToken != "" {
		// user needs to complete login flow using mfa
		return "", tokens.MfaToken, nil
	}
	if tokens.AccessToken != "" {
		// register token to minio
		return tokens.AccessToken, "", nil
	}
	return "", "", errors.New("something went wrong")
}
