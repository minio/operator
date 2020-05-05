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
	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniooperator.min.io/v1beta1"
	"github.com/minio/minio-operator/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Returns the Mcs environment variables set in configuration.
// If a user specifies a set of secrets for the MCS Environment
func mcsEnvironmentVars(mi *miniov1beta1.MinIOInstance) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0)
	if mi.HasMcsSecret() {
		var secretName string
		secretName = mi.Spec.Mcs.McsSecret.Name
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
			Value: "http://" + mi.GetServiceHost(),
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
			Value: mi.Spec.Mcs.McsAccessKey,
		})
	}
	// Return environment variables
	return envVars
}

func mcsMetadata(mi *miniov1beta1.MinIOInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if mi.HasMcsMetadata() {
		meta = *mi.Spec.Mcs.Metadata
	}
	// Initialize empty fields
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	// Add the Selector labels set by user
	if mi.HasMcsSelector() {
		for k, v := range mi.Spec.Mcs.Selector.MatchLabels {
			meta.Labels[k] = v
		}
	}
	return meta
}

// Builds the Mcs container for a MinIOInstance.
func mcsContainer(mi *miniov1beta1.MinIOInstance) corev1.Container {
	args := []string{"server"}

	return corev1.Container{
		Name:  constants.McsName,
		Image: mi.Spec.Mcs.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: constants.McsPort,
			},
		},
		ImagePullPolicy: constants.DefaultImagePullPolicy,
		Args:            args,
		Env:             mcsEnvironmentVars(mi),
		Resources:       mi.Spec.Resources,
	}
}

// NewForCluster creates a new Deployment for the given MinIO instance.
func NewForCluster(mi *miniov1beta1.MinIOInstance) *appsv1.Deployment {

	var replicas int32 = 1

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mi.Namespace,
			Name:      mi.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1beta1.SchemeGroupVersion.Group,
					Version: miniov1beta1.SchemeGroupVersion.Version,
					Kind:    constants.MinIOCRDResourceKind,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: mi.Spec.Mcs.Selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: mcsMetadata(mi),
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{mcsContainer(mi)},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	return d
}
