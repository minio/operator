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

package v2

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/miekg/dns"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/golang-jwt/jwt"
	"github.com/minio/madmin-go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

// Webhook API constants
const (
	WebhookAPIVersion       = "/webhook/v1"
	WebhookDefaultPort      = "4222"
	WebhookSecret           = "operator-webhook-secret"
	WebhookOperatorUsername = "webhookUsername"
	WebhookOperatorPassword = "webhookPassword"

	// Webhook environment variable constants
	WebhookMinIOArgs   = "MINIO_ARGS"
	WebhookMinIOBucket = "MINIO_DNS_WEBHOOK_ENDPOINT"

	MinIOServerURL          = "MINIO_SERVER_URL"
	MinIODomain             = "MINIO_DOMAIN"
	MinIOBrowserRedirectURL = "MINIO_BROWSER_REDIRECT_URL"

	MinIORootUser     = "MINIO_ROOT_USER"
	MinIORootPassword = "MINIO_ROOT_PASSWORD"

	defaultPrometheusJWTExpiry = 100 * 365 * 24 * time.Hour
)

// envGet retrieves the value of the environment variable named
// by the key. If the variable is present in the environment the
// value (which may be empty) is returned. Otherwise it returns
// the specified default value.
func envGet(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

// List of webhook APIs
const (
	WebhookAPIGetenv        = WebhookAPIVersion + "/getenv"
	WebhookAPIBucketService = WebhookAPIVersion + "/bucketsrv"
	WebhookAPIUpdate        = WebhookAPIVersion + "/update"
	WebhookCRDConversaion   = WebhookAPIVersion + "/crd-conversion"
)

type hostsTemplateValues struct {
	StatefulSet string
	CIService   string
	HLService   string
	Ellipsis    string
	Domain      string
}

var (
	once                    sync.Once
	tenantMinIOImageOnce    sync.Once
	tenantKesImageOnce      sync.Once
	monitoringIntervalOnce  sync.Once
	k8sClusterDomain        string
	tenantMinIOImage        string
	tenantKesImage          string
	monitoringInterval      int
	prometheusNamespace     string
	prometheusName          string
	prometheusNamespaceOnce sync.Once
	prometheusNameOnce      sync.Once
)

// GetPodCAFromFile assumes the operator is running inside a k8s pod and extract the
// current ca certificate from /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
func GetPodCAFromFile() []byte {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil
	}
	return namespace
}

// GetNSFromFile assumes the operator is running inside a k8s pod and extract the
// current namespace from the /var/run/secrets/kubernetes.io/serviceaccount/namespace file
func GetNSFromFile() string {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "minio-operator"
	}
	return string(namespace)
}

// ellipsis returns the host range string
func genEllipsis(start, end int) string {
	return "{" + strconv.Itoa(start) + "..." + strconv.Itoa(end) + "}"
}

// HasCredsSecret returns true if the user has provided a secret
// for a Tenant else false
func (t *Tenant) HasCredsSecret() bool {
	return t.Spec.CredsSecret != nil && t.Spec.CredsSecret.Name != ""
}

// HasConfigurationSecret returns true if the user has provided a configuration
// for a Tenant else false
func (t *Tenant) HasConfigurationSecret() bool {
	return t.Spec.Configuration != nil && t.Spec.Configuration.Name != ""
}

// HasCertConfig returns true if the user has provided a certificate
// config
func (t *Tenant) HasCertConfig() bool {
	return t.Spec.CertConfig != nil
}

// ExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key
func (t *Tenant) ExternalCert() bool {
	return len(t.Spec.ExternalCertSecret) > 0
}

// ExternalCaCerts returns true is the user has provided a
// additional CA certificates for MinIO
func (t *Tenant) ExternalCaCerts() bool {
	return len(t.Spec.ExternalCaCertSecret) > 0
}

// ExternalClientCert returns true is the user has provided a secret
// that contains CA client cert, server cert and server key
func (t *Tenant) ExternalClientCert() bool {
	return t.Spec.ExternalClientCertSecret != nil && t.Spec.ExternalClientCertSecret.Name != ""
}

// ExternalClientCerts returns true is the user has provided additional client certificates
func (t *Tenant) ExternalClientCerts() bool {
	return len(t.Spec.ExternalClientCertSecrets) > 0
}

// KESExternalCert returns true is the user has provided a secret
// that contains CA cert, server cert and server key for KES pods
func (t *Tenant) KESExternalCert() bool {
	return t.Spec.KES != nil && t.Spec.KES.ExternalCertSecret != nil && t.Spec.KES.ExternalCertSecret.Name != ""
}

// KESClientCert returns true is the user has provided a secret
// that contains CA cert, client cert and client key for KES pods
func (t *Tenant) KESClientCert() bool {
	return t.Spec.KES != nil && t.Spec.KES.ClientCertSecret != nil && t.Spec.KES.ClientCertSecret.Name != ""
}

// AutoCert is enabled by default, otherwise we return the user provided value
func (t *Tenant) AutoCert() bool {
	if t.Spec.RequestAutoCert == nil {
		return true
	}
	return *t.Spec.RequestAutoCert
}

// VolumePathForPool returns the paths for MinIO mounts based on
// total number of volumes on a given pool
func (t *Tenant) VolumePathForPool(pool *Pool) string {
	if pool.VolumesPerServer == 1 {
		// Add an extra "/" to make sure relative paths are avoided.
		return path.Join("/", t.Spec.Mountpath, "/", t.Spec.Subpath)
	}
	// Add an extra "/" to make sure relative paths are avoided.
	return path.Join("/", t.Spec.Mountpath+genEllipsis(0, int(pool.VolumesPerServer-1)), "/", t.Spec.Subpath)
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

const (
	minioReleaseTagTimeLayout       = "2006-01-02T15-04-05Z"
	minioReleaseTagTimeLayoutBackup = "2006-01-02T15:04:05Z"
)

// ReleaseTagToReleaseTime - converts a 'RELEASE.2017-09-29T19-16-56Z.hotfix' into the build time
func ReleaseTagToReleaseTime(releaseTag string) (releaseTime time.Time, err error) {
	fields := strings.Split(releaseTag, ".")
	if len(fields) < 1 {
		return releaseTime, fmt.Errorf("%s not a valid release tag", releaseTag)
	}
	releaseTimeStr := fields[0]
	if len(fields) > 1 {
		releaseTimeStr = fields[1]
	}
	releaseTime, err = time.Parse(minioReleaseTagTimeLayout, releaseTimeStr)
	if err != nil {
		return time.Parse(minioReleaseTagTimeLayoutBackup, releaseTimeStr)
	}
	return releaseTime, nil
}

// ExtractTar extracts all tar files from the list `filesToExtract` and puts the files in the `basePath` location
func ExtractTar(filesToExtract []string, basePath, tarFileName string) error {
	tarFile, err := os.Open(basePath + tarFileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = tarFile.Close()
	}()

	tr := tar.NewReader(tarFile)
	if strings.HasSuffix(tarFileName, ".gz") {
		gz, err := gzip.NewReader(tarFile)
		if err != nil {
			return err
		}
		defer func() {
			_ = gz.Close()
		}()
		tr = tar.NewReader(gz)
	}

	success := len(filesToExtract)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			// only success if we have found all files
			if success == 0 {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("Tar file extraction failed for file index: %d, with: %w", success, err)
		}
		if header.Typeflag == tar.TypeReg {
			if name := find(filesToExtract, header.Name); name != "" {
				outFile, err := os.OpenFile(basePath+path.Base(name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o777)
				if err != nil {
					return fmt.Errorf("Tar file extraction failed while opening file: %s, at index: %d, with: %w", name, success, err)
				}
				if _, err := io.Copy(outFile, tr); err != nil {
					_ = outFile.Close()
					return fmt.Errorf("Tar file extraction failed while copying file: %s, at index: %d, with: %w", name, success, err)
				}
				_ = outFile.Close()
				success--
			}
		}
	}
	return nil
}

func find(slice []string, val string) string {
	for _, item := range slice {
		if item == val {
			return item
		}
	}
	return ""
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
		t.Spec.Image = GetTenantMinIOImage()
	}

	if t.Spec.ImagePullPolicy == "" {
		t.Spec.ImagePullPolicy = DefaultImagePullPolicy
	}

	for pi, pool := range t.Spec.Pools {
		// Default names for the pools if a name is not specified
		if pool.Name == "" {
			pool.Name = fmt.Sprintf("%s-%d", StatefulSetPrefix, pi)
		}
		t.Spec.Pools[pi] = pool
	}

	if t.Spec.Mountpath == "" {
		t.Spec.Mountpath = MinIOVolumeMountPath
	}

	if t.Spec.Subpath == "" {
		t.Spec.Subpath = MinIOVolumeSubPath
	}

	if t.Spec.CertConfig != nil {
		if t.Spec.CertConfig.CommonName == "" {
			t.Spec.CertConfig.CommonName = t.MinIOWildCardName()
		}
		if t.Spec.CertConfig.DNSNames == nil || len(t.Spec.CertConfig.DNSNames) == 0 {
			t.Spec.CertConfig.DNSNames = t.MinIOHosts()
		}
		if t.Spec.CertConfig.OrganizationName == nil || len(t.Spec.CertConfig.OrganizationName) == 0 {
			t.Spec.CertConfig.OrganizationName = DefaultOrgName
		}
	} else {
		t.Spec.CertConfig = &CertificateConfig{
			CommonName:       t.MinIOWildCardName(),
			DNSNames:         t.MinIOHosts(),
			OrganizationName: DefaultOrgName,
		}
	}

	if t.HasKESEnabled() {
		if t.Spec.KES.Image == "" {
			t.Spec.KES.Image = GetTenantKesImage()
		}
		if t.Spec.KES.Replicas == 0 {
			t.Spec.KES.Replicas = DefaultKESReplicas
		}
		if t.Spec.KES.ImagePullPolicy == "" {
			t.Spec.KES.ImagePullPolicy = DefaultImagePullPolicy
		}
		if t.Spec.KES.KeyName == "" {
			t.Spec.KES.KeyName = KESMinIOKey
		}
	}

	if t.HasPrometheusEnabled() {
		if t.Spec.Prometheus.Image == "" {
			t.Spec.Prometheus.Image = PrometheusImage
		}
		if t.Spec.Prometheus.SideCarImage == "" {
			t.Spec.Prometheus.SideCarImage = PrometheusSideCarImage
		}
		if t.Spec.Prometheus.InitImage == "" {
			t.Spec.Prometheus.InitImage = PrometheusInitImage
		}
	}

	if t.HasLogSearchAPIEnabled() {
		if t.Spec.Log.Image == "" {
			t.Spec.Log.Image = DefaultLogSearchAPIImage
		}
		if t.Spec.Log.Db != nil {
			if t.Spec.Log.Db.Image == "" {
				t.Spec.Log.Db.Image = LogPgImage
			}
			if t.Spec.Log.Db.InitImage == "" {
				t.Spec.Log.Db.InitImage = InitContainerImage
			}
		}
	}

	return t
}

// MinIOEndpoints similar to MinIOHosts but as URLs
func (t *Tenant) MinIOEndpoints(hostsTemplate string) (endpoints []string) {
	hosts := t.MinIOHosts()
	if hostsTemplate != "" {
		hosts = t.TemplatedMinIOHosts(hostsTemplate)
	}

	for _, host := range hosts {
		if t.TLS() {
			endpoints = append(endpoints, "https://"+host)
		} else {
			endpoints = append(endpoints, "http://"+host)
		}
	}

	return endpoints
}

// GenBearerToken returns the JWT token for current Tenant for Prometheus authentication
func (t *Tenant) GenBearerToken(accessKey, secretKey string) string {
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: time.Now().UTC().Add(defaultPrometheusJWTExpiry).Unix(),
		Subject:   accessKey,
		Issuer:    "prometheus",
	})

	token, err := jwt.SignedString([]byte(secretKey))
	if err != nil {
		panic(fmt.Sprintf("jwt key generation: %v", err))
	}

	return token
}

// MinIOHosts returns the domain names in ellipses format created for current Tenant
func (t *Tenant) MinIOHosts() (hosts []string) {
	// Create the ellipses style URL
	for pi, pool := range t.Spec.Pools {
		// determine the proper statefulset name
		ssName := t.PoolStatefulsetName(&pool)
		if len(t.Status.Pools) > pi {
			ssName = t.Status.Pools[pi].SSName
		}

		if pool.Servers == 1 {
			hosts = append(hosts, fmt.Sprintf("%s-%s.%s.%s.svc.%s", ssName, "0", t.MinIOHLServiceName(), t.Namespace, GetClusterDomain()))
		} else {
			hosts = append(hosts, fmt.Sprintf("%s-%s.%s.%s.svc.%s", ssName, genEllipsis(0, int(pool.Servers)-1), t.MinIOHLServiceName(), t.Namespace, GetClusterDomain()))
		}
	}
	return hosts
}

// TemplatedMinIOHosts returns the domain names in ellipses format created for current Tenant without the service part
func (t *Tenant) TemplatedMinIOHosts(hostsTemplate string) (hosts []string) {
	tmpl, err := template.New("hosts").Parse(hostsTemplate)
	if err != nil {
		msg := "Invalid go template for hosts"
		klog.V(2).Infof(msg)
		return hosts
	}
	var max, index int32
	// Create the ellipses style URL
	for _, pool := range t.Spec.Pools {
		max = max + pool.Servers
		data := hostsTemplateValues{
			StatefulSet: t.MinIOStatefulSetNameForPool(&pool),
			CIService:   t.MinIOCIServiceName(),
			HLService:   t.MinIOHLServiceName(),
			Ellipsis:    genEllipsis(int(index), int(max)-1),
			Domain:      GetClusterDomain(),
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
	hosts = append(hosts, t.MinIOFQDNServiceName())
	hosts = append(hosts, t.MinIOFQDNServiceNameAndNamespace())
	hosts = append(hosts, t.MinIOFQDNShortServiceName())
	hosts = append(hosts, "*."+t.MinIOHeadlessServiceHost())
	return hosts
}

// ConsoleServerHost returns ClusterIP service Host for current Console Tenant
func (t *Tenant) ConsoleServerHost() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.ConsoleCIServiceName(), t.Namespace, GetClusterDomain())
}

// MinIOHeadlessServiceHost returns headless service Host for current Tenant
func (t *Tenant) MinIOHeadlessServiceHost() string {
	if t.Spec.Pools[0].Servers == 1 {
		msg := "Please set the server count > 1"
		klog.V(2).Infof(msg)
		return ""
	}
	return fmt.Sprintf("%s.%s.svc.%s", t.MinIOHLServiceName(), t.Namespace, GetClusterDomain())
}

// KESHosts returns the host names created for current KES StatefulSet
func (t *Tenant) KESHosts() []string {
	hosts := make([]string, 0)
	var i int32
	for i < t.Spec.KES.Replicas {
		hosts = append(hosts, fmt.Sprintf("%s-"+strconv.Itoa(int(i))+".%s.%s.svc.%s", t.KESStatefulSetName(), t.KESHLServiceName(), t.Namespace, GetClusterDomain()))
		i++
	}
	hosts = append(hosts, t.KESServiceHost())
	return hosts
}

// KESServiceEndpoint similar to KESServiceHost but a URL with current scheme
func (t *Tenant) KESServiceEndpoint() string {
	u := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(t.KESServiceHost(), strconv.Itoa(KESPort)),
	}
	return u.String()
}

// KESServiceHost returns headless service Host for KES in current Tenant
func (t *Tenant) KESServiceHost() string {
	return fmt.Sprintf("%s.%s.svc.%s", t.KESHLServiceName(), t.Namespace, GetClusterDomain())
}

// BucketDNS indicates if Bucket DNS feature is enabled.
func (t *Tenant) BucketDNS() bool {
	// we've deprecated .spec.s3 and will top working in future releases of operator
	return (t.Spec.Features != nil && t.Spec.Features.BucketDNS) || (t.Spec.S3 != nil && t.Spec.S3.BucketDNS)
}

// HasKESEnabled checks if kes configuration is provided by user
func (t *Tenant) HasKESEnabled() bool {
	return t.Spec.KES != nil
}

// HasLogSearchAPIEnabled checks if Log feature has been enabled
func (t *Tenant) HasLogSearchAPIEnabled() bool {
	return t.Spec.Log != nil
}

// HasLogDBEnabled checks if Log DB feature has been enabled
func (t *Tenant) HasLogDBEnabled() bool {
	return t.Spec.Log != nil && t.Spec.Log.Db != nil
}

// HasPrometheusEnabled checks if Prometheus metrics has been enabled
func (t *Tenant) HasPrometheusEnabled() bool {
	return t.Spec.Prometheus != nil
}

// HasPrometheusOperatorEnabled checks if Prometheus service monitor has been enabled
func (t *Tenant) HasPrometheusOperatorEnabled() bool {
	return t.Spec.PrometheusOperator
}

// GetEnvVars returns the environment variables for tenant deployment.
func (t *Tenant) GetEnvVars() (env []corev1.EnvVar) {
	return t.Spec.Env
}

// GetLogSearchAPIEnvVars returns the environment variables for Log Search Api deployment.
func (t *Tenant) GetLogSearchAPIEnvVars() (env []corev1.EnvVar) {
	if !t.HasLogSearchAPIEnabled() {
		return env
	}
	return t.Spec.Log.Env
}

// GetLogDBEnvVars returns the environment variables for Postgres deployment.
func (t *Tenant) GetLogDBEnvVars() (env []corev1.EnvVar) {
	if !t.HasLogDBEnabled() {
		return env
	}
	return t.Spec.Log.Db.Env
}

// GetPrometheusEnvVars returns the environment variables for the Prometheus deployment.
func (t *Tenant) GetPrometheusEnvVars() (env []corev1.EnvVar) {
	if !t.HasPrometheusEnabled() {
		return env
	}
	return t.Spec.Prometheus.Env
}

// GetKESEnvVars returns the environment variables for the KES deployment.
func (t *Tenant) GetKESEnvVars() (env []corev1.EnvVar) {
	if !t.HasKESEnabled() {
		return env
	}
	return t.Spec.KES.Env
}

// UpdateURL returns the URL for the sha256sum location of the new binary
func (t *Tenant) UpdateURL(ltag string, overrideURL string) (string, error) {
	if overrideURL == "" {
		overrideURL = DefaultMinIOUpdateURL
	}
	u, err := url.Parse(overrideURL)
	if err != nil {
		return "", err
	}
	u.Path = u.Path + "/minio." + ltag + ".sha256sum"
	return u.String(), nil
}

// MinIOHLPodAddress similar to MinIOFQDNServiceName but returns pod hostname with port
func (t *Tenant) MinIOHLPodAddress(podName string) string {
	var port int

	if t.TLS() {
		port = MinIOTLSPortLoadBalancerSVC
	} else {
		port = MinIOPortLoadBalancerSVC
	}

	return net.JoinHostPort(t.MinIOHLPodHostname(podName), strconv.Itoa(port))
}

// MinIOServerHostAddress similar to MinIOFQDNServiceName but returns host with port
func (t *Tenant) MinIOServerHostAddress() string {
	var port int

	if t.TLS() {
		port = MinIOTLSPortLoadBalancerSVC
	} else {
		port = MinIOPortLoadBalancerSVC
	}

	return net.JoinHostPort(t.MinIOFQDNServiceName(), strconv.Itoa(port))
}

// MinIOServerEndpoint similar to MinIOServerHostAddress but a URL with current scheme
func (t *Tenant) MinIOServerEndpoint() string {
	scheme := "http"
	if t.TLS() {
		scheme = "https"
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   t.MinIOServerHostAddress(),
	}
	return u.String()
}

// MinIOHealthCheck check MinIO cluster health
func (t *Tenant) MinIOHealthCheck(tr *http.Transport) bool {
	if tr.TLSClientConfig != nil {
		tr.TLSClientConfig.InsecureSkipVerify = true
	}

	clnt, err := madmin.NewAnonymousClient(t.MinIOServerHostAddress(), t.TLS())
	if err != nil {
		return false
	}
	clnt.SetCustomTransport(tr)

	result, err := clnt.Healthy(context.Background(), madmin.HealthOpts{})
	if err != nil {
		return false
	}

	return result.Healthy
}

// NewMinIOAdmin initializes a new madmin.Client for operator interaction
func (t *Tenant) NewMinIOAdmin(minioSecret map[string][]byte, tr *http.Transport) (*madmin.AdminClient, error) {
	return t.NewMinIOAdminForAddress("", minioSecret, tr)
}

// NewMinIOAdminForAddress initializes a new madmin.Client for operator interaction
func (t *Tenant) NewMinIOAdminForAddress(address string, minioSecret map[string][]byte, tr *http.Transport) (*madmin.AdminClient, error) {
	host, accessKey, secretKey, err := t.getMinIOTenantDetails(address, minioSecret)
	if err != nil {
		return nil, err
	}

	opts := &madmin.Options{
		Secure: t.TLS(),
		Creds:  credentials.NewStaticV4(string(accessKey), string(secretKey), ""),
	}

	madmClnt, err := madmin.NewWithOptions(host, opts)
	if err != nil {
		return nil, err
	}
	madmClnt.SetCustomTransport(tr)

	return madmClnt, nil
}

func (t *Tenant) getMinIOTenantDetails(address string, minioSecret map[string][]byte) (string, []byte, []byte, error) {
	host := address
	if host == "" {
		host = t.MinIOServerHostAddress()
		if host == "" {
			return "", nil, nil, errors.New("MinIO server host is empty")
		}
	}

	accessKey, ok := minioSecret["accesskey"]
	if !ok {
		return "", nil, nil, errors.New("MinIO server accesskey not set")
	}

	secretKey, ok := minioSecret["secretkey"]
	if !ok {
		return "", nil, nil, errors.New("MinIO server secretkey not set")
	}
	return host, accessKey, secretKey, nil
}

// NewMinIOUser initializes a new console user
func (t *Tenant) NewMinIOUser(minioSecret map[string][]byte, tr *http.Transport) (*minio.Client, error) {
	return t.NewMinIOUserForAddress("", minioSecret, tr)
}

// NewMinIOUserForAddress initializes a new console user
func (t *Tenant) NewMinIOUserForAddress(address string, minioSecret map[string][]byte, tr *http.Transport) (*minio.Client, error) {
	host, accessKey, secretKey, err := t.getMinIOTenantDetails(address, minioSecret)
	if err != nil {
		return nil, err
	}
	opts := &minio.Options{
		Transport: tr,
		Secure:    t.TLS(),
		Creds:     credentials.NewStaticV4(string(accessKey), string(secretKey), ""),
	}
	minioClient, err := minio.New(host, opts)
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

// MustGetSystemCertPool - return system CAs or empty pool in case of error (or windows)
func MustGetSystemCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return x509.NewCertPool()
	}
	return pool
}

// CreateUsers creates a list of admin users on MinIO, optionally creating users is disabled.
func (t *Tenant) CreateUsers(madmClnt *madmin.AdminClient, userCredentialSecrets []*corev1.Secret, tenantConfiguration map[string][]byte) error {
	var skipCreateUser bool // Skip creating users if LDAP is enabled.
	if ldapAddress, ok := tenantConfiguration["MINIO_IDENTITY_LDAP_SERVER_ADDR"]; ok {
		skipCreateUser = string(ldapAddress) != ""
	}
	// add user with a 20 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	for _, secret := range userCredentialSecrets {
		consoleAccessKey, ok := secret.Data["CONSOLE_ACCESS_KEY"]
		if !ok {
			return errors.New("CONSOLE_ACCESS_KEY not provided")
		}
		// remove spaces and line breaks from access key
		userAccessKey := strings.TrimSpace(string(consoleAccessKey))
		// skipCreateUser handles the scenario of LDAP users that are
		// not created in MinIO but still need to have a policy assigned
		if !skipCreateUser {
			consoleSecretKey, ok := secret.Data["CONSOLE_SECRET_KEY"]
			// remove spaces and line breaks from secret key
			userSecretKey := strings.TrimSpace(string(consoleSecretKey))
			if !ok {
				return errors.New("CONSOLE_SECRET_KEY not provided")
			}
			if err := madmClnt.AddUser(ctx, userAccessKey, userSecretKey); err != nil {
				return err
			}
		}
		if err := madmClnt.SetPolicy(ctx, ConsoleAdminPolicyName, userAccessKey, false); err != nil {
			return err
		}
	}
	return nil
}

// CreateBuckets creates buckets and skips if bucket already present
func (t *Tenant) CreateBuckets(minioClient *minio.Client, buckets ...Bucket) error {
	for _, bucket := range buckets {
		if err := s3utils.CheckValidBucketName(bucket.Name); err != nil {
			return err
		}
		// create each bucket with a 20 seconds timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		if err := minioClient.MakeBucket(ctx, bucket.Name, minio.MakeBucketOptions{
			Region:        bucket.Region,
			ObjectLocking: bucket.ObjectLocking,
		}); err != nil {
			switch minio.ToErrorResponse(err).Code {
			case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
				klog.Infof(err.Error())
				continue
			}
		}
		klog.Infof("Successfully created bucket %s", bucket.Name)
	}
	return nil
}

// Validate validate single pool as per MinIO deployment requirements
func (z *Pool) Validate(zi int) error {
	// Make sure the replicas are not 0 on any pool
	if z.Servers <= 0 {
		return fmt.Errorf("pool #%d cannot have 0 servers", zi)
	}

	// Make sure the pools don't have 0 volumes
	if z.VolumesPerServer <= 0 {
		return fmt.Errorf("pool #%d cannot have 0 volumes per server", zi)
	}

	if z.Servers*z.VolumesPerServer < 4 {
		// Erasure coding has few requirements.
		switch z.Servers {
		case 2:
			return fmt.Errorf("pool #%d with 2 servers must have at least 4 volumes in total", zi)
		case 3:
			return fmt.Errorf("pool #%d with 3 servers must have at least 6 volumes in total", zi)
		}
	}

	// Mandate a VolumeClaimTemplate
	if z.VolumeClaimTemplate == nil {
		return errors.New("a volume claim template must be specified")
	}

	// Mandate a resource request
	if z.VolumeClaimTemplate.Spec.Resources.Requests == nil {
		return errors.New("volume claim template must specify resource request")
	}

	// Mandate a request of storage
	if z.VolumeClaimTemplate.Spec.Resources.Requests.Storage() == nil {
		return errors.New("volume claim template must specify resource storage request")
	}

	// Make sure the storage request is not 0
	if z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value() <= 0 {
		return errors.New("volume size must be greater than 0")
	}

	// Make sure access mode is provided
	if len(z.VolumeClaimTemplate.Spec.AccessModes) == 0 {
		return errors.New("volume access mode must be specified")
	}

	return nil
}

// Validate returns an error if any configuration of the MinIO Tenant is invalid
func (t *Tenant) Validate() error {
	if t.Spec.Pools == nil {
		return errors.New("pools must be configured")
	}

	if !t.HasConfigurationSecret() && !t.HasCredsSecret() {
		return errors.New("please set 'configuration' secret with credentials for Tenant")
	}

	// Every pool must contain a Volume Claim Template
	for zi, pool := range t.Spec.Pools {
		if err := pool.Validate(zi); err != nil {
			return err
		}
	}
	// make sure all the domains are valid
	if err := t.ValidateDomains(); err != nil {
		return err
	}

	return nil
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

// ObjectRef returns the ObjectReference to be added to all resources created by Tenant
func (t *Tenant) ObjectRef() corev1.ObjectReference {
	return corev1.ObjectReference{
		Kind:      MinIOCRDResourceKind,
		Namespace: t.Namespace,
		Name:      t.Name,
		UID:       t.UID,
	}
}

// TLS indicates whether TLS is enabled for this tenant
func (t *Tenant) TLS() bool {
	return t.AutoCert() || t.ExternalCert()
}

// GetClusterDomain returns the Kubernetes cluster domain
func GetClusterDomain() string {
	once.Do(func() {
		k8sClusterDomain = envGet(clusterDomain, "cluster.local")
	})
	return k8sClusterDomain
}

// MergeMaps merges two maps and returns the union
func MergeMaps(a, b map[string]string) map[string]string {
	for k, v := range b {
		a[k] = v
	}
	return a
}

// ToMap converts a slice of env vars to a map of Name and value
func ToMap(envs []corev1.EnvVar) map[string]string {
	newMap := make(map[string]string)
	for i := range envs {
		newMap[envs[i].Name] = envs[i].Value
	}
	return newMap
}

// IsContainersEnvUpdated compare environment variables of existing and expected containers and
// returns true if there is a change
func IsContainersEnvUpdated(existingContainers, expectedContainers []corev1.Container) bool {
	// compare containers environment variables
	expectedContainersEnvVars := map[string][]corev1.EnvVar{}
	existingContainersEnvVars := map[string][]corev1.EnvVar{}
	for _, c := range existingContainers {
		existingContainersEnvVars[c.Name] = c.Env
	}
	for _, c := range expectedContainers {
		expectedContainersEnvVars[c.Name] = c.Env
	}
	for key := range expectedContainersEnvVars {
		if _, ok := existingContainersEnvVars[key]; !ok {
			return true
		}
		if IsEnvUpdated(ToMap(existingContainersEnvVars[key]), ToMap(expectedContainersEnvVars[key])) {
			return true
		}
	}
	return false
}

// IsEnvUpdated looks for new env vars in the old env vars and returns true if
// new env vars are not found
func IsEnvUpdated(old, new map[string]string) bool {
	if len(old) != len(new) {
		return true
	}
	if !reflect.DeepEqual(old, new) {
		return true
	}
	return false
}

// GetTenantMinIOImage returns the default MinIO image for a tenant
func GetTenantMinIOImage() string {
	tenantMinIOImageOnce.Do(func() {
		tenantMinIOImage = envGet(tenantMinIOImageEnv, DefaultMinIOImage)
	})
	return tenantMinIOImage
}

// GetTenantKesImage returns the default KES Image for a tenant
func GetTenantKesImage() string {
	tenantKesImageOnce.Do(func() {
		tenantKesImage = envGet(tenantKesImageEnv, DefaultKESImage)
	})
	return tenantKesImage
}

// GetMonitoringInterval returns how ofter we should query tenants for cluster/health
func GetMonitoringInterval() int {
	monitoringIntervalOnce.Do(func() {
		monitoringIntervalStr := envGet(monitoringIntervalEnv, "")
		if monitoringIntervalStr == "" {
			monitoringInterval = DefaultMonitoringInterval
		}
		val, err := strconv.Atoi(monitoringIntervalStr)
		if err != nil {
			monitoringInterval = DefaultMonitoringInterval
		} else {
			monitoringInterval = val
		}
	})
	return monitoringInterval
}

// GetTenantServiceURL gets tenant's service url with the proper scheme and port
func (t *Tenant) GetTenantServiceURL() (svcURL string) {
	scheme := "http"
	if t.TLS() {
		scheme = "https"
	}
	svc := t.MinIOServerHostAddress()
	return fmt.Sprintf("%s://%s", scheme, svc)
}

type envKV struct {
	Key   string
	Value string
	Skip  bool
}

func (e envKV) String() string {
	if e.Skip {
		return ""
	}
	return fmt.Sprintf("%s=%s", e.Key, e.Value)
}

func parsEnvEntry(envEntry string) (envKV, error) {
	envEntry = strings.TrimSpace(envEntry)
	if envEntry == "" {
		// Skip all empty lines
		return envKV{
			Skip: true,
		}, nil
	}
	const envSeparator = "="
	envTokens := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(envEntry, "export")), envSeparator, 2)
	if len(envTokens) != 2 {
		return envKV{}, fmt.Errorf("envEntry malformed; %s, expected to be of form 'KEY=value'", envEntry)
	}
	key := envTokens[0]
	val := envTokens[1]

	if strings.HasPrefix(key, "#") {
		// Skip commented lines
		return envKV{
			Skip: true,
		}, nil
	}

	// Remove quotes from the value if found
	if len(val) >= 2 {
		quote := val[0]
		if (quote == '"' || quote == '\'') && val[len(val)-1] == quote {
			val = val[1 : len(val)-1]
		}
	}
	return envKV{
		Key:   key,
		Value: val,
	}, nil
}

// ParseRawConfiguration map[string][]byte representation of the MinIO config.env file
func ParseRawConfiguration(configuration []byte) (config map[string][]byte) {
	config = map[string][]byte{}
	if configuration != nil {
		scanner := bufio.NewScanner(strings.NewReader(string(configuration)))
		for scanner.Scan() {
			ekv, err := parsEnvEntry(scanner.Text())
			if err != nil {
				klog.Errorf("Error parsing tenant configuration: %v", err.Error())
				continue
			}
			if ekv.Skip {
				// Skips empty lines
				continue
			}
			config[ekv.Key] = []byte(ekv.Value)
			if ekv.Key == "MINIO_ROOT_USER" || ekv.Key == "MINIO_ACCESS_KEY" {
				config["accesskey"] = config[ekv.Key]
			} else if ekv.Key == "MINIO_ROOT_PASSWORD" || ekv.Key == "MINIO_SECRET_KEY" {
				config["secretkey"] = config[ekv.Key]
			}
		}
		if err := scanner.Err(); err != nil {
			klog.Errorf("Error parsing tenant configuration: %v", err.Error())
			return config
		}
	}
	return config
}

// GetPrometheusNamespace returns namespace of the prometheus managed by prometheus operator
func GetPrometheusNamespace() string {
	prometheusNamespaceOnce.Do(func() {
		prometheusNamespace = envGet(PrometheusNamespace, DefaultPrometheusNamespace)
	})
	return prometheusNamespace
}

// GetPrometheusName returns namespace of the prometheus managed by prometheus operator
func GetPrometheusName() string {
	prometheusNameOnce.Do(func() {
		prometheusName = envGet(prometheusName, "")
	})
	return prometheusName
}

// HasMinIODomains indicates whether domains are being specified for MinIO
func (t *Tenant) HasMinIODomains() bool {
	return t.Spec.Features != nil && t.Spec.Features.Domains != nil && len(t.Spec.Features.Domains.Minio) > 0
}

// HasConsoleDomains indicates whether a domain is being specified for Console
func (t *Tenant) HasConsoleDomains() bool {
	return t.Spec.Features != nil && t.Spec.Features.Domains != nil && t.Spec.Features.Domains.Console != ""
}

// ValidateDomains checks the validity of the domains configured on the tenant
func (t *Tenant) ValidateDomains() error {
	if t.HasMinIODomains() {
		domains := t.Spec.Features.Domains.Minio
		if len(domains) != 0 {
			for _, domainName := range domains {
				_, err := url.Parse(domainName)
				if err != nil {
					return err
				}

				if _, ok := dns.IsDomainName(domainName); !ok {
					return fmt.Errorf("invalid domain `%s`", domainName)
				}
			}
			sort.Strings(domains)
			lcpSuf := lcpSuffix(domains)
			for _, domainName := range domains {
				if domainName == lcpSuf && len(domains) > 1 {
					return fmt.Errorf("overlapping domains `%s` not allowed", domainName)
				}
			}
		}
	}
	return nil
}

// GetDomainHosts returns a list of hosts in the .spec.features.domains.minio list to configure MINIO_DOMAIN
func (t *Tenant) GetDomainHosts() []string {
	if t.HasMinIODomains() {
		domains := t.Spec.Features.Domains.Minio
		var hosts []string
		for _, d := range domains {
			u, err := url.Parse(d)
			if err != nil {
				continue
			}
			// remove ports if any
			hostParts := strings.Split(u.Host, ":")
			hosts = append(hosts, hostParts[0])
		}
		return hosts
	}
	return nil
}

// HasEnv returns whether an environment variable is defined in the .spec.env field
func (t *Tenant) HasEnv(envName string) bool {
	for _, env := range t.Spec.Env {
		if env.Name == envName {
			return true
		}
	}
	return false
}

// Suffix returns the longest common suffix of the provided strings
func lcpSuffix(strs []string) string {
	return lcp(strs, false)
}

func lcp(strs []string, pre bool) string {
	// short-circuit empty list
	if len(strs) == 0 {
		return ""
	}
	xfix := strs[0]
	// short-circuit single-element list
	if len(strs) == 1 {
		return xfix
	}
	// compare first to rest
	for _, str := range strs[1:] {
		xfixl := len(xfix)
		strl := len(str)
		// short-circuit empty strings
		if xfixl == 0 || strl == 0 {
			return ""
		}
		// maximum possible length
		maxl := xfixl
		if strl < maxl {
			maxl = strl
		}
		// compare letters
		if pre {
			// prefix, iterate left to right
			for i := 0; i < maxl; i++ {
				if xfix[i] != str[i] {
					xfix = xfix[:i]
					break
				}
			}
		} else {
			// suffix, iterate right to left
			for i := 0; i < maxl; i++ {
				xi := xfixl - i - 1
				si := strl - i - 1
				if xfix[xi] != str[si] {
					xfix = xfix[xi+1:]
					break
				}
			}
		}
	}
	return xfix
}
