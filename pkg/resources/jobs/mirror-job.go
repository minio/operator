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

package jobs

import (
	"strings"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewForMirror creates a new Job for the given mirror instance.
func NewForMirror(mi *miniov1.MirrorInstance) *batchv1.Job {

	containers := []corev1.Container{mirrorJobContainer(mi)}

	d := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mi.Namespace,
			Name:      mi.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1.SchemeGroupVersion.Group,
					Version: miniov1.SchemeGroupVersion.Version,
					Kind:    miniov1.MirrorCRDResourceKind,
				}),
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: mirrorMetadata(mi),
				Spec: corev1.PodSpec{
					RestartPolicy:    miniov1.MirrorJobRestartPolicy,
					Containers:       containers,
					ImagePullSecrets: []corev1.LocalObjectReference{mi.Spec.ImagePullSecret},
				},
			},
		},
	}

	return d
}

// Builds the mc container for a MirrorInstance.
func mirrorJobContainer(mi *miniov1.MirrorInstance) corev1.Container {
	args := []string{"mirror"}

	// Add default flags
	args = append(args, miniov1.DefaultMirrorFlags...)

	for _, f := range mi.Spec.Args.Flags {
		arg := strings.Split(f, " ")
		args = append(args, arg...)
	}

	if mi.Spec.Args.Source != "" {
		args = append(args, mi.Spec.Args.Source)
	}

	if mi.Spec.Args.Target != "" {
		args = append(args, mi.Spec.Args.Target)
	}

	return corev1.Container{
		Name:            miniov1.MirrorContainerName,
		Image:           mi.Spec.Image,
		ImagePullPolicy: miniov1.DefaultImagePullPolicy,
		Args:            args,
		Env:             mirrorEnvironmentVars(mi),
	}
}

// Returns the mc mirror environment variables set in configuration.
func mirrorEnvironmentVars(mi *miniov1.MirrorInstance) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0)
	// Add all the environment variables
	for _, e := range mi.Spec.Env {
		envVars = append(envVars, e)
	}
	// Return environment variables
	return envVars
}

// Returns the MC pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func mirrorMetadata(mi *miniov1.MirrorInstance) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if mi.HasMetadata() {
		meta = *mi.Spec.Metadata
	}
	// Initialize empty fields
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	// Add the Selector labels set by user
	if mi.HasSelector() {
		for k, v := range mi.Spec.Selector.MatchLabels {
			meta.Labels[k] = v
		}
	}
	return meta
}
