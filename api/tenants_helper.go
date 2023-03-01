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
	"crypto"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/kes"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// convertModelSCToK8sSC validates and converts from models.SecurityContext to corev1.PodSecurityContext
func convertModelSCToK8sSC(sc *models.SecurityContext) (*corev1.PodSecurityContext, error) {
	if sc == nil {
		return nil, errors.New("invalid security context")
	}
	runAsUser, err := strconv.ParseInt(*sc.RunAsUser, 10, 64)
	if err != nil {
		return nil, err
	}
	runAsGroup, err := strconv.ParseInt(*sc.RunAsGroup, 10, 64)
	if err != nil {
		return nil, err
	}
	fsGroup, err := strconv.ParseInt(sc.FsGroup, 10, 64)
	if err != nil {
		return nil, err
	}
	FSGroupChangePolicy := corev1.PodFSGroupChangePolicy("Always")
	if sc.FsGroupChangePolicy != "" {
		FSGroupChangePolicy = corev1.PodFSGroupChangePolicy(sc.FsGroupChangePolicy)
	}
	return &corev1.PodSecurityContext{
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		RunAsNonRoot:        sc.RunAsNonRoot,
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &FSGroupChangePolicy,
	}, nil
}

// convertK8sSCToModelSC validates and converts from corev1.PodSecurityContext to models.SecurityContext
func convertK8sSCToModelSC(sc *corev1.PodSecurityContext) *models.SecurityContext {
	runAsUser := strconv.FormatInt(*sc.RunAsUser, 10)
	runAsGroup := strconv.FormatInt(*sc.RunAsGroup, 10)
	fsGroup := strconv.FormatInt(*sc.FSGroup, 10)
	fsGroupChangePolicy := "Always"

	if sc.FSGroupChangePolicy != nil {
		fsGroupChangePolicy = string(*sc.FSGroupChangePolicy)
	}

	return &models.SecurityContext{
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		RunAsNonRoot:        sc.RunAsNonRoot,
		FsGroup:             fsGroup,
		FsGroupChangePolicy: fsGroupChangePolicy,
	}
}

// tenantUpdateCertificates receives the keyPair certificates (public and private keys) for Minio and Console and will try
// to replace the existing kubernetes secrets with the new values, then will restart the affected pods so the new volumes can be mounted
func tenantUpdateCertificates(ctx context.Context, operatorClient OperatorClientI, clientSet K8sClientI, namespace string, params operator_api.TenantUpdateCertificateParams) error {
	tenantName := params.Tenant
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	body := params.Body
	// check if MinIO is deployed with external certs and user provided new MinIO keypair
	if tenant.ExternalCert() && body.MinioServerCertificates != nil {
		minioCertSecretName := fmt.Sprintf("%s-instance-external-certificates", tenantName)
		// update certificates
		if _, err := createOrReplaceExternalCertSecrets(ctx, clientSet, namespace, body.MinioServerCertificates, minioCertSecretName, tenantName); err != nil {
			return err
		}
	}
	return nil
}

// getTenantUpdateCertificatesResponse wrapper of tenantUpdateCertificates
func getTenantUpdateCertificatesResponse(session *models.Principal, params operator_api.TenantUpdateCertificateParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	k8sClient := k8sClient{
		client: clientSet,
	}
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	if err := tenantUpdateCertificates(ctx, &opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	return nil
}

// tenantDeleteEncryption allow user to disable tenant encryption for a particular tenant
func tenantDeleteEncryption(ctx context.Context, operatorClient OperatorClientI, namespace string, params operator_api.TenantDeleteEncryptionParams) error {
	tenantName := params.Tenant
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	tenant.EnsureDefaults()
	tenant.Spec.KES = nil
	_, err = operatorClient.TenantUpdate(ctx, tenant, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// tenantUpdateEncryption allow user to update KES server certificates, KES client certificates (used by MinIO for mTLS) and KES configuration (KMS configuration, credentials, etc)
func tenantUpdateEncryption(ctx context.Context, operatorClient OperatorClientI, clientSet K8sClientI, namespace string, params operator_api.TenantUpdateEncryptionParams) error {
	tenantName := params.Tenant
	secretName := fmt.Sprintf("%s-secret", tenantName)
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	body := params.Body
	if err != nil {
		return err
	}
	tenant.EnsureDefaults()
	// Initialize KES configuration if not present
	if !tenant.HasKESEnabled() {
		tenant.Spec.KES = &miniov2.KESConfig{}
	}
	if tenant.KESExternalCert() {
		for _, certificateToBeDeleted := range params.Body.SecretsToBeDeleted {
			if tenant.Spec.KES.ExternalCertSecret.Name == certificateToBeDeleted {
				tenant.Spec.KES.ExternalCertSecret = nil
				break
			}
		}
	}
	if tenant.ExternalClientCert() {
		for _, certificateToBeDeleted := range params.Body.SecretsToBeDeleted {
			if tenant.Spec.ExternalClientCertSecret.Name == certificateToBeDeleted {
				tenant.Spec.ExternalClientCertSecret = nil
				break
			}
		}
	}
	if body.ServerTLS != nil {
		kesExternalCertSecretName := fmt.Sprintf("%s-kes-external-cert", secretName)
		if tenant.KESExternalCert() {
			kesExternalCertSecretName = tenant.Spec.KES.ExternalCertSecret.Name
		}
		// update certificates
		certificates := []*models.KeyPairConfiguration{body.ServerTLS}
		createdCertificates, err := createOrReplaceExternalCertSecrets(ctx, clientSet, namespace, certificates, kesExternalCertSecretName, tenantName)
		if err != nil {
			return err
		}
		if len(createdCertificates) > 0 {
			tenant.Spec.KES.ExternalCertSecret = createdCertificates[0]
		}
	}
	if body.MinioMtls != nil {
		tenantExternalClientCertSecretName := fmt.Sprintf("%s-tenant-external-client-cert", secretName)
		if tenant.ExternalClientCert() {
			tenantExternalClientCertSecretName = tenant.Spec.ExternalClientCertSecret.Name
		}
		// Update certificates
		certificates := []*models.KeyPairConfiguration{body.MinioMtls}
		createdCertificates, err := createOrReplaceExternalCertSecrets(ctx, clientSet, namespace, certificates, tenantExternalClientCertSecretName, tenantName)
		if err != nil {
			return err
		}
		if len(createdCertificates) > 0 {
			tenant.Spec.ExternalClientCertSecret = createdCertificates[0]
		}
	}
	// update KES identities in kes-configuration.yaml secret
	kesConfigurationSecretName := fmt.Sprintf("%s-kes-configuration", secretName)
	kesClientCertSecretName := fmt.Sprintf("%s-kes-client-cert", secretName)
	kesConfigurationSecret, kesClientCertSecret, err := createOrReplaceKesConfigurationSecrets(ctx, clientSet, namespace, body, kesConfigurationSecretName, kesClientCertSecretName, tenantName)
	if err != nil {
		return err
	}
	tenant.Spec.KES.Configuration = kesConfigurationSecret
	tenant.Spec.KES.ClientCertSecret = kesClientCertSecret
	image := params.Body.Image
	if image == "" {
		image = miniov2.DefaultKESImage
	}
	tenant.Spec.KES.Image = image
	i, err := strconv.ParseInt(params.Body.Replicas, 10, 32)
	if err != nil {
		return err
	}
	tenant.Spec.KES.Replicas = int32(i)
	tenant.Spec.KES.SecurityContext, err = convertModelSCToK8sSC(params.Body.SecurityContext)
	if err != nil {
		return err
	}
	_, err = operatorClient.TenantUpdate(ctx, tenant, metav1.UpdateOptions{})
	return err
}

// getTenantDeleteEncryptionResponse is a wrapper for tenantDeleteEncryption
func getTenantDeleteEncryptionResponse(session *models.Principal, params operator_api.TenantDeleteEncryptionParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrDeletingEncryptionConfig)
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	if err := tenantDeleteEncryption(ctx, &opClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, ErrDeletingEncryptionConfig)
	}
	return nil
}

// getTenantUpdateEncryptionResponse is a wrapper for tenantUpdateEncryption
func getTenantUpdateEncryptionResponse(session *models.Principal, params operator_api.TenantUpdateEncryptionParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUpdatingEncryptionConfig)
	}
	k8sClient := k8sClient{
		client: clientSet,
	}
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUpdatingEncryptionConfig)
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	if err := tenantUpdateEncryption(ctx, &opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, ErrUpdatingEncryptionConfig)
	}
	return nil
}

// tenantEncryptionInfo retrieves encryption information for the current tenant
func tenantEncryptionInfo(ctx context.Context, operatorClient OperatorClientI, clientSet K8sClientI, namespace string, params operator_api.TenantEncryptionInfoParams) (*models.EncryptionConfigurationResponse, error) {
	tenantName := params.Tenant
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// Check if encryption is enabled for MinIO via KES
	if tenant.HasKESEnabled() {
		encryptConfig := &models.EncryptionConfigurationResponse{
			Image:    tenant.Spec.KES.Image,
			Replicas: fmt.Sprintf("%d", tenant.Spec.KES.Replicas),
		}
		if tenant.Spec.KES.Image == "" {
			encryptConfig.Image = miniov2.GetTenantKesImage()
		}
		if tenant.Spec.KES.SecurityContext != nil {
			encryptConfig.SecurityContext = convertK8sSCToModelSC(tenant.Spec.KES.SecurityContext)
		}
		if tenant.KESExternalCert() {
			kesExternalCerts, err := parseTenantCertificates(ctx, clientSet, tenant.Namespace, []*miniov2.LocalCertificateReference{tenant.Spec.KES.ExternalCertSecret})
			if err != nil {
				return nil, err
			}
			if len(kesExternalCerts) > 0 {
				encryptConfig.ServerTLS = kesExternalCerts[0]
			}
		}
		if tenant.ExternalClientCert() {
			clientCerts, err := parseTenantCertificates(ctx, clientSet, tenant.Namespace, []*miniov2.LocalCertificateReference{tenant.Spec.ExternalClientCertSecret})
			if err != nil {
				return nil, err
			}
			if len(clientCerts) > 0 {
				encryptConfig.MinioMtls = clientCerts[0]
			}
		}

		if tenant.Spec.KES.Configuration != nil {
			configSecret, err := clientSet.getSecret(ctx, tenant.Namespace, tenant.Spec.KES.Configuration.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if rawConfiguration, ok := configSecret.Data["server-config.yaml"]; ok {
				kesConfiguration := &kes.ServerConfig{}
				// return raw configuration in case the user wants to edit KES configuration manually
				encryptConfig.Raw = string(rawConfiguration)
				err := yaml.Unmarshal(rawConfiguration, kesConfiguration)
				if err != nil {
					return nil, err
				}
				if kesConfiguration.Keys.Vault != nil {
					vault := kesConfiguration.Keys.Vault
					vaultConfig := &models.VaultConfigurationResponse{
						Prefix:    vault.Prefix,
						Namespace: vault.Namespace,
						Engine:    vault.EnginePath,
						Endpoint:  &vault.Endpoint,
					}
					if vault.Status != nil {
						vaultConfig.Status = &models.VaultConfigurationResponseStatus{
							Ping: int64(vault.Status.Ping.Seconds()),
						}
					}
					if vault.AppRole != nil {
						vaultConfig.Approle = &models.VaultConfigurationResponseApprole{
							Engine: vault.AppRole.EnginePath,
							ID:     &vault.AppRole.ID,
							Retry:  int64(vault.AppRole.Retry.Seconds()),
							Secret: &vault.AppRole.Secret,
						}
					}
					if tenant.KESClientCert() {
						encryptConfig.KmsMtls = &models.EncryptionConfigurationResponseAO1KmsMtls{}
						clientSecretName := tenant.Spec.KES.ClientCertSecret.Name
						keyPair, err := clientSet.getSecret(ctx, namespace, clientSecretName, metav1.GetOptions{})
						if err != nil {
							return nil, err
						}
						// Extract client public certificate
						if rawCert, ok := keyPair.Data["client.crt"]; ok {
							encryptConfig.KmsMtls.Crt, err = parseCertificate(clientSecretName, rawCert)
							if err != nil {
								return nil, err
							}
						}
						// Extract client ca certificate
						if rawCert, ok := keyPair.Data["ca.crt"]; ok {
							encryptConfig.KmsMtls.Ca, err = parseCertificate(clientSecretName, rawCert)
							if err != nil {
								return nil, err
							}
						}
					}
					encryptConfig.Vault = vaultConfig
				}
				if kesConfiguration.Keys.Aws != nil {
					awsJSON, err := json.Marshal(kesConfiguration.Keys.Aws)
					if err != nil {
						return nil, err
					}
					awsConfig := &models.AwsConfiguration{}
					err = json.Unmarshal(awsJSON, awsConfig)
					if err != nil {
						return nil, err
					}
					encryptConfig.Aws = awsConfig
				}
				if kesConfiguration.Keys.Gcp != nil {
					gcpJSON, err := json.Marshal(kesConfiguration.Keys.Gcp)
					if err != nil {
						return nil, err
					}
					gcpConfig := &models.GcpConfiguration{}
					err = json.Unmarshal(gcpJSON, gcpConfig)
					if err != nil {
						return nil, err
					}
					encryptConfig.Gcp = gcpConfig
				}
				if kesConfiguration.Keys.Gemalto != nil {
					gemalto := kesConfiguration.Keys.Gemalto
					gemaltoConfig := &models.GemaltoConfigurationResponse{
						Keysecure: &models.GemaltoConfigurationResponseKeysecure{},
					}
					if gemalto.KeySecure != nil {
						gemaltoConfig.Keysecure.Endpoint = &gemalto.KeySecure.Endpoint
						if gemalto.KeySecure.Credentials != nil {
							gemaltoConfig.Keysecure.Credentials = &models.GemaltoConfigurationResponseKeysecureCredentials{
								Domain: &gemalto.KeySecure.Credentials.Domain,
								Retry:  int64(gemalto.KeySecure.Credentials.Retry.Seconds()),
								Token:  &gemalto.KeySecure.Credentials.Token,
							}
						}
						if gemalto.KeySecure.TLS != nil {
							if tenant.KESClientCert() {
								encryptConfig.KmsMtls = &models.EncryptionConfigurationResponseAO1KmsMtls{}
								clientSecretName := tenant.Spec.KES.ClientCertSecret.Name
								keyPair, err := clientSet.getSecret(ctx, namespace, clientSecretName, metav1.GetOptions{})
								if err != nil {
									return nil, err
								}
								// Extract client ca certificate
								if rawCert, ok := keyPair.Data["ca.crt"]; ok {
									encryptConfig.KmsMtls.Ca, err = parseCertificate(clientSecretName, rawCert)
									if err != nil {
										return nil, err
									}
								}
							}
						}
					}
					encryptConfig.Gemalto = gemaltoConfig
				}
				if kesConfiguration.Keys.Azure != nil {
					azureJSON, err := json.Marshal(kesConfiguration.Keys.Azure)
					if err != nil {
						return nil, err
					}
					azureConfig := &models.AzureConfiguration{}
					err = json.Unmarshal(azureJSON, azureConfig)
					if err != nil {
						return nil, err
					}
					encryptConfig.Azure = azureConfig
				}
			}
		}
		return encryptConfig, nil
	}
	return nil, ErrEncryptionConfigNotFound
}

// getTenantEncryptionResponse is a wrapper for tenantEncryptionInfo
func getTenantEncryptionInfoResponse(session *models.Principal, params operator_api.TenantEncryptionInfoParams) (*models.EncryptionConfigurationResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrEncryptionConfigNotFound)
	}
	k8sClient := k8sClient{
		client: clientSet,
	}
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrEncryptionConfigNotFound)
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	configuration, err := tenantEncryptionInfo(ctx, &opClient, &k8sClient, params.Namespace, params)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return configuration, nil
}

// getKESConfiguration will generate the KES server certificate secrets, the tenant client secrets for mTLS authentication between MinIO and KES and the
// kes-configuration.yaml file used by the KES service (how to connect to the external KMS, eg: Vault, AWS, Gemalto, etc)
func getKESConfiguration(ctx context.Context, clientSet K8sClientI, ns string, encryptionCfg *models.EncryptionConfiguration, secretName, tenantName string) (kesConfiguration *miniov2.KESConfig, err error) {
	// Secrets used by the KES service
	//
	// kesExternalCertSecretName is the name of the secret that will store the certificates for TLS in the KES server, eg: server.key and server.crt
	kesExternalCertSecretName := fmt.Sprintf("%s-kes-external-cert", secretName)
	// kesClientCertSecretName is the name of the secret that will store the certificates for mTLS between KES and the KMS, eg: mTLS with Vault or Gemalto KMS
	kesClientCertSecretName := fmt.Sprintf("%s-kes-client-cert", secretName)
	// kesConfigurationSecretName is the name of the secret that will store the configuration file, eg: kes-configuration.yaml
	kesConfigurationSecretName := fmt.Sprintf("%s-kes-configuration", secretName)

	kesConfiguration = &miniov2.KESConfig{
		Image:    KESImageVersion,
		Replicas: 1,
	}
	// Using custom image for KES
	if encryptionCfg.Image != "" {
		kesConfiguration.Image = encryptionCfg.Image
	}
	// Using custom replicas for KES
	if encryptionCfg.Replicas != "" {
		replicas, errReplicas := strconv.Atoi(encryptionCfg.Replicas)
		if errReplicas != nil {
			kesConfiguration.Replicas = int32(replicas)
		}
	}
	// Generate server certificates for KES
	if encryptionCfg.ServerTLS != nil {
		certificates := []*models.KeyPairConfiguration{encryptionCfg.ServerTLS}
		certificateSecrets, err := createOrReplaceExternalCertSecrets(ctx, clientSet, ns, certificates, kesExternalCertSecretName, tenantName)
		if err != nil {
			return nil, err
		}
		if len(certificateSecrets) > 0 {
			// External TLS certificates used by KES
			kesConfiguration.ExternalCertSecret = certificateSecrets[0]
		}
	}
	// Prepare kesConfiguration for KES
	serverConfigSecret, clientCertSecret, err := createOrReplaceKesConfigurationSecrets(ctx, clientSet, ns, encryptionCfg, kesConfigurationSecretName, kesClientCertSecretName, tenantName)
	if err != nil {
		return nil, err
	}
	// Configuration used by KES
	kesConfiguration.Configuration = serverConfigSecret
	kesConfiguration.ClientCertSecret = clientCertSecret

	return kesConfiguration, nil
}

type tenantSecret struct {
	Name    string
	Content map[string][]byte
}

// createOrReplaceSecrets receives an array of Tenant Secrets to be stored as k8s secrets
func createOrReplaceSecrets(ctx context.Context, clientSet K8sClientI, ns string, secrets []tenantSecret, tenantName string) ([]*miniov2.LocalCertificateReference, error) {
	var k8sSecrets []*miniov2.LocalCertificateReference
	for _, secret := range secrets {
		if len(secret.Content) > 0 && secret.Name != "" {
			// delete secret with same name if exists
			err := clientSet.deleteSecret(ctx, ns, secret.Name, metav1.DeleteOptions{})
			if err != nil {
				// log the errors if any and continue
				LogError("deleting secret name %s failed: %v, continuing..", secret.Name, err)
			}
			k8sSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secret.Name,
					Labels: map[string]string{
						miniov2.TenantLabel: tenantName,
					},
				},
				Type: corev1.SecretTypeOpaque,
				Data: secret.Content,
			}
			_, err = clientSet.createSecret(ctx, ns, k8sSecret, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
			k8sSecrets = append(k8sSecrets, &miniov2.LocalCertificateReference{
				Name: secret.Name,
				Type: "Opaque",
			})
		}
	}
	return k8sSecrets, nil
}

// createOrReplaceExternalCertSecrets receives an array of KeyPairs (public and private key), encoded in base64, decode it and generate an equivalent number of kubernetes
// secrets to be used by the miniov2 for TLS encryption
func createOrReplaceExternalCertSecrets(ctx context.Context, clientSet K8sClientI, ns string, keyPairs []*models.KeyPairConfiguration, secretName, tenantName string) ([]*miniov2.LocalCertificateReference, error) {
	var keyPairSecrets []*miniov2.LocalCertificateReference
	for i, keyPair := range keyPairs {
		keyPairSecretName := fmt.Sprintf("%s-%d", secretName, i)
		if keyPair == nil || keyPair.Crt == nil || keyPair.Key == nil || *keyPair.Crt == "" || *keyPair.Key == "" {
			return nil, errors.New("certificate files must not be empty")
		}
		// delete secret with same name if exists
		err := clientSet.deleteSecret(ctx, ns, keyPairSecretName, metav1.DeleteOptions{})
		if err != nil {
			// log the errors if any and continue
			LogError("deleting secret name %s failed: %v, continuing..", keyPairSecretName, err)
		}
		imm := true
		tlsCrt, err := base64.StdEncoding.DecodeString(*keyPair.Crt)
		if err != nil {
			return nil, err
		}
		tlsKey, err := base64.StdEncoding.DecodeString(*keyPair.Key)
		if err != nil {
			return nil, err
		}
		// check if the key pair is valid
		if _, err = tls.X509KeyPair(tlsCrt, tlsKey); err != nil {
			return nil, err
		}
		externalTLSCertificateSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: keyPairSecretName,
				Labels: map[string]string{
					miniov2.TenantLabel: tenantName,
				},
			},
			Type:      corev1.SecretTypeTLS,
			Immutable: &imm,
			Data: map[string][]byte{
				"tls.crt": tlsCrt,
				"tls.key": tlsKey,
			},
		}
		_, err = clientSet.createSecret(ctx, ns, externalTLSCertificateSecret, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		// Certificates used by the minio instance
		keyPairSecrets = append(keyPairSecrets, &miniov2.LocalCertificateReference{
			Name: keyPairSecretName,
			Type: "kubernetes.io/tls",
		})
	}
	return keyPairSecrets, nil
}

func createOrReplaceKesConfigurationSecrets(ctx context.Context, clientSet K8sClientI, ns string, encryptionCfg *models.EncryptionConfiguration, kesConfigurationSecretName, kesClientCertSecretName, tenantName string) (*corev1.LocalObjectReference, *miniov2.LocalCertificateReference, error) {
	// if autoCert is enabled then Operator will generate the client certificates, calculate the client cert identity
	// and pass it to KES via the ${MINIO_KES_IDENTITY} variable
	clientCrtIdentity := "${MINIO_KES_IDENTITY}"
	// If a client certificate is provided proceed to calculate the identity
	if encryptionCfg.MinioMtls != nil {
		// Client certificate for KES used by Minio to mTLS
		clientTLSCrt, err := base64.StdEncoding.DecodeString(*encryptionCfg.MinioMtls.Crt)
		if err != nil {
			return nil, nil, err
		}
		// Calculate the client cert identity based on the clientTLSCrt
		h := crypto.SHA256.New()
		certificate, err := kes.ParseCertificate(clientTLSCrt)
		if err != nil {
			return nil, nil, err
		}
		h.Write(certificate.RawSubjectPublicKeyInfo)
		clientCrtIdentity = hex.EncodeToString(h.Sum(nil))
	}
	// Default kesConfiguration for KES
	kesConfig := &kes.ServerConfig{
		Addr: "0.0.0.0:7373",
		Root: "disabled",
		TLS: kes.TLS{
			KeyPath:  "/tmp/kes/server.key",
			CertPath: "/tmp/kes/server.crt",
		},
		Policies: map[string]kes.Policy{
			"default-policy": {
				Paths: []string{
					"/v1/key/create/my-minio-key",
					"/v1/key/generate/my-minio-key",
					"/v1/key/decrypt/my-minio-key",
				},
				Identities: []kes.Identity{
					kes.Identity(clientCrtIdentity),
				},
			},
		},
		Cache: kes.Cache{
			Expiry: &kes.Expiry{
				Any:    5 * time.Minute,
				Unused: 20 * time.Second,
			},
		},
		Log: kes.Log{
			Error: "on",
			Audit: "off",
		},
		Keys: kes.Keys{},
	}
	// miniov2 will mount the mTLSCertificates in the following paths
	// therefore we set these values in the KES yaml kesConfiguration
	mTLSClientCrtPath := "/tmp/kes/client.crt"
	mTLSClientKeyPath := "/tmp/kes/client.key"
	mTLSClientCaPath := "/tmp/kes/ca.crt"
	// map to hold mTLSCertificates for KES mTLS against Vault
	mTLSCertificates := map[string][]byte{}
	// if encryption is enabled and encryption is configured to use Vault
	switch {
	case encryptionCfg.Vault != nil:
		ping := 10 // default ping
		if encryptionCfg.Vault.Status != nil {
			ping = int(encryptionCfg.Vault.Status.Ping)
		}
		// Initialize Vault Config
		kesConfig.Keys.Vault = &kes.Vault{
			Endpoint:   *encryptionCfg.Vault.Endpoint,
			EnginePath: encryptionCfg.Vault.Engine,
			Namespace:  encryptionCfg.Vault.Namespace,
			Prefix:     encryptionCfg.Vault.Prefix,
			Status: &kes.VaultStatus{
				Ping: time.Duration(ping) * time.Second,
			},
		}
		// Vault AppRole credentials
		if encryptionCfg.Vault.Approle != nil {
			retry := encryptionCfg.Vault.Approle.Retry
			kesConfig.Keys.Vault.AppRole = &kes.AppRole{
				EnginePath: encryptionCfg.Vault.Approle.Engine,
				ID:         *encryptionCfg.Vault.Approle.ID,
				Secret:     *encryptionCfg.Vault.Approle.Secret,
				Retry:      time.Duration(retry) * time.Second,
			}
		} else {
			return nil, nil, errors.New("approle credentials missing for kes")
		}
		// Vault mTLS kesConfiguration
		if encryptionCfg.KmsMtls != nil {
			vaultTLSConfig := encryptionCfg.KmsMtls
			kesConfig.Keys.Vault.TLS = &kes.VaultTLS{}
			if vaultTLSConfig.Crt != "" {
				clientCrt, err := base64.StdEncoding.DecodeString(vaultTLSConfig.Crt)
				if err != nil {
					return nil, nil, err
				}
				mTLSCertificates["client.crt"] = clientCrt
				kesConfig.Keys.Vault.TLS.CertPath = mTLSClientCrtPath
			}
			if vaultTLSConfig.Key != "" {
				clientKey, err := base64.StdEncoding.DecodeString(vaultTLSConfig.Key)
				if err != nil {
					return nil, nil, err
				}
				mTLSCertificates["client.key"] = clientKey
				kesConfig.Keys.Vault.TLS.KeyPath = mTLSClientKeyPath
			}
			if vaultTLSConfig.Ca != "" {
				caCrt, err := base64.StdEncoding.DecodeString(vaultTLSConfig.Ca)
				if err != nil {
					return nil, nil, err
				}
				mTLSCertificates["ca.crt"] = caCrt
				kesConfig.Keys.Vault.TLS.CAPath = mTLSClientCaPath
			}
		}
	case encryptionCfg.Aws != nil:
		// Initialize AWS
		kesConfig.Keys.Aws = &kes.Aws{
			SecretsManager: &kes.AwsSecretManager{},
		}
		// AWS basic kesConfiguration
		if encryptionCfg.Aws.Secretsmanager != nil {
			kesConfig.Keys.Aws.SecretsManager.Endpoint = *encryptionCfg.Aws.Secretsmanager.Endpoint
			kesConfig.Keys.Aws.SecretsManager.Region = *encryptionCfg.Aws.Secretsmanager.Region
			kesConfig.Keys.Aws.SecretsManager.KmsKey = encryptionCfg.Aws.Secretsmanager.Kmskey
			// AWS credentials
			if encryptionCfg.Aws.Secretsmanager.Credentials != nil {
				kesConfig.Keys.Aws.SecretsManager.Login = &kes.AwsSecretManagerLogin{
					AccessKey:    *encryptionCfg.Aws.Secretsmanager.Credentials.Accesskey,
					SecretKey:    *encryptionCfg.Aws.Secretsmanager.Credentials.Secretkey,
					SessionToken: encryptionCfg.Aws.Secretsmanager.Credentials.Token,
				}
			}
		}
	case encryptionCfg.Gemalto != nil:
		// Initialize Gemalto
		kesConfig.Keys.Gemalto = &kes.Gemalto{
			KeySecure: &kes.GemaltoKeySecure{},
		}
		// Gemalto Configuration
		if encryptionCfg.Gemalto.Keysecure != nil {
			kesConfig.Keys.Gemalto.KeySecure.Endpoint = *encryptionCfg.Gemalto.Keysecure.Endpoint
			// Gemalto TLS kesConfiguration
			if encryptionCfg.KmsMtls != nil {
				if encryptionCfg.KmsMtls.Ca != "" {
					caCrt, err := base64.StdEncoding.DecodeString(encryptionCfg.KmsMtls.Ca)
					if err != nil {
						return nil, nil, err
					}
					mTLSCertificates["ca.crt"] = caCrt
					kesConfig.Keys.Gemalto.KeySecure.TLS = &kes.GemaltoTLS{
						CAPath: mTLSClientCaPath,
					}
				}
			}
			// Gemalto Login
			if encryptionCfg.Gemalto.Keysecure.Credentials != nil {
				kesConfig.Keys.Gemalto.KeySecure.Credentials = &kes.GemaltoCredentials{
					Token:  *encryptionCfg.Gemalto.Keysecure.Credentials.Token,
					Domain: *encryptionCfg.Gemalto.Keysecure.Credentials.Domain,
					Retry:  15 * time.Second,
				}
			}
		}
	case encryptionCfg.Gcp != nil:
		// Initialize GCP
		kesConfig.Keys.Gcp = &kes.Gcp{
			SecretManager: &kes.GcpSecretManager{},
		}
		// GCP basic kesConfiguration
		if encryptionCfg.Gcp.Secretmanager != nil {
			kesConfig.Keys.Gcp.SecretManager.ProjectID = *encryptionCfg.Gcp.Secretmanager.ProjectID
			kesConfig.Keys.Gcp.SecretManager.Endpoint = encryptionCfg.Gcp.Secretmanager.Endpoint
			// GCP credentials
			if encryptionCfg.Gcp.Secretmanager.Credentials != nil {
				kesConfig.Keys.Gcp.SecretManager.Credentials = &kes.GcpCredentials{
					ClientEmail:  encryptionCfg.Gcp.Secretmanager.Credentials.ClientEmail,
					ClientID:     encryptionCfg.Gcp.Secretmanager.Credentials.ClientID,
					PrivateKeyID: encryptionCfg.Gcp.Secretmanager.Credentials.PrivateKeyID,
					PrivateKey:   encryptionCfg.Gcp.Secretmanager.Credentials.PrivateKey,
				}
			}
		}
	case encryptionCfg.Azure != nil:
		// Initialize Azure
		kesConfig.Keys.Azure = &kes.Azure{
			KeyVault: &kes.AzureKeyVault{},
		}
		if encryptionCfg.Azure.Keyvault != nil {
			kesConfig.Keys.Azure.KeyVault.Endpoint = *encryptionCfg.Azure.Keyvault.Endpoint
			if encryptionCfg.Azure.Keyvault.Credentials != nil {
				kesConfig.Keys.Azure.KeyVault.Credentials = &kes.AzureCredentials{
					TenantID:     *encryptionCfg.Azure.Keyvault.Credentials.TenantID,
					ClientID:     *encryptionCfg.Azure.Keyvault.Credentials.ClientID,
					ClientSecret: *encryptionCfg.Azure.Keyvault.Credentials.ClientSecret,
				}
			}
		}
	}
	imm := true
	// if mTLSCertificates contains elements we create the kubernetes secret
	var clientCertSecretReference *miniov2.LocalCertificateReference
	if len(mTLSCertificates) > 0 {
		// delete KES client cert secret only if new client certificates are provided
		if err := clientSet.deleteSecret(ctx, ns, kesClientCertSecretName, metav1.DeleteOptions{}); err != nil {
			// log the errors if any and continue
			LogError("deleting secret name %s failed: %v, continuing..", kesClientCertSecretName, err)
		}
		// Secret to store KES mTLS kesConfiguration to authenticate against a KMS
		kesClientCertSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: kesClientCertSecretName,
				Labels: map[string]string{
					miniov2.TenantLabel: tenantName,
				},
			},
			Immutable: &imm,
			Data:      mTLSCertificates,
		}
		_, err := clientSet.createSecret(ctx, ns, &kesClientCertSecret, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, err
		}
		// kubernetes generic secret
		clientCertSecretReference = &miniov2.LocalCertificateReference{
			Name: kesClientCertSecretName,
		}
	}

	var serverRawConfig []byte
	var err error

	if encryptionCfg.Raw != "" {
		serverRawConfig = []byte(encryptionCfg.Raw)
		// verify provided configuration is in valid YAML format
		var configTest kes.ServerConfig
		err = yaml.Unmarshal(serverRawConfig, &configTest)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Generate Yaml kesConfiguration for KES
		serverRawConfig, err = yaml.Marshal(kesConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	// delete KES configuration secret if exists
	if err := clientSet.deleteSecret(ctx, ns, kesConfigurationSecretName, metav1.DeleteOptions{}); err != nil {
		// log the errors if any and continue
		LogError("deleting secret name %s failed: %v, continuing..", kesConfigurationSecretName, err)
	}

	// Secret to store KES server kesConfiguration
	kesConfigurationSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: kesConfigurationSecretName,
			Labels: map[string]string{
				miniov2.TenantLabel: tenantName,
			},
		},
		Immutable: &imm,
		Data: map[string][]byte{
			"server-config.yaml": serverRawConfig,
		},
	}
	_, err = clientSet.createSecret(ctx, ns, &kesConfigurationSecret, metav1.CreateOptions{})
	return &corev1.LocalObjectReference{
		Name: kesConfigurationSecretName,
	}, clientCertSecretReference, err
}
