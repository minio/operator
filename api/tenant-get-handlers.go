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
	"fmt"

	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getTenantDetailsResponse(session *models.Principal, params operator_api.TenantDetailsParams) (*models.Tenant, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	info := getTenantInfo(minTenant)

	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	k8sClient := k8sClient{
		client: clientSet,
	}

	tenantConfiguration, err := GetTenantConfiguration(ctx, &k8sClient, minTenant)
	if err != nil {
		LogError("unable to fetch configuration for tenant %s: %v", minTenant.Name, err)
	}

	// detect if AD/LDAP is enabled
	ldapEnabled := tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_ADDR"] != ""

	// detect if OpenID is enabled
	oidcEnabled := tenantConfiguration["MINIO_IDENTITY_OPENID_CONFIG_URL"] != ""

	// detect if encryption is enabled
	info.EncryptionEnabled = minTenant.HasKESEnabled() || tenantConfiguration["MINIO_KMS_SECRET_KEY"] != ""
	info.IdpAdEnabled = ldapEnabled
	info.IdpOidcEnabled = oidcEnabled
	info.MinioTLS = minTenant.TLS()

	// attach status information
	info.Status = &models.TenantStatus{
		HealthStatus:  string(minTenant.Status.HealthStatus),
		DrivesHealing: minTenant.Status.DrivesHealing,
		DrivesOffline: minTenant.Status.DrivesOffline,
		DrivesOnline:  minTenant.Status.DrivesOnline,
		WriteQuorum:   minTenant.Status.WriteQuorum,
		Usage: &models.TenantStatusUsage{
			Raw:           minTenant.Status.Usage.RawCapacity,
			RawUsage:      minTenant.Status.Usage.RawUsage,
			Capacity:      minTenant.Status.Usage.Capacity,
			CapacityUsage: minTenant.Status.Usage.Usage,
		},
	}

	// get tenant service
	minTenant.EnsureDefaults()
	// minio service
	minSvc, err := k8sClient.getService(ctx, minTenant.Namespace, minTenant.MinIOCIServiceName(), metav1.GetOptions{})
	if err != nil {
		// we can tolerate this errors
		LogError("Unable to get MinIO service name: %v, continuing", err)
	}
	// console service
	conSvc, err := k8sClient.getService(ctx, minTenant.Namespace, minTenant.ConsoleCIServiceName(), metav1.GetOptions{})
	if err != nil {
		// we can tolerate this errors
		LogError("Unable to get MinIO console service name: %v, continuing", err)
	}

	schema := "http"
	consoleSchema := schema
	consolePort := fmt.Sprintf(":%d", miniov2.ConsolePort)
	if minTenant.TLS() {
		schema = "https"
		consoleSchema = schema
		consolePort = fmt.Sprintf(":%d", miniov2.ConsoleTLSPort)
	}

	var minioEndpoint string
	var consoleEndpoint string
	if minSvc != nil && len(minSvc.Status.LoadBalancer.Ingress) > 0 {
		if minSvc.Status.LoadBalancer.Ingress[0].IP != "" {
			minioEndpoint = fmt.Sprintf("%s://%s", schema, minSvc.Status.LoadBalancer.Ingress[0].IP)
		}

		if minSvc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			minioEndpoint = fmt.Sprintf("%s://%s", schema, minSvc.Status.LoadBalancer.Ingress[0].Hostname)
		}

	}

	if conSvc != nil && len(conSvc.Status.LoadBalancer.Ingress) > 0 {
		if conSvc.Status.LoadBalancer.Ingress[0].IP != "" {
			consoleEndpoint = fmt.Sprintf("%s://%s%s", consoleSchema, conSvc.Status.LoadBalancer.Ingress[0].IP, consolePort)
		}
		if conSvc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			consoleEndpoint = fmt.Sprintf("%s://%s%s", consoleSchema, conSvc.Status.LoadBalancer.Ingress[0].Hostname, consolePort)
		}
	}

	info.Endpoints = &models.TenantEndpoints{
		Console: consoleEndpoint,
		Minio:   minioEndpoint,
	}

	var domains models.DomainsConfiguration

	if minTenant.Spec.Features != nil {
		if minTenant.Spec.Features.EnableSFTP != nil {
			info.SftpExposed = *minTenant.Spec.Features.EnableSFTP
		}
		if minTenant.Spec.Features.Domains != nil {
			domains = models.DomainsConfiguration{
				Console: minTenant.Spec.Features.Domains.Console,
				Minio:   minTenant.Spec.Features.Domains.Minio,
			}
		}
	}

	info.Domains = &domains

	var tiers []*models.TenantTierElement

	for _, tier := range minTenant.Status.Usage.Tiers {
		tierItem := &models.TenantTierElement{
			Name: tier.Name,
			Type: tier.Type,
			Size: tier.TotalSize,
		}

		tiers = append(tiers, tierItem)
	}
	info.Tiers = tiers

	return info, nil
}
