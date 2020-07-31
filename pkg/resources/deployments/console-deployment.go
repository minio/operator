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
	"github.com/google/uuid"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Adds required Console environment variables
func consoleEnvVars(t *miniov1.Tenant) []corev1.EnvVar {
	jwtSecret := uuid.New().String()
	pkdfPass := uuid.New().String()
	pkdfSalt := uuid.New().String()
	envVars := []corev1.EnvVar{
		{
			Name:  "CONSOLE_MINIO_SERVER",
			Value: t.MinIOServerEndpoint(),
		},
		{
			Name:  "CONSOLE_HMAC_JWT_SECRET",
			Value: jwtSecret,
		},
		{
			Name:  "CONSOLE_PBKDF_PASSPHRASE",
			Value: pkdfPass,
		},
		{
			Name:  "CONSOLE_PBKDF_SALT",
			Value: pkdfSalt,
		},
		{
			Name:  "CONSOLE_ACCESS_KEY",
			Value: miniov1.DefaultConsoleAccessKey,
		},
		{
			Name:  "CONSOLE_SECRET_KEY",
			Value: miniov1.DefaultConsoleSecretKey,
		},
	}
	if t.TLS() {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "CONSOLE_MINIO_SERVER_TLS_ROOT_CAS",
			Value: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt," + miniov1.ConsoleConfigMountPath + "/public.crt",
		})
	}
	return envVars
}

func consoleMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
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
		Image: miniov1.DefaultConsoleImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.ConsolePort,
			},
		},
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             consoleEnvVars(t),
		VolumeMounts:    consoleVolumeMounts(t),
	}
}

// ConsoleVolumeMounts builds the volume mounts for Console container.
func consoleVolumeMounts(t *miniov1.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.ConsoleVolMountName(),
			MountPath: miniov1.ConsoleConfigMountPath,
		},
	}
}

// NewConsole creates a new Deployment for the given MinIO instance.
func NewConsole(t *miniov1.Tenant) *appsv1.Deployment {
	r := int32(miniov1.DefaultConsoleReplicas)
	var minioCertPaths = []corev1.KeyToPath{
		{Key: "public.crt", Path: "public.crt"},
	}

	podVolumes := []corev1.Volume{
		{
			Name: t.ConsoleVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: t.MinIOCACertSecretName(),
								},
								Items: minioCertPaths,
							},
						},
					},
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
			Replicas: &r,
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

	return d
}
