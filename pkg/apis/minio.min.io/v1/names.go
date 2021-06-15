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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// InitContainerImage name for init container.
const InitContainerImage = "busybox:1.32"

// MinIO Related Names

// MinIOStatefulSetNameForZone returns the name for MinIO StatefulSet
func (t *Tenant) MinIOStatefulSetNameForZone(z *Zone) string {
	return fmt.Sprintf("%s-%s", t.Name, z.Name)
}

// MinIOWildCardName returns the wild card name for all MinIO Pods in current StatefulSet
func (t *Tenant) MinIOWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", t.MinIOHLServiceName(), t.Namespace, ClusterDomain)
}

// MinIOTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) MinIOTLSSecretName() string {
	return t.Name + miniov2.TLSSecretSuffix
}

// MinIOClientTLSSecretName returns the name of Secret that has TLS related Info (Cert & Private Key)
// for MinIO <-> KES client side authentication.
func (t *Tenant) MinIOClientTLSSecretName() string {
	return t.Name + "-client" + miniov2.TLSSecretSuffix
}

// MinIOHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this Tenant
func (t *Tenant) MinIOHLServiceName() string {
	return t.Name + miniov2.MinIOHLSvcNameSuffix
}

// MinIOCIServiceName returns the name of Cluster IP service that is created to communicate
// with current MinIO StatefulSet pods
func (t *Tenant) MinIOCIServiceName() string {
	// DO NOT CHANGE, this should be constant
	return "minio"
}

// MinIOBucketBaseDomain returns the base domain name for buckets
func (t *Tenant) MinIOBucketBaseDomain() string {
	return fmt.Sprintf("%s.svc.%s", t.Namespace, ClusterDomain)
}

// MinIOBucketBaseWildcardDomain returns the base domain name for buckets
func (t *Tenant) MinIOBucketBaseWildcardDomain() string {
	return fmt.Sprintf("*.%s.svc.%s", t.Namespace, ClusterDomain)
}

// MinIOFQDNServiceName returns the name of the service created for the tenant.
func (t *Tenant) MinIOFQDNServiceName() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.MinIOCIServiceName(), t.Namespace, ClusterDomain)
}

// MinIOCSRName returns the name of CSR that is generated if AutoTLS is enabled
// Namespace adds uniqueness to the CSR name (single MinIO tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) MinIOCSRName() string {
	return t.Name + "-" + t.Namespace + miniov2.CSRNameSuffix
}

// MinIOClientCSRName returns the name of CSR that is generated for Client side authentication
// Used by KES Pods
func (t *Tenant) MinIOClientCSRName() string {
	return t.Name + "-client-" + t.Namespace + miniov2.CSRNameSuffix
}

// KES Related Names

// KESJobName returns the name for KES Key Job
func (t *Tenant) KESJobName() string {
	return t.Name + miniov2.KESName
}

// KESStatefulSetName returns the name for KES StatefulSet
func (t *Tenant) KESStatefulSetName() string {
	return t.Name + miniov2.KESName
}

// KESHLServiceName returns the name of headless service that is created to manage the
// StatefulSet of this Tenant
func (t *Tenant) KESHLServiceName() string {
	return t.Name + miniov2.KESHLSvcNameSuffix
}

// KESVolMountName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) KESVolMountName() string {
	return t.Name + miniov2.KESName
}

// KESWildCardName returns the wild card name managed by headless service created for
// KES StatefulSet in current Tenant
func (t *Tenant) KESWildCardName() string {
	return fmt.Sprintf("*.%s.%s.svc.%s", t.KESHLServiceName(), t.Namespace, ClusterDomain)
}

// KESTLSSecretName returns the name of Secret that has KES TLS related Info (Cert & Private Key)
func (t *Tenant) KESTLSSecretName() string {
	return t.KESStatefulSetName() + miniov2.TLSSecretSuffix
}

// KESCSRName returns the name of CSR that generated if AutoTLS is enabled for KES
// Namespace adds uniqueness to the CSR name (single KES tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) KESCSRName() string {
	return t.KESStatefulSetName() + "-" + t.Namespace + miniov2.CSRNameSuffix
}

// Console Related Names

// ConsoleDeploymentName returns the name for Console Deployment
func (t *Tenant) ConsoleDeploymentName() string {
	return t.Name + miniov2.ConsoleName
}

// ConsoleCIServiceName returns the name for Console Cluster IP Service
func (t *Tenant) ConsoleCIServiceName() string {
	return t.Name + miniov2.ConsoleName
}

// ZoneStatefulsetName returns the name of a statefulset for a given zone
func (t *Tenant) ZoneStatefulsetName(zone *Zone) string {
	return fmt.Sprintf("%s-%s", t.Name, zone.Name)
}

// ConsoleVolMountName returns the name of Secret that has TLS related Info (Cert & Private Key)
func (t *Tenant) ConsoleVolMountName() string {
	return t.Name + miniov2.ConsoleName
}

// ConsoleCommonName returns the CommonName to be used in the csr template
func (t *Tenant) ConsoleCommonName() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.ConsoleCIServiceName(), t.Namespace, ClusterDomain)
}

// ConsoleTLSSecretName returns the name of Secret that has Console TLS related Info (Cert & Private Key)
func (t *Tenant) ConsoleTLSSecretName() string {
	return t.ConsoleDeploymentName() + miniov2.TLSSecretSuffix
}

// ConsoleCSRName returns the name of CSR that generated if AutoTLS is enabled for Console
// Namespace adds uniqueness to the CSR name (single Console tenant per namsepace)
// since CSR is not a namespaced resource
func (t *Tenant) ConsoleCSRName() string {
	return t.ConsoleDeploymentName() + "-" + t.Namespace + miniov2.CSRNameSuffix
}
