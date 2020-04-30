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
	"context"
	"errors"
	"fmt"
	"net"
	"path"
	"strconv"

	"github.com/golang/glog"

	constants "github.com/minio/minio-operator/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/minio/minio/pkg/bucket/policy"
	"github.com/minio/minio/pkg/bucket/policy/condition"
	iampolicy "github.com/minio/minio/pkg/iam/policy"
	"github.com/minio/minio/pkg/madmin"
)

// HasCredsSecret returns true if the user has provided a secret
// for a MinIOInstance else false
func (mi *MinIOInstance) HasCredsSecret() bool {
	return mi.Spec.CredsSecret != nil
}

// HasMetadata returns true if the user has provided a pod metadata
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMetadata() bool {
	return mi.Spec.Metadata != nil
}

// HasSelector returns true if the user has provided a pod selector
// field (ref: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-selector)
func (mi *MinIOInstance) HasSelector() bool {
	return mi.Spec.Selector != nil
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

// PodLabels returns the default labels
func (mi *MinIOInstance) PodLabels() map[string]string {
	m := make(map[string]string, 1)
	m[constants.InstanceLabel] = mi.Name
	return m
}

// GetVolumesPath returns the paths for MinIO mounts based on
// total number of volumes per MinIO server
func (mi *MinIOInstance) GetVolumesPath() string {
	if mi.Spec.VolumesPerServer == 1 {
		return path.Join(mi.Spec.Mountpath, mi.Spec.Subpath)
	}
	return path.Join(mi.Spec.Mountpath+"{0..."+strconv.Itoa((mi.Spec.VolumesPerServer)-1)+"}", mi.Spec.Subpath)
}

// GetReplicas returns the number of total replicas
// required for this cluster
func (mi *MinIOInstance) GetReplicas() int32 {
	var replicas int32
	for _, z := range mi.Spec.Zones {
		replicas = replicas + z.Servers
	}
	return replicas
}

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
// For example a user can choose to omit the version
// and number of members.
func (mi *MinIOInstance) EnsureDefaults() *MinIOInstance {
	if mi.Spec.PodManagementPolicy == "" || (mi.Spec.PodManagementPolicy != appsv1.OrderedReadyPodManagement &&
		mi.Spec.PodManagementPolicy != appsv1.ParallelPodManagement) {
		mi.Spec.PodManagementPolicy = constants.DefaultPodManagementPolicy
	}

	if mi.Spec.Image == "" {
		mi.Spec.Image = constants.DefaultMinIOImage
	}

	for _, z := range mi.Spec.Zones {
		if z.Servers == 0 {
			z.Servers = constants.DefaultServers
		}
	}

	if mi.Spec.VolumesPerServer == 0 {
		mi.Spec.VolumesPerServer = constants.DefaultVolumesPerServer
	}

	if mi.Spec.Mountpath == "" {
		mi.Spec.Mountpath = constants.MinIOVolumeMountPath
	}

	if mi.Spec.Subpath == "" {
		mi.Spec.Subpath = constants.MinIOVolumeSubPath
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

	if !mi.HasSelector() {
		mi.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: mi.PodLabels(),
		}
	}

	if mi.HasMcsEnabled() && mi.Spec.Mcs.Image == "" {
		mi.Spec.Mcs.Image = constants.DefaultMcsImage
	}

	return mi
}

// GetHosts returns the domain names managed by headless service created for
// current MinIOInstance
func (mi *MinIOInstance) GetHosts() []string {
	hosts := make([]string, 0)
	var max, index int32
	// Create the ellipses style URL
	// mi.Name is the headless service name
	for _, z := range mi.Spec.Zones {
		max = max + z.Servers
		hosts = append(hosts, fmt.Sprintf("%s-{"+strconv.Itoa(int(index))+"..."+strconv.Itoa(int(max)-1)+"}.%s.%s.svc."+constants.ClusterDomain, mi.Name, mi.GetHeadlessServiceName(), mi.Namespace))
		index = max
	}
	return hosts
}

// GetServiceHost returns headless service Host.
// current MinIOInstance
func (mi *MinIOInstance) GetServiceHost() string {
	if mi.Spec.Zones[0].Servers == 1 {
		msg := "Please set the server count > 1"
		glog.V(2).Infof(msg)
		return ""
	}
	hostStr := fmt.Sprintf("%s.%s.svc."+constants.ClusterDomain, mi.GetHeadlessServiceName(), mi.Namespace)
	return net.JoinHostPort(hostStr, strconv.Itoa(constants.MinIOServicePortNumber))
}

// GetWildCardName returns the wild card name managed by headless service created for
// current MinIOInstance
func (mi *MinIOInstance) GetWildCardName() string {
	// mi.Name is the headless service name
	return fmt.Sprintf("*.%s.%s.svc."+constants.ClusterDomain, mi.GetHeadlessServiceName(), mi.Namespace)
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

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
func (mi *MirrorInstance) EnsureDefaults() *MirrorInstance {
	if mi.Spec.Image == "" {
		mi.Spec.Image = constants.DefaultMCImage
	}
	return mi
}

// HasMetadata returns true if the user has provided a pod metadata
// for a MinIOInstance else false
func (mi *MirrorInstance) HasMetadata() bool {
	return mi.Spec.Metadata != nil
}

// HasSelector returns true if the user has provided a pod selector
// field (ref: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-selector)
func (mi *MirrorInstance) HasSelector() bool {
	return mi.Spec.Selector != nil
}

// HasMcsEnabled checks if the mcs has been enabled by the user
func (mi *MinIOInstance) HasMcsEnabled() bool {
	return mi.Spec.Mcs != nil
}

// HasMcsSecret returns true if the user has provided an mcs secret
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMcsSecret() bool {
	return mi.Spec.Mcs != nil && mi.Spec.Mcs.McsSecret != nil
}

// HasMcsMetadata returns true if the user has provided a mcs metadata
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMcsMetadata() bool {
	return mi.Spec.Mcs != nil && mi.Spec.Mcs.Metadata != nil
}

// HasMcsSelector returns true if the user has provided a mcs selector
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMcsSelector() bool {
	return mi.Spec.Mcs != nil && mi.Spec.Mcs.Selector != nil
}

// CreateMcsUser function creates an admin user
func (mi *MinIOInstance) CreateMcsUser(minioSecret, mcsSecret map[string][]byte) error {

	var accessKey, secretKey, mcsSecretKey []byte
	var ok bool

	host := mi.GetServiceHost()
	if host == "" {
		return errors.New("MCS MINIO SERVER is empty")
	}

	accessKey, ok = minioSecret["accesskey"]
	if !ok {
		return errors.New("accesskey not provided")
	}

	secretKey, ok = minioSecret["secretkey"]
	if !ok {
		return errors.New("secretkey not provided")
	}

	mcsSecretKey, ok = mcsSecret["mcssecretkey"]
	if !ok {
		return errors.New("mcssecretkey not provided")
	}

	madmClnt, err := madmin.New(host, string(accessKey), string(secretKey), false)
	if err != nil {
		return err
	}

	if err = madmClnt.AddUser(context.Background(), string(mi.Spec.Mcs.McsAccessKey), string(mcsSecretKey)); err != nil {
		return err
	}

	// Create policy
	p := iampolicy.Policy{
		Version: iampolicy.DefaultVersion,
		Statements: []iampolicy.Statement{
			{
				SID:        policy.ID(""),
				Effect:     policy.Allow,
				Actions:    iampolicy.NewActionSet(iampolicy.AllAdminActions),
				Resources:  iampolicy.NewResourceSet(),
				Conditions: condition.NewFunctions(),
			},
			{
				SID:        policy.ID(""),
				Effect:     policy.Allow,
				Actions:    iampolicy.NewActionSet(iampolicy.AllActions),
				Resources:  iampolicy.NewResourceSet(iampolicy.NewResource("*", "")),
				Conditions: condition.NewFunctions(),
			},
		},
	}

	if err = madmClnt.AddCannedPolicy(context.Background(), constants.MCSAdminPolicyName, &p); err != nil {
		return err
	}

	if err = madmClnt.SetPolicy(context.Background(), constants.MCSAdminPolicyName, string(mi.Spec.Mcs.McsAccessKey), false); err != nil {
		return err
	}

	return nil
}
