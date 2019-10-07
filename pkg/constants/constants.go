/*
 * Copyright (C) 2019, MinIO, Inc.
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
const DefaultMinIOImage = "minio/minio:RELEASE.2019-09-11T19-53-16Z"

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
