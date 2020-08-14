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

package v1

import (
	"crypto/elliptic"
	"runtime"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MinIOOperatorVersionLabel denotes the version of the Tenant operator
// running in the cluster.
const MinIOOperatorVersionLabel = "v1.min.io/version"

// MinIOCRDResourceKind is the Kind of a Cluster.
const MinIOCRDResourceKind = "Tenant"

// DefaultPodManagementPolicy specifies default pod management policy as expllained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies
const DefaultPodManagementPolicy = appsv1.ParallelPodManagement

// DefaultUpdateStrategy specifies default pod update policy as explained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies
const DefaultUpdateStrategy = "RollingUpdate"

// DefaultImagePullPolicy specifies the policy to image pulls
const DefaultImagePullPolicy = corev1.PullAlways

// CSRNameSuffix specifies the suffix added to Tenant name to create a CSR
const CSRNameSuffix = "-csr"

// MinIO Related Constants

// MinIOCertPath is the path where all MinIO certs are mounted
const MinIOCertPath = "/tmp/certs"

// TenantLabel is applied to all components of a Tenant cluster
const TenantLabel = "v1.min.io/tenant"

// ZoneLabel is applied to all components in a Zone of a Tenant cluster
const ZoneLabel = "v1.min.io/zone"

// MinIOPort specifies the default Tenant port number.
const MinIOPort = 9000

// MinIOServicePortName specifies the default Service's port name, e.g. for automatic protocol selection in Istio
const MinIOServicePortName = "http-minio"

// MinIOVolumeName specifies the default volume name for MinIO volumes
const MinIOVolumeName = "export"

// MinIOVolumeMountPath specifies the default mount path for MinIO volumes
const MinIOVolumeMountPath = "/export"

// MinIOVolumeSubPath specifies the default sub path under mount path
const MinIOVolumeSubPath = ""

// DefaultMinIOImage specifies the default MinIO Docker hub image
const DefaultMinIOImage = "minio/minio:RELEASE.2020-08-13T02-39-50Z"

// DefaultMinIOUpdateURL specifies the default MinIO URL where binaries are
// pulled from during MinIO upgrades
const DefaultMinIOUpdateURL = "https://dl.min.io/server/minio/release/" + runtime.GOOS + "-" + runtime.GOARCH + "/archive/"

// MinIOHLSvcNameSuffix specifies the suffix added to Tenant name to create a headless service
const MinIOHLSvcNameSuffix = "-hl"

// DefaultServers specifies the default MinIO replicas to use for distributed deployment if not specified explicitly by user
const DefaultServers = 1

// DefaultVolumesPerServer specifies the default number of volumes per MinIO instance
const DefaultVolumesPerServer = 1

// DefaultZoneName specifies the default zone name
const DefaultZoneName = "zone-0"

// LivenessPath specifies the endpoint for liveness check
const LivenessPath = "/minio/health/live"

// LivenessInitialDelay specifies the initial delay in calling the liveness endpoint
const LivenessInitialDelay = 10

// LivenessPeriod specifies the interval in calling the liveness endpoint
const LivenessPeriod = 1

// LivenessTimeout specifies the timeout for the liveness probe to expect a response
const LivenessTimeout = 1

// LivenessFailureThreshold specifies the number of times the liveness probe can fail before
// the probe to be considered failed.
const LivenessFailureThreshold = 1

// Console Related Constants

// DefaultConsoleImage specifies the latest Console Docker hub image
const DefaultConsoleImage = "minio/console:v0.3.0"

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

// ConsoleServiceNameSuffix specifies the suffix added to Tenant service name to create a service for console
const ConsoleServiceNameSuffix = "-ui"

// ConsoleName specifies the default container name for Console
const ConsoleName = "-console"

// ConsoleAdminPolicyName denotes the policy name for Console user
const ConsoleAdminPolicyName = "consoleAdmin"

// ConsoleRestartPolicy defines the default restart policy for Console Containers
const ConsoleRestartPolicy = corev1.RestartPolicyAlways

// ConsoleConfigMountPath specifies the path where Console config file and all secrets are mounted
// We keep this to /tmp so it doesn't require any special permissions
const ConsoleConfigMountPath = "/tmp/console"

// DefaultConsoleReplicas specifies the default number of Console pods to be created if not specified
const DefaultConsoleReplicas = 2

// KES Related Constants

// DefaultKESImage specifies the latest KES Docker hub image
const DefaultKESImage = "minio/kes:v0.11.0"

// KESInstanceLabel is applied to the KES pods of a Tenant cluster
const KESInstanceLabel = "v1.min.io/kes"

// KESPort specifies the default KES Service's port number.
const KESPort = 7373

// KESServicePortName specifies the default KES Service's port name.
const KESServicePortName = "http-kes"

// KESMinIOKey is the name of key that KES creates on the KMS backend
const KESMinIOKey = "my-minio-key"

// KESJobRestartPolicy specifies the restart policy for the job created for key creation
const KESJobRestartPolicy = corev1.RestartPolicyOnFailure

// KESHLSvcNameSuffix specifies the suffix added to Tenant name to create a headless service for KES
const KESHLSvcNameSuffix = "-kes-hl-svc"

// KESName specifies the default container name for KES
const KESName = "-kes"

// KESConfigMountPath specifies the path where KES config file and all secrets are mounted
// We keep this to /tmp so it doesn't require any special permissions
const KESConfigMountPath = "/tmp/kes"

// DefaultKESReplicas specifies the default number of KES pods to be created if not specified
const DefaultKESReplicas = 2

// Auto TLS related constants

// DefaultEllipticCurve specifies the default elliptic curve to be used for key generation
var DefaultEllipticCurve = elliptic.P256()

// DefaultOrgName specifies the default Org name to be used in automatic certificate generation
var DefaultOrgName = []string{"Acme Co"}

// DefaultQueryInterval specifies the interval between each query for CSR Status
var DefaultQueryInterval = time.Second * 5

// DefaultQueryTimeout specifies the timeout for query for CSR Status
var DefaultQueryTimeout = time.Minute * 20

// TLSSecretSuffix is the suffix applied to Tenant name to create the TLS secret
var TLSSecretSuffix = "-tls"
