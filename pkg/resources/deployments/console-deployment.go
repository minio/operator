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
	"strings"

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
		var caCerts []string
		if t.ExternalCert() {
			for index := range t.Spec.ExternalCertSecret {
				caCerts = append(caCerts, fmt.Sprintf("%s/CAs/minio-hostname-%d.crt", miniov1.ConsoleConfigMountPath, index))
			}
		} else {
			caCerts = append(caCerts, fmt.Sprintf("%s/CAs/minio.crt", miniov1.ConsoleConfigMountPath))
		}
		envVars = append(envVars, corev1.EnvVar{
			Name:  "CONSOLE_MINIO_SERVER_TLS_ROOT_CAS",
			Value: strings.Join(caCerts, ","),
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
	meta.Labels = t.Spec.Console.Labels
	meta.Annotations = t.Spec.Console.Annotations

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
		ImagePullPolicy: t.Spec.Console.ImagePullPolicy,
		Args:            args,
		Env:             consoleEnvVars(t),
		EnvFrom:         consoleSecretEnvVars(t),
		Resources:       t.Spec.Console.Resources,
		VolumeMounts:    ConsoleVolumeMounts(t),
	}
}

// NewConsole creates a new Deployment for the given MinIO Tenant.
func NewConsole(t *miniov1.Tenant) *appsv1.Deployment {
	var certPath = "server.crt"
	var keyPath = "server.key"

	var podVolumeSources []corev1.VolumeProjection

	var serverCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: certPath},
		{Key: "private.key", Path: keyPath},
	}

	var tenantCertPath = "CAs/minio.crt"
	var tenantCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: tenantCertPath},
	}
	// External certificates will have priority over AutoCert generated certificates
	// In the future this may change when Console supports SNI
	if t.ConsoleExternalCert() {
		serverCertSecret := t.Spec.Console.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.Console.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.Console.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: certPath},
				{Key: "tls.key", Path: keyPath},
			}
		}
		podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: serverCertSecret,
				},
				Items: serverCertPaths,
			},
		})
	} else if t.AutoCert() {
		// Console certificates generated by AutoCert
		podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.ConsoleTLSSecretName(),
				},
				Items: serverCertPaths,
			},
		})
	}

	// If MinIO has AutoCert enabled load the autogenerated certificate into certs/CAS/public.crt
	if t.AutoCert() {
		// MinIO tenant certificate generated by AutoCert
		podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.MinIOTLSSecretName(),
				},
				Items: tenantCertPaths,
			},
		})
	}

	// If user provided additional certificates load those too into certs/CAs/minio-hostname-{n}.crt
	if t.ExternalCert() {
		// Iterate over all provided TLS certificates and store them on the list of
		// Volumes that will be mounted to the Pod using the following folder structure:
		//
		//	certs
		//		+ CAs
		//			 + minio-hostname-0.crt
		//			 + minio-hostname-1.crt
		//			 + minio-hostname-2.crt
		//
		for index, secret := range t.Spec.ExternalCertSecret {
			// This covers both secrets of type "kubernetes.io/tls" and
			// "cert-manager.io/v1alpha2" because of same keys in both.
			if secret.Type == "kubernetes.io/tls" || secret.Type == "cert-manager.io/v1alpha2" {
				tenantCertPaths = []corev1.KeyToPath{
					{Key: "tls.crt", Path: fmt.Sprintf("CAs/minio-hostname-%d.crt", index)},
				}
			} else {
				tenantCertPaths = []corev1.KeyToPath{
					{Key: "public.crt", Path: fmt.Sprintf("CAs/minio-hostname-%d.crt", index)},
				}
			}
			podVolumeSources = append(podVolumeSources, corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret.Name,
					},
					Items: tenantCertPaths,
				},
			})
		}
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
					ServiceAccountName: t.Spec.Console.ServiceAccountName,
					Containers:         []corev1.Container{consoleContainer(t)},
					RestartPolicy:      miniov1.ConsoleRestartPolicy,
					Volumes:            podVolumes,
					NodeSelector:       t.Spec.Console.NodeSelector,
				},
			},
		},
	}

	if t.Spec.ImagePullSecret.Name != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return d
}
