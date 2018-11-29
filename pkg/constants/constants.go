/*
 * Minio-Operator - Manage Minio clusters in Kubernetes
 *
 * Minio (C) 2018 Minio, Inc.
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

// InstanceLabel is applied to all components of a MinioInstance cluster
const InstanceLabel = "v1beta1.minio.io/instance"

// MinioOperatorVersionLabel denotes the version of the MinioInstance operator
// running in the cluster.
const MinioOperatorVersionLabel = "v1beta1.minio.io/version"

// MinioPort specifies the default MinioInstance port number.
const MinioPort = 9000

// DefaultReplicas specifies the default Minio replicas to use for distributed deployment if not specified explicitly by user
const DefaultReplicas = 4

// MinioVolumeName specifies the default volume name for Minio volumes
const MinioVolumeName = "export"

// MinioVolumeMountPath specifies the default mount path for Minio volumes
const MinioVolumeMountPath = "/export"

// MinioVolumeSubPath specifies the default sub path under mount path
const MinioVolumeSubPath = ""

// MinioImagePath specifies the Minio Docker hub path
const MinioImagePath = "minio/minio"

// DefaultMinioImageVersion specifies the latest released Minio Docker hub image
const DefaultMinioImageVersion = "RELEASE.2018-11-22T02-51-56Z"

// MinioServerName specifies the default container name for MinioInstance
const MinioServerName = "minio"

// DefaultMinioAccessKey specifies default access key for MinioInstance
const DefaultMinioAccessKey = "AKIAIOSFODNN7EXAMPLE"

//DefaultMinioSecretKey specifies default secret key for MinioInstance
const DefaultMinioSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
