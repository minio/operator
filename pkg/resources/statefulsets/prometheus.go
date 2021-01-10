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

package statefulsets

import (
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// prometheusMetadata returns the object metadata for Prometheus pods. User
// specified metadata in the spec is also included here.
func prometheusMetadata(t *miniov2.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Labels:      t.Spec.Prometheus.Labels,
		Annotations: t.Spec.Prometheus.Annotations,
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range t.PrometheusPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// prometheusSelector returns the prometheus pods selector
func prometheusSelector(t *miniov2.Tenant) *metav1.LabelSelector {
	m := t.PrometheusPodLabels()
	for k, v := range t.Spec.Prometheus.Labels {
		m[k] = v
	}
	return &metav1.LabelSelector{
		MatchLabels: m,
	}
}

func prometheusEnvVars(t *miniov2.Tenant) []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func prometheusConfigVolumeMount(t *miniov2.Tenant) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      t.PrometheusConfigVolMountName(),
		MountPath: "/etc/prometheus/prometheus.yml",
		SubPath:   "prometheus.yml",
	}
}

func prometheusVolumeMounts(t *miniov2.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.PrometheusStatefulsetName(),
			MountPath: "/prometheus",
			SubPath:   "prometheus",
		},
		prometheusConfigVolumeMount(t),
	}
}

// prometheusServerContainer returns a container for Prometheus StatefulSet.
func prometheusServerContainer(t *miniov2.Tenant) corev1.Container {
	// as per https://github.com/prometheus-operator/prometheus-operator/issues/3459 we need to set a security
	// context.
	//
	//  securityContext:
	//    fsGroup: 2000
	//    runAsNonRoot: true
	//    runAsUser: 1000
	runAsNonRoot := true
	var fsGroup int64 = 2000
	var runAsUser int64 = 1000
	return corev1.Container{
		Name:  miniov2.PrometheusContainerName,
		Image: miniov2.PrometheusImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.PrometheusPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    prometheusVolumeMounts(t),
		Env:             prometheusEnvVars(t),
		Resources:       t.Spec.Prometheus.Resources,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:    &runAsUser,
			RunAsGroup:   &fsGroup,
			RunAsNonRoot: &runAsNonRoot,
		},
	}
}

const prometheusDefaultVolumeSize = 5 * 1024 * 1024 * 1024 // 5GiB

// NewForPrometheus creates a new Prometheus StatefulSet for prometheus metrics
func NewForPrometheus(t *miniov2.Tenant, serviceName string) *appsv1.StatefulSet {
	var replicas int32 = 1
	promMeta := metav1.ObjectMeta{
		Name:            t.PrometheusStatefulsetName(),
		Namespace:       t.Namespace,
		OwnerReferences: t.OwnerRef(),
	}
	// Create a PVC to for prometheus storage
	volumeReq := corev1.ResourceList{}
	volumeSize := int64(prometheusDefaultVolumeSize)
	if t.Spec.Prometheus.DiskCapacityDB != nil && *t.Spec.Prometheus.DiskCapacityDB > 0 {
		volumeSize = int64(*t.Spec.Prometheus.DiskCapacityDB) * 1024 * 1024 * 1024
	}
	volumeReq[corev1.ResourceStorage] = *resource.NewQuantity(volumeSize, resource.BinarySI)
	volumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: promMeta,
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:   corev1.ResourceRequirements{Requests: volumeReq},
		},
	}

	podVolumes := []corev1.Volume{
		{
			Name: t.PrometheusConfigVolMountName(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.PrometheusConfigMapName(),
					},
				},
			},
		},
	}

	containers := []corev1.Container{prometheusServerContainer(t)}

	// init container per https://github.com/prometheus-operator/prometheus-operator/issues/966#issuecomment-365037713
	//
	//  initContainers:
	//  - name: "init-chown-data"
	//    image: "busybox"
	//    # 1000 is the user that prometheus uses.
	//    command: ["chown", "-R", "1000:2000", /var/prometheus/data]
	//    volumeMounts:
	//    - name: prometheus-kube-prometheus-db
	//      mountPath: /var/prometheus/data
	initContainers := []corev1.Container{
		{
			Name:  "prometheus-init-chown-data",
			Image: "busybox",
			Command: []string{
				"chown",
				"-R",
				"1000:2000",
				"/prometheus",
			},
			VolumeMounts: prometheusVolumeMounts(t),
		},
	}

	ss := &appsv1.StatefulSet{
		ObjectMeta: promMeta,
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov2.DefaultUpdateStrategy,
			},
			PodManagementPolicy:  t.Spec.PodManagementPolicy,
			Selector:             prometheusSelector(t),
			ServiceName:          serviceName,
			Replicas:             &replicas,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{volumeClaim},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: prometheusMetadata(t),
				Spec: corev1.PodSpec{
					ServiceAccountName: t.Spec.ServiceAccountName,
					Containers:         containers,
					Volumes:            podVolumes,
					RestartPolicy:      corev1.RestartPolicyAlways,
					SchedulerName:      t.Scheduler.Name,
					NodeSelector:       t.Spec.Prometheus.NodeSelector,
					InitContainers:     initContainers,
				},
			},
		},
	}
	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return ss
}
