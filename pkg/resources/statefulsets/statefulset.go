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
	"path"

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
	podLabelKey, podLabelValue := minioPodLabels(mi)
	meta.Labels[podLabelKey] = podLabelValue
	return meta
}

func minioPodLabels(mi *miniov1beta1.MinIOInstance) (string, string) {
	return constants.InstanceLabel, mi.Name
}

// Builds the volume mounts for MinIO container.
func volumeMounts(mi *miniov1beta1.MinIOInstance) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	name := constants.MinIOVolumeName
	if mi.Spec.VolumeClaimTemplate != nil {
		name = mi.Spec.VolumeClaimTemplate.Name
	}

	mounts = append(mounts, corev1.VolumeMount{
		Name:      name,
		MountPath: constants.MinIOVolumeMountPath,
	})

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
	minioPath := path.Join(mi.Spec.Mountpath, mi.Spec.Subpath)

	scheme := "http"
	if mi.RequiresAutoCertSetup() || mi.RequiresExternalCertSetup() {
		scheme = "https"
	}

	args := []string{
		"server",
	}

	if mi.Spec.Replicas == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, constants.MinIOVolumeMountPath)
	} else {
		// append all the MinIOInstance replica URLs
		hosts := mi.GetHosts()
		for _, h := range hosts {
			args = append(args, fmt.Sprintf("%s://"+h+"%s", scheme, minioPath))
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
		VolumeMounts:   volumeMounts(mi),
		Args:           args,
		Env:            minioEnvironmentVars(mi),
		Resources:      mi.Spec.Resources,
		LivenessProbe:  mi.Spec.Liveness,
		ReadinessProbe: mi.Spec.Readiness,
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

// NewForCluster creates a new StatefulSet for the given Cluster.
func NewForCluster(mi *miniov1beta1.MinIOInstance, serviceName string) *appsv1.StatefulSet {
	var secretName string

	// If a PV isn't specified just use a EmptyDir volume
	var podVolumes = []corev1.Volume{}
	if mi.Spec.VolumeClaimTemplate == nil {
		podVolumes = append(podVolumes, corev1.Volume{Name: constants.MinIOVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""}}})
	}

	if mi.RequiresAutoCertSetup() {
		secretName = mi.GetTLSSecretName()
	} else if mi.RequiresExternalCertSetup() {
		secretName = mi.Spec.ExternalCertSecret.Name
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
								Items: []corev1.KeyToPath{
									{
										Key:  "public.crt",
										Path: "public.crt",
									},
									{
										Key:  "private.key",
										Path: "private.key",
									},
									{
										Key:  "public.crt",
										Path: "CAs/public.crt",
									},
								},
							},
						},
					},
				},
			},
		})
	}

	containers := []corev1.Container{minioServerContainer(mi, serviceName)}
	podLabelKey, podLabelValue := minioPodLabels(mi)
	podLabels := map[string]string{
		podLabelKey: podLabelValue,
	}

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
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			ServiceName: serviceName,
			Replicas:    &mi.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: minioMetadata(mi),
				Spec: corev1.PodSpec{
					Containers:    containers,
					Volumes:       podVolumes,
					Affinity:      mi.Spec.Affinity,
					SchedulerName: mi.Scheduler.Name,
					Tolerations:   minioTolerations(mi),
				},
			},
		},
	}

	if mi.Spec.VolumeClaimTemplate != nil {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *mi.Spec.VolumeClaimTemplate)
	}
	return ss
}
