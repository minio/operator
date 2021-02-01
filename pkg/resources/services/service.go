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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// NewClusterIPForMinIO will return a new headless Kubernetes service for a Tenant
func NewClusterIPForMinIO(t *miniov2.Tenant) *corev1.Service {
	var port int32 = miniov2.MinIOPortLoadBalancerSVC
	var name string = miniov2.MinIOServiceHTTPPortName
	var internalLabels, labels, annotations map[string]string

	internalLabels = t.MinIOPodLabels()
	if t.TLS() {
		port = miniov2.MinIOTLSPortLoadBalancerSVC
		name = miniov2.MinIOServiceHTTPSPortName
	}
	if t.Spec.ServiceMetadata != nil && t.Spec.ServiceMetadata.MinIOServiceLabels != nil {
		labels = miniov2.MergeMaps(internalLabels, t.Spec.ServiceMetadata.MinIOServiceLabels)
		annotations = t.Spec.ServiceMetadata.MinIOServiceAnnotations
	}
	minioPort := corev1.ServicePort{
		Port:       port,
		Name:       name,
		TargetPort: intstr.FromInt(miniov2.MinIOPort),
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          labels,
			Name:            t.MinIOCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
			Annotations:     annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{minioPort},
			Selector: t.MinIOPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	// check if the service is meant to be exposed
	if t.Spec.ExposeServices != nil && t.Spec.ExposeServices.MinIO {
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	}

	return svc
}

// ServiceForBucket will return a external name based service
func ServiceForBucket(t *miniov2.Tenant, bucket string) *corev1.Service {
	var port int32 = miniov2.MinIOPortLoadBalancerSVC
	var name string = miniov2.MinIOServiceHTTPPortName
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
	minioPort := corev1.ServicePort{Port: miniov2.MinIOPort, Name: miniov2.MinIOServiceHTTPPortName}
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
func NewHeadlessForKES(t *miniov2.Tenant) *corev1.Service {
	kesPort := corev1.ServicePort{Port: miniov2.KESPort, Name: miniov2.KESServicePortName}
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

// NewHeadlessForLog returns a k8s Headless service object for Log
func NewHeadlessForLog(t *miniov2.Tenant) *corev1.Service {
	searchPort := corev1.ServicePort{Port: miniov2.LogPgPort, Name: miniov2.LogPgPortName}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.LogPgPodLabels(),
			Name:            t.LogHLServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{searchPort},
			Selector:  t.LogPgPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	return svc
}

// NewHeadlessForPrometheus returns a k8s Headless service object for the
// Prometheus StatefulSet.
func NewHeadlessForPrometheus(t *miniov2.Tenant) *corev1.Service {
	promPort := corev1.ServicePort{Port: miniov2.PrometheusPort, Name: miniov2.PrometheusPortName}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.PrometheusPodLabels(),
			Name:            t.PrometheusHLServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports:     []corev1.ServicePort{promPort},
			Selector:  t.PrometheusPodLabels(),
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

// NewClusterIPForConsole will return a new cluster IP service for Console Deployment
func NewClusterIPForConsole(t *miniov2.Tenant) *corev1.Service {
	var internalLabels, labels, annotations map[string]string
	internalLabels = t.ConsolePodLabels()

	consolePort := corev1.ServicePort{Port: miniov2.ConsolePort, Name: miniov2.ConsoleServicePortName}
	if t.TLS() || t.ConsoleExternalCert() {
		consolePort = corev1.ServicePort{Port: miniov2.ConsoleTLSPort, Name: miniov2.ConsoleServiceTLSPortName}
	}
	if t.Spec.ServiceMetadata != nil && t.Spec.ServiceMetadata.ConsoleServiceLabels != nil {
		labels = miniov2.MergeMaps(internalLabels, t.Spec.ServiceMetadata.ConsoleServiceLabels)
		annotations = t.Spec.ServiceMetadata.ConsoleServiceAnnotations
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          labels,
			Name:            t.ConsoleCIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
			Annotations:     annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				consolePort,
			},
			Selector: t.ConsolePodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	// check if the service is meant to be exposed
	if t.Spec.ExposeServices != nil && t.Spec.ExposeServices.Console {
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	}

	return svc
}

// NewClusterIPForLogSearchAPI will return a new cluster IP service object for log-search-api deployment
func NewClusterIPForLogSearchAPI(t *miniov2.Tenant) *corev1.Service {
	logSearchAPIPort := corev1.ServicePort{Port: miniov2.LogSearchAPIPort, Name: miniov2.LogSearchAPIPortName}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          t.LogSearchAPIPodLabels(),
			Name:            t.LogSearchAPIServiceName(),
			Namespace:       t.Namespace,
			OwnerReferences: t.OwnerRef(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				logSearchAPIPort,
			},
			Selector: t.LogSearchAPIPodLabels(),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

// GetLogSearchDBAddr returns the tenant's Postgres DB server address
func GetLogSearchDBAddr(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s.%s.svc.%s:%d", t.LogHLServiceName(), t.Namespace, miniov2.GetClusterDomain(), miniov2.LogPgPort)
}

// GetLogSearchAPIAddr returns the tenant's log-search-api server address
func GetLogSearchAPIAddr(t *miniov2.Tenant) string {
	return fmt.Sprintf("http://%s.%s.svc.%s:%d", t.LogSearchAPIServiceName(), t.Namespace, miniov2.GetClusterDomain(), miniov2.LogSearchAPIPort)
}
