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

package services

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
)

// NewClusterIPForMinIO will return a new headless Kubernetes service for a MinIOInstance
func NewClusterIPForMinIO(mi *miniov1.MinIOInstance) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.MinIOPort, Name: miniov1.MinIOServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          mi.MinIOPodLabels(),
			Name:            mi.MinIOCIServiceName(),
			Namespace:       mi.Namespace,
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{minioPort},
			Selector: mi.MinIOPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	return svc
}

// NewHeadlessForMinIO will return a new headless Kubernetes service for a MinIOInstance
func NewHeadlessForMinIO(mi *miniov1.MinIOInstance) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.MinIOPort, Name: miniov1.MinIOServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          mi.MinIOPodLabels(),
			Name:            mi.MinIOHLServiceName(),
			Namespace:       mi.Namespace,
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{minioPort},
			Selector:  mi.MinIOPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

// NewHeadlessForKES will return a new headless Kubernetes service for a KES StatefulSet
func NewHeadlessForKES(mi *miniov1.MinIOInstance) *corev1.Service {
	kesPort := corev1.ServicePort{Port: miniov1.KESPort, Name: miniov1.KESServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          mi.KESPodLabels(),
			Name:            mi.KESHLServiceName(),
			Namespace:       mi.Namespace,
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{kesPort},
			Selector:  mi.KESPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

// NewClusterIPForMCS will return a new cluster IP service for MCS Deployment
func NewClusterIPForMCS(mi *miniov1.MinIOInstance) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.MCSPort, Name: miniov1.MCSServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          mi.MCSPodLabels(),
			Name:            mi.MCSCIServiceName(),
			Namespace:       mi.Namespace,
			OwnerReferences: mi.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{minioPort},
			Selector: mi.MCSPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	return svc
}
