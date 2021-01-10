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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Adds required log-search-api environment variables
func logSearchAPIEnvVars(t *miniov2.Tenant) []corev1.EnvVar {
	var diskCapacityGB int
	if t.Spec.Log.Audit.DiskCapacityGB != nil {
		diskCapacityGB = *t.Spec.Log.Audit.DiskCapacityGB
	}
	return []corev1.EnvVar{
		{
			Name:  miniov2.LogSearchDiskCapacityGB,
			Value: fmt.Sprintf("%d", diskCapacityGB),
		},
		{
			Name: miniov2.LogPgConnStr,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.LogSecretName(),
					},
					Key: miniov2.LogPgConnStr,
				},
			},
		},
		{
			Name: miniov2.LogAuditTokenKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.LogSecretName(),
					},
					Key: miniov2.LogAuditTokenKey,
				},
			},
		},
		{
			Name: miniov2.LogQueryTokenKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.LogSecretName(),
					},
					Key: miniov2.LogQueryTokenKey,
				},
			},
		},
	}
}

func logSearchAPIContainer(t *miniov2.Tenant) corev1.Container {
	logSearchAPIImage := miniov2.DefaultLogSearchAPIImage
	if t.Spec.Log.Image != "" {
		logSearchAPIImage = t.Spec.Log.Image
	}
	container := corev1.Container{
		Name:  miniov2.LogSearchAPIContainerName,
		Image: logSearchAPIImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.LogSearchAPIPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		Env:             logSearchAPIEnvVars(t),
		Resources:       t.Spec.Log.Resources,
	}

	return container
}

func logSearchAPIMeta(t *miniov2.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	meta.Labels = make(map[string]string)
	for k, v := range t.LogSearchAPIPodLabels() {
		meta.Labels[k] = v
	}

	// attach any labels
	for k, v := range t.Spec.Log.Labels {
		meta.Labels[k] = v
	}
	// attach any annotations
	if len(t.Spec.Log.Annotations) > 0 {
		meta.Annotations = make(map[string]string)
		for k, v := range t.Spec.Log.Annotations {
			meta.Annotations[k] = v
		}
	}

	return meta
}

// logSearchAPISelector Returns the Log search API Pod selector
func logSearchAPISelector(t *miniov2.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.LogSearchAPIPodLabels(),
	}
}

// NewForLogSearchAPI returns k8s deployment object for Log Search API server
func NewForLogSearchAPI(t *miniov2.Tenant) *appsv1.Deployment {
	var replicas int32 = 1

	apiPod := corev1.PodTemplateSpec{
		ObjectMeta: logSearchAPIMeta(t),
		Spec: corev1.PodSpec{
			ServiceAccountName: t.Spec.ServiceAccountName,
			Containers:         []corev1.Container{logSearchAPIContainer(t)},
			RestartPolicy:      corev1.RestartPolicyAlways,
		},
	}

	if t.Spec.Log.Db != nil {
		// attach affinity clauses
		if t.Spec.Log.Db.Affinity != nil {
			apiPod.Spec.Affinity = t.Spec.Log.Db.Affinity
		}
		// attach node selector clauses
		apiPod.Spec.NodeSelector = t.Spec.Log.Db.NodeSelector
		// attach tolerations
		apiPod.Spec.Tolerations = t.Spec.Log.Db.Tolerations
	}

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.LogSearchAPIDeploymentName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: logSearchAPISelector(t),
			Template: apiPod,
		},
	}

	if t.Spec.ImagePullSecret.Name != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return d
}
