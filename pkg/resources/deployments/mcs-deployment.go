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

package deployments

import (
	"net"
	"strconv"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Returns the MCS environment variables set in configuration.
func mcsEnvironmentVars(mi *miniov1.MinIOInstance) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0)
	if mi.HasMCSSecret() {
		var secretName string
		secretName = mi.Spec.MCS.MCSSecret.Name
		envVars = append(envVars, corev1.EnvVar{
			Name: "MCS_HMAC_JWT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "mcshmacjwt",
				},
			},
		}, corev1.EnvVar{
			Name: "MCS_PBKDF_PASSPHRASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "mcspbkdfpassphrase",
				},
			},
		}, corev1.EnvVar{
			Name: "MCS_PBKDF_SALT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "mcspbkdfsalt",
				},
			},
		}, corev1.EnvVar{
			Name:  "MCS_MINIO_SERVER",
			Value: miniov1.Scheme + "://" + net.JoinHostPort(mi.MinIOCIServiceHost(), strconv.Itoa(miniov1.MinIOPort)),
		}, corev1.EnvVar{
			Name: "MCS_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: "mcssecretkey",
				},
			},
		}, corev1.EnvVar{
			Name:  "MCS_ACCESS_KEY",
			Value: mi.Spec.MCS.MCSAccessKey,
		})
	}
	return envVars
}

func mcsMetadata(mi *miniov1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	// Initialize empty fields
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	if mi.HasMCSMetadata() {
		meta = *mi.Spec.MCS.Metadata
	}
	for k, v := range mi.MCSPodLabels() {
		meta.Labels[k] = v
	}
	// Add the Selector labels set by user
	if mi.HasMCSSelector() {
		for k, v := range mi.Spec.MCS.Selector.MatchLabels {
			meta.Labels[k] = v
		}
	}
	return meta
}

// Builds the MCS container for a MinIOInstance.
func mcsContainer(mi *miniov1.MinIOInstance) corev1.Container {
	args := []string{"server"}

	return corev1.Container{
		Name:  miniov1.MCSContainerName,
		Image: mi.Spec.MCS.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.MCSPort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             mcsEnvironmentVars(mi),
		Resources:       mi.Spec.Resources,
	}
}

// NewForMCS creates a new Deployment for the given MinIO instance.
func NewForMCS(mi *miniov1.MinIOInstance) *appsv1.Deployment {

	var replicas int32 = 1

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       mi.Namespace,
			Name:            mi.MCSDeploymentName(),
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: mi.Spec.MCS.Selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: mcsMetadata(mi),
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{mcsContainer(mi)},
					RestartPolicy: miniov1.MCSRestartPolicy,
				},
			},
		},
	}

	return d
}
