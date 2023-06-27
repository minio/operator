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

package statefulsets

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/minio/operator/pkg/common"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	bucketDNSEnv = "MINIO_DNS_WEBHOOK_ENDPOINT"
)

// Returns the MinIO environment variables set in configuration.
// If a user specifies a secret in the spec (for MinIO credentials) we use
// that to set MINIO_ROOT_USER & MINIO_ROOT_PASSWORD.
func minioEnvironmentVars(t *miniov2.Tenant, skipEnvVars map[string][]byte, opVersion string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	// Enable `mc admin update` style updates to MinIO binaries
	// within the container, only operator is supposed to perform
	// these operations.
	envVarsMap := map[string]corev1.EnvVar{
		"MINIO_UPDATE": {
			Name:  "MINIO_UPDATE",
			Value: "on",
		},
		"MINIO_UPDATE_MINISIGN_PUBKEY": {
			Name:  "MINIO_UPDATE_MINISIGN_PUBKEY",
			Value: "RWTx5Zr1tiHQLwG9keckT0c45M3AGeHD6IvimQHpyRywVWGbP1aVSGav",
		},
		"MINIO_OPERATOR_VERSION": {
			Name:  "MINIO_OPERATOR_VERSION",
			Value: opVersion,
		},
		"MINIO_PROMETHEUS_JOB_ID": {
			Name:  "MINIO_PROMETHEUS_JOB_ID",
			Value: t.PrometheusConfigJobName(),
		},
	}
	// Specific case of bug in runtimeClass crun where $HOME is not set
	for _, pool := range t.Spec.Pools {
		if pool.RuntimeClassName != nil && *pool.RuntimeClassName == "crun" {
			// Set HOME to /
			envVarsMap["HOME"] = corev1.EnvVar{
				Name:  "HOME",
				Value: "/",
			}
		}
	}

	var domains []string
	// Enable Bucket DNS only if asked for by default turned off
	if t.BucketDNS() {
		domains = append(domains, t.MinIOBucketBaseDomain())
		sidecarBucketURL := fmt.Sprintf("http://127.0.0.1:%s%s/%s/%s",
			common.WebhookDefaultPort,
			common.WebhookAPIBucketService,
			t.Namespace,
			t.Name)
		envVarsMap[bucketDNSEnv] = corev1.EnvVar{
			Name:  bucketDNSEnv,
			Value: sidecarBucketURL,
		}
	}
	// Check if any domains are configured
	if t.HasMinIODomains() {
		domains = append(domains, t.GetDomainHosts()...)
	}
	// tell MinIO about all the domains meant to hit it if they are not passed manually via .spec.env
	if len(domains) > 0 {
		envVarsMap[miniov2.MinIODomain] = corev1.EnvVar{
			Name:  miniov2.MinIODomain,
			Value: strings.Join(domains, ","),
		}
	}
	// If no specific server URL is specified we will specify the internal k8s url, but if a list of domains was
	// provided we will use the first domain.
	serverURL := t.MinIOServerEndpoint()
	if t.HasMinIODomains() {
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(t.Spec.Features.Domains.Minio[0], "http") {
			useSchema := "http"
			if t.TLS() {
				useSchema = "https"
			}
			serverURL = fmt.Sprintf("%s://%s", useSchema, t.Spec.Features.Domains.Minio[0])
		} else {
			serverURL = t.Spec.Features.Domains.Minio[0]
		}
	}
	envVarsMap[miniov2.MinIOServerURL] = corev1.EnvVar{
		Name:  miniov2.MinIOServerURL,
		Value: serverURL,
	}

	// Set the redirect url for console
	if t.HasConsoleDomains() {
		consoleDomain := t.Spec.Features.Domains.Console
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(consoleDomain, "http") {
			useSchema := "http"
			if t.TLS() {
				useSchema = "https"
			}
			consoleDomain = fmt.Sprintf("%s://%s", useSchema, consoleDomain)
		}
		envVarsMap[miniov2.MinIOBrowserRedirectURL] = corev1.EnvVar{
			Name:  miniov2.MinIOBrowserRedirectURL,
			Value: consoleDomain,
		}
	}

	if t.HasKESEnabled() {
		envVarsMap["MINIO_KMS_KES_ENDPOINT"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_ENDPOINT",
			Value: t.KESServiceEndpoint(),
		}
		envVarsMap["MINIO_KMS_KES_CERT_FILE"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CERT_FILE",
			Value: miniov2.MinIOCertPath + "/client.crt",
		}
		envVarsMap["MINIO_KMS_KES_KEY_FILE"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_FILE",
			Value: miniov2.MinIOCertPath + "/client.key",
		}
		envVarsMap["MINIO_KMS_KES_CA_PATH"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CA_PATH",
			Value: miniov2.MinIOCertPath + "/CAs/kes.crt",
		}
		envVarsMap["MINIO_KMS_KES_KEY_NAME"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_NAME",
			Value: t.Spec.KES.KeyName,
		}
	}

	if t.HasConfigurationSecret() {
		envVarsMap["MINIO_CONFIG_ENV_FILE"] = corev1.EnvVar{
			Name:  "MINIO_CONFIG_ENV_FILE",
			Value: miniov2.CfgFile,
		}
	}

	// Add all the tenant.spec.env environment variables
	// User defined environment variables will take precedence over default environment variables
	for _, env := range t.GetEnvVars() {
		envVarsMap[env.Name] = env
	}

	// transform map to array and skip configurations from config.env
	for _, env := range envVarsMap {
		if _, ok := skipEnvVars[env.Name]; !ok {
			envVars = append(envVars, env)
		}
	}
	// sort the array to produce the same result everytime
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	// Return environment variables
	return envVars
}

// PodMetadata Returns the MinIO pods metadata set in configuration.
// If a user specifies metadata in the spec we return that metadata.
func PodMetadata(t *miniov2.Tenant, pool *miniov2.Pool) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	// Copy Labels and Annotations from Tenant
	labels := t.ObjectMeta.Labels
	annotations := t.ObjectMeta.Annotations

	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[miniov2.Revision] = fmt.Sprintf("%d", t.Status.Revision)

	if labels == nil {
		labels = make(map[string]string)
	}
	// Add the additional label used by StatefulSet spec selector
	for k, v := range t.MinIOPodLabels() {
		labels[k] = v
	}
	// Add information labels, such as which pool we are building this pod about
	labels[miniov2.PoolLabel] = pool.Name
	// Add the additional label used by Console spec selector
	for k, v := range t.ConsolePodLabels() {
		labels[k] = v
	}

	// Add user specific annotations
	if pool.Annotations != nil {
		annotations = miniov2.MergeMaps(annotations, pool.Annotations)
	}

	if pool.Labels != nil {
		labels = miniov2.MergeMaps(labels, pool.Labels)
	}

	meta.Labels = labels
	meta.Annotations = annotations

	return meta
}

// ContainerMatchLabels Returns the labels that match the Pods in the statefulset
func ContainerMatchLabels(t *miniov2.Tenant, pool *miniov2.Pool) *metav1.LabelSelector {
	labels := miniov2.MergeMaps(t.MinIOPodLabels(), t.ConsolePodLabels())
	// Add pool information so it's passed down to the underlying PVCs
	labels[miniov2.PoolLabel] = pool.Name
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

// CfgVolumeMount is the volume mount used by `minio`, `sidecar` and `validate-arguments` containers
var CfgVolumeMount = corev1.VolumeMount{
	Name:      CfgVol,
	MountPath: miniov2.CfgPath,
}

// TmpCfgVolumeMount is the temporary location
var TmpCfgVolumeMount = corev1.VolumeMount{
	Name:      "configuration",
	MountPath: miniov2.TmpPath + "/minio-config",
}

// Builds the volume mounts for MinIO container.
func volumeMounts(t *miniov2.Tenant, pool *miniov2.Pool, certVolumeSources []v1.VolumeProjection) (mounts []v1.VolumeMount) {
	// Default volume name, unless another one was provided
	name := miniov2.MinIOVolumeName
	if pool.VolumeClaimTemplate != nil {
		name = pool.VolumeClaimTemplate.Name
	}

	// shared configuration Volume
	mounts = append(mounts, CfgVolumeMount)

	if pool.VolumesPerServer == 1 {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name + strconv.Itoa(0),
			MountPath: t.Spec.Mountpath,
		})
	} else {
		for i := 0; i < int(pool.VolumesPerServer); i++ {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      name + strconv.Itoa(i),
				MountPath: t.Spec.Mountpath + strconv.Itoa(i),
			})
		}
	}

	// CertPath (/tmp/certs) will always be mounted even if the tenant doesn't have any TLS certificate
	// operator will still mount the operator public cert under /tmp/certs/CAs/operator.crt
	if len(certVolumeSources) > 0 {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      t.MinIOTLSSecretName(),
			MountPath: miniov2.MinIOCertPath,
		})
	}

	return mounts
}

// Builds the MinIO container for a Tenant.
func poolMinioServerContainer(t *miniov2.Tenant, skipEnvVars map[string][]byte, pool *miniov2.Pool, hostsTemplate string, opVersion string, certVolumeSources []v1.VolumeProjection) v1.Container {
	consolePort := miniov2.ConsolePort
	if t.TLS() {
		consolePort = miniov2.ConsoleTLSPort
	}
	args := []string{"server", "--certs-dir", miniov2.MinIOCertPath, "--console-address", ":" + strconv.Itoa(consolePort)}
	if t.Spec.Logging != nil {
		// If logging is specified, expect users to
		// provide the right set of settings to toggle
		// various flags.
		if t.Spec.Logging.JSON {
			args = append(args, "--json")
		}
		if t.Spec.Logging.Anonymous {
			args = append(args, "--anonymous")
		}
		if t.Spec.Logging.Quiet {
			args = append(args, "--quiet")
		}
	}

	return corev1.Container{
		Name:  miniov2.MinIOServerName,
		Image: t.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.MinIOPort,
			},
			{
				ContainerPort: int32(consolePort),
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    volumeMounts(t, pool, certVolumeSources),
		Args:            args,
		Env:             minioEnvironmentVars(t, skipEnvVars, opVersion),
		Resources:       pool.Resources,
		LivenessProbe:   t.Spec.Liveness,
		ReadinessProbe:  t.Spec.Readiness,
		StartupProbe:    t.Spec.Startup,
		SecurityContext: poolContainerSecurityContext(pool),
	}
}

// GetContainerArgs returns the arguments that the MinIO container receives
func GetContainerArgs(t *miniov2.Tenant, hostsTemplate string) []string {
	var args []string
	if len(t.Spec.Pools) == 1 && t.Spec.Pools[0].Servers == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, t.VolumePathForPool(&t.Spec.Pools[0]))
	} else {
		for index, endpoint := range t.MinIOEndpoints(hostsTemplate) {
			args = append(args, fmt.Sprintf("%s%s", endpoint, t.VolumePathForPool(&t.Spec.Pools[index])))
		}
	}
	return args
}

// Builds the tolerations for a Pool.
func poolTolerations(z *miniov2.Pool) []corev1.Toleration {
	var tolerations []corev1.Toleration
	return append(tolerations, z.Tolerations...)
}

// Builds the topology spread constraints for a Pool.
func poolTopologySpreadConstraints(z *miniov2.Pool) []corev1.TopologySpreadConstraint {
	var constraints []corev1.TopologySpreadConstraint
	return append(constraints, z.TopologySpreadConstraints...)
}

// Builds the security context for a Pool
func poolSecurityContext(pool *miniov2.Pool, status *miniov2.PoolStatus) *v1.PodSecurityContext {
	runAsNonRoot := true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	var fsGroup int64 = 1000
	fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch

	securityContext := corev1.PodSecurityContext{
		RunAsNonRoot:        &runAsNonRoot,
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &fsGroupChangePolicy,
	}

	if pool != nil && pool.SecurityContext != nil {
		securityContext = *pool.SecurityContext
		// if the pool has no security context, and it's market as legacy security context return nil
	} else if status.LegacySecurityContext {
		return nil
	}
	// Prevents high CPU usage of kubelet by preventing chown on the entire CSI
	if securityContext.FSGroupChangePolicy == nil {
		fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch
		securityContext.FSGroupChangePolicy = &fsGroupChangePolicy
	}
	return &securityContext
}

// Builds the security context for containers in a Pool
func poolContainerSecurityContext(pool *miniov2.Pool) *v1.SecurityContext {
	// Default values:
	// By default, values should be totally empty if not provided
	// This is specially needed in OpenShift where Security Context Constraints restrict them
	// if let empty then OCP can pick the values from the constraints defined.
	containerSecurityContext := corev1.SecurityContext{}
	runAsNonRoot := true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	poolSCSet := false

	// Values from pool.SecurityContext ONLY if provided
	if pool.SecurityContext != nil {
		if pool.SecurityContext.RunAsNonRoot != nil {
			runAsNonRoot = *pool.SecurityContext.RunAsNonRoot
			poolSCSet = true
		}
		if pool.SecurityContext.RunAsUser != nil {
			runAsUser = *pool.SecurityContext.RunAsUser
			poolSCSet = true
		}
		if pool.SecurityContext.RunAsGroup != nil {
			runAsGroup = *pool.SecurityContext.RunAsGroup
			poolSCSet = true
		}
		if poolSCSet {
			// Only set values if one of above is set otherwise let it empty
			containerSecurityContext = corev1.SecurityContext{
				RunAsNonRoot: &runAsNonRoot,
				RunAsUser:    &runAsUser,
				RunAsGroup:   &runAsGroup,
			}
		}
	}

	// Values from pool.ContainerSecurityContext if provided
	if pool != nil && pool.ContainerSecurityContext != nil {
		containerSecurityContext = *pool.ContainerSecurityContext
	}
	return &containerSecurityContext
}

// CfgVol is the name of the configuration volume we will use
const CfgVol = "cfg-vol"

// NewPoolArgs arguments used to create a new pool
type NewPoolArgs struct {
	Tenant          *miniov2.Tenant
	SkipEnvVars     map[string][]byte
	Pool            *miniov2.Pool
	PoolStatus      *miniov2.PoolStatus
	ServiceName     string
	HostsTemplate   string
	OperatorVersion string
	OperatorCATLS   bool
	OperatorImage   string
}

// NewPool creates a new StatefulSet for the given Cluster.
func NewPool(args *NewPoolArgs) *appsv1.StatefulSet {
	t := args.Tenant
	skipEnvVars := args.SkipEnvVars
	pool := args.Pool
	poolStatus := args.PoolStatus
	serviceName := args.ServiceName
	hostsTemplate := args.HostsTemplate
	operatorVersion := args.OperatorVersion
	operatorCATLS := args.OperatorCATLS
	operatorImage := args.OperatorImage

	var podVolumes []corev1.Volume
	replicas := pool.Servers
	var certVolumeSources []corev1.VolumeProjection

	var clientCertSecret string
	clientCertPaths := []corev1.KeyToPath{
		{Key: "public.crt", Path: "client.crt"},
		{Key: "private.key", Path: "client.key"},
	}
	var kesCertSecret string
	KESCertPath := []corev1.KeyToPath{
		{Key: "public.crt", Path: "CAs/kes.crt"},
	}

	// Create an empty dir volume to share the configuration between the main container and side-car

	podVolumes = append(podVolumes, corev1.Volume{
		Name: CfgVol,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Multiple certificates will be mounted using the following folder structure:
	//
	//	certs
	//		+ public.crt
	//		+ private.key
	//		|
	//		+ hostname-0
	//		|			+ public.crt
	//		|			+ private.key
	//		+ hostname-1
	//		|			+ public.crt
	//		|			+ private.key
	//		+ hostname-2
	//		|			+ public.crt
	//		|			+ private.key
	//		+ CAs
	//			 + hostname-0.crt
	//			 + hostname-1.crt
	//			 + hostname-2.crt
	//			 + public.crt
	//
	// Iterate over all provided TLS certificates and store them on the list of Volumes that will be mounted to the Pod
	for index, secret := range t.Spec.ExternalCertSecret {
		crtMountPath := fmt.Sprintf("hostname-%d/public.crt", index)
		keyMountPath := fmt.Sprintf("hostname-%d/private.key", index)
		caMountPath := fmt.Sprintf("CAs/hostname-%d.crt", index)
		// MinIO requires to have at least 1 certificate keyPair under the `certs` folder, by default
		// we will take the first secret as the default certificate
		//
		//	certs
		//		+ public.crt
		//		+ private.key
		if index == 0 {
			crtMountPath = "public.crt"
			keyMountPath = "private.key"
			caMountPath = "CAs/public.crt"
		}

		var serverCertPaths []corev1.KeyToPath
		if secret.Type == "kubernetes.io/tls" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: crtMountPath},
				{Key: "tls.key", Path: keyMountPath},
				{Key: "tls.crt", Path: caMountPath},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: crtMountPath},
				{Key: "tls.key", Path: keyMountPath},
				{Key: "ca.crt", Path: caMountPath},
			}
		} else {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "public.crt", Path: crtMountPath},
				{Key: "private.key", Path: keyMountPath},
				{Key: "public.crt", Path: caMountPath},
			}
		}
		certVolumeSources = append(certVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Items: serverCertPaths,
			},
		})
	}
	// AutoCert certificates will be used for internal communication if requested
	if t.AutoCert() {
		crtMountPath := "public.crt"
		keyMountPath := "private.key"
		caMountPath := "CAs/public.crt"
		if len(t.Spec.ExternalCertSecret) > 0 {
			index := len(t.Spec.ExternalCertSecret)
			crtMountPath = fmt.Sprintf("hostname-%d/public.crt", index)
			keyMountPath = fmt.Sprintf("hostname-%d/private.key", index)
			caMountPath = fmt.Sprintf("CAs/hostname-%d.crt", index)
		}
		certVolumeSources = append(certVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.MinIOTLSSecretName(),
				},
				Items: []corev1.KeyToPath{
					{Key: "public.crt", Path: crtMountPath},
					{Key: "private.key", Path: keyMountPath},
					{Key: "public.crt", Path: caMountPath},
				},
			},
		})
	}
	// Multiple client certificates will be mounted using the following folder structure:
	//
	//	certs
	//		|
	//		+ client-0
	//		|			+ client.crt
	//		|			+ client.key
	//		+ client-1
	//		|			+ client.crt
	//		|			+ client.key
	//		+ client-2
	//		|			+ client.crt
	//		|			+ client.key
	//
	// Iterate over all provided client TLS certificates and store them on the list of Volumes that will be mounted to the Pod
	for index, secret := range t.Spec.ExternalClientCertSecrets {
		crtMountPath := fmt.Sprintf("client-%d/client.crt", index)
		keyMountPath := fmt.Sprintf("client-%d/client.key", index)
		var clientKeyPairPaths []corev1.KeyToPath
		if secret.Type == "kubernetes.io/tls" {
			clientKeyPairPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: crtMountPath},
				{Key: "tls.key", Path: keyMountPath},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			clientKeyPairPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: crtMountPath},
				{Key: "tls.key", Path: keyMountPath},
				{Key: "ca.crt", Path: fmt.Sprintf("CAs/client-ca-%d.crt", index)},
			}
		} else {
			clientKeyPairPaths = []corev1.KeyToPath{
				{Key: "public.crt", Path: crtMountPath},
				{Key: "private.key", Path: keyMountPath},
			}
		}
		certVolumeSources = append(certVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Items: clientKeyPairPaths,
			},
		})
	}

	// Will mount into ~/.minio/certs/CAs folder the user provided CA certificates.
	// This is used for MinIO to verify TLS connections with other applications.
	//	certs
	//		+ CAs
	//			 + ca-0.crt
	//			 + ca-1.crt
	//			 + ca-2.crt
	for index, secret := range t.Spec.ExternalCaCertSecret {
		var caCertPaths []corev1.KeyToPath
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if secret.Type == "kubernetes.io/tls" {
			caCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: fmt.Sprintf("CAs/ca-%d.crt", index)},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			caCertPaths = []corev1.KeyToPath{
				{Key: "ca.crt", Path: fmt.Sprintf("CAs/ca-%d.crt", index)},
			}
		} else {
			caCertPaths = []corev1.KeyToPath{
				{Key: "public.crt", Path: fmt.Sprintf("CAs/ca-%d.crt", index)},
			}
		}
		certVolumeSources = append(certVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Items: caCertPaths,
			},
		})
	}

	if operatorCATLS {
		// Mount Operator CA TLS certificate to MinIO ~/cert/CAs
		operatorCATLSSecretName := "operator-ca-tls"
		certVolumeSources = append(certVolumeSources, []corev1.VolumeProjection{
			{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: operatorCATLSSecretName,
					},
					Items: []corev1.KeyToPath{
						{Key: "public.crt", Path: "CAs/operator-ca.crt"},
					},
				},
			},
		}...)
	}
	// If KES is enable mount TLS certificate secrets
	if t.HasKESEnabled() {
		// External Client certificates will have priority over AutoCert generated certificates
		if t.ExternalClientCert() {
			clientCertSecret = t.Spec.ExternalClientCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" / cert-manager.io/v1 because of same keys in both.
			if t.Spec.ExternalClientCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalClientCertSecret.Type == "cert-manager.io/v1alpha2" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1" {
				clientCertPaths = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "client.crt"},
					{Key: "tls.key", Path: "client.key"},
				}
			} else {
				clientCertPaths = []corev1.KeyToPath{
					{Key: "public.crt", Path: "client.crt"},
					{Key: "private.key", Path: "client.key"},
				}
			}
		} else {
			clientCertSecret = t.MinIOClientTLSSecretName()
		}

		// KES External certificates will have priority over AutoCert generated certificates
		if t.KESExternalCert() {
			kesCertSecret = t.Spec.KES.ExternalCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if t.Spec.KES.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1" {
				KESCertPath = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "CAs/kes.crt"},
				}
			}
		} else {
			kesCertSecret = t.KESTLSSecretName()
		}

		certVolumeSources = append(certVolumeSources, []corev1.VolumeProjection{
			{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: clientCertSecret,
					},
					Items: clientCertPaths,
				},
			},
			{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: kesCertSecret,
					},
					Items: KESCertPath,
				},
			},
		}...)
	}

	if len(certVolumeSources) > 0 {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: t.MinIOTLSSecretName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: certVolumeSources,
				},
			},
		})
	}

	if t.HasConfigurationSecret() {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: "configuration",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: t.Spec.Configuration.Name,
								},
							},
						},
					},
				},
			},
		})
	}

	ssMeta := metav1.ObjectMeta{
		Namespace: t.Namespace,
		Name:      t.PoolStatefulsetName(pool),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(t, schema.GroupVersionKind{
				Group:   miniov2.SchemeGroupVersion.Group,
				Version: miniov2.SchemeGroupVersion.Version,
				Kind:    miniov2.MinIOCRDResourceKind,
			}),
		},
	}
	// Copy labels and annotations from the Tenant.Spec.Metadata
	ssMeta.Labels = t.ObjectMeta.Labels
	ssMeta.Annotations = t.ObjectMeta.Annotations

	if ssMeta.Labels == nil {
		ssMeta.Labels = make(map[string]string)
	}

	// Add information labels, such as which pool we are building this pod about
	ssMeta.Labels[miniov2.PoolLabel] = pool.Name
	ssMeta.Labels[miniov2.TenantLabel] = t.Name

	// Add user specific annotations
	if pool.Annotations != nil {
		ssMeta.Annotations = miniov2.MergeMaps(ssMeta.Annotations, pool.Annotations)
	}

	if pool.Labels != nil {
		ssMeta.Labels = miniov2.MergeMaps(ssMeta.Labels, pool.Labels)
	}

	containers := []corev1.Container{
		poolMinioServerContainer(t, skipEnvVars, pool, hostsTemplate, operatorVersion, certVolumeSources),
		getSideCarContainer(t, operatorImage, pool),
	}

	// attach any sidecar containers and volumes
	if t.Spec.SideCars != nil && len(t.Spec.SideCars.Containers) > 0 {
		containers = append(containers, t.Spec.SideCars.Containers...)
		podVolumes = append(podVolumes, t.Spec.SideCars.Volumes...)
	}

	moreServers := pool.Servers > 4
	unavailable := intstr.FromInt(1)
	if moreServers {
		// if we have more servers than 4 nodes in a statefulset,
		// we can allow rolling update strategy to roll two pods
		// at a time instead of just '1' (default), MinIO servers
		// beyond 4 servers are capable of tolerating a minimum
		// of 2 replicas being down - this allows for faster
		// rolling update strategy for all our deployments.
		unavailable = intstr.FromInt(2)
	}

	initContainer := getInitContainer(t, operatorImage, pool)

	ss := &appsv1.StatefulSet{
		ObjectMeta: ssMeta,
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov2.DefaultUpdateStrategy,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &unavailable,
				},
			},
			PodManagementPolicy: t.Spec.PodManagementPolicy,
			Selector:            ContainerMatchLabels(t, pool),
			ServiceName:         serviceName,
			Replicas:            &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: PodMetadata(t, pool),
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						initContainer,
					},
					Containers:                containers,
					Volumes:                   podVolumes,
					RestartPolicy:             corev1.RestartPolicyAlways,
					Affinity:                  pool.Affinity,
					NodeSelector:              pool.NodeSelector,
					SchedulerName:             t.Scheduler.Name,
					Tolerations:               poolTolerations(pool),
					TopologySpreadConstraints: poolTopologySpreadConstraints(pool),
					SecurityContext:           poolSecurityContext(pool, poolStatus),
					ServiceAccountName:        t.Spec.ServiceAccountName,
					PriorityClassName:         t.Spec.PriorityClassName,
				},
			},
		},
	}

	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	if pool.VolumeClaimTemplate != nil {
		pvClaim := *pool.VolumeClaimTemplate
		name := pvClaim.Name
		for i := 0; i < int(pool.VolumesPerServer); i++ {
			pvClaim.Name = name + strconv.Itoa(i)
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, pvClaim)
		}
	}
	// attach any sidecar containers and volumes
	if t.Spec.SideCars != nil && len(t.Spec.SideCars.VolumeClaimTemplates) > 0 {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, t.Spec.SideCars.VolumeClaimTemplates...)
	}

	// pass on RuntimeClassName
	if pool.RuntimeClassName != nil && *pool.RuntimeClassName != "" {
		ss.Spec.Template.Spec.RuntimeClassName = pool.RuntimeClassName
	}
	// add customs initContainers to StatefulSet
	if len(t.Spec.InitContainers) != 0 {
		ss.Spec.Template.Spec.InitContainers = append(ss.Spec.Template.Spec.InitContainers, t.Spec.InitContainers...)
	}
	return ss
}

func getInitContainer(t *miniov2.Tenant, operatorImage string, pool *miniov2.Pool) v1.Container {
	initContainer := corev1.Container{
		Name:  "validate-arguments",
		Image: operatorImage,
		Args: []string{
			"validate",
			"--tenant",
			t.Name,
		},
		Env: []corev1.EnvVar{
			{
				Name:  "CLUSTER_DOMAIN",
				Value: miniov2.GetClusterDomain(),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			CfgVolumeMount,
		},
		SecurityContext: poolContainerSecurityContext(pool),
	}
	if t.HasConfigurationSecret() {
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, TmpCfgVolumeMount)
	}
	return initContainer
}

func getSideCarContainer(t *miniov2.Tenant, operatorImage string, pool *miniov2.Pool) v1.Container {
	sidecarContainer := corev1.Container{
		Name:  "sidecar",
		Image: operatorImage,
		Args: []string{
			"sidecar",
			"--tenant",
			t.Name,
			"--config-name",
			t.Spec.Configuration.Name,
		},
		Env: []corev1.EnvVar{
			{
				Name:  "CLUSTER_DOMAIN",
				Value: miniov2.GetClusterDomain(),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			CfgVolumeMount,
		},
		SecurityContext: poolContainerSecurityContext(pool),
	}
	if t.HasConfigurationSecret() {
		sidecarContainer.VolumeMounts = append(sidecarContainer.VolumeMounts, TmpCfgVolumeMount)
	}
	return sidecarContainer
}
