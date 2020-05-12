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

package jobs

import (
	"net"
	"strconv"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewForKES creates a new Job to create KES Key
func NewForKES(mi *miniov1.MinIOInstance) *batchv1.Job {

	containers := []corev1.Container{kesJobContainer(mi)}

	var MinIOCertKeyPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "minio.crt"},
		{Key: "private.key", Path: "minio.key"},
	}

	podVolumes := []corev1.Volume{
		{
			Name: mi.KESVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: mi.MinIOClientTLSSecretName(),
								},
								Items: MinIOCertKeyPaths,
							},
						},
					},
				},
			},
		},
	}

	d := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       mi.Namespace,
			Name:            mi.KESJobName(),
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: kesMetadata(mi),
				Spec: corev1.PodSpec{
					RestartPolicy:    miniov1.KESJobRestartPolicy,
					Containers:       containers,
					ImagePullSecrets: []corev1.LocalObjectReference{mi.Spec.ImagePullSecret},
					Volumes:          podVolumes,
				},
			},
		},
	}

	return d
}

// returns the KES job container
func kesJobContainer(mi *miniov1.MinIOInstance) corev1.Container {
	args := []string{"key", "create", miniov1.KESMinIOKey, "-k"}

	return corev1.Container{
		Name:            miniov1.KESContainerName,
		Image:           mi.Spec.KES.Image,
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             kesEnvironmentVars(mi),
		VolumeMounts:    kesVolumeMounts(mi),
	}
}

// Returns the KES environment variables required for the Job.
func kesEnvironmentVars(mi *miniov1.MinIOInstance) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "KES_SERVER",
			Value: "https://" + net.JoinHostPort(mi.KESServiceHost(), strconv.Itoa(miniov1.KESPort)),
		},
		{
			Name:  "KES_CLIENT_CERT",
			Value: miniov1.KESConfigMountPath + "/minio.crt",
		},
		{
			Name:  "KES_CLIENT_KEY",
			Value: miniov1.KESConfigMountPath + "/minio.key",
		},
	}
}

// KESMetadata Returns the KES pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func kesMetadata(mi *miniov1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}

	if mi.HasKESMetadata() {
		meta = *mi.Spec.KES.Metadata
	}

	// Add the additional label from spec
	for k, v := range mi.KESPodLabels() {
		meta.Labels[k] = v
	}

	// Add the Selector labels set by user
	if mi.HasKESSelector() {
		for k, v := range mi.Spec.KES.Selector.MatchLabels {
			meta.Labels[k] = v
		}
	}
	return meta
}

// kesVolumeMounts builds the volume mounts for KES container.
func kesVolumeMounts(mi *miniov1.MinIOInstance) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      mi.KESVolMountName(),
			MountPath: miniov1.KESConfigMountPath,
		},
	}
}
