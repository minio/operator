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

package servicemonitor

import (
	"errors"
	"log"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	constants "github.com/minio/minio-operator/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func hasServiceMonitor() (bool, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error getting cluster config: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error getting cluster config: %s", err.Error())
	}

	serviceMonitorGV := schema.GroupVersion{
		Group:   monitoringv1.SchemeGroupVersion.Group,
		Version: monitoringv1.SchemeGroupVersion.Version,
	}

	if err := discovery.ServerSupportsVersion(clientset, serviceMonitorGV); err != nil {
		// The service monitor API is not present
		return false, errors.New("ServiceMonitor API does not exist")
	}

	return true, nil
}

func createServiceMonitor(mi *miniov1beta1.MinIOInstance) monitoringv1.ServiceMonitor {
	return monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mi.Name + "-servicemonitor",
			Namespace: mi.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1beta1.SchemeGroupVersion.Group,
					Version: miniov1beta1.SchemeGroupVersion.Version,
					Kind:    miniov1beta1.ClusterCRDResourceKind,
				}),
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.InstanceLabel: mi.Name,
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					TargetPort: &intstr.IntOrString{
						IntVal: 9000,
					},
					Path:   "/minio/prometheus/metrics",
					Scheme: "http",
				},
			},
		},
	}
}
