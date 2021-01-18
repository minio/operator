/*
 * This file is part of MinIO Operator
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

package resources

import (
	"github.com/minio/kubectl-minio/cmd/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperatorOptions encapsulates the CLI options for a MinIO Operator
type OperatorOptions struct {
	Name            string
	Image           string
	Namespace       string
	NSToWatch       string
	ClusterDomain   string
	ImagePullSecret string
}

func operatorLabels() map[string]string {
	m := make(map[string]string)
	m["name"] = "minio-operator"
	return m
}

// Adds required Operator environment variables
func envVars(clusterDomain, nsToWatch string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CLUSTER_DOMAIN",
			Value: clusterDomain,
		},
		{
			Name:  "WATCHED_NAMESPACE",
			Value: nsToWatch,
		},
	}
}

// Builds the MinIO Operator container.
func container(img, clusterDomain, nsToWatch string) corev1.Container {
	return corev1.Container{
		Name:            helpers.ContainerName,
		Image:           img,
		ImagePullPolicy: helpers.DefaultImagePullPolicy,
		Env:             envVars(clusterDomain, nsToWatch),
	}
}

// NewDeploymentForOperator will return a new deployment for a MinIO Operator
func NewDeploymentForOperator(opts OperatorOptions) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helpers.DeploymentName,
			Namespace: opts.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &helpers.DeploymentReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: operatorLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: operatorLabels(),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: helpers.DefaultServiceAccount,
					Containers:         []corev1.Container{container(opts.Image, opts.ClusterDomain, opts.NSToWatch)},
					ImagePullSecrets:   []corev1.LocalObjectReference{{Name: opts.ImagePullSecret}},
				},
			},
		},
	}
}
