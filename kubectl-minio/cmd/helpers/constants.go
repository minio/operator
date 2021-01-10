/*
 * This file is part of MinIO Operator
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

package helpers

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// ClusterRoleBindingName is the name for CRB
	ClusterRoleBindingName = "minio-operator-binding"

	// ClusterRoleName is the name for Cluster Role for operator
	ClusterRoleName = "minio-operator-role"

	// ContainerName is the name of operator container
	ContainerName = "minio-operator"

	// DeploymentName is the name of operator deployment
	DeploymentName = "minio-operator"

	// DefaultNamespace is the default namespace for all operations
	DefaultNamespace = "minio-operator"

	// DefaultStorageclass is the default storage class for MinIO Tenant PVC
	DefaultStorageclass = "default"

	// DefaultServiceAccount is the service account for all
	DefaultServiceAccount = "minio-operator"

	// DefaultClusterDomain is the default domain of the Kubernetes cluster
	DefaultClusterDomain = "cluster.local"

	// DefaultSecretNameSuffix is the suffix added to tenant name to create the
	// credentials secret for this tenant
	DefaultSecretNameSuffix = "-creds-secret"

	// DefaultServiceNameSuffix is the suffix added to tenant name to create the
	// internal clusterIP service for this tenant
	DefaultServiceNameSuffix = "-internal-service"

	// MinIOPrometheusPath is the path where MinIO tenant exposes Prometheus metrics
	MinIOPrometheusPath = "/minio/prometheus/metrics"

	// MinIOPrometheusPort is the port where MinIO tenant exposes Prometheus metrics
	MinIOPrometheusPort = "9000"

	// MinIOMountPath is the path where MinIO related PVs are mounted in a container
	MinIOMountPath = "/export"

	// MinIOAccessMode is the default access mode to be used for PVC / PV binding
	MinIOAccessMode = "ReadWriteOnce"

	// DefaultImagePullPolicy specifies the policy to image pulls
	DefaultImagePullPolicy = corev1.PullIfNotPresent

	// DefaultOperatorImage is the default operator image to be used
	DefaultOperatorImage = "minio/operator:v4.0.0"

	// DefaultTenantImage is the default MinIO image used while creating tenant
	DefaultTenantImage = "minio/minio:RELEASE.2021-01-16T02-19-44Z"

	// DefaultKESImage is the default KES image used while creating tenant
	DefaultKESImage = "minio/kes:v0.13.4"

	// DefaultConsoleImage is the default console image used while creating tenant
	DefaultConsoleImage = "minio/console:v0.5.2"
)

var (
	// DeploymentReplicas is the number of replicas for MinIO Operator
	DeploymentReplicas int32 = 1

	// KESReplicas is the number of replicas for MinIO KES
	KESReplicas int32 = 2

	// ConsoleReplicas is the number of replicas for MinIO Console
	ConsoleReplicas int32 = 2
)
