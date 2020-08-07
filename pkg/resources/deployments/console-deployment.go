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
	"fmt"

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
			Name:  "CONSOLE_MINIO_SERVER_TLS_ROOT_CAS",
			Value: fmt.Sprintf("%s/CAs/minio.crt", miniov1.ConsoleConfigMountPath),
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

// ConsoleVolumeMounts builds the volume mounts for Console container.
func ConsoleVolumeMounts(t *miniov1.Tenant) (mounts []corev1.VolumeMount) {
	if t.TLS() || t.ConsoleExternalCert() {
		mounts = []corev1.VolumeMount{
			{
				Name:      t.ConsoleVolMountName(),
				MountPath: miniov1.ConsoleConfigMountPath,
			},
		}
	}
	return mounts
}

// Builds the Console container for a Tenant.
func consoleContainer(t *miniov1.Tenant) corev1.Container {
	args := []string{"server"}

	if t.AutoCert() || t.ConsoleExternalCert() {
		args = append(args, "--tls-certificate=/tmp/console/server.crt", "--tls-key=/tmp/console/server.key")
	}

	return corev1.Container{
		Name:  miniov1.ConsoleContainerName,
		Image: t.Spec.Console.Image,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: miniov1.ConsolePort,
			},
			{
				Name:          "https",
				ContainerPort: miniov1.ConsoleTLSPort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             consoleEnvVars(t),
		EnvFrom:         consoleSecretEnvVars(t),
		Resources:       t.Spec.Console.Resources,
		VolumeMounts:    ConsoleVolumeMounts(t),
	}
}

// NewConsole creates a new Deployment for the given MinIO instance.
func NewConsole(t *miniov1.Tenant) *appsv1.Deployment {
	var certPath = "server.crt"
	var keyPath = "server.key"
	var serverCertSecret string
	var tenantCertSecret string
	var podVolumeSources []corev1.VolumeProjection
	var serverCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: certPath},
		{Key: "private.key", Path: keyPath},
	}
	var tenantCertPath = "CAs/minio.crt"
	var tenantCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: tenantCertPath},
	}

	if t.AutoCert() {
		serverCertSecret = t.ConsoleTLSSecretName()
		tenantCertSecret = t.MinIOTLSSecretName()
	} else if t.ConsoleExternalCert() {
		serverCertSecret = t.Spec.Console.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.Console.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.Console.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: certPath},
				{Key: "tls.key", Path: keyPath},
			}
		}
	}

	// Add MinIO certificate to the CAs pool of Console
	if t.ExternalCert() {
		tenantCertSecret = t.Spec.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			tenantCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: tenantCertPath},
			}
		}
	}

	if serverCertSecret != "" {
		podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: serverCertSecret,
				},
				Items: serverCertPaths,
			},
		})
	}

	if tenantCertSecret != "" {
		podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: tenantCertSecret,
				},
				Items: tenantCertPaths,
			},
		})
	}

	podVolumes := []corev1.Volume{
		{
			Name: t.ConsoleVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: podVolumeSources,
				},
			},
		},
	}

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
					Volumes:       podVolumes,
				},
			},
		},
	}

	if t.Spec.ImagePullSecret.Name != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return d
}
