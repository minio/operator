/*
 * MinIO-Operator - Manage MinIO clusters in Kubernetes
 *
 * MinIO Cloud Storage, (C) 2018, 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package statefulsets

import (
	"fmt"
	"path"
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

	if mi.RequiresSSLSetup() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      mi.Name + "TLS",
			MountPath: "/root/.minio/certs",
		})
	}

	return mounts
}

// Builds the MinIO container for a MinIOInstance.
func minioServerContainer(mi *miniov1beta1.MinIOInstance, serviceName, imagePath string) corev1.Container {
	replicas := int(mi.Spec.Replicas)
	minioPath := path.Join(mi.Spec.Mountpath, mi.Spec.Subpath)

	scheme := "http"
	if mi.RequiresSSLSetup() {
		scheme = "https"
	}

	args := []string{
		"server",
	}

	if replicas == 1 {
		// to run in standalone mode we must pass the path
		args = append(args, constants.MinIOVolumeMountPath)
	} else {
		// append all the MinIOInstance replica URLs
		for i := 0; i < replicas; i++ {
			args = append(args, fmt.Sprintf("%s://%s-"+strconv.Itoa(i)+".%s.%s.svc.cluster.local%s", scheme, mi.Name, serviceName, mi.Namespace, minioPath))
		}
	}

	return corev1.Container{
		Name:  constants.MinIOServerName,
		Image: fmt.Sprintf("%s:%s", imagePath, mi.Spec.Version),
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

// NewForCluster creates a new StatefulSet for the given Cluster.
func NewForCluster(mi *miniov1beta1.MinIOInstance, serviceName, imagePath string) *appsv1.StatefulSet {
	// If a PV isn't specified just use a EmptyDir volume
	var podVolumes = []corev1.Volume{}
	if mi.Spec.VolumeClaimTemplate == nil {
		podVolumes = append(podVolumes, corev1.Volume{Name: constants.MinIOVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""}}})
	}

	// Add SSL volume from SSL secret to the podVolumes
	if mi.RequiresSSLSetup() {
		podVolumes = append(podVolumes, corev1.Volume{
			Name: mi.Name + "TLS",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: mi.Spec.SSLSecret.Name,
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

	containers := []corev1.Container{minioServerContainer(mi, serviceName, imagePath)}
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
				},
			},
		},
	}

	if mi.Spec.VolumeClaimTemplate != nil {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *mi.Spec.VolumeClaimTemplate)
	}
	return ss
}
