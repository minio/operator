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
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Adds required Console environment variables
func consoleEnvVars(t *miniov1.Tenant) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "CONSOLE_MINIO_SERVER",
			Value: t.MinIOServerEndpoint(),
		},
	}
	if t.TLS() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "CONSOLE_MINIO_SERVER_TLS_SKIP_VERIFICATION",
			Value: "on", // FIXME: should be trusted
		})
	}
	// Add all the environment variables
	envVars = append(envVars, t.Spec.Console.Env...)
	return envVars
}

// Returns the Console environment variables set in configuration.
func consoleSecretEnvVars(t *miniov1.Tenant) []corev1.EnvFromSource {
	envVars := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.Spec.Console.ConsoleSecret.Name,
				},
			},
		},
	}
	return envVars
}

func consoleMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if t.HasConsoleMetadata() {
		meta = *t.Spec.Console.Metadata
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range t.ConsolePodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// consoleSelector Returns the Console pods selector
func consoleSelector(t *miniov1.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.ConsolePodLabels(),
	}
}

// Builds the Console container for a Tenant.
func consoleContainer(t *miniov1.Tenant) corev1.Container {
	args := []string{"server"}

	return corev1.Container{
		Name:  miniov1.ConsoleContainerName,
		Image: t.Spec.Console.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.ConsolePort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             consoleEnvVars(t),
		EnvFrom:         consoleSecretEnvVars(t),
		Resources:       t.Spec.Console.Resources,
	}
}

// NewConsole creates a new Deployment for the given MinIO instance.
func NewConsole(t *miniov1.Tenant) *appsv1.Deployment {

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.ConsoleDeploymentName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &t.Spec.Console.Replicas,
			// Console is always matched via Tenant Name + console prefix
			Selector: consoleSelector(t),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: consoleMetadata(t),
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{consoleContainer(t)},
					RestartPolicy: miniov1.ConsoleRestartPolicy,
				},
			},
		},
	}

	return d
}
