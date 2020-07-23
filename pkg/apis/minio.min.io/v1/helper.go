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
// for a Tenant else false
func (t *Tenant) HasCredsSecret() bool {
	return t.Spec.CredsSecret != nil
}

// HasMetadata returns true if the user has provided a pod metadata
// for a Tenant else false
func (t *Tenant) HasMetadata() bool {
	return t.Spec.Metadata != nil
}

// HasCertConfig returns true if the user has provided a certificate
// config
func (t *Tenant) HasCertConfig() bool {
	return t.Spec.CertConfig != nil
}

// ExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key
func (t *Tenant) ExternalCert() bool {
	return t.Spec.ExternalCertSecret != nil
}

// ExternalClientCert returns true is the user has provided a secret
// that contains CA client cert, server cert and server key
func (t *Tenant) ExternalClientCert() bool {
	return t.Spec.ExternalClientCertSecret != nil
}

// KESExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key for KES pods
func (t *Tenant) KESExternalCert() bool {
	return t.Spec.KES != nil && t.Spec.KES.ExternalCertSecret != nil
}

// AutoCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (t *Tenant) AutoCert() bool {
	return t.Spec.RequestAutoCert
}

// VolumePathForZone returns the paths for MinIO mounts based on
// total number of volumes on a given zone
func (t *Tenant) VolumePathForZone(zone *Zone) string {
	if zone.VolumesPerServer == 1 {
		return path.Join(t.Spec.Mountpath, t.Spec.Subpath)
	}
	return path.Join(t.Spec.Mountpath+ellipsis(0, int(zone.VolumesPerServer-1)), t.Spec.Subpath)
}

// KESReplicas returns the number of total KES replicas
// required for this cluster
func (t *Tenant) KESReplicas() int32 {
	var replicas int32
	if t.Spec.KES != nil && t.Spec.KES.Replicas != 0 {
		replicas = t.Spec.KES.Replicas
	}
	return replicas
}

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
// For example a user can choose to omit the version
// and number of members.
func (t *Tenant) EnsureDefaults() *Tenant {
	if t.Spec.PodManagementPolicy == "" || (t.Spec.PodManagementPolicy != appsv1.OrderedReadyPodManagement &&
		t.Spec.PodManagementPolicy != appsv1.ParallelPodManagement) {
		t.Spec.PodManagementPolicy = DefaultPodManagementPolicy
	}

	if t.Spec.Image == "" {
		t.Spec.Image = DefaultMinIOImage
	}

	// Default an empty service name to the instance name
	if t.Spec.ServiceName == "" {
		t.Spec.ServiceName = t.Name
	}

	for zi, z := range t.Spec.Zones {
		if z.Name == "" {
			z.Name = fmt.Sprintf("zone-%d", zi)
		}
		t.Spec.Zones[zi] = z
	}

	if t.Spec.Mountpath == "" {
		t.Spec.Mountpath = MinIOVolumeMountPath
	}

	if t.Spec.Subpath == "" {
		t.Spec.Subpath = MinIOVolumeSubPath
	}

	if t.AutoCert() {
		if t.Spec.CertConfig != nil {
			if t.Spec.CertConfig.CommonName == "" {
				t.Spec.CertConfig.CommonName = t.MinIOWildCardName()
			}
			if t.Spec.CertConfig.DNSNames == nil {
				t.Spec.CertConfig.DNSNames = t.MinIOHosts()
			}
			if t.Spec.CertConfig.OrganizationName == nil {
				t.Spec.CertConfig.OrganizationName = DefaultOrgName
			}
		} else {
			t.Spec.CertConfig = &CertificateConfig{
				CommonName:       t.MinIOWildCardName(),
				DNSNames:         t.MinIOHosts(),
				OrganizationName: DefaultOrgName,
			}
		}
	}

	if t.Spec.Liveness != nil {
		if t.Spec.Liveness.InitialDelaySeconds == 0 {
			t.Spec.Liveness.InitialDelaySeconds = LivenessInitialDelay
		}
		if t.Spec.Liveness.PeriodSeconds == 0 {
			t.Spec.Liveness.PeriodSeconds = LivenessPeriod
		}
		if t.Spec.Liveness.TimeoutSeconds == 0 {
			t.Spec.Liveness.TimeoutSeconds = LivenessTimeout
		}
	}

	if t.HasConsoleEnabled() {
		if t.Spec.Console.Image == "" {
			t.Spec.Console.Image = DefaultConsoleImage
		}
		if t.Spec.Console.Replicas == 0 {
			t.Spec.Console.Replicas = DefaultMCSReplicas
		}
	}

	if t.HasKESEnabled() {
		if t.Spec.KES.Image == "" {
			t.Spec.KES.Image = DefaultKESImage
		}
		if t.Spec.KES.Replicas == 0 {
			t.Spec.KES.Replicas = DefaultKESReplicas
		}
	}

	return t
}

// MinIOHosts returns the domain names in ellipses format created for current Tenant
func (t *Tenant) MinIOHosts() []string {
	hosts := make([]string, 0)
	// Create the ellipses style URL
	for _, z := range t.Spec.Zones {
		if z.Servers == 1 {
			hosts = append(hosts, fmt.Sprintf("%s-%s.%s.%s.svc.%s", t.MinIOStatefulSetNameForZone(&z), "0", t.MinIOHLServiceName(), t.Namespace, ClusterDomain))
		} else {
			hosts = append(hosts, fmt.Sprintf("%s-%s.%s.%s.svc.%s", t.MinIOStatefulSetNameForZone(&z), ellipsis(0, int(z.Servers)-1), t.MinIOHLServiceName(), t.Namespace, ClusterDomain))
		}
	}
	return hosts
}

// TemplatedMinIOHosts returns the domain names in ellipses format created for current Tenant without the service part
func (t *Tenant) TemplatedMinIOHosts(hostsTemplate string) []string {
	hosts := make([]string, 0)
	tmpl, err := template.New("hosts").Parse(hostsTemplate)
	if err != nil {
		msg := "Invalid go template for hosts"
		klog.V(2).Infof(msg)
		return hosts
	}
	var max, index int32
	// Create the ellipses style URL
	for _, z := range t.Spec.Zones {
		max = max + z.Servers
		data := hostsTemplateValues{
			StatefulSet: t.MinIOStatefulSetNameForZone(&z),
			CIService:   t.MinIOCIServiceName(),
			HLService:   t.MinIOHLServiceName(),
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

// AllMinIOHosts returns the all the individual domain names relevant for current Tenant
func (t *Tenant) AllMinIOHosts() []string {
	hosts := make([]string, 0)
	hosts = append(hosts, t.MinIOCIServiceHost())
	hosts = append(hosts, "*."+t.MinIOHeadlessServiceHost())
	return hosts
}

// MinIOCIServiceHost returns ClusterIP service Host for current Tenant
func (t *Tenant) MinIOCIServiceHost() string {
	if t.Spec.Zones[0].Servers == 1 {
		msg := "Please set the server count > 1"
		klog.V(2).Infof(msg)
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.%s", t.MinIOCIServiceName(), t.Namespace, ClusterDomain)
}

// MinIOHeadlessServiceHost returns headless service Host for current Tenant
func (t *Tenant) MinIOHeadlessServiceHost() string {
	if t.Spec.Zones[0].Servers == 1 {
		msg := "Please set the server count > 1"
		klog.V(2).Infof(msg)
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.%s", t.MinIOHLServiceName(), t.Namespace, ClusterDomain)
}

// KESHosts returns the host names created for current KES StatefulSet
func (t *Tenant) KESHosts() []string {
	hosts := make([]string, 0)
	var i int32 = 0
	for i < t.Spec.KES.Replicas {
		hosts = append(hosts, fmt.Sprintf("%s-"+strconv.Itoa(int(i))+".%s.%s.svc.%s", t.KESStatefulSetName(), t.KESHLServiceName(), t.Namespace, ClusterDomain))
		i++
	}
	hosts = append(hosts, t.KESServiceHost())
	return hosts
}

// KESServiceHost returns headless service Host for KES in current Tenant
func (t *Tenant) KESServiceHost() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.KESHLServiceName(), t.Namespace, ClusterDomain)
}

// HasKESEnabled checks if kes configuration is provided by user
func (t *Tenant) HasKESEnabled() bool {
	return t.Spec.KES != nil
}

// HasConsoleEnabled checks if the mcs has been enabled by the user
func (t *Tenant) HasConsoleEnabled() bool {
	return t.Spec.Console != nil
}

// HasConsoleSecret returns true if the user has provided an mcs secret
// for a Tenant else false
func (t *Tenant) HasConsoleSecret() bool {
	return t.Spec.Console != nil && t.Spec.Console.ConsoleSecret != nil
}

// HasConsoleMetadata returns true if the user has provided a mcs metadata
// for a Tenant else false
func (t *Tenant) HasConsoleMetadata() bool {
	return t.Spec.Console != nil && t.Spec.Console.Metadata != nil
}

// HasKESMetadata returns true if the user has provided KES metadata
// for a Tenant else false
func (t *Tenant) HasKESMetadata() bool {
	return t.Spec.KES != nil && t.Spec.KES.Metadata != nil
}

// CreateMCSUser function creates an admin user
func (t *Tenant) CreateMCSUser(minioSecret, mcsSecret map[string][]byte) error {

	var accessKey, secretKey, mcsAccessKey, mcsSecretKey []byte
	var ok bool

	host := net.JoinHostPort(t.MinIOCIServiceHost(), strconv.Itoa(MinIOPort))
	if host == "" {
		return errors.New("Console MINIO SERVER is empty")
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
func (t *Tenant) Validate() error {
	if t.Spec.Zones == nil {
		return errors.New("zones must be configured")
	}
	// every zone must contain a Volume Claim Template
	for zi, zone := range t.Spec.Zones {
		// Make sure the replicas are not 0 on any zone
		if zone.Servers == 0 {
			return fmt.Errorf("zone #%d cannot have 0 servers", zi)
		}
		// Make sure the zones don't have 0 volumes
		if zone.VolumesPerServer == 0 {
			return fmt.Errorf("zone #%d cannot have 0 volumes per server", zi)
		}
		// Distributed Setup can't have 2 or 3 disks only. Either 1 or 4+ volumes
		if zone.Servers == 1 && (zone.VolumesPerServer == 2 || zone.VolumesPerServer == 3) {
			return fmt.Errorf("distributed setup must have 4 or more volumes")
		}
		// Mandate a VolumeClaimTemplate
		if zone.VolumeClaimTemplate == nil {
			return errors.New("a volume claim template must be specified")
		}
		// Mandate a resource request
		if zone.VolumeClaimTemplate.Spec.Resources.Requests == nil {
			return errors.New("volume claim template must specify resource request")
		}
		// Mandate a request of storage
		if zone.VolumeClaimTemplate.Spec.Resources.Requests.Storage() == nil {
			return errors.New("volume claim template must specify resource storage request")
		}
		// Make sure the storage request is not 0
		if zone.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value() <= 0 {
			return errors.New("volume size must be greater than 0")
		}
		// Make sure access mode is provided
		if len(zone.VolumeClaimTemplate.Spec.AccessModes) == 0 {
			return errors.New("volume access mode must be specified")
		}
	}

	if t.Spec.CredsSecret == nil {
		return errors.New("please set credsSecret secret with credentials for Tenant")
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

// OwnerRef returns the OwnerReference to be added to all resources created by Tenant
func (t *Tenant) OwnerRef() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(t, schema.GroupVersionKind{
			Group:   SchemeGroupVersion.Group,
			Version: SchemeGroupVersion.Version,
			Kind:    MinIOCRDResourceKind,
		}),
	}
}
