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

	miniov1 "github.com/minio/minio-operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Returns the MinIO environment variables set in configuration.
// If a user specifies a secret in the spec (for MinIO credentials) we use
// that to set MINIO_ACCESS_KEY & MINIO_SECRET_KEY.
func minioEnvironmentVars(t *miniov1.Tenant) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	// Add all the environment variables
	envVars = append(envVars, t.Spec.Env...)
	// Add env variables from credentials secret, if no secret provided, dont use
	// env vars. MinIO server automatically creates default credentials
	if t.HasCredsSecret() {
		secretName := t.Spec.CredsSecret.Name
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
	if t.HasKESEnabled() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MINIO_KMS_KES_ENDPOINT",
			Value: "https://" + net.JoinHostPort(t.KESServiceHost(), strconv.Itoa(miniov1.KESPort)),
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
func minioMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if t.HasMetadata() {
		meta = *t.Spec.Metadata
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	// Add the additional label used by StatefulSet spec selector
	for k, v := range t.MinIOPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// MinIOSelector Returns the MinIO pods selector set in configuration.
func MinIOSelector(t *miniov1.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.MinIOPodLabels(),
	}
}

// Builds the volume mounts for MinIO container.
func volumeMounts(t *miniov1.Tenant, zone *miniov1.Zone) (mounts []corev1.VolumeMount) {
	// This is the case where user didn't provide a zone and we deploy a EmptyDir based
	// single node single drive (FS) MinIO deployment
	name := miniov1.MinIOVolumeName
	if zone.VolumeClaimTemplate != nil {
		name = zone.VolumeClaimTemplate.Name
	}

	if zone.VolumesPerServer == 1 {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name + strconv.Itoa(0),
			MountPath: t.Spec.Mountpath,
		})
	} else {
		for i := 0; i < int(zone.VolumesPerServer); i++ {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      name + strconv.Itoa(i),
				MountPath: t.Spec.Mountpath + strconv.Itoa(i),
			})
		}
	}

	if t.AutoCert() || t.ExternalCert() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      t.MinIOTLSSecretName(),
			MountPath: miniov1.MinIOCertPath,
		})
	}

	return mounts
}

func probes(t *miniov1.Tenant) (liveness *corev1.Probe) {
	scheme := corev1.URIScheme(strings.ToUpper(miniov1.Scheme))
	port := intstr.IntOrString{
		IntVal: int32(miniov1.MinIOPort),
	}

	if t.Spec.Liveness != nil {
		liveness = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   miniov1.LivenessPath,
					Port:   port,
					Scheme: scheme,
				},
			},
			InitialDelaySeconds: t.Spec.Liveness.InitialDelaySeconds,
			PeriodSeconds:       t.Spec.Liveness.PeriodSeconds,
			TimeoutSeconds:      t.Spec.Liveness.TimeoutSeconds,
		}
	}

	return liveness
}

// Builds the MinIO container for a Tenant.
func zoneMinioServerContainer(t *miniov1.Tenant, zone *miniov1.Zone, hostsTemplate string) corev1.Container {
	args := GetContainerArgs(t, hostsTemplate)

	liveProbe := probes(t)

	return corev1.Container{
		Name:  miniov1.MinIOServerName,
		Image: t.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.MinIOPort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		VolumeMounts:    volumeMounts(t, zone),
		Args:            args,
		Env:             minioEnvironmentVars(t),
		Resources:       zone.Resources,
		LivenessProbe:   liveProbe,
	}
}

// GetContainerArgs returns the arguments that the MinIO container receives
func GetContainerArgs(t *miniov1.Tenant, hostsTemplate string) []string {
	args := []string{"server", "--certs-dir", miniov1.MinIOCertPath}

	if len(t.Spec.Zones) == 1 && t.Spec.Zones[0].Servers == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, t.VolumePathForZone(&t.Spec.Zones[0]))
	} else {
		// append all the Tenant replica URLs
		hosts := t.MinIOHosts()
		if hostsTemplate != "" {
			hosts = t.TemplatedMinIOHosts(hostsTemplate)
		}
		for hi, h := range hosts {
			args = append(args, fmt.Sprintf("%s://"+h+"%s", miniov1.Scheme, t.VolumePathForZone(&t.Spec.Zones[hi])))
		}
	}
	return args
}

// Builds the tolerations for a Zone.
func minioZoneTolerations(z *miniov1.Zone) []corev1.Toleration {
	var tolerations []corev1.Toleration
	return append(tolerations, z.Tolerations...)
}

// Builds the security context for a Tenant
func minioSecurityContext(t *miniov1.Tenant) *corev1.PodSecurityContext {
	var securityContext = corev1.PodSecurityContext{}
	if t.Spec.SecurityContext != nil {
		securityContext = *t.Spec.SecurityContext
	}
	return &securityContext
}

// NewForMinIOZone creates a new StatefulSet for the given Cluster.
func NewForMinIOZone(t *miniov1.Tenant, zone *miniov1.Zone, serviceName string, hostsTemplate string) *appsv1.StatefulSet {
	var podVolumes []corev1.Volume
	var replicas = zone.Servers
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

	if t.AutoCert() {
		serverCertSecret = t.MinIOTLSSecretName()
		clientCertSecret = t.MinIOClientTLSSecretName()
		kesCertSecret = t.KESTLSSecretName()
	} else if t.ExternalCert() {
		serverCertSecret = t.Spec.ExternalCertSecret.Name
		if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "tls.crt", Path: "CAs/public.crt"},
			}
		} else if t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "ca.crt", Path: "CAs/public.crt"},
			}
		}
		if t.ExternalClientCert() {
			clientCertSecret = t.Spec.ExternalClientCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
				clientCertPaths = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "client.crt"},
					{Key: "tls.key", Path: "client.key"},
				}
			}
		}
		if t.KESExternalCert() {
			kesCertSecret = t.Spec.KES.ExternalCertSecret.Name
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
				KESCertPath = []corev1.KeyToPath{
					{Key: "tls.crt", Path: "CAs/kes.crt"},
				}
			}
		}
	}

	// Add SSL volume from SSL secret to the podVolumes
	if t.AutoCert() || t.ExternalCert() {
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
		if t.HasKESEnabled() {
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
			Name: t.MinIOTLSSecretName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: sources,
				},
			},
		})
	}

	containers := []corev1.Container{zoneMinioServerContainer(t, zone, hostsTemplate)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Namespace,
			Name:      t.ZoneStatefulsetName(zone),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(t, schema.GroupVersionKind{
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
			PodManagementPolicy: t.Spec.PodManagementPolicy,
			Selector:            MinIOSelector(t),
			ServiceName:         serviceName,
			Replicas:            &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: minioMetadata(t),
				Spec: corev1.PodSpec{
					Containers:         containers,
					Volumes:            podVolumes,
					RestartPolicy:      corev1.RestartPolicyAlways,
					Affinity:           zone.Affinity,
					SchedulerName:      t.Scheduler.Name,
					Tolerations:        minioZoneTolerations(zone),
					SecurityContext:    minioSecurityContext(t),
					ServiceAccountName: t.Spec.ServiceAccountName,
				},
			},
		},
	}

	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	if zone.VolumeClaimTemplate != nil {
		pvClaim := *zone.VolumeClaimTemplate
		name := pvClaim.Name
		for i := 0; i < int(zone.VolumesPerServer); i++ {
			pvClaim.Name = name + strconv.Itoa(i)
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, pvClaim)
		}
	}
	return ss
}
