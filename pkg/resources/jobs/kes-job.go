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

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewForKES creates a new Job to create KES Key
func NewForKES(t *miniov1.Tenant) *batchv1.Job {

	containers := []corev1.Container{kesJobContainer(t)}

	var clientCertSecret string
	var clientCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "minio.crt"},
		{Key: "private.key", Path: "minio.key"},
	}

	if t.AutoCert() {
		clientCertSecret = t.MinIOClientTLSSecretName()
	} else if t.ExternalClientCert() {
		clientCertSecret = t.Spec.ExternalClientCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			clientCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: "minio.crt"},
				{Key: "tls.key", Path: "minio.key"},
			}
		}
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
					RestartPolicy: miniov1.KESJobRestartPolicy,
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
func kesJobContainer(t *miniov1.Tenant) corev1.Container {
	args := []string{"key", "create", miniov1.KESMinIOKey, "-k"}

	return corev1.Container{
		Name:            miniov1.KESContainerName,
		Image:           t.Spec.KES.Image,
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             kesEnvironmentVars(t),
		VolumeMounts:    kesVolumeMounts(t),
	}
}

// Returns the KES environment variables required for the Job.
func kesEnvironmentVars(t *miniov1.Tenant) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "KES_SERVER",
			Value: "https://" + net.JoinHostPort(t.KESServiceHost(), strconv.Itoa(miniov1.KESPort)),
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
func kesMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if t.HasKESMetadata() {
		meta = *t.Spec.KES.Metadata
	}
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
func kesVolumeMounts(t *miniov1.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.KESVolMountName(),
			MountPath: miniov1.KESConfigMountPath,
		},
	}
}
