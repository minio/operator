/*
 * MinIO-Operator - Manage MinIO clusters in Kubernetes
 *
 * MinIO Cloud Storage, (C) 2018, 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

import (
	"crypto/elliptic"
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

// InstanceLabel is applied to all components of a MinIOInstance cluster
const InstanceLabel = "v1beta1.min.io/instance"

// MinIOOperatorVersionLabel denotes the version of the MinIOInstance operator
// running in the cluster.
const MinIOOperatorVersionLabel = "v1beta1.min.io/version"

// MinIOPort specifies the default MinIOInstance port number.
const MinIOPort = 9000

// DefaultReplicas specifies the default MinIO replicas to use for distributed deployment if not specified explicitly by user
const DefaultReplicas = 4

// MinIOVolumeName specifies the default volume name for MinIO volumes
const MinIOVolumeName = "export"

// MinIOVolumeMountPath specifies the default mount path for MinIO volumes
const MinIOVolumeMountPath = "/export"

// MinIOVolumeSubPath specifies the default sub path under mount path
const MinIOVolumeSubPath = ""

// DefaultMinIOImage specifies the default MinIO Docker hub image
const DefaultMinIOImage = "minio/minio:RELEASE.2019-08-21T19-40-07Z"

// MinIOServerName specifies the default container name for MinIOInstance
const MinIOServerName = "minio"

// DefaultMinIOAccessKey specifies default access key for MinIOInstance
const DefaultMinIOAccessKey = "AKIAIOSFODNN7EXAMPLE"

// DefaultMinIOSecretKey specifies default secret key for MinIOInstance
const DefaultMinIOSecretKey = "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY"

// DefaultPodManagementPolicy specifies default pod management policy as expllained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies
const DefaultPodManagementPolicy = appsv1.ParallelPodManagement

// DefaultUpdateStrategy specifies default pod update policy as explained here
// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies
const DefaultUpdateStrategy = "RollingUpdate"

// HeadlessServiceNameSuffix specifies the suffix added to MinIOInstance name to create a headless service
const HeadlessServiceNameSuffix = "-hl-svc"

// CSRNameSuffix specifies the suffix added to MinIOInstance name to create a CSR
const CSRNameSuffix = "-csr"

// Auto TLS related constants

// DefaultEllipticCurve specifies the default elliptic curve to be used for key generation
var DefaultEllipticCurve = elliptic.P256()

// DefaultOrgName specifies the default Org name to be used in automatic certificate generation
var DefaultOrgName = []string{"Acme Co"}

// DefaultQueryInterval specifies the interval between each query for CSR Status
var DefaultQueryInterval = time.Second * 5

// DefaultQueryTimeout specifies the timeout for query for CSR Status
var DefaultQueryTimeout = time.Minute * 20

// TLSSecretSuffix is the suffix applied to MinIOInstance name to create the TLS secret
var TLSSecretSuffix = "-tls"
