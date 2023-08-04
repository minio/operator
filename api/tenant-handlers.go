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
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	utils2 "github.com/minio/operator/pkg/http"

	"github.com/minio/madmin-go/v2"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type imageRegistry struct {
	Auths map[string]imageRegistryCredentials `json:"auths"`
}

type imageRegistryCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

func registerTenantHandlers(api *operations.OperatorAPI) {
	// Add Tenant
	api.OperatorAPICreateTenantHandler = operator_api.CreateTenantHandlerFunc(func(params operator_api.CreateTenantParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantCreatedResponse(session, params)
		if err != nil {
			return operator_api.NewCreateTenantDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewCreateTenantOK().WithPayload(resp)
	})
	// List All Tenants of all namespaces
	api.OperatorAPIListAllTenantsHandler = operator_api.ListAllTenantsHandlerFunc(func(params operator_api.ListAllTenantsParams, session *models.Principal) middleware.Responder {
		resp, err := getListAllTenantsResponse(session, params)
		if err != nil {
			return operator_api.NewListTenantsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewListTenantsOK().WithPayload(resp)
	})
	// List Tenants by namespace
	api.OperatorAPIListTenantsHandler = operator_api.ListTenantsHandlerFunc(func(params operator_api.ListTenantsParams, session *models.Principal) middleware.Responder {
		resp, err := getListTenantsResponse(session, params)
		if err != nil {
			return operator_api.NewListTenantsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewListTenantsOK().WithPayload(resp)
	})
	// Detail Tenant
	api.OperatorAPITenantDetailsHandler = operator_api.TenantDetailsHandlerFunc(func(params operator_api.TenantDetailsParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantDetailsResponse(session, params)
		if err != nil {
			return operator_api.NewTenantDetailsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantDetailsOK().WithPayload(resp)
	})

	// Delete Tenant
	api.OperatorAPIDeleteTenantHandler = operator_api.DeleteTenantHandlerFunc(func(params operator_api.DeleteTenantParams, session *models.Principal) middleware.Responder {
		err := getDeleteTenantResponse(session, params)
		if err != nil {
			return operator_api.NewDeleteTenantDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDeleteTenantNoContent()
	})

	// Update Tenant
	api.OperatorAPIUpdateTenantHandler = operator_api.UpdateTenantHandlerFunc(func(params operator_api.UpdateTenantParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantCreated()
	})

	// Get Tenant Usage
	api.OperatorAPIGetTenantUsageHandler = operator_api.GetTenantUsageHandlerFunc(func(params operator_api.GetTenantUsageParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantUsageResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantUsageDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantUsageOK().WithPayload(payload)
	})
}

// getDeleteTenantResponse gets the output of deleting a minio instance
func getDeleteTenantResponse(session *models.Principal, params operator_api.DeleteTenantParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	// get Kubernetes Client
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	deleteTenantPVCs := false
	if params.Body != nil {
		deleteTenantPVCs = params.Body.DeletePvcs
	}

	tenant, err := opClient.TenantGet(params.HTTPRequest.Context(), params.Namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	tenant.EnsureDefaults()

	if err = deleteTenantAction(params.HTTPRequest.Context(), opClient, clientset.CoreV1(), tenant, deleteTenantPVCs); err != nil {
		return ErrorWithContext(ctx, err)
	}
	return nil
}

// deleteTenantAction performs the actions of deleting a tenant
//
// It also adds the option of deleting the tenant's underlying pvcs if deletePvcs set
func deleteTenantAction(
	ctx context.Context,
	operatorClient OperatorClientI,
	clientset v1.CoreV1Interface,
	tenant *miniov2.Tenant,
	deletePvcs bool,
) error {
	err := operatorClient.TenantDelete(ctx, tenant.Namespace, tenant.Name, metav1.DeleteOptions{})
	if err != nil {
		// try to delete pvc even if the tenant doesn't exist anymore but only if deletePvcs is set to true,
		// else, we return the errors
		if (deletePvcs && !k8sErrors.IsNotFound(err)) || !deletePvcs {
			return err
		}
	}

	if deletePvcs {

		// delete MinIO PVCs
		opts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
		}
		err = clientset.PersistentVolumeClaims(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, opts)
		if err != nil {
			return err
		}

		// delete all tenant's secrets only if deletePvcs = true
		return clientset.Secrets(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, opts)
	}
	return nil
}

// GetTenantServiceURL gets tenant's service url with the proper scheme and port
func GetTenantServiceURL(mi *miniov2.Tenant) (svcURL string) {
	scheme := "http"
	port := miniov2.MinIOPortLoadBalancerSVC
	if mi.AutoCert() || mi.ExternalCert() {
		scheme = "https"
		port = miniov2.MinIOTLSPortLoadBalancerSVC
	}
	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(mi.MinIOFQDNServiceName(), strconv.Itoa(port)))
}

func getTenantAdminClient(ctx context.Context, client K8sClientI, tenant *miniov2.Tenant, svcURL string) (*madmin.AdminClient, error) {
	tenantCreds, err := getTenantCreds(ctx, client, tenant)
	if err != nil {
		return nil, err
	}
	sessionToken := ""
	mAdmin, pErr := NewAdminClientWithInsecure(svcURL, tenantCreds.accessKey, tenantCreds.secretKey, sessionToken, true)
	if pErr != nil {
		return nil, pErr.Cause
	}
	return mAdmin, nil
}

type tenantKeys struct {
	accessKey string
	secretKey string
}

func getTenantCreds(ctx context.Context, client K8sClientI, tenant *miniov2.Tenant) (*tenantKeys, error) {
	tenantConfiguration, err := GetTenantConfiguration(ctx, client, tenant)
	if err != nil {
		return nil, err
	}
	tenantAccessKey, ok := tenantConfiguration["accesskey"]
	if !ok {
		LogError("tenant's secret doesn't contain accesskey")
		return nil, ErrDefault
	}
	tenantSecretKey, ok := tenantConfiguration["secretkey"]
	if !ok {
		LogError("tenant's secret doesn't contain secretkey")
		return nil, ErrDefault
	}
	return &tenantKeys{accessKey: tenantAccessKey, secretKey: tenantSecretKey}, nil
}

func getTenant(ctx context.Context, operatorClient OperatorClientI, namespace, tenantName string) (*miniov2.Tenant, error) {
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func getTenantInfo(tenant *miniov2.Tenant) *models.Tenant {
	var pools []*models.Pool
	var totalSize int64
	for _, p := range tenant.Spec.Pools {
		pools = append(pools, parseTenantPool(&p))
		poolSize := int64(p.Servers) * int64(p.VolumesPerServer) * p.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value()
		totalSize += poolSize
	}
	var deletion string
	if tenant.ObjectMeta.DeletionTimestamp != nil {
		deletion = tenant.ObjectMeta.DeletionTimestamp.Format(time.RFC3339)
	}

	return &models.Tenant{
		CreationDate: tenant.ObjectMeta.CreationTimestamp.Format(time.RFC3339),
		DeletionDate: deletion,
		Name:         tenant.Name,
		TotalSize:    totalSize,
		CurrentState: tenant.Status.CurrentState,
		Pools:        pools,
		Namespace:    tenant.ObjectMeta.Namespace,
		Image:        tenant.Spec.Image,
	}
}

func parseCertificate(name string, rawCert []byte) (*models.CertificateInfo, error) {
	block, _ := pem.Decode(rawCert)
	if block == nil {
		return nil, errors.New("certificate failed to decode")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	domains := []string{}
	// append certificate domain names
	if len(cert.DNSNames) > 0 {
		domains = append(domains, cert.DNSNames...)
	}
	// append certificate IPs
	if len(cert.IPAddresses) > 0 {
		for _, ip := range cert.IPAddresses {
			domains = append(domains, ip.String())
		}
	}
	return &models.CertificateInfo{
		SerialNumber: cert.SerialNumber.String(),
		Name:         name,
		Domains:      domains,
		Expiry:       cert.NotAfter.Format(time.RFC3339),
	}, nil
}

var secretTypePublicKeyNameMap = map[string]string{
	"kubernetes.io/tls":        "tls.crt",
	"cert-manager.io/v1":       "tls.crt",
	"cert-manager.io/v1alpha2": "tls.crt",
	// Add newer secretTypes and their corresponding values in future
}

// parseTenantCertificates convert public key pem certificates stored in k8s secrets for a given Tenant into x509 certificates
func parseTenantCertificates(ctx context.Context, clientSet K8sClientI, namespace string, secrets []*miniov2.LocalCertificateReference) ([]*models.CertificateInfo, error) {
	var certificates []*models.CertificateInfo
	publicKey := "public.crt"
	// Iterate over TLS secrets and build array of CertificateInfo structure
	// that will be used to display information about certs in the UI
	for _, secret := range secrets {
		keyPair, err := clientSet.getSecret(ctx, namespace, secret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if v, ok := secretTypePublicKeyNameMap[secret.Type]; ok {
			publicKey = v
		}
		var rawCert []byte
		if _, ok := keyPair.Data[publicKey]; !ok {
			return nil, fmt.Errorf("public key: %v not found inside certificate secret %v", publicKey, secret.Name)
		}
		rawCert = keyPair.Data[publicKey]
		var blocks []byte
		for {
			var block *pem.Block
			block, rawCert = pem.Decode(rawCert)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				blocks = append(blocks, block.Bytes...)
			}
		}
		// parse all certificates we found on this k8s secret
		certs, err := x509.ParseCertificates(blocks)
		if err != nil {
			return nil, err
		}
		for _, cert := range certs {
			var domains []string
			if cert.Subject.CommonName != "" {
				domains = append(domains, cert.Subject.CommonName)
			}
			// append certificate domain names
			if len(cert.DNSNames) > 0 {
				domains = append(domains, cert.DNSNames...)
			}
			// append certificate IPs
			if len(cert.IPAddresses) > 0 {
				for _, ip := range cert.IPAddresses {
					domains = append(domains, ip.String())
				}
			}
			certificates = append(certificates, &models.CertificateInfo{
				SerialNumber: cert.SerialNumber.String(),
				Name:         secret.Name,
				Domains:      domains,
				Expiry:       cert.NotAfter.Format(time.RFC3339),
			})
		}
	}
	return certificates, nil
}

func getTenantSecurity(ctx context.Context, clientSet K8sClientI, tenant *miniov2.Tenant) (response *models.TenantSecurityResponse, err error) {
	var minioExternalServerCertificates []*models.CertificateInfo
	var minioExternalClientCertificates []*models.CertificateInfo
	var minioExternalCaCertificates []*models.CertificateInfo
	var tenantSecurityContext *models.SecurityContext
	// Server certificates used by MinIO
	if minioExternalServerCertificates, err = parseTenantCertificates(ctx, clientSet, tenant.Namespace, tenant.Spec.ExternalCertSecret); err != nil {
		return nil, err
	}
	// Client certificates used by MinIO
	if minioExternalClientCertificates, err = parseTenantCertificates(ctx, clientSet, tenant.Namespace, tenant.Spec.ExternalClientCertSecrets); err != nil {
		return nil, err
	}
	// CA Certificates used by MinIO
	if minioExternalCaCertificates, err = parseTenantCertificates(ctx, clientSet, tenant.Namespace, tenant.Spec.ExternalCaCertSecret); err != nil {
		return nil, err
	}
	// Security Context used by MinIO server
	if len(tenant.Spec.Pools) > 0 && tenant.Spec.Pools[0].SecurityContext != nil {
		tenantSecurityContext = convertK8sSCToModelSC(tenant.Spec.Pools[0].SecurityContext)
	}
	return &models.TenantSecurityResponse{
		AutoCert: tenant.AutoCert(),
		CustomCertificates: &models.TenantSecurityResponseCustomCertificates{
			Minio:    minioExternalServerCertificates,
			MinioCAs: minioExternalCaCertificates,
			Client:   minioExternalClientCertificates,
		},
		SecurityContext: tenantSecurityContext,
	}, nil
}

func getTenantIdentityProvider(ctx context.Context, clientSet K8sClientI, tenant *miniov2.Tenant) (response *models.IdpConfiguration, err error) {
	tenantConfiguration, err := GetTenantConfiguration(ctx, clientSet, tenant)
	if err != nil {
		return nil, err
	}

	var idpConfiguration *models.IdpConfiguration

	if tenantConfiguration["MINIO_IDENTITY_OPENID_CONFIG_URL"] != "" {

		callbackURL := tenantConfiguration["MINIO_IDENTITY_OPENID_REDIRECT_URI"]
		claimName := tenantConfiguration["MINIO_IDENTITY_OPENID_CLAIM_NAME"]
		clientID := tenantConfiguration["MINIO_IDENTITY_OPENID_CLIENT_ID"]
		configurationURL := tenantConfiguration["MINIO_IDENTITY_OPENID_CONFIG_URL"]
		scopes := tenantConfiguration["MINIO_IDENTITY_OPENID_SCOPES"]
		secretID := tenantConfiguration["MINIO_IDENTITY_OPENID_CLIENT_SECRET"]

		idpConfiguration = &models.IdpConfiguration{
			Oidc: &models.IdpConfigurationOidc{
				CallbackURL:      callbackURL,
				ClaimName:        &claimName,
				ClientID:         &clientID,
				ConfigurationURL: &configurationURL,
				Scopes:           scopes,
				SecretID:         &secretID,
			},
		}
	}
	if tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_ADDR"] != "" {

		groupSearchBaseDN := tenantConfiguration["MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN"]
		groupSearchFilter := tenantConfiguration["MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER"]
		lookupBindDN := tenantConfiguration["MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN"]
		lookupBindPassword := tenantConfiguration["MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD"]
		serverInsecure := tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_INSECURE"] == "on"
		serverStartTLS := tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_STARTTLS"] == "on"
		tlsSkipVerify := tenantConfiguration["MINIO_IDENTITY_LDAP_TLS_SKIP_VERIFY"] == "on"
		serverAddress := tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_ADDR"]
		userDNSearchBaseDN := tenantConfiguration["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN"]
		userDNSearchFilter := tenantConfiguration["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER"]

		idpConfiguration = &models.IdpConfiguration{
			ActiveDirectory: &models.IdpConfigurationActiveDirectory{
				GroupSearchBaseDn:   groupSearchBaseDN,
				GroupSearchFilter:   groupSearchFilter,
				LookupBindDn:        &lookupBindDN,
				LookupBindPassword:  lookupBindPassword,
				ServerInsecure:      serverInsecure,
				ServerStartTLS:      serverStartTLS,
				SkipTLSVerification: tlsSkipVerify,
				URL:                 &serverAddress,
				UserDnSearchBaseDn:  userDNSearchBaseDN,
				UserDnSearchFilter:  userDNSearchFilter,
			},
		}
	}
	return idpConfiguration, nil
}

func updateTenantIdentityProvider(ctx context.Context, operatorClient OperatorClientI, client K8sClientI, namespace string, params operator_api.UpdateTenantIdentityProviderParams) error {
	tenant, err := operatorClient.TenantGet(ctx, namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return err
	}
	tenantConfiguration, err := GetTenantConfiguration(ctx, client, tenant)
	if err != nil {
		return err
	}

	delete(tenantConfiguration, "accesskey")
	delete(tenantConfiguration, "secretkey")

	oidcConfig := params.Body.Oidc
	// set new oidc configuration fields
	if oidcConfig != nil {
		configurationURL := *oidcConfig.ConfigurationURL
		clientID := *oidcConfig.ClientID
		secretID := *oidcConfig.SecretID
		claimName := *oidcConfig.ClaimName
		scopes := oidcConfig.Scopes
		callbackURL := oidcConfig.CallbackURL
		// oidc config
		tenantConfiguration["MINIO_IDENTITY_OPENID_CONFIG_URL"] = configurationURL
		tenantConfiguration["MINIO_IDENTITY_OPENID_CLIENT_ID"] = clientID
		tenantConfiguration["MINIO_IDENTITY_OPENID_CLIENT_SECRET"] = secretID
		tenantConfiguration["MINIO_IDENTITY_OPENID_CLAIM_NAME"] = claimName
		tenantConfiguration["MINIO_IDENTITY_OPENID_REDIRECT_URI"] = callbackURL
		if scopes == "" {
			scopes = "openid,profile,email"
		}
		tenantConfiguration["MINIO_IDENTITY_OPENID_SCOPES"] = scopes
	} else {
		// reset oidc configuration fields
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_CLAIM_NAME")
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_CLIENT_ID")
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_CONFIG_URL")
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_SCOPES")
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_CLIENT_SECRET")
		delete(tenantConfiguration, "MINIO_IDENTITY_OPENID_REDIRECT_URI")
	}
	ldapConfig := params.Body.ActiveDirectory
	// set new active directory configuration fields
	if ldapConfig != nil {
		// ldap config
		serverAddress := *ldapConfig.URL
		tlsSkipVerify := ldapConfig.SkipTLSVerification
		serverInsecure := ldapConfig.ServerInsecure
		lookupBindDN := *ldapConfig.LookupBindDn
		lookupBindPassword := ldapConfig.LookupBindPassword
		userDNSearchBaseDN := ldapConfig.UserDnSearchBaseDn
		userDNSearchFilter := ldapConfig.UserDnSearchFilter
		groupSearchBaseDN := ldapConfig.GroupSearchBaseDn
		groupSearchFilter := ldapConfig.GroupSearchFilter
		serverStartTLS := ldapConfig.ServerStartTLS
		// LDAP Server
		tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_ADDR"] = serverAddress
		if tlsSkipVerify {
			tenantConfiguration["MINIO_IDENTITY_LDAP_TLS_SKIP_VERIFY"] = "on"
		}
		if serverInsecure {
			tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_INSECURE"] = "on"
		}
		if serverStartTLS {
			tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_STARTTLS"] = "on"
		}
		// LDAP Lookup
		tenantConfiguration["MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN"] = lookupBindDN
		tenantConfiguration["MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD"] = lookupBindPassword
		// LDAP User DN
		tenantConfiguration["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN"] = userDNSearchBaseDN
		tenantConfiguration["MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER"] = userDNSearchFilter
		// LDAP Group
		tenantConfiguration["MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN"] = groupSearchBaseDN
		tenantConfiguration["MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER"] = groupSearchFilter
	} else {
		// reset active directory configuration fields
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_SERVER_INSECURE")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_SERVER_STARTTLS")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_TLS_SKIP_VERIFY")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_SERVER_ADDR")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN")
		delete(tenantConfiguration, "MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER")
	}
	// write tenant configuration to secret that contains config.env
	tenantConfigurationName := fmt.Sprintf("%s-env-configuration", tenant.Name)
	_, err = createOrReplaceSecrets(ctx, client, tenant.Namespace, []tenantSecret{
		{
			Name: tenantConfigurationName,
			Content: map[string][]byte{
				"config.env": []byte(GenerateTenantConfigurationFile(tenantConfiguration)),
			},
		},
	}, tenant.Name)
	if err != nil {
		return err
	}
	tenant.Spec.Configuration = &corev1.LocalObjectReference{Name: tenantConfigurationName}
	tenant.EnsureDefaults()
	// update tenant CRD
	_, err = operatorClient.TenantUpdate(ctx, tenant, metav1.UpdateOptions{})
	return err
}

func listTenants(ctx context.Context, operatorClient OperatorClientI, namespace string, limit *int32) (*models.ListTenantsResponse, error) {
	listOpts := metav1.ListOptions{
		Limit: 10,
	}

	if limit != nil {
		listOpts.Limit = int64(*limit)
	}

	minTenants, err := operatorClient.TenantList(ctx, namespace, listOpts)
	if err != nil {
		return nil, err
	}

	var tenants []*models.TenantList

	for _, tenant := range minTenants.Items {
		var totalSize int64
		var instanceCount int64
		var volumeCount int64
		for _, pool := range tenant.Spec.Pools {
			instanceCount += int64(pool.Servers)
			volumeCount += int64(pool.Servers * pool.VolumesPerServer)
			if pool.VolumeClaimTemplate != nil {
				poolSize := int64(pool.VolumesPerServer) * int64(pool.Servers) * pool.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value()
				totalSize += poolSize
			}
		}

		var deletion string
		if tenant.ObjectMeta.DeletionTimestamp != nil {
			deletion = tenant.ObjectMeta.DeletionTimestamp.Format(time.RFC3339)
		}

		var tiers []*models.TenantTierElement

		for _, tier := range tenant.Status.Usage.Tiers {
			tierItem := &models.TenantTierElement{
				Name: tier.Name,
				Type: tier.Type,
				Size: tier.TotalSize,
			}

			tiers = append(tiers, tierItem)
		}

		var domains models.DomainsConfiguration

		if tenant.Spec.Features != nil && tenant.Spec.Features.Domains != nil {
			domains = models.DomainsConfiguration{
				Console: tenant.Spec.Features.Domains.Console,
				Minio:   tenant.Spec.Features.Domains.Minio,
			}
		}

		tenants = append(tenants, &models.TenantList{
			CreationDate:     tenant.ObjectMeta.CreationTimestamp.Format(time.RFC3339),
			DeletionDate:     deletion,
			Name:             tenant.ObjectMeta.Name,
			PoolCount:        int64(len(tenant.Spec.Pools)),
			InstanceCount:    instanceCount,
			VolumeCount:      volumeCount,
			CurrentState:     tenant.Status.CurrentState,
			Namespace:        tenant.ObjectMeta.Namespace,
			TotalSize:        totalSize,
			HealthStatus:     string(tenant.Status.HealthStatus),
			CapacityRaw:      tenant.Status.Usage.RawCapacity,
			CapacityRawUsage: tenant.Status.Usage.RawUsage,
			Capacity:         tenant.Status.Usage.Capacity,
			CapacityUsage:    tenant.Status.Usage.Usage,
			Tiers:            tiers,
			Domains:          &domains,
		})
	}

	return &models.ListTenantsResponse{
		Tenants: tenants,
		Total:   int64(len(tenants)),
	}, nil
}

func getListAllTenantsResponse(session *models.Principal, params operator_api.ListAllTenantsParams) (*models.ListTenantsResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	listT, err := listTenants(ctx, opClient, "", params.Limit)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return listT, nil
}

// getListTenantsResponse list tenants by namespace
func getListTenantsResponse(session *models.Principal, params operator_api.ListTenantsParams) (*models.ListTenantsResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	listT, err := listTenants(ctx, opClient, params.Namespace, params.Limit)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return listT, nil
}

// setImageRegistry creates a secret to store the private registry credentials, if one exist it updates the existing one
// returns the name of the secret created/updated
func setImageRegistry(ctx context.Context, req *models.ImageRegistry, clientset K8sClientI, namespace, tenantName string) (string, error) {
	if req == nil || req.Registry == nil || req.Username == nil || req.Password == nil {
		return "", nil
	}

	credentials := make(map[string]imageRegistryCredentials)
	// username:password encoded
	authData := []byte(fmt.Sprintf("%s:%s", *req.Username, *req.Password))
	authStr := base64.StdEncoding.EncodeToString(authData)

	credentials[*req.Registry] = imageRegistryCredentials{
		Username: *req.Username,
		Password: *req.Password,
		Auth:     authStr,
	}
	imRegistry := imageRegistry{
		Auths: credentials,
	}
	imRegistryJSON, err := json.Marshal(imRegistry)
	if err != nil {
		return "", err
	}

	pullSecretName := fmt.Sprintf("%s-regcred", tenantName)
	secretCredentials := map[string][]byte{
		corev1.DockerConfigJsonKey: []byte(string(imRegistryJSON)),
	}
	// Get or Create secret if it doesn't exist
	currentSecret, err := clientset.getSecret(ctx, namespace, pullSecretName, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			instanceSecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: pullSecretName,
					Labels: map[string]string{
						miniov2.TenantLabel: tenantName,
					},
				},
				Data: secretCredentials,
				Type: corev1.SecretTypeDockerConfigJson,
			}
			_, err = clientset.createSecret(ctx, namespace, &instanceSecret, metav1.CreateOptions{})
			if err != nil {
				return "", err
			}
			return pullSecretName, nil
		}
		return "", err
	}
	currentSecret.Data = secretCredentials
	_, err = clientset.updateSecret(ctx, namespace, currentSecret, metav1.UpdateOptions{})
	if err != nil {
		return "", err
	}
	return pullSecretName, nil
}

func getUpdateTenantResponse(session *models.Principal, params operator_api.UpdateTenantParams) *models.Error {
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
	k8sClient := &k8sClient{
		client: clientSet,
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	client := GetConsoleHTTPClient("")
	client.Timeout = 4 * time.Second
	httpC := &utils2.Client{
		Client: client,
	}
	if err := updateTenantAction(ctx, opClient, k8sClient, httpC, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant"))
	}
	return nil
}

// addTenantPool creates a pool to a defined tenant
func addTenantPool(ctx context.Context, operatorClient OperatorClientI, params operator_api.TenantAddPoolParams) error {
	tenant, err := operatorClient.TenantGet(ctx, params.Namespace, params.Tenant, metav1.GetOptions{})
	if err != nil {
		return err
	}

	poolParams := params.Body
	pool, err := parseTenantPoolRequest(poolParams)
	if err != nil {
		return err
	}
	tenant.Spec.Pools = append(tenant.Spec.Pools, *pool)
	payloadBytes, err := json.Marshal(tenant)
	if err != nil {
		return err
	}

	_, err = operatorClient.TenantPatch(ctx, tenant.Namespace, tenant.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

// getTenantUsageResponse returns the usage of a tenant
func getTenantUsageResponse(session *models.Principal, params operator_api.GetTenantUsageParams) (*models.TenantUsage, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}
	k8sClient := &k8sClient{
		client: clientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	return getTenantUsage(ctx, minTenant, k8sClient)
}

func getTenantUsage(ctx context.Context, minTenant *miniov2.Tenant, k8sClient K8sClientI) (*models.TenantUsage, *models.Error) {
	minTenant.EnsureDefaults()

	svcURL := GetTenantServiceURL(minTenant)
	// getTenantAdminClient will use all certificates under ~/.console/certs/CAs to trust the TLS connections with MinIO tenants
	mAdmin, err := getTenantAdminClient(
		ctx,
		k8sClient,
		minTenant,
		svcURL,
	)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	return _getTenantUsage(ctx, AdminClient{Client: mAdmin})
}

func _getTenantUsage(ctx context.Context, adminClient MinioAdmin) (*models.TenantUsage, *models.Error) {
	adminInfo, err := GetAdminInfo(ctx, adminClient)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	return &models.TenantUsage{Used: adminInfo.Usage, DiskUsed: adminInfo.DisksUsage}, nil
}

// parseTenantPoolRequest parse pool request and returns the equivalent
// miniov2.Pool object
func parseTenantPoolRequest(poolParams *models.Pool) (*miniov2.Pool, error) {
	if poolParams.VolumeConfiguration == nil {
		return nil, errors.New("a volume configuration must be specified")
	}

	if poolParams.VolumeConfiguration.Size == nil || *poolParams.VolumeConfiguration.Size <= int64(0) {
		return nil, errors.New("volume size must be greater than 0")
	}

	if poolParams.Servers == nil || *poolParams.Servers <= 0 {
		return nil, errors.New("number of servers must be greater than 0")
	}

	if poolParams.VolumesPerServer == nil || *poolParams.VolumesPerServer <= 0 {
		return nil, errors.New("number of volumes per server must be greater than 0")
	}

	volumeSize := resource.NewQuantity(*poolParams.VolumeConfiguration.Size, resource.DecimalExponent)
	volTemp := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: *volumeSize,
			},
		},
	}
	if poolParams.VolumeConfiguration.StorageClassName != "" {
		volTemp.StorageClassName = &poolParams.VolumeConfiguration.StorageClassName
	}

	// parse resources' requests
	resourcesRequests := make(corev1.ResourceList)
	resourcesLimits := make(corev1.ResourceList)
	if poolParams.Resources != nil {
		for key, val := range poolParams.Resources.Requests {
			resourcesRequests[corev1.ResourceName(key)] = *resource.NewQuantity(val, resource.BinarySI)
		}
		for key, val := range poolParams.Resources.Limits {
			resourcesLimits[corev1.ResourceName(key)] = *resource.NewQuantity(val, resource.BinarySI)
		}
	}

	// parse Node Affinity
	nodeSelectorTerms := []corev1.NodeSelectorTerm{}
	preferredSchedulingTerm := []corev1.PreferredSchedulingTerm{}
	if poolParams.Affinity != nil && poolParams.Affinity.NodeAffinity != nil {
		if poolParams.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			for _, elem := range poolParams.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				term := parseModelsNodeSelectorTerm(elem)
				nodeSelectorTerms = append(nodeSelectorTerms, term)
			}
		}
		for _, elem := range poolParams.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			pst := corev1.PreferredSchedulingTerm{
				Weight:     *elem.Weight,
				Preference: parseModelsNodeSelectorTerm(elem.Preference),
			}
			preferredSchedulingTerm = append(preferredSchedulingTerm, pst)
		}
	}
	var nodeAffinity *corev1.NodeAffinity
	if len(nodeSelectorTerms) > 0 || len(preferredSchedulingTerm) > 0 {
		nodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: mergeNodeSelectorTerms(nodeSelectorTerms),
			},
			PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTerm,
		}
	}

	// parse Pod Affinity
	podAffinityTerms := []corev1.PodAffinityTerm{}
	weightedPodAffinityTerms := []corev1.WeightedPodAffinityTerm{}
	if poolParams.Affinity != nil && poolParams.Affinity.PodAffinity != nil {
		for _, elem := range poolParams.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			podAffinityTerms = append(podAffinityTerms, parseModelPodAffinityTerm(elem))
		}
		for _, elem := range poolParams.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			wAffinityTerm := corev1.WeightedPodAffinityTerm{
				Weight:          *elem.Weight,
				PodAffinityTerm: parseModelPodAffinityTerm(elem.PodAffinityTerm),
			}
			weightedPodAffinityTerms = append(weightedPodAffinityTerms, wAffinityTerm)
		}
	}
	var podAffinity *corev1.PodAffinity
	if len(podAffinityTerms) > 0 || len(weightedPodAffinityTerms) > 0 {
		podAffinity = &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  podAffinityTerms,
			PreferredDuringSchedulingIgnoredDuringExecution: weightedPodAffinityTerms,
		}
	}

	// parse Pod Anti Affinity
	podAntiAffinityTerms := []corev1.PodAffinityTerm{}
	weightedPodAntiAffinityTerms := []corev1.WeightedPodAffinityTerm{}
	if poolParams.Affinity != nil && poolParams.Affinity.PodAntiAffinity != nil {
		for _, elem := range poolParams.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			podAntiAffinityTerms = append(podAntiAffinityTerms, parseModelPodAffinityTerm(elem))
		}
		for _, elem := range poolParams.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			wAffinityTerm := corev1.WeightedPodAffinityTerm{
				Weight:          *elem.Weight,
				PodAffinityTerm: parseModelPodAffinityTerm(elem.PodAffinityTerm),
			}
			weightedPodAntiAffinityTerms = append(weightedPodAntiAffinityTerms, wAffinityTerm)
		}
	}
	var podAntiAffinity *corev1.PodAntiAffinity
	if len(podAntiAffinityTerms) > 0 || len(weightedPodAntiAffinityTerms) > 0 {
		podAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  podAntiAffinityTerms,
			PreferredDuringSchedulingIgnoredDuringExecution: weightedPodAntiAffinityTerms,
		}
	}

	var affinity *corev1.Affinity
	if nodeAffinity != nil || podAffinity != nil || podAntiAffinity != nil {
		affinity = &corev1.Affinity{
			NodeAffinity:    nodeAffinity,
			PodAffinity:     podAffinity,
			PodAntiAffinity: podAntiAffinity,
		}
	}

	// parse tolerations
	tolerations := []corev1.Toleration{}
	for _, elem := range poolParams.Tolerations {
		var tolerationSeconds *int64
		if elem.TolerationSeconds != nil {
			// elem.TolerationSeconds.Seconds is allowed to be nil
			tolerationSeconds = elem.TolerationSeconds.Seconds
		}

		toleration := corev1.Toleration{
			Key:               elem.Key,
			Operator:          corev1.TolerationOperator(elem.Operator),
			Value:             elem.Value,
			Effect:            corev1.TaintEffect(elem.Effect),
			TolerationSeconds: tolerationSeconds,
		}
		tolerations = append(tolerations, toleration)
	}

	// Pass annotations to the volume
	vct := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "data",
			Labels:      poolParams.VolumeConfiguration.Labels,
			Annotations: poolParams.VolumeConfiguration.Annotations,
		},
		Spec: volTemp,
	}

	pool := &miniov2.Pool{
		Name:                poolParams.Name,
		Servers:             int32(*poolParams.Servers),
		VolumesPerServer:    *poolParams.VolumesPerServer,
		VolumeClaimTemplate: vct,
		Resources: corev1.ResourceRequirements{
			Requests: resourcesRequests,
			Limits:   resourcesLimits,
		},
		NodeSelector:     poolParams.NodeSelector,
		Affinity:         affinity,
		Tolerations:      tolerations,
		RuntimeClassName: &poolParams.RuntimeClassName,
	}
	// if security context for Tenant is present, configure it.
	if poolParams.SecurityContext != nil {
		sc, err := convertModelSCToK8sSC(poolParams.SecurityContext)
		if err != nil {
			return nil, err
		}
		pool.SecurityContext = sc
	}
	return pool, nil
}

// mergeNodeSelectorTerms is for matchExpressions merge, when matchExpressions/matchFields have the same Key+Operator , we merge values.
// When matchExpressions/matchFields have the same Key+Operator, Set it 'And'
func mergeNodeSelectorTerms(terms []corev1.NodeSelectorTerm) []corev1.NodeSelectorTerm {
	mergedTerms := []corev1.NodeSelectorTerm{}
	for _, term := range terms {
		nodeSelectorTerm := corev1.NodeSelectorTerm{}
		nodeSelectorMatchExpressionsMap := map[string]*corev1.NodeSelectorRequirement{}
		for _, exp := range term.MatchExpressions {
			if selector, ok := nodeSelectorMatchExpressionsMap[fmt.Sprintf("%s-%s", exp.Key, exp.Operator)]; ok {
				selector.Values = append(selector.Values, exp.Values...)
			} else {
				nodeSelectorMatchExpressionsMap[fmt.Sprintf("%s-%s", exp.Key, exp.Operator)] = &corev1.NodeSelectorRequirement{
					Key:      exp.Key,
					Operator: exp.Operator,
					Values:   exp.Values,
				}
			}
		}
		nodeSelectorMatchMatchFieldsMap := map[string]*corev1.NodeSelectorRequirement{}
		for _, field := range term.MatchFields {
			if selector, ok := nodeSelectorMatchMatchFieldsMap[fmt.Sprintf("%s-%s", field.Key, field.Operator)]; ok {
				selector.Values = append(selector.Values, field.Values...)
			} else {
				nodeSelectorMatchMatchFieldsMap[fmt.Sprintf("%s-%s", field.Key, field.Operator)] = &corev1.NodeSelectorRequirement{
					Key:      field.Key,
					Operator: field.Operator,
					Values:   field.Values,
				}
			}
		}
		for _, exp := range nodeSelectorMatchExpressionsMap {
			nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, *exp)
		}
		for _, field := range nodeSelectorMatchMatchFieldsMap {
			nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchFields, *field)
		}
		if len(nodeSelectorMatchExpressionsMap) != 0 || len(nodeSelectorMatchMatchFieldsMap) != 0 {
			mergedTerms = append(mergedTerms, nodeSelectorTerm)
		}
	}
	return mergedTerms
}

func parseModelPodAffinityTerm(term *models.PodAffinityTerm) corev1.PodAffinityTerm {
	labelMatchExpressions := []metav1.LabelSelectorRequirement{}
	for _, exp := range term.LabelSelector.MatchExpressions {
		labelSelectorReq := metav1.LabelSelectorRequirement{
			Key:      *exp.Key,
			Operator: metav1.LabelSelectorOperator(*exp.Operator),
			Values:   exp.Values,
		}
		labelMatchExpressions = append(labelMatchExpressions, labelSelectorReq)
	}

	podAffinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchExpressions: labelMatchExpressions,
			MatchLabels:      term.LabelSelector.MatchLabels,
		},
		Namespaces:  term.Namespaces,
		TopologyKey: *term.TopologyKey,
	}
	return podAffinityTerm
}

func parseModelsNodeSelectorTerm(elem *models.NodeSelectorTerm) corev1.NodeSelectorTerm {
	var term corev1.NodeSelectorTerm
	for _, matchExpression := range elem.MatchExpressions {
		matchExp := corev1.NodeSelectorRequirement{
			Key:      *matchExpression.Key,
			Operator: corev1.NodeSelectorOperator(*matchExpression.Operator),
			Values:   matchExpression.Values,
		}
		term.MatchExpressions = append(term.MatchExpressions, matchExp)
	}
	for _, matchField := range elem.MatchFields {
		matchF := corev1.NodeSelectorRequirement{
			Key:      *matchField.Key,
			Operator: corev1.NodeSelectorOperator(*matchField.Operator),
			Values:   matchField.Values,
		}
		term.MatchFields = append(term.MatchFields, matchF)
	}
	return term
}

// parseTenantPool miniov2 pool object and returns the equivalent
// models.Pool object
func parseTenantPool(pool *miniov2.Pool) *models.Pool {
	var size *int64
	var storageClassName string
	if pool.VolumeClaimTemplate != nil {
		size = swag.Int64(pool.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value())
		if pool.VolumeClaimTemplate.Spec.StorageClassName != nil {
			storageClassName = *pool.VolumeClaimTemplate.Spec.StorageClassName
		}
	}

	// parse resources' requests
	var resources *models.PoolResources
	resourcesRequests := make(map[string]int64)
	resourcesLimits := make(map[string]int64)
	for key, val := range pool.Resources.Requests {
		resourcesRequests[key.String()] = val.Value()
	}
	for key, val := range pool.Resources.Limits {
		resourcesLimits[key.String()] = val.Value()
	}
	if len(resourcesRequests) > 0 || len(resourcesLimits) > 0 {
		resources = &models.PoolResources{
			Limits:   resourcesLimits,
			Requests: resourcesRequests,
		}
	}

	// parse Node Affinity
	nodeSelectorTerms := []*models.NodeSelectorTerm{}
	preferredSchedulingTerm := []*models.PoolAffinityNodeAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{}

	if pool.Affinity != nil && pool.Affinity.NodeAffinity != nil {
		if pool.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			for _, elem := range pool.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				term := parseNodeSelectorTerm(&elem)
				nodeSelectorTerms = append(nodeSelectorTerms, term)
			}
		}
		for _, elem := range pool.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			pst := &models.PoolAffinityNodeAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{
				Weight:     swag.Int32(elem.Weight),
				Preference: parseNodeSelectorTerm(&elem.Preference),
			}
			preferredSchedulingTerm = append(preferredSchedulingTerm, pst)
		}
	}

	var nodeAffinity *models.PoolAffinityNodeAffinity
	if len(nodeSelectorTerms) > 0 || len(preferredSchedulingTerm) > 0 {
		nodeAffinity = &models.PoolAffinityNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &models.PoolAffinityNodeAffinityRequiredDuringSchedulingIgnoredDuringExecution{
				NodeSelectorTerms: nodeSelectorTerms,
			},
			PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTerm,
		}
	}

	// parse Pod Affinity
	podAffinityTerms := []*models.PodAffinityTerm{}
	weightedPodAffinityTerms := []*models.PoolAffinityPodAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{}

	if pool.Affinity != nil && pool.Affinity.PodAffinity != nil {
		for _, elem := range pool.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			podAffinityTerms = append(podAffinityTerms, parsePodAffinityTerm(&elem))
		}
		for _, elem := range pool.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			wAffinityTerm := &models.PoolAffinityPodAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{
				Weight:          swag.Int32(elem.Weight),
				PodAffinityTerm: parsePodAffinityTerm(&elem.PodAffinityTerm),
			}
			weightedPodAffinityTerms = append(weightedPodAffinityTerms, wAffinityTerm)
		}
	}
	var podAffinity *models.PoolAffinityPodAffinity
	if len(podAffinityTerms) > 0 || len(weightedPodAffinityTerms) > 0 {
		podAffinity = &models.PoolAffinityPodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  podAffinityTerms,
			PreferredDuringSchedulingIgnoredDuringExecution: weightedPodAffinityTerms,
		}
	}

	// parse Pod Anti Affinity
	podAntiAffinityTerms := []*models.PodAffinityTerm{}
	weightedPodAntiAffinityTerms := []*models.PoolAffinityPodAntiAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{}

	if pool.Affinity != nil && pool.Affinity.PodAntiAffinity != nil {
		for _, elem := range pool.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			podAntiAffinityTerms = append(podAntiAffinityTerms, parsePodAffinityTerm(&elem))
		}
		for _, elem := range pool.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			wAffinityTerm := &models.PoolAffinityPodAntiAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{
				Weight:          swag.Int32(elem.Weight),
				PodAffinityTerm: parsePodAffinityTerm(&elem.PodAffinityTerm),
			}
			weightedPodAntiAffinityTerms = append(weightedPodAntiAffinityTerms, wAffinityTerm)
		}
	}

	var podAntiAffinity *models.PoolAffinityPodAntiAffinity
	if len(podAntiAffinityTerms) > 0 || len(weightedPodAntiAffinityTerms) > 0 {
		podAntiAffinity = &models.PoolAffinityPodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  podAntiAffinityTerms,
			PreferredDuringSchedulingIgnoredDuringExecution: weightedPodAntiAffinityTerms,
		}
	}

	// build affinity object
	var affinity *models.PoolAffinity
	if nodeAffinity != nil || podAffinity != nil || podAntiAffinity != nil {
		affinity = &models.PoolAffinity{
			NodeAffinity:    nodeAffinity,
			PodAffinity:     podAffinity,
			PodAntiAffinity: podAntiAffinity,
		}
	}

	// parse tolerations
	var tolerations models.PoolTolerations
	for _, elem := range pool.Tolerations {
		var tolerationSecs *models.PoolTolerationSeconds
		if elem.TolerationSeconds != nil {
			tolerationSecs = &models.PoolTolerationSeconds{
				Seconds: elem.TolerationSeconds,
			}
		}
		toleration := &models.PoolTolerationsItems0{
			Key:               elem.Key,
			Operator:          string(elem.Operator),
			Value:             elem.Value,
			Effect:            string(elem.Effect),
			TolerationSeconds: tolerationSecs,
		}
		tolerations = append(tolerations, toleration)
	}

	var securityContext models.SecurityContext

	if pool.SecurityContext != nil {
		var fsGroup string
		var runAsGroup string
		var runAsUser string
		var fsGroupChangePolicy string
		if pool.SecurityContext.FSGroup != nil {
			fsGroup = strconv.Itoa(int(*pool.SecurityContext.FSGroup))
		}
		if pool.SecurityContext.RunAsGroup != nil {
			runAsGroup = strconv.Itoa(int(*pool.SecurityContext.RunAsGroup))
		}
		if pool.SecurityContext.RunAsUser != nil {
			runAsUser = strconv.Itoa(int(*pool.SecurityContext.RunAsUser))
		}
		if pool.SecurityContext.FSGroupChangePolicy != nil {
			fsGroupChangePolicy = string(*pool.SecurityContext.FSGroupChangePolicy)
		}
		securityContext = models.SecurityContext{
			FsGroup:             fsGroup,
			RunAsGroup:          &runAsGroup,
			RunAsNonRoot:        pool.SecurityContext.RunAsNonRoot,
			RunAsUser:           &runAsUser,
			FsGroupChangePolicy: fsGroupChangePolicy,
		}
	}

	var runtimeClassName string
	if pool.RuntimeClassName != nil {
		runtimeClassName = *pool.RuntimeClassName
	}

	poolModel := &models.Pool{
		Name:             pool.Name,
		Servers:          swag.Int64(int64(pool.Servers)),
		VolumesPerServer: swag.Int32(pool.VolumesPerServer),
		VolumeConfiguration: &models.PoolVolumeConfiguration{
			Size:             size,
			StorageClassName: storageClassName,
		},
		NodeSelector:     pool.NodeSelector,
		Resources:        resources,
		Affinity:         affinity,
		Tolerations:      tolerations,
		SecurityContext:  &securityContext,
		RuntimeClassName: runtimeClassName,
	}
	return poolModel
}

func parsePodAffinityTerm(term *corev1.PodAffinityTerm) *models.PodAffinityTerm {
	labelMatchExpressions := []*models.PodAffinityTermLabelSelectorMatchExpressionsItems0{}
	for _, exp := range term.LabelSelector.MatchExpressions {
		labelSelectorReq := &models.PodAffinityTermLabelSelectorMatchExpressionsItems0{
			Key:      swag.String(exp.Key),
			Operator: swag.String(string(exp.Operator)),
			Values:   exp.Values,
		}
		labelMatchExpressions = append(labelMatchExpressions, labelSelectorReq)
	}

	podAffinityTerm := &models.PodAffinityTerm{
		LabelSelector: &models.PodAffinityTermLabelSelector{
			MatchExpressions: labelMatchExpressions,
			MatchLabels:      term.LabelSelector.MatchLabels,
		},
		Namespaces:  term.Namespaces,
		TopologyKey: swag.String(term.TopologyKey),
	}
	return podAffinityTerm
}

func parseNodeSelectorTerm(term *corev1.NodeSelectorTerm) *models.NodeSelectorTerm {
	var t models.NodeSelectorTerm
	for _, matchExpression := range term.MatchExpressions {
		matchExp := &models.NodeSelectorTermMatchExpressionsItems0{
			Key:      swag.String(matchExpression.Key),
			Operator: swag.String(string(matchExpression.Operator)),
			Values:   matchExpression.Values,
		}
		t.MatchExpressions = append(t.MatchExpressions, matchExp)
	}
	for _, matchField := range term.MatchFields {
		matchF := &models.NodeSelectorTermMatchFieldsItems0{
			Key:      swag.String(matchField.Key),
			Operator: swag.String(string(matchField.Operator)),
			Values:   matchField.Values,
		}
		t.MatchFields = append(t.MatchFields, matchF)
	}
	return &t
}
