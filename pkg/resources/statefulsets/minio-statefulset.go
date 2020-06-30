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

package statefulsets

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Returns the MinIO environment variables set in configuration.
// If a user specifies a secret in the spec (for MinIO credentials) we use
// that to set MINIO_ACCESS_KEY & MINIO_SECRET_KEY.
func minioEnvironmentVars(mi *miniov1.MinIOInstance) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	// Add all the environment variables
	envVars = append(envVars, mi.Spec.Env...)
	// Add env variables from credentials secret, if no secret provided, dont use
	// env vars. MinIO server automatically creates default credentials
	if mi.HasCredsSecret() {
		secretName := mi.Spec.CredsSecret.Name
		envVars = append(envVars, corev1.EnvVar{
			Name: "MINIO_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "accesskey",
				},
			},
		}, corev1.EnvVar{
			Name: "MINIO_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "secretkey",
				},
			},
		})
	}
	if mi.HasKESEnabled() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_ENDPOINT",
			Value: "https://" + net.JoinHostPort(mi.KESServiceHost(), strconv.Itoa(miniov1.KESPort)),
		}, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CERT_FILE",
			Value: miniov1.MinIOCertPath + "/client.crt",
		}, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_FILE",
			Value: miniov1.MinIOCertPath + "/client.key",
		}, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CA_PATH",
			Value: miniov1.MinIOCertPath + "/CAs/kes.crt",
		}, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_NAME",
			Value: miniov1.KESMinIOKey,
		})
	}

	// Return environment variables
	return envVars
}

// Returns the MinIO pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func minioMetadata(mi *miniov1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if mi.HasMetadata() {
		meta = *mi.Spec.Metadata
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	// Add the additional label used by StatefulSet spec selector
	for k, v := range mi.MinIOPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// MinIOSelector Returns the MinIO pods selector set in configuration.
func MinIOSelector(mi *miniov1.MinIOInstance) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: mi.MinIOPodLabels(),
	}
}

// Builds the volume mounts for MinIO container.
func volumeMounts(mi *miniov1.MinIOInstance) (mounts []corev1.VolumeMount) {
	// This is the case where user didn't provide a zone and we deploy a EmptyDir based
	// single node single drive (FS) MinIO deployment
	name := miniov1.MinIOVolumeName
	if mi.Spec.VolumeClaimTemplate != nil {
		name = mi.Spec.VolumeClaimTemplate.Name
	}

	if mi.Spec.VolumesPerServer == 1 {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name + strconv.Itoa(0),
			MountPath: mi.Spec.Mountpath,
		})
	} else {
		for i := 0; i < mi.Spec.VolumesPerServer; i++ {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      name + strconv.Itoa(i),
				MountPath: mi.Spec.Mountpath + strconv.Itoa(i),
			})
		}
	}

	if mi.AutoCert() || mi.ExternalCert() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      mi.MinIOTLSSecretName(),
			MountPath: miniov1.MinIOCertPath,
		})
	}

	return mounts
}

func probes(mi *miniov1.MinIOInstance) (liveness *corev1.Probe) {
	scheme := corev1.URIScheme(strings.ToUpper(miniov1.Scheme))
	port := intstr.IntOrString{
		IntVal: int32(miniov1.MinIOPort),
	}

	if mi.Spec.Liveness != nil {
		liveness = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   miniov1.LivenessPath,
					Port:   port,
					Scheme: scheme,
				},
			},
			InitialDelaySeconds: mi.Spec.Liveness.InitialDelaySeconds,
			PeriodSeconds:       mi.Spec.Liveness.PeriodSeconds,
			TimeoutSeconds:      mi.Spec.Liveness.TimeoutSeconds,
		}
	}

	return liveness
}

// Builds the MinIO container for a MinIOInstance.
func minioServerContainer(mi *miniov1.MinIOInstance, serviceName string, hostsTemplate string) corev1.Container {
	args := []string{"server", "--certs-dir", miniov1.MinIOCertPath}

	if mi.Spec.Zones[0].Servers == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, mi.VolumePath())
	} else {
		// append all the MinIOInstance replica URLs
		hosts := mi.MinIOHosts()
		if hostsTemplate != "" {
			hosts = mi.TemplatedMinIOHosts(hostsTemplate)
		}
		for _, h := range hosts {
			args = append(args, fmt.Sprintf("%s://"+h+"%s.minio.local", miniov1.Scheme, mi.VolumePath()))
		}
	}

	liveProbe := probes(mi)

	return corev1.Container{
		Name:  miniov1.MinIOServerName,
		Image: mi.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.MinIOPort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		VolumeMounts:    volumeMounts(mi),
		Args:            args,
		Env:             minioEnvironmentVars(mi),
		Resources:       mi.Spec.Resources,
		LivenessProbe:   liveProbe,
	}
}

// Builds the tolerations for a MinIOInstance.
func minioTolerations(mi *miniov1.MinIOInstance) []corev1.Toleration {
	var tolerations []corev1.Toleration
	return append(tolerations, mi.Spec.Tolerations...)
}

// Builds the security context for a MinIOInstance
func minioSecurityContext(mi *miniov1.MinIOInstance) *corev1.PodSecurityContext {
	var securityContext = corev1.PodSecurityContext{}
	if mi.Spec.SecurityContext != nil {
		securityContext = *mi.Spec.SecurityContext
	}
	return &securityContext
}

func getVolumesForContainer(mi *miniov1.MinIOInstance) []corev1.Volume {
	var podVolumes = []corev1.Volume{}
	// This is the case where user didn't provide a volume claim template and we deploy a
	// EmptyDir based MinIO deployment
	if mi.Spec.VolumeClaimTemplate == nil {
		for _, z := range mi.Spec.Zones {
			podVolumes = append(podVolumes, corev1.Volume{Name: z.Name,
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""}}})
		}
	}
	return podVolumes
}

// NewForMinIO creates a new StatefulSet for the given Cluster.
func NewForMinIO(mi *miniov1.MinIOInstance, serviceName string, hostsTemplate string, discoIP string) *appsv1.StatefulSet {
	// If a PV isn't specified just use a EmptyDir volume
	var podVolumes = getVolumesForContainer(mi)
	var replicas = mi.MinIOReplicas()
	var serverCertSecret string
	var serverCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "public.crt"},
		{Key: "private.key", Path: "private.key"},
		{Key: "public.crt", Path: "CAs/public.crt"},
	}
	var clientCertSecret string
	var clientCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "client.crt"},
		{Key: "private.key", Path: "client.key"},
	}
	var kesCertSecret string
	var KESCertPath = []corev1.KeyToPath{
		{Key: "public.crt", Path: "CAs/kes.crt"},
	}

	if mi.AutoCert() {
		serverCertSecret = mi.MinIOTLSSecretName()
		clientCertSecret = mi.MinIOClientTLSSecretName()
		kesCertSecret = mi.KESTLSSecretName()
	} else if mi.ExternalCert() {
		serverCertSecret = mi.Spec.ExternalCertSecret.Name
		if mi.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "tls.crt", Path: "CAs/public.crt"},
			}
		} else if mi.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "ca.crt", Path: "CAs/public.crt"},
			}
		}
		if mi.ExternalClientCert() {
			clientCertSecret = mi.Spec.ExternalClientCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if mi.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || mi.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
				clientCertPaths = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "client.crt"},
					{Key: "tls.key", Path: "client.key"},
				}
			}
		}
		if mi.KESExternalCert() {
			kesCertSecret = mi.Spec.KES.ExternalCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if mi.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || mi.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
				KESCertPath = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "CAs/kes.crt"},
				}
			}
		}
	}

	// Add SSL volume from SSL secret to the podVolumes
	if mi.AutoCert() || mi.ExternalCert() {
		sources := []corev1.VolumeProjection{
			{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: serverCertSecret,
					},
					Items: serverCertPaths,
				},
			},
		}
		if mi.HasKESEnabled() {
			sources = append(sources, []corev1.VolumeProjection{
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
		podVolumes = append(podVolumes, corev1.Volume{
			Name: mi.MinIOTLSSecretName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: sources,
				},
			},
		})
	}

	containers := []corev1.Container{minioServerContainer(mi, serviceName, hostsTemplate)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mi.Namespace,
			Name:      mi.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1.SchemeGroupVersion.Group,
					Version: miniov1.SchemeGroupVersion.Version,
					Kind:    miniov1.MinIOCRDResourceKind,
				}),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov1.DefaultUpdateStrategy,
			},
			PodManagementPolicy: mi.Spec.PodManagementPolicy,
			Selector:            MinIOSelector(mi),
			ServiceName:         serviceName,
			Replicas:            &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: minioMetadata(mi),
				Spec: corev1.PodSpec{
					Containers:         containers,
					Volumes:            podVolumes,
					RestartPolicy:      corev1.RestartPolicyAlways,
					Affinity:           mi.Spec.Affinity,
					SchedulerName:      mi.Scheduler.Name,
					Tolerations:        minioTolerations(mi),
					SecurityContext:    minioSecurityContext(mi),
					ServiceAccountName: mi.Spec.ServiceAccountName,
				},
			},
		},
	}

	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if mi.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{mi.Spec.ImagePullSecret}
	}

	if mi.Spec.VolumeClaimTemplate != nil {
		pvClaim := *mi.Spec.VolumeClaimTemplate
		name := pvClaim.Name
		for i := 0; i < mi.Spec.VolumesPerServer; i++ {
			pvClaim.Name = name + strconv.Itoa(i)
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, pvClaim)
		}
	}

	// if the instance has io.min.disco annotation, then we know we should configure the dns for the pods as well
	if _, ok := ss.Spec.Template.ObjectMeta.Annotations["io.min.disco"]; ok {
		// configure the pod to register into disco
		ss.Spec.Template.ObjectMeta.Annotations["io.min.disco"] = fmt.Sprintf("{.metadata.name}.%s.minio.local", mi.MinIOHLServiceName())
		// set the DNS policy for the pod, this is not needed if disco is configured for the whole cluster
		ss.Spec.Template.Spec.DNSPolicy = corev1.DNSNone
		ss.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{discoIP},
		}

	}
	return ss
}
