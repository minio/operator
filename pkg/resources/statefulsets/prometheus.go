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
	"fmt"
	"strconv"

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
		MountPath: "/etc/prometheus",
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
	return corev1.Container{
		Name:  miniov2.PrometheusContainerName,
		Image: t.Spec.Prometheus.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.PrometheusPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    prometheusVolumeMounts(t),
		Env:             prometheusEnvVars(t),
		Resources:       t.Spec.Prometheus.Resources,
		Args: []string{
			"--config.file=/etc/prometheus/prometheus.yml",
			"--storage.tsdb.path=/prometheus",
			"--web.console.libraries=/usr/share/prometheus/console_libraries",
			"--web.console.templates=/usr/share/prometheus/consoles",
			"--web.enable-lifecycle",
		},
	}
}

// prometheusSidecarContainer returns a container for Prometheus sidecar. It
// simply listens for changes to the config file (mounted via the configmap) and
// calls Prometheus server's reload API.
func prometheusSidecarContainer(t *miniov2.Tenant) corev1.Container {
	return corev1.Container{
		Name:            miniov2.PrometheusContainerName + "-sidecar",
		Image:           t.Spec.Prometheus.SideCarImage,
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    prometheusVolumeMounts(t),
		Env:             prometheusEnvVars(t),
		Resources:       t.Spec.Prometheus.Resources,
		Command:         []string{"/bin/sh"},
		Args: []string{
			"-c",
			`echo -e '#!/bin/sh\n\nset -e\nset -x\necho "POST /-/reload HTTP/1.1\r\nHost:localhost:9090\r\nConnection: close\r\n\r\n" | nc localhost 9090\n' > /tmp/run.sh && echo "ok" && chmod +x /tmp/run.sh && inotifyd /tmp/run.sh /etc/prometheus/prometheus.yml:w`,
		},
	}

}

const prometheusDefaultVolumeSize = 5 * 1024 * 1024 * 1024 // 5GiB

// prometheusSecurityContext builds the security context for prometheus pods
func prometheusSecurityContext(t *miniov2.Tenant) *corev1.PodSecurityContext {
	var runAsNonRoot = true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	var fsGroup int64 = 1000
	var securityContext = corev1.PodSecurityContext{
		RunAsNonRoot: &runAsNonRoot,
		RunAsUser:    &runAsUser,
		RunAsGroup:   &runAsGroup,
		FSGroup:      &fsGroup,
	}
	if t.HasPrometheusEnabled() && t.Spec.Prometheus.SecurityContext != nil {
		securityContext = *t.Spec.Prometheus.SecurityContext
	}
	return &securityContext
}

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
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:        corev1.ResourceRequirements{Requests: volumeReq},
			StorageClassName: t.Spec.Prometheus.StorageClassName,
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

	containers := []corev1.Container{prometheusServerContainer(t), prometheusSidecarContainer(t)}

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

	// Attach security Policy
	securityContext := prometheusSecurityContext(t)

	var initContainerSecurityContext corev1.SecurityContext
	var initContainers []corev1.Container
	// If securityContext is present InitContainer still requires running with elevated privileges
	// and user will have to provide a serviceAccount that allows this
	if securityContext != nil && securityContext.RunAsUser != nil && securityContext.RunAsGroup != nil {
		var runAsUser int64
		var runAsNonRoot = false
		var allowPrivilegeEscalation = true
		initContainerSecurityContext = corev1.SecurityContext{
			RunAsUser:                &runAsUser,
			RunAsNonRoot:             &runAsNonRoot,
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		}
		initContainers = []corev1.Container{
			{
				Name:  "prometheus-init-chown-data",
				Image: t.Spec.Prometheus.InitImage,
				Command: []string{
					"chown",
					"-R",
					fmt.Sprintf("%s:%s", strconv.FormatInt(*securityContext.RunAsUser, 10), strconv.FormatInt(*securityContext.RunAsGroup, 10)),
					"/prometheus",
				},
				SecurityContext: &initContainerSecurityContext,
				VolumeMounts:    prometheusVolumeMounts(t),
			},
		}
	}

	serviceAccount := t.Spec.ServiceAccountName
	if t.Spec.Prometheus.ServiceAccountName != "" {
		serviceAccount = t.Spec.Prometheus.ServiceAccountName
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
					ServiceAccountName:        serviceAccount,
					Containers:                containers,
					Volumes:                   podVolumes,
					RestartPolicy:             corev1.RestartPolicyAlways,
					SchedulerName:             t.Scheduler.Name,
					NodeSelector:              t.Spec.Prometheus.NodeSelector,
					Affinity:                  t.Spec.Prometheus.Affinity,
					TopologySpreadConstraints: t.Spec.Prometheus.TopologySpreadConstraints,
					InitContainers:            initContainers,
					SecurityContext:           securityContext,
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
