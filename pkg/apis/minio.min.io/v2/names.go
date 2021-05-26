/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package v2

import (
	"fmt"
	"strings"
)

// MinIOServerName specifies the default container name for Tenant
const MinIOServerName = "minio"

// MinIODNSInitContainer Init Container for DNS
const MinIODNSInitContainer = "minio-dns-wait"

// MinIOVolumeInitContainer Init Container for DNS
const MinIOVolumeInitContainer = "minio-vol-wait"

// KESContainerName specifies the default container name for KES
const KESContainerName = "kes"

// ConsoleContainerName specifies the default container name for Console
const ConsoleContainerName = "console"

// LogPgContainerName is the default name for the Log (PostgreSQL) server
// container
const LogPgContainerName = "log-search-pg"

// LogSearchAPIContainerName is the name for the log search API server container
const LogSearchAPIContainerName = "log-search-api"

// PrometheusContainerName is the name of the prometheus server container
const PrometheusContainerName = "prometheus"

// InitContainerImage name for init container.
const InitContainerImage = "busybox:1.32"

// MinIO Related Names

// MinIOStatefulSetNameForPool returns the name for MinIO StatefulSet
func (t *Tenant) MinIOStatefulSetNameForPool(z *Pool) string {
	return fmt.Sprintf("%s-%s", t.Name, z.Name)
}

// MinIOWildCardName returns the wild card name for all MinIO Pods in current StatefulSet
func (t *Tenant) MinIOWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", t.MinIOHLServiceName(), t.Namespace, GetClusterDomain())
}

// MinIOTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) MinIOTLSSecretName() string {
	return t.Name + TLSSecretSuffix
}

// MinIOClientTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
// for MinIO <-> KES client side authentication.
func (t *Tenant) MinIOClientTLSSecretName() string {
	return t.Name + "-client" + TLSSecretSuffix
}

// MinIOHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this Tenant
func (t *Tenant) MinIOHLServiceName() string {
	return t.Name + MinIOHLSvcNameSuffix
}

// MinIOCIServiceName returns the name of Cluster IP service that is created to communicate
// with current MinIO StatefulSet pods
func (t *Tenant) MinIOCIServiceName() string {
	// DO NOT CHANGE, this should be constant
	// This is possible because each namespace has only one Tenant
	return "minio"
}

// MinIOBucketBaseDomain returns the base domain name for buckets
func (t *Tenant) MinIOBucketBaseDomain() string {
	return fmt.Sprintf("%s.svc.%s", t.Namespace, GetClusterDomain())
}

// MinIOHLPodHostname returns the full address of a particular MinIO pod.
func (t *Tenant) MinIOHLPodHostname(podName string) string {
	return fmt.Sprintf("%s.%s.%s.svc.%s", podName, t.MinIOHLServiceName(), t.Namespace, GetClusterDomain())
}

// MinIOBucketBaseWildcardDomain returns the base domain name for buckets
func (t *Tenant) MinIOBucketBaseWildcardDomain() string {
	return fmt.Sprintf("*.%s.svc.%s", t.Namespace, GetClusterDomain())
}

// MinIOFQDNServiceName returns the name of the service created for the tenant.
func (t *Tenant) MinIOFQDNServiceName() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.MinIOCIServiceName(), t.Namespace, GetClusterDomain())
}

// MinIOFQDNServiceNameAndNamespace returns the name of the service created for the tenant up to namespace, ie: minio.default
func (t *Tenant) MinIOFQDNServiceNameAndNamespace() string {
	return fmt.Sprintf("%s.%s", t.MinIOCIServiceName(), t.Namespace)
}

// MinIOFQDNShortServiceName returns the name of the service created for the tenant up to svc, ie: minio.default.svc
func (t *Tenant) MinIOFQDNShortServiceName() string {
	return fmt.Sprintf("%s.svc", t.MinIOFQDNServiceNameAndNamespace())
}

// MinIOCSRName returns the name of CSR that is generated if AutoTLS is enabled
// Namespace adds uniqueness to the CSR name (single MinIO tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) MinIOCSRName() string {
	return t.Name + "-" + t.Namespace + CSRNameSuffix
}

// MinIOClientCSRName returns the name of CSR that is generated for Client side authentication
// Used by KES Pods
func (t *Tenant) MinIOClientCSRName() string {
	return t.Name + "-client-" + t.Namespace + CSRNameSuffix
}

// KES Related Names

// KESJobName returns the name for KES Key Job
func (t *Tenant) KESJobName() string {
	return t.Name + KESName
}

// KESStatefulSetName returns the name for KES StatefulSet
func (t *Tenant) KESStatefulSetName() string {
	return t.Name + KESName
}

// KESHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this Tenant
func (t *Tenant) KESHLServiceName() string {
	return t.Name + KESHLSvcNameSuffix
}

// KESVolMountName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) KESVolMountName() string {
	return t.Name + KESName
}

// KESWildCardName returns the wild card name managed by headless service created for
// KES StatefulSet in current Tenant
func (t *Tenant) KESWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", t.KESHLServiceName(), t.Namespace, GetClusterDomain())
}

// KESTLSSecretName returns the name of Secret that has KES TLS related Info (Cert & Private Key)
func (t *Tenant) KESTLSSecretName() string {
	return t.KESStatefulSetName() + TLSSecretSuffix
}

// KESCSRName returns the name of CSR that generated if AutoTLS is enabled for KES
// Namespace adds uniqueness to the CSR name (single KES tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) KESCSRName() string {
	return t.KESStatefulSetName() + "-" + t.Namespace + CSRNameSuffix
}

// Console Related Names

// ConsoleDeploymentName returns the name for Console Deployment
func (t *Tenant) ConsoleDeploymentName() string {
	return t.Name + ConsoleName
}

// ConsoleCIServiceName returns the name for Console Cluster IP Service
func (t *Tenant) ConsoleCIServiceName() string {
	return t.Name + ConsoleName
}

// PoolStatefulsetName returns the name of a statefulset for a given pool
func (t *Tenant) PoolStatefulsetName(pool *Pool) string {
	return fmt.Sprintf("%s-%s", t.Name, pool.Name)
}

// LegacyStatefulsetName returns the name of a statefulset for a given pool
func (t *Tenant) LegacyStatefulsetName(pool *Pool) string {
	zoneName := strings.Replace(pool.Name, StatefulSetPrefix, StatefulSetLegacyPrefix, 1)
	return fmt.Sprintf("%s-%s", t.Name, zoneName)
}

// ConsoleVolMountName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) ConsoleVolMountName() string {
	return t.Name + ConsoleName
}

// ConsoleCommonName returns the CommonName to be used in the csr template
func (t *Tenant) ConsoleCommonName() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.ConsoleCIServiceName(), t.Namespace, GetClusterDomain())
}

// ConsoleTLSSecretName returns the name of Secret that has Console TLS related Info (Cert & Private Key)
func (t *Tenant) ConsoleTLSSecretName() string {
	return t.ConsoleDeploymentName() + TLSSecretSuffix
}

// ConsoleCSRName returns the name of CSR that generated if AutoTLS is enabled for Console
// Namespace adds uniqueness to the CSR name (single Console tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) ConsoleCSRName() string {
	return t.ConsoleDeploymentName() + "-" + t.Namespace + CSRNameSuffix
}

// LogStatefulsetName returns name of statefulsets meant for Log feature
func (t *Tenant) LogStatefulsetName() string {
	return fmt.Sprintf("%s-%s", t.Name, "log")
}

// LogHLServiceName returns name of Headless service for the Log statefulsets
func (t *Tenant) LogHLServiceName() string {
	return t.Name + LogHLSvcNameSuffix
}

// LogSecretName returns name of secret shared by Log PG server and log-search-api server
func (t *Tenant) LogSecretName() string {
	return fmt.Sprintf("%s-%s", t.Name, "log-secret")
}

// PromServiceMonitorSecret returns name of secret with jwt for Prometheus service monitor
func (t *Tenant) PromServiceMonitorSecret() string {
	return fmt.Sprintf("%s-%s", t.Name, "prom-sm-secret")
}

// LogSearchAPIDeploymentName returns name of Log Search API server deployment
func (t *Tenant) LogSearchAPIDeploymentName() string {
	return fmt.Sprintf("%s-%s", t.Name, LogSearchAPIContainerName)
}

// LogSearchAPIServiceName returns name of Log Search API service name
func (t *Tenant) LogSearchAPIServiceName() string {
	return fmt.Sprintf("%s-%s", t.Name, LogSearchAPIContainerName)
}

// PrometheusStatefulsetName returns name of statefulset meant for Prometheus
// metrics.
func (t *Tenant) PrometheusStatefulsetName() string {
	return fmt.Sprintf("%s-%s", t.Name, "prometheus")
}

// PrometheusServiceMonitorName returns name of service monitor meant
// for Prometheus metrics.
func (t *Tenant) PrometheusServiceMonitorName() string {
	return fmt.Sprintf("%s-%s", t.Name, "prometheus")
}

// PrometheusConfigMapName returns name of the config map for Prometheus.
func (t *Tenant) PrometheusConfigMapName() string {
	return fmt.Sprintf("%s-%s", t.Name, "prometheus-config-map")
}

// PrometheusConfigVolMountName returns name of the prometheus config volume.
func (t *Tenant) PrometheusConfigVolMountName() string {
	return fmt.Sprintf("%s-prometheus-config-volmount", t.Name)
}

// PrometheusServiceName returns name of the Prometheus service
func (t *Tenant) PrometheusServiceName() string {
	return fmt.Sprintf("%s-%s", t.Name, PrometheusContainerName)
}

// PrometheusHLServiceName returns name of Headless service for the Log
// statefulsets
func (t *Tenant) PrometheusHLServiceName() string {
	return t.Name + PrometheusHLSvcNameSuffix
}
