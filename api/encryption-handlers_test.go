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
	"net/http"
	"time"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/kes"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *TenantTestSuite) TestTenantUpdateEncryptionHandlerWithError() {
	params, api := suite.initTenantUpdateEncryptionRequest()
	response := api.OperatorAPITenantUpdateEncryptionHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.TenantUpdateEncryptionDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionWithExternalCertError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	params.Body = &models.EncryptionConfiguration{
		ServerTLS: &models.KeyPairConfiguration{},
	}
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					ExternalCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-crt",
					},
				},
			},
		}, nil
	}
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionWithExternalClientCertError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	params.Body = &models.EncryptionConfiguration{
		MinioMtls: &models.KeyPairConfiguration{},
	}
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				ExternalClientCertSecret: &miniov2.LocalCertificateReference{
					Name: "mock-crt",
				},
			},
		}, nil
	}
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionAWSWithoutError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	endpoint := "mock-endpoint"
	region := "mock-region"
	ak := "mock-accesskey"
	sk := "mock-secretkey"
	params.Body = &models.EncryptionConfiguration{
		Replicas:           "1",
		SecurityContext:    suite.createMockModelsSecurityContext(),
		SecretsToBeDeleted: []string{"mock-crt"},
		Aws: &models.AwsConfiguration{
			Secretsmanager: &models.AwsConfigurationSecretsmanager{
				Endpoint: &endpoint,
				Region:   &region,
				Kmskey:   "mock-kmskey",
				Credentials: &models.AwsConfigurationSecretsmanagerCredentials{
					Accesskey: &ak,
					Secretkey: &sk,
				},
			},
		},
	}
	k8sClientCreateSecretMock = func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
		return nil, nil
	}
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				ExternalClientCertSecret: &miniov2.LocalCertificateReference{
					Name: "mock-crt",
				},
				KES: &miniov2.KESConfig{
					ExternalCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-crt",
					},
				},
			},
		}, nil
	}
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return nil, nil
	}
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(err)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionGemaltoWithoutError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	endpoint := "mock-endpoint"
	token := "mock-token"
	domain := "mock-domain"
	params.Body = &models.EncryptionConfiguration{
		Replicas:        "1",
		SecurityContext: suite.createMockModelsSecurityContext(),
		Gemalto: &models.GemaltoConfiguration{
			Keysecure: &models.GemaltoConfigurationKeysecure{
				Endpoint: &endpoint,
				Credentials: &models.GemaltoConfigurationKeysecureCredentials{
					Token:  &token,
					Domain: &domain,
				},
			},
		},
		KmsMtls: &models.EncryptionConfigurationAO1KmsMtls{
			Ca: "bW9jaw==",
		},
	}
	suite.prepareEncryptionUpdateMocksNoError()
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(err)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionGCPWithoutError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	project := "mock-project"
	params.Body = &models.EncryptionConfiguration{
		Replicas:        "1",
		SecurityContext: suite.createMockModelsSecurityContext(),
		Gcp: &models.GcpConfiguration{
			Secretmanager: &models.GcpConfigurationSecretmanager{
				ProjectID: &project,
				Endpoint:  "mock-endpoint",
				Credentials: &models.GcpConfigurationSecretmanagerCredentials{
					ClientEmail:  "mock",
					ClientID:     "mock",
					PrivateKey:   "mock",
					PrivateKeyID: "mock",
				},
			},
		},
	}
	suite.prepareEncryptionUpdateMocksNoError()
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(err)
}

func (suite *TenantTestSuite) TestTenantUpdateEncryptionAzureWithoutError() {
	params, _ := suite.initTenantUpdateEncryptionRequest()
	endpoint := "mock-endpoint"
	tenant := "mock-tenant"
	clientID := "mock-client-id"
	clientSecret := "mock-client-secret"
	params.Body = &models.EncryptionConfiguration{
		Replicas:        "1",
		SecurityContext: suite.createMockModelsSecurityContext(),
		Azure: &models.AzureConfiguration{
			Keyvault: &models.AzureConfigurationKeyvault{
				Endpoint: &endpoint,
				Credentials: &models.AzureConfigurationKeyvaultCredentials{
					TenantID:     &tenant,
					ClientID:     &clientID,
					ClientSecret: &clientSecret,
				},
			},
		},
	}
	suite.prepareEncryptionUpdateMocksNoError()
	err := tenantUpdateEncryption(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(err)
}

func (suite *TenantTestSuite) prepareEncryptionUpdateMocksNoError() {
	k8sClientCreateSecretMock = func(ctx context.Context, namespace string, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
		return nil, nil
	}
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{Spec: miniov2.TenantSpec{}}, nil
	}
	opClientTenantUpdateMock = func(ctx context.Context, tenant *miniov2.Tenant, opts metav1.UpdateOptions) (*miniov2.Tenant, error) {
		return nil, nil
	}
}

func (suite *TenantTestSuite) initTenantUpdateEncryptionRequest() (params operator_api.TenantUpdateEncryptionParams, api operations.OperatorAPI) {
	registerEncryptionHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}

func (suite *TenantTestSuite) TestTenantDeleteEncryptionHandlerWithError() {
	params, api := suite.initTenantDeleteEncryptionRequest()
	response := api.OperatorAPITenantDeleteEncryptionHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.TenantDeleteEncryptionDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initTenantDeleteEncryptionRequest() (params operator_api.TenantDeleteEncryptionParams, api operations.OperatorAPI) {
	registerEncryptionHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoHandlerWithError() {
	params, api := suite.initTenantEncryptionInfoRequest()
	response := api.OperatorAPITenantEncryptionInfoHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.TenantEncryptionInfoDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWitNoKesError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{Spec: miniov2.TenantSpec{}}, nil
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithExtCertError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					ExternalCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-crt",
					},
					SecurityContext: suite.createTenantPodSecurityContext(),
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithClientCertError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{},
				ExternalClientCertSecret: &miniov2.LocalCertificateReference{
					Name: "mock-crt",
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithKesClientCertError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					ClientCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-kes-crt",
					},
					Configuration: &corev1.LocalObjectReference{
						Name: "mock-kes-config",
					},
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		if secretName == "mock-kes-config" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"server-config.yaml": suite.getKesYamlMock(false),
				},
			}, nil
		}
		if secretName == "mock-kes-crt" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"client.crt": []byte("mock-client-crt"),
				},
			}, nil
		}
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithKesClientCACertError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					ClientCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-kes-crt",
					},
					Configuration: &corev1.LocalObjectReference{
						Name: "mock-kes-config",
					},
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		if secretName == "mock-kes-config" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"server-config.yaml": suite.getKesYamlMock(false),
				},
			}, nil
		}
		if secretName == "mock-kes-crt" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"ca.crt": []byte("mock-client-crt"),
				},
			}, nil
		}
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithGemaltoError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					ClientCertSecret: &miniov2.LocalCertificateReference{
						Name: "mock-kes-crt",
					},
					Configuration: &corev1.LocalObjectReference{
						Name: "mock-kes-config",
					},
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		if secretName == "mock-kes-config" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"server-config.yaml": suite.getKesYamlMock(true),
				},
			}, nil
		}
		if secretName == "mock-kes-crt" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"ca.crt": []byte("mock-client-crt"),
				},
			}, nil
		}
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.Nil(res)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestTenantEncryptionInfoWithoutError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				KES: &miniov2.KESConfig{
					Configuration: &corev1.LocalObjectReference{
						Name: "mock-kes-config",
					},
				},
			},
		}, nil
	}
	k8sclientGetSecretMock = func(ctx context.Context, namespace, secretName string, opts metav1.GetOptions) (*corev1.Secret, error) {
		if secretName == "mock-kes-config" {
			return &corev1.Secret{
				Data: map[string][]byte{
					"server-config.yaml": suite.getKesYamlMock(false),
				},
			}, nil
		}
		return nil, errors.New("mock-get-error")
	}
	params, _ := suite.initTenantEncryptionInfoRequest()
	res, err := tenantEncryptionInfo(context.Background(), suite.opClient, suite.k8sclient, params.Namespace, params)
	suite.assert.NotNil(res)
	suite.assert.Nil(err)
}

func (suite *TenantTestSuite) getKesYamlMock(noVault bool) []byte {
	kesConfig := &kes.ServerConfig{
		Keys: kes.Keys{
			Vault: &kes.Vault{
				Prefix:     "mock-prefix",
				Namespace:  "mock-namespace",
				EnginePath: "mock-engine-path",
				Endpoint:   "mock-endpoint",
				Status: &kes.VaultStatus{
					Ping: 5 * time.Second,
				},
				AppRole: &kes.AppRole{
					EnginePath: "mock-engine-path",
					ID:         "mock-id",
					Retry:      5 * time.Second,
					Secret:     "mock-secret",
				},
			},
			Aws: &kes.Aws{},
			Gcp: &kes.Gcp{},
			Gemalto: &kes.Gemalto{
				KeySecure: &kes.GemaltoKeySecure{
					Endpoint: "mock-endpoint",
					Credentials: &kes.GemaltoCredentials{
						Domain: "mock-domain",
						Retry:  5 * time.Second,
						Token:  "mock-token",
					},
					TLS: &kes.GemaltoTLS{},
				},
			},
			Azure: &kes.Azure{},
		},
	}
	if noVault {
		kesConfig.Keys.Vault = nil
	}
	kesConfigBytes, _ := yaml.Marshal(kesConfig)
	return kesConfigBytes
}

func (suite *TenantTestSuite) initTenantEncryptionInfoRequest() (params operator_api.TenantEncryptionInfoParams, api operations.OperatorAPI) {
	registerEncryptionHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}
