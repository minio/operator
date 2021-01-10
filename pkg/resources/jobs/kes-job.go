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
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewForKES creates a new Job to create KES Key
func NewForKES(t *miniov2.Tenant) *batchv1.Job {

	containers := []corev1.Container{kesJobContainer(t)}

	var clientCertSecret string
	var clientCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "minio.crt"},
		{Key: "private.key", Path: "minio.key"},
	}

	if t.ExternalClientCert() {
		clientCertSecret = t.Spec.ExternalClientCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.ExternalClientCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalClientCertSecret.Type == "cert-manager.io/v1alpha2" {
			clientCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "minio.crt"},
				{Key: "tls.key", Path: "minio.key"},
			}
		}
	} else if t.AutoCert() {
		clientCertSecret = t.MinIOClientTLSSecretName()
	}

	podVolumes := []corev1.Volume{
		{
			Name: t.KESVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: clientCertSecret,
								},
								Items: clientCertPaths,
							},
						},
					},
				},
			},
		},
	}

	d := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.KESJobName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: kesMetadata(t),
				Spec: corev1.PodSpec{
					RestartPolicy: miniov2.KESJobRestartPolicy,
					Containers:    containers,
					Volumes:       podVolumes,
				},
			},
		},
	}
	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return d
}

// returns the KES job container
func kesJobContainer(t *miniov2.Tenant) corev1.Container {
	args := []string{"key", "create", "-k", miniov2.KESMinIOKey} // KES CLI expects flags before command args

	return corev1.Container{
		Name:            miniov2.KESContainerName,
		Image:           t.Spec.KES.Image,
		ImagePullPolicy: t.Spec.KES.ImagePullPolicy,
		Args:            args,
		Env:             kesEnvironmentVars(t),
		VolumeMounts:    kesVolumeMounts(t),
	}
}

// Returns the KES environment variables required for the Job.
func kesEnvironmentVars(t *miniov2.Tenant) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "KES_SERVER",
			Value: t.KESServiceEndpoint(),
		},
		{
			Name:  "KES_CLIENT_CERT",
			Value: miniov2.KESConfigMountPath + "/minio.crt",
		},
		{
			Name:  "KES_CLIENT_KEY",
			Value: miniov2.KESConfigMountPath + "/minio.key",
		},
	}
}

// KESMetadata Returns the KES pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func kesMetadata(t *miniov2.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	meta.Labels = t.Spec.KES.Labels
	meta.Annotations = t.Spec.KES.Annotations

	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	// Add the additional label from spec
	for k, v := range t.KESPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// kesVolumeMounts builds the volume mounts for KES container.
func kesVolumeMounts(t *miniov2.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.KESVolMountName(),
			MountPath: miniov2.KESConfigMountPath,
		},
	}
}
