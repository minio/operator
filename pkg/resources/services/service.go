// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package services

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// NewClusterIPForMinIO will return a new ClusterIP Kubernetes service for a Tenant
func NewClusterIPForMinIO(t *miniov2.Tenant) *corev1.Service {
	var port int32 = miniov2.MinIOPortLoadBalancerSVC
	name := miniov2.MinIOServiceHTTPPortName
	if t.TLS() {
		port = miniov2.MinIOTLSPortLoadBalancerSVC
		name = miniov2.MinIOServiceHTTPSPortName
	}

	minioPort := corev1.ServicePort{
		Port:       port,
		Name:       name,
		TargetPort: intstr.FromInt(miniov2.MinIOPort),
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.MinIOCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:                    []corev1.ServicePort{minioPort},
			Selector:                 t.MinIOPodLabels(),
			Type:                     corev1.ServiceTypeClusterIP,
			PublishNotReadyAddresses: false,
		},
	}

	if t.Spec.ServiceMetadata != nil {
		svc.Labels = t.Spec.ServiceMetadata.MinIOServiceLabels
		svc.Annotations = t.Spec.ServiceMetadata.MinIOServiceAnnotations
	}

	// check if the service is meant to be exposed
	if t.Spec.ExposeServices != nil && t.Spec.ExposeServices.MinIO {
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	}
	return svc
}

// NewClusterIPForConsole will return a new cluster IP service for Console Deployment
func NewClusterIPForConsole(t *miniov2.Tenant) *corev1.Service {
	consolePort := corev1.ServicePort{
		Port:       miniov2.ConsolePort,
		TargetPort: intstr.FromInt(miniov2.ConsolePort),
		Name:       miniov2.ConsoleServicePortName,
	}
	if t.TLS() {
		consolePort = corev1.ServicePort{
			Port:       miniov2.ConsoleTLSPort,
			TargetPort: intstr.FromInt(miniov2.ConsoleTLSPort),
			Name:       miniov2.ConsoleServiceTLSPortName,
		}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.ConsoleCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				consolePort,
			},
			Selector: t.MinIOPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	if t.Spec.ServiceMetadata != nil {
		svc.Labels = t.Spec.ServiceMetadata.ConsoleServiceLabels
		svc.Annotations = t.Spec.ServiceMetadata.ConsoleServiceAnnotations
	}

	// check if the service is meant to be exposed
	if t.Spec.ExposeServices != nil && t.Spec.ExposeServices.Console {
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	}

	return svc
}

// ServiceForBucket will return a external name based service
func ServiceForBucket(t *miniov2.Tenant, bucket string) *corev1.Service {
	var port int32 = miniov2.MinIOPortLoadBalancerSVC
	name := miniov2.MinIOServiceHTTPPortName
	if t.TLS() {
		port = miniov2.MinIOTLSPortLoadBalancerSVC
		name = miniov2.MinIOServiceHTTPSPortName
	}
	minioPort := corev1.ServicePort{
		Port:       port,
		Name:       name,
		TargetPort: intstr.FromInt(miniov2.MinIOPort),
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            bucket,
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			ExternalName: t.MinIOFQDNServiceName(),
			Ports:        []corev1.ServicePort{minioPort},
			Type:         corev1.ServiceTypeExternalName,
		},
	}
	return svc
}

// NewHeadlessForMinIO will return a new headless Kubernetes service for a Tenant
func NewHeadlessForMinIO(t *miniov2.Tenant) *corev1.Service {
	name := miniov2.MinIOServiceHTTPPortName
	if t.TLS() {
		name = miniov2.MinIOServiceHTTPSPortName
	}
	minioPort := corev1.ServicePort{Port: miniov2.MinIOPort, Name: name}
	ports := []corev1.ServicePort{minioPort}

	if t.Spec.Features != nil && t.Spec.Features.EnableSFTP != nil && *t.Spec.Features.EnableSFTP {
		minioSFTPPort := corev1.ServicePort{Port: miniov2.MinIOSFTPPort, Name: miniov2.MinIOServiceSFTPPortName}
		ports = append(ports, minioSFTPPort)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            t.MinIOHLServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:                    ports,
			Selector:                 t.MinIOPodLabels(),
			Type:                     corev1.ServiceTypeClusterIP,
			ClusterIP:                corev1.ClusterIPNone,
			PublishNotReadyAddresses: true,
		},
	}

	return svc
}

// NewHeadlessForKES will return a new headless Kubernetes service for a KES StatefulSet
func NewHeadlessForKES(t *miniov2.Tenant) *corev1.Service {
	kesPort := corev1.ServicePort{Port: miniov2.KESPort, Name: miniov2.KESServicePortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
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
	if t.Spec.ServiceMetadata != nil {
		svc.Labels = t.Spec.ServiceMetadata.KESServiceLabels
		svc.Annotations = t.Spec.ServiceMetadata.KESServiceAnnotations
	}
	return svc
}
