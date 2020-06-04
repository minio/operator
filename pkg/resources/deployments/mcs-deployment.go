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

// Adds required MCS environment variables
func mcsEnvVars(mi *miniov1.MinIOInstance) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "MCS_MINIO_SERVER",
			Value: miniov1.Scheme + "://" + net.JoinHostPort(mi.MinIOCIServiceHost(), strconv.Itoa(miniov1.MinIOPort)),
		},
	}
	if miniov1.Scheme == "https" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MCS_MINIO_SERVER_TLS_SKIP_VERIFICATION",
			Value: "on",
		})
	}
	return envVars
}

// Returns the MCS environment variables set in configuration.
func mcsSecretEnvVars(mi *miniov1.MinIOInstance) []corev1.EnvFromSource {
	envVars := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: mi.Spec.MCS.MCSSecret.Name,
				},
			},
		},
	}
	return envVars
}

func mcsMetadata(mi *miniov1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if mi.HasMCSMetadata() {
		meta = *mi.Spec.MCS.Metadata
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range mi.MCSPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// mcsSelector Returns the MCS pods selector
func mcsSelector(mi *miniov1.MinIOInstance) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: mi.MCSPodLabels(),
	}
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
		Env:             mcsEnvVars(mi),
		EnvFrom:         mcsSecretEnvVars(mi),
		Resources:       mi.Spec.Resources,
	}
}

// NewForMCS creates a new Deployment for the given MinIO instance.
func NewForMCS(mi *miniov1.MinIOInstance) *appsv1.Deployment {

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       mi.Namespace,
			Name:            mi.MCSDeploymentName(),
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mi.Spec.MCS.Replicas,
			// MCS is always matched via MinIOInstance Name + mcs prefix
			Selector: mcsSelector(mi),
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
