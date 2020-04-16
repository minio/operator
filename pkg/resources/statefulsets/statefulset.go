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

package statefulsets

import (
	"fmt"
	"strconv"

	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	constants "github.com/minio/minio-operator/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Returns the MinIO environment variables set in configuration.
// If a user specifies a secret in the spec (for MinIO credentials) we use
// that to set MINIO_ACCESS_KEY & MINIO_SECRET_KEY.
func minioEnvironmentVars(mi *miniov1beta1.MinIOInstance) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0)
	// Add all the environment variables
	for _, e := range mi.Spec.Env {
		envVars = append(envVars, e)
	}
	// Add env variables from credentials secret, if no secret provided, dont use
	// env vars. MinIO server automatically creates default credentials
	if mi.HasCredsSecret() {
		var secretName string
		secretName = mi.Spec.CredsSecret.Name
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
	// Return environment variables
	return envVars
}

// Returns the MinIO pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func minioMetadata(mi *miniov1beta1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if mi.HasMetadata() {
		meta = *mi.Spec.Metadata
	}
	// Initialize empty fields
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	// Add the additional label used by StatefulSet spec
	for k, v := range mi.PodLabels() {
		meta.Labels[k] = v
	}
	// Add the Selector labels set by user
	if mi.HasSelector() {
		for k, v := range mi.Spec.Selector.MatchLabels {
			meta.Labels[k] = v
		}
	}
	return meta
}

// Builds the volume mounts for MinIO container.
func volumeMounts(mi *miniov1beta1.MinIOInstance) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	// This is the case where user didn't provide a zone and we deploy a EmptyDir based
	// single node single drive (FS) MinIO deployment
	name := constants.MinIOVolumeName
	if mi.Spec.VolumeClaimTemplate != nil {
		name = mi.Spec.VolumeClaimTemplate.Name
	}

	if mi.Spec.VolumesPerServer == 1 {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      name + strconv.Itoa(0),
			MountPath: constants.MinIOVolumeMountPath,
		})
	} else {
		for i := 0; i < mi.Spec.VolumesPerServer; i++ {
			mounts = append(mounts, corev1.VolumeMount{
				Name:      name + strconv.Itoa(i),
				MountPath: constants.MinIOVolumeMountPath + strconv.Itoa(i),
			})
		}
	}

	if mi.RequiresAutoCertSetup() || mi.RequiresExternalCertSetup() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      mi.GetTLSSecretName(),
			MountPath: "/root/.minio/certs",
		})
	}

	return mounts
}

// Builds the MinIO container for a MinIOInstance.
func minioServerContainer(mi *miniov1beta1.MinIOInstance, serviceName string) corev1.Container {
	scheme := "http"
	if mi.RequiresAutoCertSetup() || mi.RequiresExternalCertSetup() {
		scheme = "https"
	}

	args := []string{"server"}

	if mi.Spec.Zones[0].Servers == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, constants.MinIOVolumeMountPath)
	} else {
		// append all the MinIOInstance replica URLs
		hosts := mi.GetHosts()
		for _, h := range hosts {
			args = append(args, fmt.Sprintf("%s://"+h+"%s", scheme, mi.GetVolumesPath()))
		}
	}

	return corev1.Container{
		Name:  constants.MinIOServerName,
		Image: mi.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: constants.MinIOPort,
			},
		},
		ImagePullPolicy: constants.DefaultImagePullPolicy,
		VolumeMounts:    volumeMounts(mi),
		Args:            args,
		Env:             minioEnvironmentVars(mi),
		Resources:       mi.Spec.Resources,
		LivenessProbe:   mi.Spec.Liveness,
		ReadinessProbe:  mi.Spec.Readiness,
	}
}

// Builds the tolerations for a MinIOInstance.
func minioTolerations(mi *miniov1beta1.MinIOInstance) []corev1.Toleration {
	tolerations := make([]corev1.Toleration, 0)
	// Add all the tolerations
	for _, t := range mi.Spec.Tolerations {
		tolerations = append(tolerations, t)
	}
	// Return tolerations
	return tolerations
}

func getVolumesForContainer(mi *miniov1beta1.MinIOInstance) []corev1.Volume {
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

// NewForCluster creates a new StatefulSet for the given Cluster.
func NewForCluster(mi *miniov1beta1.MinIOInstance, serviceName string) *appsv1.StatefulSet {
	var secretName string

	// If a PV isn't specified just use a EmptyDir volume
	var podVolumes = getVolumesForContainer(mi)
	var replicas = mi.GetReplicas()

	var keyPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "public.crt"},
		{Key: "private.key", Path: "private.key"},
		{Key: "public.crt", Path: "CAs/public.crt"},
	}

	if mi.RequiresAutoCertSetup() {
		secretName = mi.GetTLSSecretName()
	} else if mi.RequiresExternalCertSetup() {
		secretName = mi.Spec.ExternalCertSecret.Name
		if mi.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" {
			keyPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "tls.crt", Path: "CAs/public.crt"},
			}
		} else if mi.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			keyPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "public.crt"},
				{Key: "tls.key", Path: "private.key"},
				{Key: "ca.crt", Path: "CAs/public.crt"},
			}
		}
	}

	// Add SSL volume from SSL secret to the podVolumes
	if mi.RequiresAutoCertSetup() || mi.RequiresExternalCertSetup() {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: mi.GetTLSSecretName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
								Items: keyPaths,
							},
						},
					},
				},
			},
		})
	}

	containers := []corev1.Container{minioServerContainer(mi, serviceName)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mi.Namespace,
			Name:      mi.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1beta1.SchemeGroupVersion.Group,
					Version: miniov1beta1.SchemeGroupVersion.Version,
					Kind:    miniov1beta1.ClusterCRDResourceKind,
				}),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: constants.DefaultUpdateStrategy,
			},
			PodManagementPolicy: mi.Spec.PodManagementPolicy,
			Selector:            mi.Spec.Selector,
			ServiceName:         serviceName,
			Replicas:            &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: minioMetadata(mi),
				Spec: corev1.PodSpec{
					Containers:       containers,
					Volumes:          podVolumes,
					ImagePullSecrets: []corev1.LocalObjectReference{mi.Spec.ImagePullSecret},
					Affinity:         mi.Spec.Affinity,
					SchedulerName:    mi.Scheduler.Name,
					Tolerations:      minioTolerations(mi),
				},
			},
		},
	}

	if mi.Spec.VolumeClaimTemplate != nil {
		pvClaim := *mi.Spec.VolumeClaimTemplate
		name := pvClaim.Name
		for i := 0; i < mi.Spec.VolumesPerServer; i++ {
			pvClaim.Name = name + strconv.Itoa(i)
			ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, pvClaim)
		}
	}
	return ss
}
