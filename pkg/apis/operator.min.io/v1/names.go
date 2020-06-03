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
	"fmt"
)

// MinIOServerName specifies the default container name for MinIOInstance
const MinIOServerName = "minio"

// KESContainerName specifies the default container name for KES
const KESContainerName = "kes"

// MCSContainerName specifies the default container name for MCS
const MCSContainerName = "mcs"

// MinIO Related Names

// MinIOStatefulSetName returns the name for MinIO StatefulSet
func (mi *MinIOInstance) MinIOStatefulSetName() string {
	return mi.Name
}

// MinIOWildCardName returns the wild card name for all MinIO Pods in current StatefulSet
func (mi *MinIOInstance) MinIOWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", mi.MinIOHLServiceName(), mi.Namespace, ClusterDomain)
}

// MinIOTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (mi *MinIOInstance) MinIOTLSSecretName() string {
	return mi.Name + TLSSecretSuffix
}

// MinIOClientTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
// for MinIO <-> KES client side authentication.
func (mi *MinIOInstance) MinIOClientTLSSecretName() string {
	return mi.Name + "-client" + TLSSecretSuffix
}

// MinIOHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this MinIOInstance
func (mi *MinIOInstance) MinIOHLServiceName() string {
	return mi.Name + MinIOHLSvcNameSuffix
}

// MinIOCIServiceName returns the name of Cluster IP service that is created to communicate
// with current MinIO StatefulSet pods
func (mi *MinIOInstance) MinIOCIServiceName() string {
	return mi.Spec.ServiceName
}

// MinIOCSRName returns the name of CSR that is generated if AutoTLS is enabled
func (mi *MinIOInstance) MinIOCSRName() string {
	return mi.Name + CSRNameSuffix
}

// MinIOClientCSRName returns the name of CSR that is generated for Client side authentication
// Used by KES Pods
func (mi *MinIOInstance) MinIOClientCSRName() string {
	return mi.Name + "-client" + CSRNameSuffix
}

// KES Related Names

// KESJobName returns the name for KES Key Job
func (mi *MinIOInstance) KESJobName() string {
	return mi.Name + KESName
}

// KESStatefulSetName returns the name for KES StatefulSet
func (mi *MinIOInstance) KESStatefulSetName() string {
	return mi.Name + KESName
}

// KESHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this MinIOInstance
func (mi *MinIOInstance) KESHLServiceName() string {
	return mi.Name + KESHLSvcNameSuffix
}

// KESVolMountName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (mi *MinIOInstance) KESVolMountName() string {
	return mi.Name + KESName
}

// KESWildCardName returns the wild card name managed by headless service created for
// KES StatefulSet in current MinIOInstance
func (mi *MinIOInstance) KESWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", mi.KESHLServiceName(), mi.Namespace, ClusterDomain)
}

// KESTLSSecretName returns the name of Secret that has KES TLS related Info (Cert & Private Key)
func (mi *MinIOInstance) KESTLSSecretName() string {
	return mi.KESStatefulSetName() + TLSSecretSuffix
}

// KESCSRName returns the name of CSR that generated if AutoTLS is enabled for KES
func (mi *MinIOInstance) KESCSRName() string {
	return mi.KESStatefulSetName() + CSRNameSuffix
}

// MCS Related Names

// MCSDeploymentName returns the name for MCS Deployment
func (mi *MinIOInstance) MCSDeploymentName() string {
	return mi.Name + MCSName
}

// MCSCIServiceName returns the name for MCS Cluster IP Service
func (mi *MinIOInstance) MCSCIServiceName() string {
	return mi.Name + MCSName
}
