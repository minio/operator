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

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
)

// NewClusterIPForMinIO will return a new headless Kubernetes service for a Tenant
func NewClusterIPForMinIO(t *miniov1.Tenant) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.MinIOPort, Name: miniov1.MinIOServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.MinIOPodLabels(),
			Name:            t.MinIOCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{minioPort},
			Selector: t.MinIOPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	return svc
}

// NewHeadlessForMinIO will return a new headless Kubernetes service for a Tenant
func NewHeadlessForMinIO(t *miniov1.Tenant) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.MinIOPort, Name: miniov1.MinIOServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.MinIOPodLabels(),
			Name:            t.MinIOHLServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{minioPort},
			Selector:  t.MinIOPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

// NewHeadlessForKES will return a new headless Kubernetes service for a KES StatefulSet
func NewHeadlessForKES(t *miniov1.Tenant) *corev1.Service {
	kesPort := corev1.ServicePort{Port: miniov1.KESPort, Name: miniov1.KESServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.KESPodLabels(),
			Name:            t.KESHLServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{kesPort},
			Selector:  t.KESPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

// NewClusterIPForMCS will return a new cluster IP service for Console Deployment
func NewClusterIPForMCS(t *miniov1.Tenant) *corev1.Service {
	minioPort := corev1.ServicePort{Port: miniov1.ConsolePort, Name: miniov1.ConsoleServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.ConsolePodLabels(),
			Name:            t.ConsoleCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{minioPort},
			Selector: t.ConsolePodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	return svc
}
