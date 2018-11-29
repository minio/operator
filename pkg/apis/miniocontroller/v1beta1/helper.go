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

package v1beta1

import (
	"path"

	constants "github.com/minio/minio-operator/pkg/constants"
)

// HasCredsSecret returns true if the user has provided a secret
// for a MinioInstance else false
func (mi *MinioInstance) HasCredsSecret() bool {
	return mi.Spec.CredsSecret != nil
}

// RequiresSSLSetup returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (mi *MinioInstance) RequiresSSLSetup() bool {
	return mi.Spec.SSLSecret != nil
}

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
// For example a user can choose to omit the version
// and number of members.
func (mi *MinioInstance) EnsureDefaults() *MinioInstance {
	if mi.Spec.Replicas == 0 {
		mi.Spec.Replicas = constants.DefaultReplicas
	}

	if mi.Spec.Version == "" {
		mi.Spec.Version = constants.DefaultMinioImageVersion
	}

	if mi.Spec.Mountpath == "" {
		mi.Spec.Mountpath = constants.MinioVolumeMountPath
	} else {
		// Ensure there is no trailing `/`
		mi.Spec.Mountpath = path.Clean(mi.Spec.Mountpath)
	}

	if mi.Spec.Subpath == "" {
		mi.Spec.Subpath = constants.MinioVolumeSubPath
	} else {
		// Ensure there is no `/` in beginning
		mi.Spec.Subpath = path.Clean(mi.Spec.Subpath)
	}

	return mi
}
