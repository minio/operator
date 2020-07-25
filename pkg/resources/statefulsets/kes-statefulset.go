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
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KESMetadata Returns the KES pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func KESMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if t.HasKESMetadata() {
		meta = *t.Spec.KES.Metadata
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range t.KESPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// KESSelector Returns the KES pods selector set in configuration.
func KESSelector(t *miniov1.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.KESPodLabels(),
	}
}

// KESVolumeMounts builds the volume mounts for MinIO container.
func KESVolumeMounts(t *miniov1.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.KESVolMountName(),
			MountPath: miniov1.KESConfigMountPath,
		},
	}
}

// KESEnvironmentVars returns the KES environment variables set in configuration.
func KESEnvironmentVars(t *miniov1.Tenant) []corev1.EnvVar {
	// pass the identity created while generating the MinIO client cert
	return []corev1.EnvVar{
		{
			Name:  "MINIO_ID",
			Value: miniov1.Identity,
		},
	}
}

// KESServerContainer returns the KES container for a KES StatefulSet.
func KESServerContainer(t *miniov1.Tenant) corev1.Container {

	// Args to start KES with config mounted at miniov1.KESConfigMountPath and require but don't verify mTLS authentication
	args := []string{"server", "--config=" + miniov1.KESConfigMountPath + "/server-config.yaml", "--auth=off"}

	return corev1.Container{
		Name:  miniov1.KESContainerName,
		Image: t.Spec.KES.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.KESPort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		VolumeMounts:    KESVolumeMounts(t),
		Args:            args,
		Env:             KESEnvironmentVars(t),
	}
}

// NewForKES creates a new KES StatefulSet for the given Cluster.
func NewForKES(t *miniov1.Tenant, serviceName string) *appsv1.StatefulSet {
	var replicas = t.KESReplicas()
	var certPath = "server.crt"
	var keyPath = "server.key"
	var serverCertSecret string
	var serverCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: certPath},
		{Key: "private.key", Path: keyPath},
	}

	var configPath = []corev1.KeyToPath{
		{Key: "server-config.yaml", Path: "server-config.yaml"},
	}

	if t.AutoCert() {
		serverCertSecret = t.KESTLSSecretName()
	} else if t.KESExternalCert() {
		serverCertSecret = t.Spec.KES.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: certPath},
				{Key: "tls.key", Path: keyPath},
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
									Name: serverCertSecret,
								},
								Items: serverCertPaths,
							},
						},
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: t.Spec.KES.Configuration.Name,
								},
								Items: configPath,
							},
						},
					},
				},
			},
		},
	}

	containers := []corev1.Container{KESServerContainer(t)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.KESStatefulSetName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov1.DefaultUpdateStrategy,
			},
			PodManagementPolicy: t.Spec.PodManagementPolicy,
			// KES is always matched via Tenant Name + KES prefix
			Selector:    KESSelector(t),
			ServiceName: serviceName,
			Replicas:    &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: KESMetadata(t),
				Spec: corev1.PodSpec{
					Containers:       containers,
					Volumes:          podVolumes,
					ImagePullSecrets: []corev1.LocalObjectReference{t.Spec.ImagePullSecret},
					RestartPolicy:    corev1.RestartPolicyAlways,
					SchedulerName:    t.Scheduler.Name,
				},
			},
		},
	}

	return ss
}
