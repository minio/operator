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
	"encoding/base64"
	"fmt"

	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/pkg/auth/utils"

	corev1 "k8s.io/api/core/v1"

	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getTenantCreatedResponse(session *models.Principal, params operator_api.CreateTenantParams) (response *models.CreateTenantResponse, mError *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	k8sClient := &k8sClient{
		client: clientSet,
	}
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	return createTenant(ctx, params, k8sClient, session)
}

func createTenant(ctx context.Context, params operator_api.CreateTenantParams, clientSet K8sClientI, session *models.Principal) (response *models.CreateTenantResponse, mError *models.Error) {
	tenantReq := params.Body
	minioImage := getTenantMinIOImage(tenantReq.Image)

	ns := *tenantReq.Namespace

	accessKey, secretKey := getTenantCredentials(tenantReq.AccessKey, tenantReq.SecretKey)
	tenantName := *tenantReq.Name
	var users []*corev1.LocalObjectReference

	// delete secrets created if an errors occurred during tenant creation,
	defer func() {
		deleteSecretsIfTenantCreationFails(ctx, mError, tenantName, ns, clientSet)
	}()

	err := createTenantCredentialsSecret(ctx, ns, tenantName, clientSet)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	tenantConfigurationENV := map[string]string{}
	tenantConfigurationENV["MINIO_BROWSER"] = isTenantConsoleEnabled(tenantReq.EnableConsole)
	tenantConfigurationENV["MINIO_ROOT_USER"] = accessKey
	tenantConfigurationENV["MINIO_ROOT_PASSWORD"] = secretKey

	// Check the Erasure Coding Parity for validity and pass it to Tenant
	if tenantReq.ErasureCodingParity > 0 {
		if tenantReq.ErasureCodingParity < 2 || tenantReq.ErasureCodingParity > 8 {
			return nil, ErrorWithContext(ctx, ErrInvalidErasureCodingValue)
		}
		tenantConfigurationENV["MINIO_STORAGE_CLASS_STANDARD"] = fmt.Sprintf("EC:%d", tenantReq.ErasureCodingParity)
	}

	// Construct a MinIO Instance with everything we are getting from parameters
	minInst := miniov2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tenantName,
			Labels: tenantReq.Labels,
		},
		Spec: miniov2.TenantSpec{
			Image:     minioImage,
			Mountpath: "/export",
			CredsSecret: &corev1.LocalObjectReference{
				Name: fmt.Sprintf("%s-secret", tenantName),
			},
		},
	}
	var tenantExternalIDPConfigured bool
	if tenantReq.Idp != nil {
		// Enable IDP (Active Directory) for MinIO
		switch {
		case tenantReq.Idp.ActiveDirectory != nil:
			tenantExternalIDPConfigured = true
			tenantConfigurationENV, users, err = setTenantActiveDirectoryConfig(ctx, clientSet, tenantReq, tenantConfigurationENV, users)
			if err != nil {
				return nil, ErrorWithContext(ctx, err)
			}
			// attach the users to the tenant
			minInst.Spec.Users = users
		case tenantReq.Idp.Oidc != nil:
			tenantExternalIDPConfigured = true
			tenantConfigurationENV = setTenantOIDCConfig(tenantReq, tenantConfigurationENV)
		case len(tenantReq.Idp.Keys) > 0:
			users, err = setTenantBuiltInUsers(ctx, clientSet, tenantReq, users)
			if err != nil {
				return nil, ErrorWithContext(ctx, err)
			}
			// attach the users to the tenant
			minInst.Spec.Users = users
		}
	}

	canEncryptionBeEnabled := false

	if tenantReq.EnableTLS != nil {
		// if enableTLS is defined in the create tenant request we assign the value
		// to the RequestAutoCert attribute in the tenant spec
		minInst.Spec.RequestAutoCert = tenantReq.EnableTLS
		if *tenantReq.EnableTLS {
			// requestAutoCert is enabled, MinIO will be deployed with TLS enabled and encryption can be enabled
			canEncryptionBeEnabled = true
		}
	}
	// External server TLS certificates for MinIO
	if tenantReq.TLS != nil && len(tenantReq.TLS.MinioServerCertificates) > 0 {
		canEncryptionBeEnabled = true
		// Certificates used by the MinIO instance
		externalCertSecretName := fmt.Sprintf("%s-external-server-certificate", tenantName)
		externalCertSecret, err := createOrReplaceExternalCertSecrets(ctx, clientSet, ns, tenantReq.TLS.MinioServerCertificates, externalCertSecretName, tenantName)
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		minInst.Spec.ExternalCertSecret = externalCertSecret
	}
	// External client TLS certificates for MinIO
	if tenantReq.TLS != nil && len(tenantReq.TLS.MinioClientCertificates) > 0 {
		// Client certificates used by the MinIO instance
		externalClientCertSecretName := fmt.Sprintf("%s-external-client-certificate", tenantName)
		externalClientCertSecret, err := createOrReplaceExternalCertSecrets(ctx, clientSet, ns, tenantReq.TLS.MinioClientCertificates, externalClientCertSecretName, tenantName)
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		minInst.Spec.ExternalClientCertSecrets = externalClientCertSecret
	}
	// If encryption configuration is present and TLS will be enabled (using AutoCert or External certificates)
	if tenantReq.Encryption != nil && canEncryptionBeEnabled {
		// KES client mTLSCertificates used by MinIO instance
		if tenantReq.Encryption.MinioMtls != nil {
			tenantExternalClientCertSecretName := fmt.Sprintf("%s-external-client-certificate-kes", tenantName)
			certificates := []*models.KeyPairConfiguration{tenantReq.Encryption.MinioMtls}
			certificateSecrets, err := createOrReplaceExternalCertSecrets(ctx, clientSet, ns, certificates, tenantExternalClientCertSecretName, tenantName)
			if err != nil {
				return nil, ErrorWithContext(ctx, ErrDefault)
			}
			if len(certificateSecrets) > 0 {
				minInst.Spec.ExternalClientCertSecret = certificateSecrets[0]
			}
		}

		// KES configuration for Tenant instance
		minInst.Spec.KES, err = getKESConfiguration(ctx, clientSet, ns, tenantReq.Encryption, fmt.Sprintf("%s-secret", tenantName), tenantName)
		if err != nil {
			return nil, ErrorWithContext(ctx, ErrDefault)
		}
		// Set Labels, Annotations and Node Selector for KES
		minInst.Spec.KES.Labels = tenantReq.Encryption.Labels
		minInst.Spec.KES.Annotations = tenantReq.Encryption.Annotations
		minInst.Spec.KES.NodeSelector = tenantReq.Encryption.NodeSelector

		if tenantReq.Encryption.SecurityContext != nil {
			sc, err := convertModelSCToK8sSC(tenantReq.Encryption.SecurityContext)
			if err != nil {
				return nil, ErrorWithContext(ctx, err)
			}
			minInst.Spec.KES.SecurityContext = sc
		}
	}
	// External TLS CA certificates for MinIO
	if tenantReq.TLS != nil && len(tenantReq.TLS.MinioCAsCertificates) > 0 {
		var caCertificates []tenantSecret
		for i, caCertificate := range tenantReq.TLS.MinioCAsCertificates {
			certificateContent, err := base64.StdEncoding.DecodeString(caCertificate)
			if err != nil {
				return nil, ErrorWithContext(ctx, ErrDefault, nil, err)
			}
			caCertificates = append(caCertificates, tenantSecret{
				Name: fmt.Sprintf("%s-ca-certificate-%d", tenantName, i),
				Content: map[string][]byte{
					"public.crt": certificateContent,
				},
			})
		}
		if len(caCertificates) > 0 {
			certificateSecrets, err := createOrReplaceSecrets(ctx, clientSet, ns, caCertificates, tenantName)
			if err != nil {
				return nil, ErrorWithContext(ctx, ErrDefault, nil, err)
			}
			minInst.Spec.ExternalCaCertSecret = certificateSecrets
		}
	}

	// add annotations
	var annotations map[string]string

	if len(tenantReq.Annotations) > 0 {
		annotations = tenantReq.Annotations
		minInst.Annotations = annotations
	}
	// set the pools if they are provided
	for _, pool := range tenantReq.Pools {
		pool, err := parseTenantPoolRequest(pool)
		if err != nil {
			LogError("parseTenantPoolRequest failed: %v", err)
			return nil, ErrorWithContext(ctx, err)
		}
		minInst.Spec.Pools = append(minInst.Spec.Pools, *pool)
	}

	// Set Mount Path if provided
	if tenantReq.MountPath != "" {
		minInst.Spec.Mountpath = tenantReq.MountPath
	}

	// We accept either `image_pull_secret` or the individual details of the `image_registry` but not both
	var imagePullSecret string

	if tenantReq.ImagePullSecret != "" {
		imagePullSecret = tenantReq.ImagePullSecret
	} else if imagePullSecret, err = setImageRegistry(ctx, tenantReq.ImageRegistry, clientSet, ns, tenantName); err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	// pass the image pull secret to the Tenant
	if imagePullSecret != "" {
		minInst.Spec.ImagePullSecret = corev1.LocalObjectReference{
			Name: imagePullSecret,
		}
	}

	// expose services
	minInst.Spec.ExposeServices = &miniov2.ExposeServices{
		MinIO:   tenantReq.ExposeMinio,
		Console: tenantReq.ExposeConsole,
	}

	// set custom environment variables in configuration file
	for _, envVar := range tenantReq.EnvironmentVariables {
		tenantConfigurationENV[envVar.Key] = envVar.Value
	}

	// write tenant configuration to secret that contains config.env
	tenantConfigurationName := fmt.Sprintf("%s-env-configuration", tenantName)
	_, err = createOrReplaceSecrets(ctx, clientSet, ns, []tenantSecret{
		{
			Name: tenantConfigurationName,
			Content: map[string][]byte{
				"config.env": []byte(GenerateTenantConfigurationFile(tenantConfigurationENV)),
			},
		},
	}, tenantName)
	if err != nil {
		return nil, ErrorWithContext(ctx, ErrDefault, nil, err)
	}
	minInst.Spec.Configuration = &corev1.LocalObjectReference{Name: tenantConfigurationName}

	var features miniov2.Features
	if tenantReq.Domains != nil {
		var domains miniov2.TenantDomains

		// tenant domains
		if tenantReq.Domains.Console != "" {
			domains.Console = tenantReq.Domains.Console
		}

		if tenantReq.Domains.Minio != nil {
			domains.Minio = tenantReq.Domains.Minio
		}

		features.Domains = &domains
	}
	if tenantReq.ExposeSftp {
		features.EnableSFTP = &tenantReq.ExposeSftp
	}
	minInst.Spec.Features = &features

	opClient, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	_, err = opClient.MinioV2().Tenants(ns).Create(context.Background(), &minInst, metav1.CreateOptions{})
	if err != nil {
		LogError("Creating new tenant failed with: %v", err)
		return nil, ErrorWithContext(ctx, err)
	}

	response = &models.CreateTenantResponse{
		ExternalIDP: tenantExternalIDPConfigured,
	}
	thisClient := &operatorClient{
		client: opClient,
	}

	minTenant, _ := getTenant(ctx, thisClient, ns, tenantName)

	if tenantReq.Idp != nil && !tenantExternalIDPConfigured {
		for _, credential := range tenantReq.Idp.Keys {
			response.Console = append(response.Console, &models.TenantResponseItem{
				AccessKey: *credential.AccessKey,
				SecretKey: *credential.SecretKey,
				URL:       GetTenantServiceURL(minTenant),
			})
		}
	}
	return response, nil
}

func getTenantMinIOImage(minioImage string) string {
	if minioImage == "" {
		minImg, err := GetMinioImage()
		// we can live without figuring out the latest version of MinIO, Operator will use a hardcoded value
		if err == nil {
			minioImage = *minImg
		}
	}
	return minioImage
}

func getTenantCredentials(accessKey, secretKey string) (string, string) {
	defaultAccessKey := utils.RandomCharString(16)
	defaultSecretKey := utils.RandomCharString(32)
	if accessKey != "" {
		defaultAccessKey = accessKey
	}
	if secretKey != "" {
		defaultSecretKey = secretKey
	}
	return defaultAccessKey, defaultSecretKey
}

func createTenantCredentialsSecret(ctx context.Context, ns, tenantName string, clientSet K8sClientI) error {
	imm := true
	// Create the secret for the root credentials (deprecated)
	secretName := fmt.Sprintf("%s-secret", tenantName)
	instanceSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				miniov2.TenantLabel: tenantName,
			},
		},
		Immutable: &imm,
		Data: map[string][]byte{
			"accesskey": []byte(""),
			"secretkey": []byte(""),
		},
	}
	_, err := clientSet.createSecret(ctx, ns, &instanceSecret, metav1.CreateOptions{})
	return err
}

func isTenantConsoleEnabled(enable *bool) string {
	enabledConsole := "on"
	if enable != nil && !*enable {
		enabledConsole = "off"
	}
	return enabledConsole
}

func deleteSecretsIfTenantCreationFails(ctx context.Context, mError *models.Error, tenantName, ns string, clientSet K8sClientI) {
	if mError != nil {
		LogError("deleting secrets created for failed tenant: %s if any: %v", tenantName, mError)
		opts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenantName),
		}
		err := clientSet.deleteSecretsCollection(ctx, ns, metav1.DeleteOptions{}, opts)
		if err != nil {
			LogError("error deleting tenant's secrets: %v", err)
		}
	}
}

func setTenantActiveDirectoryConfig(ctx context.Context, clientSet K8sClientI, tenantReq *models.CreateTenantRequest, tenantConfigurationENV map[string]string, users []*corev1.LocalObjectReference) (map[string]string, []*corev1.LocalObjectReference, error) {
	imm := true
	serverAddress := *tenantReq.Idp.ActiveDirectory.URL
	tlsSkipVerify := tenantReq.Idp.ActiveDirectory.SkipTLSVerification
	serverInsecure := tenantReq.Idp.ActiveDirectory.ServerInsecure
	lookupBindDN := *tenantReq.Idp.ActiveDirectory.LookupBindDn
	lookupBindPassword := tenantReq.Idp.ActiveDirectory.LookupBindPassword
	userDNSearchBaseDN := tenantReq.Idp.ActiveDirectory.UserDnSearchBaseDn
	userDNSearchFilter := tenantReq.Idp.ActiveDirectory.UserDnSearchFilter
	groupSearchBaseDN := tenantReq.Idp.ActiveDirectory.GroupSearchBaseDn
	groupSearchFilter := tenantReq.Idp.ActiveDirectory.GroupSearchFilter
	serverStartTLS := tenantReq.Idp.ActiveDirectory.ServerStartTLS

	// LDAP Server
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_SERVER_ADDR"] = serverAddress
	if tlsSkipVerify {
		tenantConfigurationENV["MINIO_IDENTITY_LDAP_TLS_SKIP_VERIFY"] = "on"
	}
	if serverInsecure {
		tenantConfigurationENV["MINIO_IDENTITY_LDAP_SERVER_INSECURE"] = "on"
	}
	if serverStartTLS {
		tenantConfigurationENV["MINIO_IDENTITY_LDAP_SERVER_STARTTLS"] = "on"
	}

	// LDAP Lookup
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN"] = lookupBindDN
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD"] = lookupBindPassword

	// LDAP User DN
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN"] = userDNSearchBaseDN
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER"] = userDNSearchFilter

	// LDAP Group
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN"] = groupSearchBaseDN
	tenantConfigurationENV["MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER"] = groupSearchFilter

	// Attach the list of LDAP user DNs that will be administrator for the Tenant
	for i, userDN := range tenantReq.Idp.ActiveDirectory.UserDNS {
		userSecretName := fmt.Sprintf("%s-user-%d", *tenantReq.Name, i)
		users = append(users, &corev1.LocalObjectReference{Name: userSecretName})

		userSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: userSecretName,
				Labels: map[string]string{
					miniov2.TenantLabel: *tenantReq.Name,
				},
			},
			Immutable: &imm,
			Data: map[string][]byte{
				"CONSOLE_ACCESS_KEY": []byte(userDN),
			},
		}
		_, err := clientSet.createSecret(ctx, *tenantReq.Namespace, &userSecret, metav1.CreateOptions{})
		if err != nil {
			return tenantConfigurationENV, users, err
		}
	}
	return tenantConfigurationENV, users, nil
}

func setTenantOIDCConfig(tenantReq *models.CreateTenantRequest, tenantConfigurationENV map[string]string) map[string]string {
	configurationURL := *tenantReq.Idp.Oidc.ConfigurationURL
	clientID := *tenantReq.Idp.Oidc.ClientID
	secretID := *tenantReq.Idp.Oidc.SecretID
	claimName := *tenantReq.Idp.Oidc.ClaimName
	scopes := tenantReq.Idp.Oidc.Scopes
	callbackURL := tenantReq.Idp.Oidc.CallbackURL
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_CONFIG_URL"] = configurationURL
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_CLIENT_ID"] = clientID
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_CLIENT_SECRET"] = secretID
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_CLAIM_NAME"] = claimName
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_REDIRECT_URI"] = callbackURL
	if scopes == "" {
		scopes = "openid,profile,email"
	}
	tenantConfigurationENV["MINIO_IDENTITY_OPENID_SCOPES"] = scopes
	return tenantConfigurationENV
}

func setTenantBuiltInUsers(ctx context.Context, clientSet K8sClientI, tenantReq *models.CreateTenantRequest, users []*corev1.LocalObjectReference) ([]*corev1.LocalObjectReference, error) {
	imm := true
	for i := 0; i < len(tenantReq.Idp.Keys); i++ {
		userSecretName := fmt.Sprintf("%s-user-%d", *tenantReq.Name, i)
		users = append(users, &corev1.LocalObjectReference{Name: userSecretName})
		userSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: userSecretName,
				Labels: map[string]string{
					miniov2.TenantLabel: *tenantReq.Name,
				},
			},
			Immutable: &imm,
			Data: map[string][]byte{
				"CONSOLE_ACCESS_KEY": []byte(*tenantReq.Idp.Keys[i].AccessKey),
				"CONSOLE_SECRET_KEY": []byte(*tenantReq.Idp.Keys[i].SecretKey),
			},
		}
		_, err := clientSet.createSecret(ctx, *tenantReq.Namespace, &userSecret, metav1.CreateOptions{})
		if err != nil {
			return users, err
		}
	}
	return users, nil
}
