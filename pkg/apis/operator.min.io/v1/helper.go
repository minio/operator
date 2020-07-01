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
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"text/template"
	"time"

	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	"github.com/minio/minio/pkg/bucket/policy"
	"github.com/minio/minio/pkg/bucket/policy/condition"
	iampolicy "github.com/minio/minio/pkg/iam/policy"
	"github.com/minio/minio/pkg/madmin"
)

type hostsTemplateValues struct {
	StatefulSet string
	CIService   string
	HLService   string
	Ellipsis    string
	Domain      string
}

// ellipsis returns the host range string
func ellipsis(start, end int) string {
	return "{" + strconv.Itoa(start) + "..." + strconv.Itoa(end) + "}"
}

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

// HasCertConfig returns true if the user has provided a certificate
// config
func (mi *MinIOInstance) HasCertConfig() bool {
	return mi.Spec.CertConfig != nil
}

// ExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key
func (mi *MinIOInstance) ExternalCert() bool {
	return mi.Spec.ExternalCertSecret != nil
}

// ExternalClientCert returns true is the user has provided a secret
// that contains CA client cert, server cert and server key
func (mi *MinIOInstance) ExternalClientCert() bool {
	return mi.Spec.ExternalClientCertSecret != nil
}

// KESExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key for KES pods
func (mi *MinIOInstance) KESExternalCert() bool {
	return mi.Spec.KES != nil && mi.Spec.KES.ExternalCertSecret != nil
}

// AutoCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (mi *MinIOInstance) AutoCert() bool {
	return mi.Spec.RequestAutoCert
}

// VolumePath returns the paths for MinIO mounts based on
// total number of volumes per MinIO server
func (mi *MinIOInstance) VolumePath() string {
	if mi.Spec.VolumesPerServer == 1 {
		return path.Join(mi.Spec.Mountpath, mi.Spec.Subpath)
	}
	return path.Join(mi.Spec.Mountpath+ellipsis(0, mi.Spec.VolumesPerServer-1), mi.Spec.Subpath)
}

// MinIOReplicas returns the number of total replicas
// required for this cluster
func (mi *MinIOInstance) MinIOReplicas() int32 {
	var replicas int32
	for _, z := range mi.Spec.Zones {
		replicas = replicas + z.Servers
	}
	return replicas
}

// KESReplicas returns the number of total KES replicas
// required for this cluster
func (mi *MinIOInstance) KESReplicas() int32 {
	var replicas int32
	if mi.Spec.KES != nil && mi.Spec.KES.Replicas != 0 {
		replicas = mi.Spec.KES.Replicas
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
		mi.Spec.PodManagementPolicy = DefaultPodManagementPolicy
	}

	if mi.Spec.Image == "" {
		mi.Spec.Image = DefaultMinIOImage
	}

	// Default an empty service name to the instance name
	if mi.Spec.ServiceName == "" {
		mi.Spec.ServiceName = mi.Name
	}

	for _, z := range mi.Spec.Zones {
		if z.Servers == 0 {
			z.Servers = DefaultServers
		}
	}

	if mi.Spec.VolumesPerServer == 0 {
		mi.Spec.VolumesPerServer = DefaultVolumesPerServer
	}

	if mi.Spec.Mountpath == "" {
		mi.Spec.Mountpath = MinIOVolumeMountPath
	}

	if mi.Spec.Subpath == "" {
		mi.Spec.Subpath = MinIOVolumeSubPath
	}

	if mi.AutoCert() {
		if mi.Spec.CertConfig != nil {
			if mi.Spec.CertConfig.CommonName == "" {
				mi.Spec.CertConfig.CommonName = mi.MinIOWildCardName()
			}
			if mi.Spec.CertConfig.DNSNames == nil {
				mi.Spec.CertConfig.DNSNames = mi.MinIOHosts()
			}
			if mi.Spec.CertConfig.OrganizationName == nil {
				mi.Spec.CertConfig.OrganizationName = DefaultOrgName
			}
		} else {
			mi.Spec.CertConfig = &CertificateConfig{
				CommonName:       mi.MinIOWildCardName(),
				DNSNames:         mi.MinIOHosts(),
				OrganizationName: DefaultOrgName,
			}
		}
	}

	if mi.Spec.Liveness != nil {
		if mi.Spec.Liveness.InitialDelaySeconds == 0 {
			mi.Spec.Liveness.InitialDelaySeconds = LivenessInitialDelay
		}
		if mi.Spec.Liveness.PeriodSeconds == 0 {
			mi.Spec.Liveness.PeriodSeconds = LivenessPeriod
		}
		if mi.Spec.Liveness.TimeoutSeconds == 0 {
			mi.Spec.Liveness.TimeoutSeconds = LivenessTimeout
		}
	}

	if mi.HasMCSEnabled() {
		if mi.Spec.MCS.Image == "" {
			mi.Spec.MCS.Image = DefaultMCSImage
		}
		if mi.Spec.MCS.Replicas == 0 {
			mi.Spec.MCS.Replicas = DefaultMCSReplicas
		}
	}

	if mi.HasKESEnabled() {
		if mi.Spec.KES.Image == "" {
			mi.Spec.KES.Image = DefaultKESImage
		}
		if mi.Spec.KES.Replicas == 0 {
			mi.Spec.KES.Replicas = DefaultKESReplicas
		}
	}

	return mi
}

// MinIOHosts returns the domain names in ellipses format created for current MinIOInstance
func (mi *MinIOInstance) MinIOHosts() []string {
	hosts := make([]string, 0)
	var max, index int32
	hostPostfix := mi.HostPostfix()
	// Create the ellipses style URL
	for _, z := range mi.Spec.Zones {
		max = max + z.Servers
		hosts = append(hosts, fmt.Sprintf("%s-%s.%s", mi.MinIOStatefulSetName(), ellipsis(int(index), int(max)-1), hostPostfix))
		index = max
	}
	return hosts
}

// TemplatedMinIOHosts returns the domain names in ellipses format created for current MinIOInstance without the service part
func (mi *MinIOInstance) TemplatedMinIOHosts(hostsTemplate string) []string {
	hosts := make([]string, 0)
	tmpl, err := template.New("hosts").Parse(hostsTemplate)
	if err != nil {
		msg := "Invalid go template for hosts"
		klog.V(2).Infof(msg)
		return hosts
	}
	var max, index int32
	// Create the ellipses style URL
	for _, z := range mi.Spec.Zones {
		max = max + z.Servers
		data := hostsTemplateValues{
			StatefulSet: mi.MinIOStatefulSetName(),
			CIService:   mi.MinIOCIServiceName(),
			HLService:   mi.MinIOHLServiceName(),
			Ellipsis:    ellipsis(int(index), int(max)-1),
			Domain:      ClusterDomain,
		}
		output := new(bytes.Buffer)
		if err = tmpl.Execute(output, data); err != nil {
			continue
		}
		hosts = append(hosts, output.String())
		index = max
	}
	return hosts
}

// HostPostfix returns the last part of the host `service.minio.local` used ie: `instance.service.minio.local`
func (mi *MinIOInstance) HostPostfix() string {
	return fmt.Sprintf("%s.%s", mi.MinIOHLServiceName(), ClusterDomain)
}

// AllMinIOHosts returns the all the individual domain names relevant for current MinIOInstance
func (mi *MinIOInstance) AllMinIOHosts() []string {
	topLevelHostURL := mi.HostPostfix()
	hosts := make([]string, 0)
	var max, index int32
	for _, z := range mi.Spec.Zones {
		max = max + z.Servers
		for index < max {
			hosts = append(hosts, fmt.Sprintf("%s-"+strconv.Itoa(int(index))+".%s", mi.MinIOStatefulSetName(), topLevelHostURL))
			index++
		}
	}
	hosts = append(hosts, mi.MinIOCIServiceHost())
	hosts = append(hosts, mi.MinIOHeadlessServiceHost())
	return hosts
}

// MinIOCIServiceHost returns ClusterIP service Host for current MinIOInstance
func (mi *MinIOInstance) MinIOCIServiceHost() string {
	if mi.Spec.Zones[0].Servers == 1 {
		msg := "Please set the server count > 1"
		klog.V(2).Infof(msg)
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.%s", mi.MinIOCIServiceName(), mi.Namespace, ClusterDomain)
}

// MinIOHeadlessServiceHost returns headless service Host for current MinIOInstance
func (mi *MinIOInstance) MinIOHeadlessServiceHost() string {
	if mi.Spec.Zones[0].Servers == 1 {
		msg := "Please set the server count > 1"
		klog.V(2).Infof(msg)
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.%s", mi.MinIOHLServiceName(), mi.Namespace, ClusterDomain)
}

// KESHosts returns the host names created for current KES StatefulSet
func (mi *MinIOInstance) KESHosts() []string {
	hosts := make([]string, 0)
	var i int32 = 0
	for i < mi.Spec.KES.Replicas {
		hosts = append(hosts, fmt.Sprintf("%s-"+strconv.Itoa(int(i))+".%s.%s.svc.%s", mi.KESStatefulSetName(), mi.KESHLServiceName(), mi.Namespace, ClusterDomain))
		i++
	}
	hosts = append(hosts, mi.KESServiceHost())
	return hosts
}

// KESServiceHost returns headless service Host for KES in current MinIOInstance
func (mi *MinIOInstance) KESServiceHost() string {
	return fmt.Sprintf("%s.%s.svc.%s", mi.KESHLServiceName(), mi.Namespace, ClusterDomain)
}

// HasKESEnabled checks if kes configuration is provided by user
func (mi *MinIOInstance) HasKESEnabled() bool {
	return mi.Spec.KES != nil
}

// HasMCSEnabled checks if the mcs has been enabled by the user
func (mi *MinIOInstance) HasMCSEnabled() bool {
	return mi.Spec.MCS != nil
}

// HasMCSSecret returns true if the user has provided an mcs secret
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMCSSecret() bool {
	return mi.Spec.MCS != nil && mi.Spec.MCS.MCSSecret != nil
}

// HasMCSMetadata returns true if the user has provided a mcs metadata
// for a MinIOInstance else false
func (mi *MinIOInstance) HasMCSMetadata() bool {
	return mi.Spec.MCS != nil && mi.Spec.MCS.Metadata != nil
}

// HasKESMetadata returns true if the user has provided KES metadata
// for a MinIOInstance else false
func (mi *MinIOInstance) HasKESMetadata() bool {
	return mi.Spec.KES != nil && mi.Spec.KES.Metadata != nil
}

// CreateMCSUser function creates an admin user
func (mi *MinIOInstance) CreateMCSUser(minioSecret, mcsSecret map[string][]byte) error {

	var accessKey, secretKey, mcsAccessKey, mcsSecretKey []byte
	var ok bool

	host := net.JoinHostPort(mi.MinIOCIServiceHost(), strconv.Itoa(MinIOPort))
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

	mcsAccessKey, ok = mcsSecret["MCS_ACCESS_KEY"]
	if !ok {
		return errors.New("MCS_ACCESS_KEY not provided")
	}

	mcsSecretKey, ok = mcsSecret["MCS_SECRET_KEY"]
	if !ok {
		return errors.New("MCS_SECRET_KEY not provided")
	}

	madmClnt, err := madmin.New(host, string(accessKey), string(secretKey), Scheme == "https")
	if err != nil {
		return err
	}

	if Scheme == "https" {
		madmClnt = setUpInsecureTLS(madmClnt)
	}

	// add user with a 20 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	if err = madmClnt.AddUser(ctx, string(mcsAccessKey), string(mcsSecretKey)); err != nil {
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

	if err = madmClnt.AddCannedPolicy(context.Background(), MCSAdminPolicyName, &p); err != nil {
		return err
	}

	if err = madmClnt.SetPolicy(context.Background(), MCSAdminPolicyName, string(mcsAccessKey), false); err != nil {
		return err
	}

	return nil
}

// Validate returns an error if any configuration of the MinIO instance is invalid
func (mi *MinIOInstance) Validate() error {
	// Make sure the storage request is not 0
	if mi.Spec.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value() <= 0 {
		return errors.New("volume size must be greater than 0")
	}
	// Make sure the replicas are not 0 on any zone
	for _, z := range mi.Spec.Zones {
		if z.Servers == 0 {
			return fmt.Errorf("zone '%s' cannot have 0 servers", z.Name)
		}
	}

	return nil
}

// Set up admin client to use self certificates
func setUpInsecureTLS(api *madmin.AdminClient) *madmin.AdminClient {
	// Keep TLS config.
	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	var transport http.RoundTripper = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 15 * time.Second,
		}).DialContext,
		TLSClientConfig: tlsConfig,
	}

	// Set custom transport.
	api.SetCustomTransport(transport)
	return api
}

// OwnerRef returns the OwnerReference to be added to all resources created by MinIOInstance
func (mi *MinIOInstance) OwnerRef() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(mi, schema.GroupVersionKind{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Kind:    MinIOCRDResourceKind,
		}),
	}
}

// ProbeDisco probe Service Discovery via the `probe.minio.local` domain
func (mi *MinIOInstance) ProbeDisco(clientset kubernetes.Interface) (string, error) {
	// if we get a response, it means disco is
	// properly configured in the cluster, else we fallback to get the `minio-disco` service IP and configure
	// that on the dnsPolicy for the pods of the statefulset
	var discoSvcIP string
	_, err := net.LookupIP(discoProbeDomain)
	if err != nil {
		discoSvc, err := clientset.
			CoreV1().
			Services(mi.Namespace).
			Get(context.Background(), discoSvcName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		discoSvcIP = discoSvc.Spec.ClusterIP
	}
	return discoSvcIP, nil
}
