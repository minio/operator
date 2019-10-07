/*
 * Copyright (C) 2019, MinIO, Inc.
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

package services

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	constants "github.com/minio/minio-operator/pkg/constants"
)

// NewForCluster will return a new headless Kubernetes service for a MinIOInstance
func NewForCluster(mi *miniov1beta1.MinIOInstance) *corev1.Service {
	minioPort := corev1.ServicePort{Port: constants.MinIOPort}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    map[string]string{constants.InstanceLabel: mi.Name},
			Name:      mi.GetHeadlessServiceName(),
			Namespace: mi.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1beta1.SchemeGroupVersion.Group,
					Version: miniov1beta1.SchemeGroupVersion.Version,
					Kind:    miniov1beta1.ClusterCRDResourceKind,
				}),
			},
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{minioPort},
			Selector: map[string]string{
				constants.InstanceLabel: mi.Name,
			},
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}
