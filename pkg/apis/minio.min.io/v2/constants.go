// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package v2

import (
	"crypto/elliptic"
	"runtime"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MinIOCRDResourceKind is the Kind of Cluster.
const MinIOCRDResourceKind = "Tenant"

// DefaultPodManagementPolicy specifies default pod management policy as expllained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies
const DefaultPodManagementPolicy = appsv1.ParallelPodManagement

// DefaultUpdateStrategy specifies default pod update policy as explained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies
const DefaultUpdateStrategy = "RollingUpdate"

// DefaultImagePullPolicy specifies the policy to image pulls
const DefaultImagePullPolicy = corev1.PullIfNotPresent

// CSRNameSuffix specifies the suffix added to Tenant name to create a CSR
const CSRNameSuffix = "-csr"

// MinIO Related Constants

// MinIOCertPath is the path where all MinIO certs are mounted
const MinIOCertPath = "/tmp/certs"

// TmpPath /tmp path inside the container file system
const TmpPath = "/tmp"

// CfgPath is the location of the MinIO Configuration File
const CfgPath = "/tmp/minio/"

// CfgFile is the Configuration File for MinIO
const CfgFile = CfgPath + "config.env"

// TenantLabel is applied to all components of a Tenant cluster
const TenantLabel = "v1.min.io/tenant"

// PoolLabel is applied to all components in a Pool of a Tenant cluster
const PoolLabel = "v1.min.io/pool"

// ZoneLabel is used for compatibility with tenants deployed prior to operator 4.0.0
const ZoneLabel = "v1.min.io/zone"

// Revision is applied to all statefulsets
const Revision = "min.io/revision"

// MinIOPort specifies the default Tenant port number.
const MinIOPort = 9000

// MinIOPortLoadBalancerSVC specifies the default Service port number for the load balancer service.
const MinIOPortLoadBalancerSVC = 80

// MinIOTLSPortLoadBalancerSVC specifies the default Service TLS port number for the load balancer service.
const MinIOTLSPortLoadBalancerSVC = 443

// MinIOServiceHTTPPortName specifies the default Service's http port name, e.g. for automatic protocol selection in Istio
const MinIOServiceHTTPPortName = "http-minio"

// MinIOServiceHTTPSPortName specifies the default Service's https port name, e.g. for automatic protocol selection in Istio
const MinIOServiceHTTPSPortName = "https-minio"

// MinIOVolumeName specifies the default volume name for MinIO volumes
const MinIOVolumeName = "export"

// MinIOVolumeMountPath specifies the default mount path for MinIO volumes
const MinIOVolumeMountPath = "/export"

// MinIOVolumeSubPath specifies the default sub path under mount path
const MinIOVolumeSubPath = ""

// DefaultMinIOImage specifies the default MinIO Docker hub image
const DefaultMinIOImage = "minio/minio:RELEASE.2023-06-23T20-26-00Z"

// DefaultMinIOUpdateURL specifies the default MinIO URL where binaries are
// pulled from during MinIO upgrades
const DefaultMinIOUpdateURL = "https://dl.min.io/server/minio/release/" + runtime.GOOS + "-" + runtime.GOARCH + "/archive/"

// MinIOHLSvcNameSuffix specifies the suffix added to Tenant name to create a headless service
const MinIOHLSvcNameSuffix = "-hl"

// TenantConfigurationSecretSuffix specifies the suffix added to tenant name to create the configuration secret name
const TenantConfigurationSecretSuffix = "-configuration"

// Console Related Constants

// ConsoleTenantLabel is applied to the Console pods of a Tenant cluster
const ConsoleTenantLabel = "v1.min.io/console"

// ConsolePort specifies the default Console port number.
const ConsolePort = 9090

// ConsoleServicePortName specifies the default Console Service's port name.
const ConsoleServicePortName = "http-console"

// ConsoleTLSPort specifies the default Console port number for HTTPS.
const ConsoleTLSPort = 9443

// ConsoleServiceTLSPortName specifies the default Console Service's port name.
const ConsoleServiceTLSPortName = "https-console"

// ConsoleName specifies the default container name for Console
const ConsoleName = "-console"

// ConsoleAdminPolicyName denotes the policy name for Console user
const ConsoleAdminPolicyName = "consoleAdmin"

// KES Related Constants

// DefaultKESImage specifies the latest KES Docker hub image
const DefaultKESImage = "minio/kes:2023-05-02T22-48-10Z"

// KESInstanceLabel is applied to the KES pods of a Tenant cluster
const KESInstanceLabel = "v1.min.io/kes"

// KESPort specifies the default KES Service's port number.
const KESPort = 7373

// KESServicePortName specifies the default KES Service's port name.
const KESServicePortName = "http-kes"

// KESMinIOKey is the name of key that KES creates on the KMS backend
const KESMinIOKey = "my-minio-key"

// KESHLSvcNameSuffix specifies the suffix added to Tenant name to create a headless service for KES
const KESHLSvcNameSuffix = "-kes-hl-svc"

// KESName specifies the default container name for KES
const KESName = "-kes"

// KESConfigMountPath specifies the path where KES config file and all secrets are mounted
// We keep this to /tmp, so it doesn't require any special permissions
const KESConfigMountPath = "/tmp/kes"

// DefaultKESReplicas specifies the default number of KES pods to be created if not specified
const DefaultKESReplicas = 2

// Auto TLS related constants

// DefaultEllipticCurve specifies the default elliptic curve to be used for key generation
var DefaultEllipticCurve = elliptic.P256()

// DefaultOrgName specifies the default Org name to be used in automatic certificate generation
var DefaultOrgName = []string{"system:nodes"}

// DefaultQueryInterval specifies the interval between each query for CSR Status
var DefaultQueryInterval = time.Second * 5

// DefaultQueryTimeout specifies the timeout for query for CSR Status
var DefaultQueryTimeout = time.Minute * 20

// TLSSecretSuffix is the suffix applied to Tenant name to create the TLS secret
var TLSSecretSuffix = "-tls"

// Cluster Domain
const clusterDomain = "CLUSTER_DOMAIN"

// StatefulSetPrefix used by statefulsets
const StatefulSetPrefix = "ss"

// StatefulSetLegacyPrefix by old operators
const StatefulSetLegacyPrefix = "zone"

// MinIOPrometheusPathCluster is the path where MinIO tenant exposes cluster Prometheus metrics
const MinIOPrometheusPathCluster = "/minio/v2/metrics/cluster"

// MinIOPrometheusScrapeInterval defines how frequently to scrape targets.
const MinIOPrometheusScrapeInterval = 30 * time.Second

const tenantMinIOImageEnv = "TENANT_MINIO_IMAGE"

const tenantKesImageEnv = "TENANT_KES_IMAGE"

const monitoringIntervalEnv = "MONITORING_INTERVAL"

// DefaultMonitoringInterval is how often we run monitoring on tenants
const DefaultMonitoringInterval = 3

// PrometheusNamespace is the namespace of the prometheus
const PrometheusNamespace = "PROMETHEUS_NAMESPACE"

// DefaultPrometheusNamespace is the default namespace for prometheus
const DefaultPrometheusNamespace = "default"

// PrometheusAddlScrapeConfigSecret is the name of the secrets which contains the scrape config
const PrometheusAddlScrapeConfigSecret = "minio-prom-additional-scrape-config"

// PrometheusAddlScrapeConfigKey is the key in secret data
const PrometheusAddlScrapeConfigKey = "prometheus-additional.yaml"
