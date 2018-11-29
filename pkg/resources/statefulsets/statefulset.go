/*
 * Minio-Operator - Manage Minio clusters in Kubernetes
 *
 * Minio (C) 2018 Minio, Inc.
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

// Returns the Minio credential environment variables
// If a user specifies a secret in the spec we use that
// else we create a secret with a default password
func minioCredentials(mi *miniov1beta1.MinioInstance) []corev1.EnvVar {
	var secretName string
	if mi.HasCredsSecret() {
		secretName = mi.Spec.CredsSecret.Name
		return []corev1.EnvVar{
			{
				Name: "MINIO_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: "accesskey",
					},
				},
			},
			{
				Name: "MINIO_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: "secretkey",
					},
				},
			},
		}
	}
	// If no secret provided, dont use env vars. Minio server automatically creates default
	// credentials
	return []corev1.EnvVar{}
}

// Builds the volume mounts for Minio container.
func volumeMounts(mi *miniov1beta1.MinioInstance) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	name := constants.MinioVolumeName
	if mi.Spec.VolumeClaimTemplate != nil {
		name = mi.Spec.VolumeClaimTemplate.Name
	}

	mounts = append(mounts, corev1.VolumeMount{
		Name:      name,
		MountPath: constants.MinioVolumeMountPath,
	})

	if mi.RequiresSSLSetup() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      mi.Name + "TLS",
			MountPath: "/root/.minio/certs",
		})
	}

	return mounts
}

// Builds the Minio container for a MinioInstance.
func minioServerContainer(mi *miniov1beta1.MinioInstance, serviceName string) corev1.Container {
	replicas := int(mi.Spec.Replicas)
	minioPath := path.Join(mi.Spec.Mountpath, mi.Spec.Subpath)

	scheme := "http"
	if mi.RequiresSSLSetup() {
		scheme = "https"
	}

	args := []string{
		"server",
	}

	// append all the MinioInstance replica URLs
	for i := 0; i < replicas; i++ {
		args = append(args, fmt.Sprintf("%s://%s-"+strconv.Itoa(i)+".%s.%s.svc.cluster.local%s", scheme, mi.Name, serviceName, mi.Namespace, minioPath))
	}

	return corev1.Container{
		Name:  constants.MinioServerName,
		Image: fmt.Sprintf("%s:%s", constants.MinioImagePath, mi.Spec.Version),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: constants.MinioPort,
			},
		},
		VolumeMounts: volumeMounts(mi),
		Args:         args,
		Env:          minioCredentials(mi),
	}
}

// NewForCluster creates a new StatefulSet for the given Cluster.
func NewForCluster(mi *miniov1beta1.MinioInstance, serviceName string) *appsv1.StatefulSet {
	// If a PV isn't specified just use a EmptyDir volume
	var podVolumes = []corev1.Volume{}
	if mi.Spec.VolumeClaimTemplate == nil {
		podVolumes = append(podVolumes, corev1.Volume{Name: constants.MinioVolumeName,
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

	containers := []corev1.Container{minioServerContainer(mi, serviceName)}

	podLabels := map[string]string{
		constants.InstanceLabel: mi.Name,
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
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: containers,
					Volumes:    podVolumes,
				},
			},
		},
	}

	if mi.Spec.VolumeClaimTemplate != nil {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *mi.Spec.VolumeClaimTemplate)
	}
	return ss
}
