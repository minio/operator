// This file is part of MinIO Operator
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

package helpers

const (
	// DefaultNamespace is the default namespace for all operations
	DefaultNamespace = "minio-operator"

	// DefaultStorageclass is empty so the cluster pick its own native storage class as default
	DefaultStorageclass = ""

	// DefaultClusterDomain is the default domain of the Kubernetes cluster
	DefaultClusterDomain = "cluster.local"

	// DefaultServiceNameSuffix is the suffix added to tenant name to create the
	// internal clusterIP service for this tenant
	DefaultServiceNameSuffix = "-internal-service"

	// MinIOMountPath is the path where MinIO related PVs are mounted in a container
	MinIOMountPath = "/export"

	// MinIOAccessMode is the default access mode to be used for PVC / PV binding
	MinIOAccessMode = "ReadWriteOnce"

	// DefaultOperatorImage is the default operator image to be used
	DefaultOperatorImage = "minio/operator:v5.0.6"

	// DefaultTenantImage is the default MinIO image used while creating tenant
	DefaultTenantImage = "minio/minio:RELEASE.2023-06-23T20-26-00Z"

	// DefaultKESImage is the default KES image used while creating tenant
	DefaultKESImage = "minio/kes:2023-05-02T22-48-10Z"
)

// KESReplicas is the number of replicas for MinIO KES
var KESReplicas int32 = 2
