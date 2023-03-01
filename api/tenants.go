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
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	utils2 "github.com/minio/operator/pkg/http"

	"github.com/minio/madmin-go/v2"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/minio/operator/pkg/auth/utils"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
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

	// Tenant Configuration details
	// Tenant Security details
	api.OperatorAPITenantConfigurationHandler = operator_api.TenantConfigurationHandlerFunc(func(params operator_api.TenantConfigurationParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantConfigurationResponse(session, params)
		if err != nil {
			return operator_api.NewTenantConfigurationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantConfigurationOK().WithPayload(resp)
	})
	// Update Tenant Configuration
	api.OperatorAPIUpdateTenantConfigurationHandler = operator_api.UpdateTenantConfigurationHandlerFunc(func(params operator_api.UpdateTenantConfigurationParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantConfigurationResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantConfigurationDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantConfigurationNoContent()
	})

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

	// Set Tenant Administrators
	api.OperatorAPISetTenantAdministratorsHandler = operator_api.SetTenantAdministratorsHandlerFunc(func(params operator_api.SetTenantAdministratorsParams, session *models.Principal) middleware.Responder {
		err := getSetTenantAdministratorsResponse(session, params)
		if err != nil {
			return operator_api.NewSetTenantAdministratorsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewSetTenantAdministratorsNoContent()
	})

	// Tenant identity provider details
	api.OperatorAPITenantIdentityProviderHandler = operator_api.TenantIdentityProviderHandlerFunc(func(params operator_api.TenantIdentityProviderParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantIdentityProviderResponse(session, params)
		if err != nil {
			return operator_api.NewTenantIdentityProviderDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantIdentityProviderOK().WithPayload(resp)
	})

	// Update Tenant identity provider configuration
	api.OperatorAPIUpdateTenantIdentityProviderHandler = operator_api.UpdateTenantIdentityProviderHandlerFunc(func(params operator_api.UpdateTenantIdentityProviderParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantIdentityProviderResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantIdentityProviderDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantIdentityProviderNoContent()
	})

	// Delete Tenant
	api.OperatorAPIDeleteTenantHandler = operator_api.DeleteTenantHandlerFunc(func(params operator_api.DeleteTenantParams, session *models.Principal) middleware.Responder {
		err := getDeleteTenantResponse(session, params)
		if err != nil {
			return operator_api.NewDeleteTenantDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDeleteTenantNoContent()
	})

	// Delete Pod
	api.OperatorAPIDeletePodHandler = operator_api.DeletePodHandlerFunc(func(params operator_api.DeletePodParams, session *models.Principal) middleware.Responder {
		err := getDeletePodResponse(session, params)
		if err != nil {
			return operator_api.NewDeletePodDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDeletePodNoContent()
	})

	// Update Tenant
	api.OperatorAPIUpdateTenantHandler = operator_api.UpdateTenantHandlerFunc(func(params operator_api.UpdateTenantParams, session *models.Principal) middleware.Responder {
		err := getUpdateTenantResponse(session, params)
		if err != nil {
			return operator_api.NewUpdateTenantDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantCreated()
	})

	// Add Tenant Pools
	api.OperatorAPITenantAddPoolHandler = operator_api.TenantAddPoolHandlerFunc(func(params operator_api.TenantAddPoolParams, session *models.Principal) middleware.Responder {
		err := getTenantAddPoolResponse(session, params)
		if err != nil {
			return operator_api.NewTenantAddPoolDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantAddPoolCreated()
	})

	// Get Tenant Usage
	api.OperatorAPIGetTenantUsageHandler = operator_api.GetTenantUsageHandlerFunc(func(params operator_api.GetTenantUsageParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantUsageResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantUsageDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantUsageOK().WithPayload(payload)
	})

	registerTenantLogsHandlers(api)

	api.OperatorAPIGetTenantPodsHandler = operator_api.GetTenantPodsHandlerFunc(func(params operator_api.GetTenantPodsParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantPodsResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantPodsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantPodsOK().WithPayload(payload)
	})

	api.OperatorAPIGetPodLogsHandler = operator_api.GetPodLogsHandlerFunc(func(params operator_api.GetPodLogsParams, session *models.Principal) middleware.Responder {
		payload, err := getPodLogsResponse(session, params)
		if err != nil {
			return operator_api.NewGetPodLogsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetPodLogsOK().WithPayload(payload)
	})

	api.OperatorAPIGetPodEventsHandler = operator_api.GetPodEventsHandlerFunc(func(params operator_api.GetPodEventsParams, session *models.Principal) middleware.Responder {
		payload, err := getPodEventsResponse(session, params)
		if err != nil {
			return operator_api.NewGetPodEventsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetPodEventsOK().WithPayload(payload)
	})

	api.OperatorAPIDescribePodHandler = operator_api.DescribePodHandlerFunc(func(params operator_api.DescribePodParams, session *models.Principal) middleware.Responder {
		payload, err := getDescribePodResponse(session, params)
		if err != nil {
			return operator_api.NewDescribePodDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDescribePodOK().WithPayload(payload)
	})

	// Get tenant monitoring info
	api.OperatorAPIGetTenantMonitoringHandler = operator_api.GetTenantMonitoringHandlerFunc(func(params operator_api.GetTenantMonitoringParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantMonitoringResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantMonitoringDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantMonitoringOK().WithPayload(payload)
	})
	// Set configuration fields for Prometheus monitoring on a tenant
	api.OperatorAPISetTenantMonitoringHandler = operator_api.SetTenantMonitoringHandlerFunc(func(params operator_api.SetTenantMonitoringParams, session *models.Principal) middleware.Responder {
		_, err := setTenantMonitoringResponse(session, params)
		if err != nil {
			return operator_api.NewSetTenantMonitoringDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewSetTenantMonitoringCreated()
	})

	// Update Tenant Pools
	api.OperatorAPITenantUpdatePoolsHandler = operator_api.TenantUpdatePoolsHandlerFunc(func(params operator_api.TenantUpdatePoolsParams, session *models.Principal) middleware.Responder {
		resp, err := getTenantUpdatePoolResponse(session, params)
		if err != nil {
			return operator_api.NewTenantUpdatePoolsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantUpdatePoolsOK().WithPayload(resp)
	})

	// Update Tenant Certificates
	api.OperatorAPITenantUpdateCertificateHandler = operator_api.TenantUpdateCertificateHandlerFunc(func(params operator_api.TenantUpdateCertificateParams, session *models.Principal) middleware.Responder {
		err := getTenantUpdateCertificatesResponse(session, params)
		if err != nil {
			return operator_api.NewTenantUpdateCertificateDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantUpdateCertificateCreated()
	})

	// Update Tenant Encryption Configuration
	api.OperatorAPITenantUpdateEncryptionHandler = operator_api.TenantUpdateEncryptionHandlerFunc(func(params operator_api.TenantUpdateEncryptionParams, session *models.Principal) middleware.Responder {
		err := getTenantUpdateEncryptionResponse(session, params)
		if err != nil {
			return operator_api.NewTenantUpdateEncryptionDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantUpdateEncryptionCreated()
	})

	// Delete tenant Encryption Configuration
	api.OperatorAPITenantDeleteEncryptionHandler = operator_api.TenantDeleteEncryptionHandlerFunc(func(params operator_api.TenantDeleteEncryptionParams, session *models.Principal) middleware.Responder {
		err := getTenantDeleteEncryptionResponse(session, params)
		if err != nil {
			return operator_api.NewTenantDeleteEncryptionDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantDeleteEncryptionNoContent()
	})

	// Get Tenant Encryption Configuration
	api.OperatorAPITenantEncryptionInfoHandler = operator_api.TenantEncryptionInfoHandlerFunc(func(params operator_api.TenantEncryptionInfoParams, session *models.Principal) middleware.Responder {
		configuration, err := getTenantEncryptionInfoResponse(session, params)
		if err != nil {
			return operator_api.NewTenantEncryptionInfoDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewTenantEncryptionInfoOK().WithPayload(configuration)
	})

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
	// Get Tenant Events
	api.OperatorAPIGetTenantEventsHandler = operator_api.GetTenantEventsHandlerFunc(func(params operator_api.GetTenantEventsParams, principal *models.Principal) middleware.Responder {
		payload, err := getTenantEventsResponse(principal, params)
		if err != nil {
			return operator_api.NewGetTenantEventsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantEventsOK().WithPayload(payload)
	})
	// Update Tenant Domains
	api.OperatorAPIUpdateTenantDomainsHandler = operator_api.UpdateTenantDomainsHandlerFunc(func(params operator_api.UpdateTenantDomainsParams, principal *models.Principal) middleware.Responder {
		err := getUpdateDomainsResponse(principal, params)
		if err != nil {
			return operator_api.NewUpdateTenantDomainsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewUpdateTenantDomainsNoContent()
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
		// delete postgres PVCs

		logOpts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", miniov2.LogDBInstanceLabel, tenant.LogStatefulsetName()),
		}
		err := clientset.PersistentVolumeClaims(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, logOpts)
		if err != nil {
			return err
		}

		// delete prometheus PVCs

		promOpts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", miniov2.PrometheusInstanceLabel, tenant.PrometheusStatefulsetName()),
		}

		if err := clientset.PersistentVolumeClaims(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, promOpts); err != nil {
			return err
		}

		// delete all tenant's secrets only if deletePvcs = true
		return clientset.Secrets(tenant.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, opts)
	}
	return nil
}

// getDeletePodResponse gets the output of deleting a minio instance
func getDeletePodResponse(session *models.Principal, params operator_api.DeletePodParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("v1.min.io/tenant=%s", params.Tenant),
		FieldSelector: fmt.Sprintf("metadata.name=%s%s", params.Tenant, params.PodName[len(params.Tenant):]),
	}
	if err = clientset.CoreV1().Pods(params.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOpts); err != nil {
		return ErrorWithContext(ctx, err)
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

func isPrometheusEnabled(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	// if one of the following prometheus annotations are not present
	// we consider the tenant as not integrated with prometheus
	if _, ok := annotations[prometheusPath]; !ok {
		return false
	}
	if _, ok := annotations[prometheusPort]; !ok {
		return false
	}
	if _, ok := annotations[prometheusScrape]; !ok {
		return false
	}
	return true
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
		CreationDate:     tenant.ObjectMeta.CreationTimestamp.Format(time.RFC3339),
		DeletionDate:     deletion,
		Name:             tenant.Name,
		TotalSize:        totalSize,
		CurrentState:     tenant.Status.CurrentState,
		Pools:            pools,
		Namespace:        tenant.ObjectMeta.Namespace,
		Image:            tenant.Spec.Image,
		EnablePrometheus: isPrometheusEnabled(tenant.Annotations),
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

func getTenantIdentityProviderResponse(session *models.Principal, params operator_api.TenantIdentityProviderParams) (*models.IdpConfiguration, *models.Error) {
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
	info, err := getTenantIdentityProvider(ctx, &k8sClient, minTenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return info, nil
}

func getUpdateTenantIdentityProviderResponse(session *models.Principal, params operator_api.UpdateTenantIdentityProviderParams) *models.Error {
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
	if err := updateTenantIdentityProvider(ctx, opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant"))
	}
	return nil
}

func getSetTenantAdministratorsResponse(session *models.Principal, params operator_api.SetTenantAdministratorsParams) *models.Error {
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

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	return setTenantAdministrators(ctx, minTenant, k8sClient, params)
}

func setTenantAdministrators(ctx context.Context, minTenant *miniov2.Tenant, k8sClient K8sClientI, params operator_api.SetTenantAdministratorsParams) *models.Error {
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
		return ErrorWithContext(ctx, err)
	}
	// create a minioClient interface implementation
	// defining the client to be used
	adminClient := AdminClient{Client: mAdmin}
	for _, user := range params.Body.UserDNS {
		if err := SetPolicy(ctx, adminClient, "consoleAdmin", user, "user"); err != nil {
			return ErrorWithContext(ctx, err)
		}
	}
	for _, group := range params.Body.GroupDNS {
		if err := SetPolicy(ctx, adminClient, "consoleAdmin", group, "group"); err != nil {
			return ErrorWithContext(ctx, err)
		}
	}
	return nil
}

func getTenantConfigurationResponse(session *models.Principal, params operator_api.TenantConfigurationParams) (*models.TenantConfigurationResponse, *models.Error) {
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
	k8sClient := &k8sClient{
		client: clientSet,
	}
	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return parseTenantConfiguration(ctx, k8sClient, minTenant)
}

func parseTenantConfiguration(ctx context.Context, k8sClient K8sClientI, minTenant *miniov2.Tenant) (*models.TenantConfigurationResponse, *models.Error) {
	tenantConfiguration, err := GetTenantConfiguration(ctx, k8sClient, minTenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	delete(tenantConfiguration, "accesskey")
	delete(tenantConfiguration, "secretkey")
	var envVars []*models.EnvironmentVariable
	for key, value := range tenantConfiguration {
		envVars = append(envVars, &models.EnvironmentVariable{
			Key:   key,
			Value: value,
		})
	}
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Key < envVars[j].Key
	})
	configurationInfo := &models.TenantConfigurationResponse{EnvironmentVariables: envVars}
	return configurationInfo, nil
}

func getUpdateTenantConfigurationResponse(session *models.Principal, params operator_api.UpdateTenantConfigurationParams) *models.Error {
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
	if err := updateTenantConfigurationFile(ctx, opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to update tenant configuration"))
	}
	return nil
}

func updateTenantConfigurationFile(ctx context.Context, operatorClient OperatorClientI, client K8sClientI, namespace string, params operator_api.UpdateTenantConfigurationParams) error {
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

	requestBody := params.Body
	if requestBody == nil {
		return errors.New("missing request body")
	}
	// Patch tenant configuration file with the new values provided by the user
	for _, envVar := range requestBody.EnvironmentVariables {
		if envVar.Key == "" {
			continue
		}
		tenantConfiguration[envVar.Key] = envVar.Value
	}
	// Remove existing values from configuration file
	for _, keyToBeDeleted := range requestBody.KeysToBeDeleted {
		delete(tenantConfiguration, keyToBeDeleted)
	}

	if !tenant.HasConfigurationSecret() {
		return errors.New("tenant configuration file not found")
	}
	tenantConfigurationSecret, err := client.getSecret(ctx, tenant.Namespace, tenant.Spec.Configuration.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	tenantConfigurationSecret.Data["config.env"] = []byte(GenerateTenantConfigurationFile(tenantConfiguration))
	_, err = client.updateSecret(ctx, namespace, tenantConfigurationSecret, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	// Restart all MinIO pods at the same time for they to take the new configuration
	err = client.deletePodCollection(ctx, namespace, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	})
	if err != nil {
		return err
	}
	return nil
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

// addAnnotations will merge two annotation maps
func addAnnotations(annotationsOne, annotationsTwo map[string]string) map[string]string {
	if annotationsOne == nil {
		annotationsOne = map[string]string{}
	}
	for key, value := range annotationsTwo {
		annotationsOne[key] = value
	}
	return annotationsOne
}

// removeAnnotations will remove keys from the first annotations map based on the second one
func removeAnnotations(annotationsOne, annotationsTwo map[string]string) map[string]string {
	if annotationsOne == nil {
		annotationsOne = map[string]string{}
	}
	for key := range annotationsTwo {
		delete(annotationsOne, key)
	}
	return annotationsOne
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

func getTenantAddPoolResponse(session *models.Principal, params operator_api.TenantAddPoolParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: opClientClientSet,
	}
	if err := addTenantPool(ctx, opClient, params); err != nil {
		return ErrorWithContext(ctx, err, errors.New("unable to add pool"))
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

func getTenantPodsResponse(session *models.Principal, params operator_api.GetTenantPodsParams) ([]*models.TenantPod, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, params.Tenant),
	}
	pods, err := clientset.CoreV1().Pods(params.Namespace).List(ctx, listOpts)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return getTenantPods(pods), nil
}

func getTenantPods(pods *corev1.PodList) []*models.TenantPod {
	retval := []*models.TenantPod{}
	for _, pod := range pods.Items {
		var restarts int64
		if len(pod.Status.ContainerStatuses) > 0 {
			restarts = int64(pod.Status.ContainerStatuses[0].RestartCount)
		}
		status := string(pod.Status.Phase)
		if pod.DeletionTimestamp != nil {
			status = "Terminating"
		}
		retval = append(retval, &models.TenantPod{
			Name:        swag.String(pod.Name),
			Status:      status,
			TimeCreated: pod.CreationTimestamp.Unix(),
			PodIP:       pod.Status.PodIP,
			Restarts:    restarts,
			Node:        pod.Spec.NodeName,
		})
	}
	return retval
}

func getPodLogsResponse(session *models.Principal, params operator_api.GetPodLogsParams) (string, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	listOpts := &corev1.PodLogOptions{}
	logs := clientset.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, listOpts)
	buff, err := logs.DoRaw(ctx)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	return string(buff), nil
}

func getPodEventsResponse(session *models.Principal, params operator_api.GetPodEventsParams) (models.EventListWrapper, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	pod, err := clientset.CoreV1().Pods(params.Namespace).Get(ctx, params.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	events, err := clientset.CoreV1().Events(params.Namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pod.UID)})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	retval := models.EventListWrapper{}
	for i := 0; i < len(events.Items); i++ {
		retval = append(retval, &models.EventListElement{
			Namespace: events.Items[i].Namespace,
			LastSeen:  events.Items[i].LastTimestamp.Unix(),
			Message:   events.Items[i].Message,
			EventType: events.Items[i].Type,
			Reason:    events.Items[i].Reason,
		})
	}
	sort.SliceStable(retval, func(i int, j int) bool {
		return retval[i].LastSeen < retval[j].LastSeen
	})
	return retval, nil
}

func getDescribePodResponse(session *models.Principal, params operator_api.DescribePodParams) (*models.DescribePodWrapper, *models.Error) {
	ctx := context.Background()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	pod, err := clientset.CoreV1().Pods(params.Namespace).Get(ctx, params.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	retval := &models.DescribePodWrapper{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		PriorityClassName: pod.Spec.PriorityClassName,
		NodeName:          pod.Spec.NodeName,
	}
	if pod.Spec.Priority != nil {
		retval.Priority = int64(*pod.Spec.Priority)
	}
	if pod.Status.StartTime != nil {
		retval.StartTime = pod.Status.StartTime.Time.String()
	}
	labelArray := make([]*models.Label, len(pod.Labels))
	i := 0
	for key := range pod.Labels {
		labelArray[i] = &models.Label{Key: key, Value: pod.Labels[key]}
		i++
	}
	retval.Labels = labelArray
	annotationArray := make([]*models.Annotation, len(pod.Annotations))
	i = 0
	for key := range pod.Annotations {
		annotationArray[i] = &models.Annotation{Key: key, Value: pod.Annotations[key]}
		i++
	}
	retval.Annotations = annotationArray
	if pod.DeletionTimestamp != nil {
		retval.DeletionTimestamp = translateTimestampSince(*pod.DeletionTimestamp)
		retval.DeletionGracePeriodSeconds = *pod.DeletionGracePeriodSeconds
	}
	retval.Phase = string(pod.Status.Phase)
	retval.Reason = pod.Status.Reason
	retval.Message = pod.Status.Message
	retval.PodIP = pod.Status.PodIP
	retval.ControllerRef = metav1.GetControllerOf(pod).String()
	retval.Containers = make([]*models.Container, len(pod.Spec.Containers))
	statusMap := map[string]corev1.ContainerStatus{}
	statusKeys := make([]string, len(pod.Status.ContainerStatuses))
	for i, status := range pod.Status.ContainerStatuses {
		statusMap[status.Name] = status
		statusKeys[i] = status.Name

	}
	for i := range pod.Spec.Containers {
		retval.Containers[i] = &models.Container{
			Name:      pod.Spec.Containers[i].Name,
			Image:     pod.Spec.Containers[i].Image,
			Ports:     describeContainerPorts(pod.Spec.Containers[i].Ports),
			HostPorts: describeContainerHostPorts(pod.Spec.Containers[i].Ports),
			Args:      pod.Spec.Containers[i].Args,
		}
		if slices.Contains(statusKeys, pod.Spec.Containers[i].Name) {
			retval.Containers[i].ContainerID = statusMap[pod.Spec.Containers[i].Name].ContainerID
			retval.Containers[i].ImageID = statusMap[pod.Spec.Containers[i].Name].ImageID
			retval.Containers[i].Ready = statusMap[pod.Spec.Containers[i].Name].Ready
			retval.Containers[i].RestartCount = int64(statusMap[pod.Spec.Containers[i].Name].RestartCount)
			retval.Containers[i].State, retval.Containers[i].LastState = describeStatus(statusMap[pod.Spec.Containers[i].Name])
		}
		retval.Containers[i].EnvironmentVariables = make([]*models.EnvironmentVariable, len(pod.Spec.Containers[0].Env))
		for j := range pod.Spec.Containers[i].Env {
			retval.Containers[i].EnvironmentVariables[j] = &models.EnvironmentVariable{
				Key:   pod.Spec.Containers[i].Env[j].Name,
				Value: pod.Spec.Containers[i].Env[j].Value,
			}
		}
		retval.Containers[i].Mounts = make([]*models.Mount, len(pod.Spec.Containers[i].VolumeMounts))
		for j := range pod.Spec.Containers[i].VolumeMounts {
			retval.Containers[i].Mounts[j] = &models.Mount{
				Name:      pod.Spec.Containers[i].VolumeMounts[j].Name,
				MountPath: pod.Spec.Containers[i].VolumeMounts[j].MountPath,
				SubPath:   pod.Spec.Containers[i].VolumeMounts[j].SubPath,
				ReadOnly:  pod.Spec.Containers[i].VolumeMounts[j].ReadOnly,
			}
		}
	}
	retval.Conditions = make([]*models.Condition, len(pod.Status.Conditions))
	for i := range pod.Status.Conditions {
		retval.Conditions[i] = &models.Condition{
			Type:   string(pod.Status.Conditions[i].Type),
			Status: string(pod.Status.Conditions[i].Status),
		}
	}
	retval.Volumes = make([]*models.Volume, len(pod.Spec.Volumes))
	for i := range pod.Spec.Volumes {
		retval.Volumes[i] = &models.Volume{
			Name: pod.Spec.Volumes[i].Name,
		}
		if pod.Spec.Volumes[i].PersistentVolumeClaim != nil {
			retval.Volumes[i].Pvc = &models.Pvc{
				ReadOnly:  pod.Spec.Volumes[i].PersistentVolumeClaim.ReadOnly,
				ClaimName: pod.Spec.Volumes[i].PersistentVolumeClaim.ClaimName,
			}
		} else if pod.Spec.Volumes[i].Projected != nil {
			retval.Volumes[i].Projected = &models.ProjectedVolume{}
			retval.Volumes[i].Projected.Sources = make([]*models.ProjectedVolumeSource, len(pod.Spec.Volumes[i].Projected.Sources))
			for j := range pod.Spec.Volumes[i].Projected.Sources {
				retval.Volumes[i].Projected.Sources[j] = &models.ProjectedVolumeSource{}
				if pod.Spec.Volumes[i].Projected.Sources[j].Secret != nil {
					retval.Volumes[i].Projected.Sources[j].Secret = &models.Secret{
						Name:     pod.Spec.Volumes[i].Projected.Sources[j].Secret.Name,
						Optional: pod.Spec.Volumes[i].Projected.Sources[j].Secret.Optional != nil,
					}
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].DownwardAPI != nil {
					retval.Volumes[i].Projected.Sources[j].DownwardAPI = true
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap != nil {
					retval.Volumes[i].Projected.Sources[j].ConfigMap = &models.ConfigMap{
						Name:     pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap.Name,
						Optional: pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap.Optional != nil,
					}
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].ServiceAccountToken != nil {
					retval.Volumes[i].Projected.Sources[j].ServiceAccountToken = &models.ServiceAccountToken{ExpirationSeconds: *pod.Spec.Volumes[i].Projected.Sources[j].ServiceAccountToken.ExpirationSeconds}
				}
			}
		}
	}
	retval.QosClass = string(getPodQOS(pod))
	nodeSelectorArray := make([]*models.NodeSelector, len(pod.Spec.NodeSelector))
	i = 0
	for key := range pod.Spec.NodeSelector {
		nodeSelectorArray[i] = &models.NodeSelector{Key: key, Value: pod.Spec.NodeSelector[key]}
		i++
	}
	retval.NodeSelector = nodeSelectorArray
	retval.Tolerations = make([]*models.Toleration, len(pod.Spec.Tolerations))
	for i := range pod.Spec.Tolerations {
		retval.Tolerations[i] = &models.Toleration{
			Effect:            string(pod.Spec.Tolerations[i].Effect),
			Key:               pod.Spec.Tolerations[i].Key,
			Value:             pod.Spec.Tolerations[i].Value,
			Operator:          string(pod.Spec.Tolerations[i].Operator),
			TolerationSeconds: *pod.Spec.Tolerations[i].TolerationSeconds,
		}
	}
	return retval, nil
}

func describeStatus(status corev1.ContainerStatus) (*models.State, *models.State) {
	retval := &models.State{}
	last := &models.State{}
	state := status.State
	lastState := status.LastTerminationState
	switch {
	case state.Running != nil:
		retval.State = "Running"
		retval.Started = state.Running.StartedAt.Time.Format(time.RFC1123Z)
	case state.Waiting != nil:
		retval.State = "Waiting"
		retval.Reason = state.Waiting.Reason
	case state.Terminated != nil:
		retval.State = "Terminated"
		retval.Message = state.Terminated.Message
		retval.ExitCode = int64(state.Terminated.ExitCode)
		retval.Signal = int64(state.Terminated.Signal)
		retval.Started = state.Terminated.StartedAt.Time.Format(time.RFC1123Z)
		retval.Finished = state.Terminated.FinishedAt.Time.Format(time.RFC1123Z)
		switch {
		case lastState.Running != nil:
			last.State = "Running"
			last.Started = lastState.Running.StartedAt.Time.Format(time.RFC1123Z)
		case lastState.Waiting != nil:
			last.State = "Waiting"
			last.Reason = lastState.Waiting.Reason
		case lastState.Terminated != nil:
			last.State = "Terminated"
			last.Message = lastState.Terminated.Message
			last.ExitCode = int64(lastState.Terminated.ExitCode)
			last.Signal = int64(lastState.Terminated.Signal)
			last.Started = lastState.Terminated.StartedAt.Time.Format(time.RFC1123Z)
			last.Finished = lastState.Terminated.FinishedAt.Time.Format(time.RFC1123Z)
		default:
			last.State = "Waiting"
		}
	default:
		retval.State = "Waiting"
	}
	return retval, last
}

func describeContainerPorts(cPorts []corev1.ContainerPort) []string {
	ports := make([]string, 0, len(cPorts))
	for _, cPort := range cPorts {
		ports = append(ports, fmt.Sprintf("%d/%s", cPort.ContainerPort, cPort.Protocol))
	}
	return ports
}

func describeContainerHostPorts(cPorts []corev1.ContainerPort) []string {
	ports := make([]string, 0, len(cPorts))
	for _, cPort := range cPorts {
		ports = append(ports, fmt.Sprintf("%d/%s", cPort.HostPort, cPort.Protocol))
	}
	return ports
}

func getPodQOS(pod *corev1.Pod) corev1.PodQOSClass {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}
	zeroQuantity := resource.MustParse("0")
	isGuaranteed := true
	allContainers := []corev1.Container{}
	allContainers = append(allContainers, pod.Spec.Containers...)
	allContainers = append(allContainers, pod.Spec.InitContainers...)
	for _, container := range allContainers {
		// process requests
		for name, quantity := range container.Resources.Requests {
			if !isSupportedQoSComputeResource(name) {
				continue
			}
			if quantity.Cmp(zeroQuantity) == 1 {
				delta := quantity.DeepCopy()
				if _, exists := requests[name]; !exists {
					requests[name] = delta
				} else {
					delta.Add(requests[name])
					requests[name] = delta
				}
			}
		}
		// process limits
		qosLimitsFound := sets.NewString()
		for name, quantity := range container.Resources.Limits {
			if !isSupportedQoSComputeResource(name) {
				continue
			}
			if quantity.Cmp(zeroQuantity) == 1 {
				qosLimitsFound.Insert(string(name))
				delta := quantity.DeepCopy()
				if _, exists := limits[name]; !exists {
					limits[name] = delta
				} else {
					delta.Add(limits[name])
					limits[name] = delta
				}
			}
		}

		if !qosLimitsFound.HasAll(string(corev1.ResourceMemory), string(corev1.ResourceCPU)) {
			isGuaranteed = false
		}
	}
	if len(requests) == 0 && len(limits) == 0 {
		return corev1.PodQOSBestEffort
	}
	// Check is requests match limits for all resources.
	if isGuaranteed {
		for name, req := range requests {
			if lim, exists := limits[name]; !exists || lim.Cmp(req) != 0 {
				isGuaranteed = false
				break
			}
		}
	}
	if isGuaranteed &&
		len(requests) == len(limits) {
		return corev1.PodQOSGuaranteed
	}
	return corev1.PodQOSBurstable
}

var supportedQoSComputeResources = sets.NewString(string(corev1.ResourceCPU), string(corev1.ResourceMemory))

func isSupportedQoSComputeResource(name corev1.ResourceName) bool {
	return supportedQoSComputeResources.Has(string(name))
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// get values for prometheus metrics
func getTenantMonitoringResponse(session *models.Principal, params operator_api.GetTenantMonitoringParams) (*models.TenantMonitoringInfo, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minInst, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return getTenantMonitoring(minInst), nil
}

func getTenantMonitoring(minInst *miniov2.Tenant) *models.TenantMonitoringInfo {
	monitoringInfo := &models.TenantMonitoringInfo{}

	if minInst.Spec.Prometheus != nil {
		monitoringInfo.PrometheusEnabled = true
	} else {
		monitoringInfo.PrometheusEnabled = false
		return monitoringInfo
	}

	var storageClassName string
	if minInst.Spec.Prometheus.StorageClassName != nil {
		storageClassName = *minInst.Spec.Prometheus.StorageClassName
		monitoringInfo.StorageClassName = storageClassName
	}

	var requestedCPU string
	var requestedMem string

	if minInst.Spec.Prometheus.Resources.Requests != nil {
		// Parse cpu request
		if requestedCPUQ, ok := minInst.Spec.Prometheus.Resources.Requests["cpu"]; ok && requestedCPUQ.Value() != 0 {
			requestedCPU = strconv.FormatInt(requestedCPUQ.Value(), 10)
			monitoringInfo.MonitoringCPURequest = requestedCPU
		}
		// Parse memory request
		if requestedMemQ, ok := minInst.Spec.Prometheus.Resources.Requests["memory"]; ok && requestedMemQ.Value() != 0 {
			requestedMem = strconv.FormatInt(requestedMemQ.Value(), 10)
			monitoringInfo.MonitoringMemRequest = requestedMem
		}
	}

	if len(minInst.Spec.Prometheus.Labels) != 0 && minInst.Spec.Prometheus.Labels != nil {
		mLabels := []*models.Label{}
		for k, v := range minInst.Spec.Prometheus.Labels {
			mLabels = append(mLabels, &models.Label{Key: k, Value: v})
		}
		monitoringInfo.Labels = mLabels
	}

	if len(minInst.Spec.Prometheus.Annotations) != 0 && minInst.Spec.Prometheus.Annotations != nil {
		mAnnotations := []*models.Annotation{}
		for k, v := range minInst.Spec.Prometheus.Annotations {
			mAnnotations = append(mAnnotations, &models.Annotation{Key: k, Value: v})
		}
		monitoringInfo.Annotations = mAnnotations
	}

	if len(minInst.Spec.Prometheus.NodeSelector) != 0 && minInst.Spec.Prometheus.NodeSelector != nil {
		mNodeSelector := []*models.NodeSelector{}
		for k, v := range minInst.Spec.Prometheus.NodeSelector {
			mNodeSelector = append(mNodeSelector, &models.NodeSelector{Key: k, Value: v})
		}
		monitoringInfo.NodeSelector = mNodeSelector
	}

	if *minInst.Spec.Prometheus.DiskCapacityDB != 0 {
		monitoringInfo.DiskCapacityGB = strconv.Itoa(*minInst.Spec.Prometheus.DiskCapacityDB)
	}
	if len(minInst.Spec.Prometheus.Image) != 0 {
		monitoringInfo.Image = minInst.Spec.Prometheus.Image
	}
	if len(minInst.Spec.Prometheus.InitImage) != 0 {
		monitoringInfo.InitImage = minInst.Spec.Prometheus.InitImage
	}
	if len(minInst.Spec.Prometheus.ServiceAccountName) != 0 {
		monitoringInfo.ServiceAccountName = minInst.Spec.Prometheus.ServiceAccountName
	}
	if len(minInst.Spec.Prometheus.SideCarImage) != 0 {
		monitoringInfo.SidecarImage = minInst.Spec.Prometheus.SideCarImage
	}
	if minInst.Spec.Prometheus.SecurityContext != nil {
		monitoringInfo.SecurityContext = convertK8sSCToModelSC(minInst.Spec.Prometheus.SecurityContext)
	}
	return monitoringInfo
}

// sets tenant Prometheus monitoring cofiguration fields to values provided
func setTenantMonitoringResponse(session *models.Principal, params operator_api.SetTenantMonitoringParams) (bool, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return false, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return false, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	return setTenantMonitoring(ctx, minTenant, opClient, params)
}

func setTenantMonitoring(ctx context.Context, minTenant *miniov2.Tenant, opClient OperatorClientI, params operator_api.SetTenantMonitoringParams) (bool, *models.Error) {
	if params.Data.Toggle {
		if params.Data.PrometheusEnabled {
			minTenant.Spec.Prometheus = nil
		} else {
			promDiskSpaceGB := 5
			promImage := ""
			minTenant.Spec.Prometheus = &miniov2.PrometheusConfig{
				DiskCapacityDB: swag.Int(promDiskSpaceGB),
				Image:          promImage,
			}
		}
		_, err := opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		return true, nil
	}

	labels := make(map[string]string)
	for i := 0; i < len(params.Data.Labels); i++ {
		if params.Data.Labels[i] != nil {
			labels[params.Data.Labels[i].Key] = params.Data.Labels[i].Value
		}
	}
	annotations := make(map[string]string)
	for i := 0; i < len(params.Data.Annotations); i++ {
		if params.Data.Annotations[i] != nil {
			annotations[params.Data.Annotations[i].Key] = params.Data.Annotations[i].Value
		}
	}
	nodeSelector := make(map[string]string)
	for i := 0; i < len(params.Data.NodeSelector); i++ {
		if params.Data.NodeSelector[i] != nil {
			nodeSelector[params.Data.NodeSelector[i].Key] = params.Data.NodeSelector[i].Value
		}
	}

	monitoringResourceRequest := make(corev1.ResourceList)
	if params.Data.MonitoringCPURequest != "" {
		cpuQuantity, err := resource.ParseQuantity(params.Data.MonitoringCPURequest)
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		monitoringResourceRequest["cpu"] = cpuQuantity
	}

	if params.Data.MonitoringMemRequest != "" {
		memQuantity, err := resource.ParseQuantity(params.Data.MonitoringMemRequest)
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		monitoringResourceRequest["memory"] = memQuantity
	}

	minTenant.Spec.Prometheus.Resources.Requests = monitoringResourceRequest
	minTenant.Spec.Prometheus.Labels = labels
	minTenant.Spec.Prometheus.Annotations = annotations
	minTenant.Spec.Prometheus.NodeSelector = nodeSelector
	minTenant.Spec.Prometheus.Image = params.Data.Image
	minTenant.Spec.Prometheus.SideCarImage = params.Data.SidecarImage
	minTenant.Spec.Prometheus.InitImage = params.Data.InitImage
	if params.Data.StorageClassName == "" {
		minTenant.Spec.Prometheus.StorageClassName = nil
	} else {
		minTenant.Spec.Prometheus.StorageClassName = &params.Data.StorageClassName
	}

	diskCapacityGB, err := strconv.Atoi(params.Data.DiskCapacityGB)
	if err == nil {
		*minTenant.Spec.Prometheus.DiskCapacityDB = diskCapacityGB
	}

	minTenant.Spec.Prometheus.ServiceAccountName = params.Data.ServiceAccountName
	minTenant.Spec.Prometheus.SecurityContext, err = convertModelSCToK8sSC(params.Data.SecurityContext)
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	_, err = opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	return true, nil
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
				NodeSelectorTerms: nodeSelectorTerms,
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

func getTenantUpdatePoolResponse(session *models.Principal, params operator_api.TenantUpdatePoolsParams) (*models.Tenant, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	tenant, err := updateTenantPools(ctx, opClient, params.Namespace, params.Tenant, params.Body.Pools)
	if err != nil {
		LogError("error updating Tenant's pools: %v", err)
		return nil, ErrorWithContext(ctx, err)
	}
	return tenant, nil
}

// updateTenantPools Sets the Tenant's pools to the ones provided by the request
//
// It does the equivalent to a PUT request on Tenant's pools
func updateTenantPools(
	ctx context.Context,
	operatorClient OperatorClientI,
	namespace string,
	tenantName string,
	poolsReq []*models.Pool,
) (*models.Tenant, error) {
	minInst, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// set the pools if they are provided
	var newPoolArray []miniov2.Pool
	for _, pool := range poolsReq {
		pool, err := parseTenantPoolRequest(pool)
		if err != nil {
			return nil, err
		}
		newPoolArray = append(newPoolArray, *pool)
	}

	// replace pools array
	minInst.Spec.Pools = newPoolArray

	minInst = minInst.DeepCopy()
	minInst.EnsureDefaults()

	payloadBytes, err := json.Marshal(minInst)
	if err != nil {
		return nil, err
	}
	tenantUpdated, err := operatorClient.TenantPatch(ctx, namespace, minInst.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return getTenantInfo(tenantUpdated), nil
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

func getUpdateDomainsResponse(session *models.Principal, params operator_api.UpdateTenantDomainsParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	operatorCli, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: operatorCli,
	}

	err = updateTenantDomains(ctx, opClient, params.Namespace, params.Tenant, params.Body.Domains)

	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	return nil
}

func updateTenantDomains(ctx context.Context, operatorClient OperatorClientI, namespace string, tenantName string, domainConfig *models.DomainsConfiguration) error {
	minTenant, err := getTenant(ctx, operatorClient, namespace, tenantName)
	if err != nil {
		return err
	}

	var features miniov2.Features
	var domains miniov2.TenantDomains

	// We include current value for BucketDNS. Domains will be overwritten as we are passing all the values that must be saved.
	if minTenant.Spec.Features != nil {
		features = miniov2.Features{
			BucketDNS: minTenant.Spec.Features.BucketDNS,
		}
	}

	if domainConfig != nil {
		// tenant domains
		if domainConfig.Console != "" {
			domains.Console = domainConfig.Console
		}

		if domainConfig.Minio != nil {
			domains.Minio = domainConfig.Minio
		}

		features.Domains = &domains
	}

	minTenant.Spec.Features = &features

	_, err = operatorClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})

	return err
}
