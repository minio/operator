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

package v1beta1

import (
	"fmt"
	"path"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"

	constants "github.com/minio/minio-operator/pkg/constants"
)

// HasCredsSecret returns true if the user has provided a secret
// for a MinIOInstance else false
func (mi *MinIOInstance) HasCredsSecret() bool {
	return mi.Spec.CredsSecret != nil
}

// HasMetadata returns true if the user has provided a object metadata
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMetadata() bool {
	return mi.Spec.Metadata != nil
}

// HasCertConfig returns true if the user has provided a certificate
// config
func (mi *MinIOInstance) HasCertConfig() bool {
	return mi.Spec.CertConfig != nil
}

// RequiresExternalCertSetup returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (mi *MinIOInstance) RequiresExternalCertSetup() bool {
	return mi.Spec.ExternalCertSecret != nil
}

// RequiresAutoCertSetup returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (mi *MinIOInstance) RequiresAutoCertSetup() bool {
	return mi.Spec.RequestAutoCert == true
}

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
// For example a user can choose to omit the version
// and number of members.
func (mi *MinIOInstance) EnsureDefaults() *MinIOInstance {
	if mi.Spec.Replicas == 0 {
		mi.Spec.Replicas = constants.DefaultReplicas
	}

	if mi.Spec.PodManagementPolicy == "" || (mi.Spec.PodManagementPolicy != appsv1.OrderedReadyPodManagement &&
		mi.Spec.PodManagementPolicy != appsv1.ParallelPodManagement) {
		mi.Spec.PodManagementPolicy = constants.DefaultPodManagementPolicy
	}

	if mi.Spec.Image == "" {
		mi.Spec.Image = constants.DefaultMinIOImage
	}

	if mi.Spec.Mountpath == "" {
		mi.Spec.Mountpath = constants.MinIOVolumeMountPath
	} else {
		// Ensure there is no trailing `/`
		mi.Spec.Mountpath = path.Clean(mi.Spec.Mountpath)
	}

	if mi.Spec.Subpath == "" {
		mi.Spec.Subpath = constants.MinIOVolumeSubPath
	} else {
		// Ensure there is no `/` in beginning
		mi.Spec.Subpath = path.Clean(mi.Spec.Subpath)
	}

	if mi.RequiresAutoCertSetup() == true {
		if mi.Spec.CertConfig != nil {
			if mi.Spec.CertConfig.CommonName == "" {
				mi.Spec.CertConfig.CommonName = mi.GetWildCardName()
			}
			if mi.Spec.CertConfig.DNSNames == nil {
				mi.Spec.CertConfig.DNSNames = mi.GetHosts()
			}
			if mi.Spec.CertConfig.OrganizationName == nil {
				mi.Spec.CertConfig.OrganizationName = constants.DefaultOrgName
			}
		} else {
			mi.Spec.CertConfig = &CertificateConfig{
				CommonName:       mi.GetWildCardName(),
				DNSNames:         mi.GetHosts(),
				OrganizationName: constants.DefaultOrgName,
			}
		}
	}

	return mi
}

// GetHosts returns the domain names managed by headless service created for
// current MinIOInstance
func (mi *MinIOInstance) GetHosts() []string {
	hosts := make([]string, 0)
	// append all the MinIOInstance replica URLs
	// mi.Name is the headless service name
	for i := 0; i < int(mi.Spec.Replicas); i++ {
		hosts = append(hosts, fmt.Sprintf("%s-"+strconv.Itoa(i)+".%s.%s.svc.cluster.local", mi.Name, mi.GetHeadlessServiceName(), mi.Namespace))
	}
	return hosts
}

// GetWildCardName returns the wild card name managed by headless service created for
// current MinIOInstance
func (mi *MinIOInstance) GetWildCardName() string {
	// mi.Name is the headless service name
	return fmt.Sprintf("*.%s.%s.svc.cluster.local", mi.GetHeadlessServiceName(), mi.Namespace)
}

// GetTLSSecretName returns the name of Secret that has TLS related Info (Cert & Prviate Key)
func (mi *MinIOInstance) GetTLSSecretName() string {
	return mi.Name + constants.TLSSecretSuffix
}

// GetHeadlessServiceName returns the name of headless service that is created to manage the
// StatefulSet of this MinIOInstance
func (mi *MinIOInstance) GetHeadlessServiceName() string {
	return mi.Name + constants.HeadlessServiceNameSuffix
}

// GetCSRName returns the name of CSR that generated if AutoTLS is enabled
func (mi *MinIOInstance) GetCSRName() string {
	return mi.Name + constants.CSRNameSuffix
}
