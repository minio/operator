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
	"path/filepath"
	"sort"
	"strconv"

	"k8s.io/apimachinery/pkg/util/intstr"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/certs"
	"github.com/minio/operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Returns the MinIO environment variables set in configuration.
// If a user specifies a secret in the spec (for MinIO credentials) we use
// that to set MINIO_ROOT_USER & MINIO_ROOT_PASSWORD.
func minioEnvironmentVars(t *miniov2.Tenant, skipEnvVars map[string][]byte) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	envVarsMap := map[string]corev1.EnvVar{}

	envVarsMap["MINIO_CONFIG_ENV_FILE"] = corev1.EnvVar{
		Name:  "MINIO_CONFIG_ENV_FILE",
		Value: miniov2.CfgFile,
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
	if t.Spec.PoolsMetadata != nil {
		meta.Labels = t.Spec.PoolsMetadata.Labels
		meta.Annotations = t.Spec.PoolsMetadata.Annotations
	}
	meta.Labels = utils.MergeMaps(meta.Labels, pool.Labels, t.MinIOPodLabels(), t.ConsolePodLabels())
	meta.Annotations = utils.MergeMaps(meta.Annotations, pool.Annotations)

	// Set specific information
	meta.Labels[miniov2.PoolLabel] = pool.Name
	meta.Annotations[miniov2.Revision] = fmt.Sprintf("%d", t.Status.Revision)

	return meta
}

// ContainerMatchLabels Returns the labels that match the Pods in the statefulset
func ContainerMatchLabels(t *miniov2.Tenant, pool *miniov2.Pool) *metav1.LabelSelector {
	labels := utils.MergeMaps(t.MinIOPodLabels(), t.ConsolePodLabels())
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
func volumeMounts(t *miniov2.Tenant, pool *miniov2.Pool, certVolumeSources []corev1.VolumeProjection) (mounts []corev1.VolumeMount) {
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
func poolMinioServerContainer(t *miniov2.Tenant, skipEnvVars map[string][]byte, pool *miniov2.Pool, certVolumeSources []corev1.VolumeProjection) corev1.Container {
	consolePort := miniov2.ConsolePort
	if t.TLS() {
		consolePort = miniov2.ConsoleTLSPort
	}
	args := []string{
		"server",
		"--certs-dir", miniov2.MinIOCertPath,
		"--console-address", ":" + strconv.Itoa(consolePort),
	}

	containerPorts := []corev1.ContainerPort{
		{
			Name:          miniov2.MinIOPortName,
			ContainerPort: miniov2.MinIOPort,
		},
		{
			Name:          miniov2.ConsolePortName,
			ContainerPort: int32(consolePort),
		},
	}

	if t.Spec.Features != nil && t.Spec.Features.EnableSFTP != nil && *t.Spec.Features.EnableSFTP {
		pkFile := filepath.Join(miniov2.MinIOCertPath, certs.PrivateKeyFile)
		args = append(args, []string{
			"--sftp", fmt.Sprintf("address=:%d", miniov2.MinIOSFTPPort),
			"--sftp", "ssh-private-key=" + pkFile,
		}...)
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          miniov2.MinIOSFTPPortName,
			ContainerPort: miniov2.MinIOSFTPPort,
		})
	}

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
		Name:            miniov2.MinIOServerName,
		Image:           t.Spec.Image,
		Ports:           containerPorts,
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    volumeMounts(t, pool, certVolumeSources),
		Args:            args,
		Env:             minioEnvironmentVars(t, skipEnvVars),
		Resources:       pool.Resources,
		LivenessProbe:   t.Spec.Liveness,
		ReadinessProbe:  t.Spec.Readiness,
		StartupProbe:    t.Spec.Startup,
		Lifecycle:       t.Spec.Lifecycle,
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
func poolSecurityContext(pool *miniov2.Pool, status *miniov2.PoolStatus) *corev1.PodSecurityContext {
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
func poolContainerSecurityContext(pool *miniov2.Pool) *corev1.SecurityContext {
	// By default, we are opinionated and set the following values to request
	// kubernetes to run our pods as a non-root user intentionally, we don't need to be root
	// if the user needs a special security context, it should be specified on the pool's
	// securityContext
	runAsNonRoot := true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	if pool != nil && pool.SecurityContext != nil {
		if pool.SecurityContext.RunAsNonRoot != nil {
			runAsNonRoot = *pool.SecurityContext.RunAsNonRoot
		}
		if pool.SecurityContext.RunAsUser != nil {
			runAsUser = *pool.SecurityContext.RunAsUser
		}
		if pool.SecurityContext.RunAsGroup != nil {
			runAsGroup = *pool.SecurityContext.RunAsGroup
		}
	}

	containerSecurityContext := corev1.SecurityContext{
		RunAsNonRoot: &runAsNonRoot,
		RunAsUser:    &runAsUser,
		RunAsGroup:   &runAsGroup,
	}

	// Values from pool.ContainerSecurityContext if provided
	if pool.ContainerSecurityContext != nil {
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
}

// NewPool creates a new StatefulSet for the given Cluster.
func NewPool(args *NewPoolArgs) *appsv1.StatefulSet {
	t := args.Tenant.DeepCopy()
	skipEnvVars := args.SkipEnvVars
	pool := args.Pool
	poolStatus := args.PoolStatus
	serviceName := args.ServiceName

	var podVolumes []corev1.Volume
	replicas := pool.Servers
	var certVolumeSources []corev1.VolumeProjection

	var clientCertSecret string
	clientCertPaths := []corev1.KeyToPath{
		{Key: certs.PublicCertFile, Path: "client.crt"},
		{Key: certs.PrivateKeyFile, Path: "client.key"},
	}
	var kesCertSecret string
	KESCertPath := []corev1.KeyToPath{
		{Key: certs.PublicCertFile, Path: "CAs/kes.crt"},
	}

	// Create an empty dir volume to share the configuration between the main container and side-car

	podVolumes = append(podVolumes, corev1.Volume{
		Name: CfgVol,
		VolumeSource: corev1.VolumeSource{
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
		crtMountPath := fmt.Sprintf("hostname-%d/%s", index, certs.PublicCertFile)
		keyMountPath := fmt.Sprintf("hostname-%d/%s", index, certs.PrivateKeyFile)
		caMountPath := fmt.Sprintf("CAs/hostname-%d.crt", index)
		// MinIO requires to have at least 1 certificate keyPair under the `certs` folder, by default
		// we will take the first secret as the default certificate
		//
		//	certs
		//		+ public.crt
		//		+ private.key
		if index == 0 {
			crtMountPath = certs.PublicCertFile
			keyMountPath = certs.PrivateKeyFile
			caMountPath = fmt.Sprintf("%s/%s", certs.CertsCADir, certs.PublicCertFile)
		}

		var serverCertPaths []corev1.KeyToPath
		if secret.Type == "kubernetes.io/tls" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: certs.TLSCertFile, Path: crtMountPath},
				{Key: certs.TLSKeyFile, Path: keyMountPath},
				{Key: certs.TLSCertFile, Path: caMountPath},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: certs.TLSCertFile, Path: crtMountPath},
				{Key: certs.TLSKeyFile, Path: keyMountPath},
				{Key: certs.CAPublicCertFile, Path: caMountPath},
			}
		} else {
			serverCertPaths = []corev1.KeyToPath{
				{Key: certs.PublicCertFile, Path: crtMountPath},
				{Key: certs.PrivateKeyFile, Path: keyMountPath},
				{Key: certs.PublicCertFile, Path: caMountPath},
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
		crtMountPath := certs.PublicCertFile
		keyMountPath := certs.PrivateKeyFile
		caMountPath := fmt.Sprintf("%s/%s", certs.CertsCADir, certs.PublicCertFile)
		if len(t.Spec.ExternalCertSecret) > 0 {
			index := len(t.Spec.ExternalCertSecret)
			crtMountPath = fmt.Sprintf("hostname-%d/%s", index, certs.PublicCertFile)
			keyMountPath = fmt.Sprintf("hostname-%d/%s", index, certs.PrivateKeyFile)
			caMountPath = fmt.Sprintf("CAs/hostname-%d.crt", index)
		}
		certVolumeSources = append(certVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.MinIOTLSSecretName(),
				},
				Items: []corev1.KeyToPath{
					{Key: certs.PublicCertFile, Path: crtMountPath},
					{Key: certs.PrivateKeyFile, Path: keyMountPath},
					{Key: certs.PublicCertFile, Path: caMountPath},
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
				{Key: certs.TLSCertFile, Path: crtMountPath},
				{Key: certs.TLSKeyFile, Path: keyMountPath},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			clientKeyPairPaths = []corev1.KeyToPath{
				{Key: certs.TLSCertFile, Path: crtMountPath},
				{Key: certs.TLSKeyFile, Path: keyMountPath},
				{Key: certs.CAPublicCertFile, Path: fmt.Sprintf("%s/client-ca-%d.crt", certs.CertsCADir, index)},
			}
		} else {
			clientKeyPairPaths = []corev1.KeyToPath{
				{Key: certs.PublicCertFile, Path: crtMountPath},
				{Key: certs.PrivateKeyFile, Path: keyMountPath},
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
				{Key: certs.TLSCertFile, Path: fmt.Sprintf("%s/ca-%d.crt", certs.CertsCADir, index)},
			}
		} else if secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
			caCertPaths = []corev1.KeyToPath{
				{Key: certs.CAPublicCertFile, Path: fmt.Sprintf("%s/ca-%d.crt", certs.CertsCADir, index)},
			}
		} else {
			caCertPaths = []corev1.KeyToPath{
				{Key: certs.PublicCertFile, Path: fmt.Sprintf("%s/ca-%d.crt", certs.CertsCADir, index)},
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

	// If KES is enable mount TLS certificate secrets
	if t.HasKESEnabled() {
		// External Client certificates will have priority over AutoCert generated certificates
		if t.ExternalClientCert() {
			clientCertSecret = t.Spec.ExternalClientCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" / cert-manager.io/v1 because of same keys in both.
			if t.Spec.ExternalClientCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalClientCertSecret.Type == "cert-manager.io/v1alpha2" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1" {
				clientCertPaths = []corev1.KeyToPath{
					{Key: certs.TLSCertFile, Path: "client.crt"},
					{Key: certs.TLSKeyFile, Path: "client.key"},
				}
			} else {
				clientCertPaths = []corev1.KeyToPath{
					{Key: certs.PublicCertFile, Path: "client.crt"},
					{Key: certs.PrivateKeyFile, Path: "client.key"},
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
					{Key: certs.TLSCertFile, Path: fmt.Sprintf("%s/kes.crt", certs.CertsCADir)},
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
	// unless `StatefulSetMetadata` is defined, then we'll copy it
	// from there.
	if t.Spec.PoolsMetadata != nil {
		ssMeta.Labels = t.Spec.PoolsMetadata.Labels
		ssMeta.Annotations = t.Spec.PoolsMetadata.Annotations
	}

	// Add pool specific annotations
	ssMeta.Annotations = utils.MergeMaps(ssMeta.Annotations, pool.Annotations)
	ssMeta.Labels = utils.MergeMaps(ssMeta.Labels, pool.Labels)

	// Add information labels, such as which pool we are building this pod about
	ssMeta.Labels[miniov2.PoolLabel] = pool.Name
	ssMeta.Labels[miniov2.TenantLabel] = t.Name

	containers := []corev1.Container{
		poolMinioServerContainer(t, skipEnvVars, pool, certVolumeSources),
		getSideCarContainer(t, pool),
	}

	// attach any sidecar containers and volumes
	if t.Spec.SideCars != nil && len(t.Spec.SideCars.Containers) > 0 {
		containers = append(containers, t.Spec.SideCars.Containers...)
		podVolumes = append(podVolumes, t.Spec.SideCars.Volumes...)
	}

	initContainer := getInitContainer(t, pool)

	ss := &appsv1.StatefulSet{
		ObjectMeta: ssMeta,
		Spec: appsv1.StatefulSetSpec{
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
	// add customs VolumeMounts and Volumes.
	if len(t.Spec.AdditionalVolumeMounts) != 0 && len(t.Spec.AdditionalVolumes) == len(t.Spec.AdditionalVolumeMounts) {
		for i := range ss.Spec.Template.Spec.InitContainers {
			ss.Spec.Template.Spec.InitContainers[i].VolumeMounts = append(ss.Spec.Template.Spec.InitContainers[i].VolumeMounts, t.Spec.AdditionalVolumeMounts...)
		}
		for i := range ss.Spec.Template.Spec.Containers {
			ss.Spec.Template.Spec.Containers[i].VolumeMounts = append(ss.Spec.Template.Spec.Containers[i].VolumeMounts, t.Spec.AdditionalVolumeMounts...)
		}
		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, t.Spec.AdditionalVolumes...)
	}
	return ss
}

func getInitContainer(t *miniov2.Tenant, pool *miniov2.Pool) corev1.Container {
	initContainer := corev1.Container{
		Name:  "validate-arguments",
		Image: getSidecarImage(),
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
	// That's ok to use the sidecar's resource
	if t.Spec.SideCars != nil && t.Spec.SideCars.Resources != nil {
		initContainer.Resources = *t.Spec.SideCars.Resources
	}
	if t.HasConfigurationSecret() {
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, TmpCfgVolumeMount)
	}
	return initContainer
}

func getSideCarContainer(t *miniov2.Tenant, pool *miniov2.Pool) corev1.Container {
	scheme := corev1.URISchemeHTTP

	readinessProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/ready",
				Port: intstr.IntOrString{
					IntVal: 4444,
				},
				// Host:        "localhost",
				Scheme:      scheme,
				HTTPHeaders: nil,
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       1,
		FailureThreshold:    1,
	}

	sidecarContainer := corev1.Container{
		Name:  "sidecar",
		Image: getSidecarImage(),
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
		ReadinessProbe:  readinessProbe,
	}
	if t.Spec.SideCars != nil && t.Spec.SideCars.Resources != nil {
		sidecarContainer.Resources = *t.Spec.SideCars.Resources
	}
	if t.HasConfigurationSecret() {
		sidecarContainer.VolumeMounts = append(sidecarContainer.VolumeMounts, TmpCfgVolumeMount)
	}
	return sidecarContainer
}
