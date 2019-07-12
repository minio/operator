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

// DefaultMinIOImagePath specifies the MinIO Docker hub path
const DefaultMinIOImagePath = "minio/minio"

// DefaultMinIOImageVersion specifies the latest released MinIO Docker hub image
const DefaultMinIOImageVersion = "RELEASE.2019-07-10T00-34-56Z"

// MinIOServerName specifies the default container name for MinIOInstance
const MinIOServerName = "minio"

// DefaultMinIOAccessKey specifies default access key for MinIOInstance
const DefaultMinIOAccessKey = "AKIAIOSFODNN7EXAMPLE"

//DefaultMinIOSecretKey specifies default secret key for MinIOInstance
const DefaultMinIOSecretKey = "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY"
