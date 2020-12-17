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
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// logSelector returns the Log pods selector
func logSelector(t *miniov1.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.LogPgPodLabels(),
	}
}

// logMetadata returns the object metadata for Log pods
func logMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
	labels := make(map[string]string)
	labels[miniov1.TenantLabel] = t.Name
	for k, v := range t.LogPgPodLabels() {
		labels[k] = v
	}

	return metav1.ObjectMeta{
		Labels: labels,
	}
}

// logEnvVars returns env with POSTGRES_DB set to log database, POSTGRES_USER and POSTGRES_PASSWORD from Log's k8s secret
func logEnvVars(t *miniov1.Tenant) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  miniov1.LogAuditDBKey,
			Value: miniov1.LogAuditDB,
		},
		{
			Name:  miniov1.LogPgUserKey,
			Value: miniov1.LogPgUser,
		},
		{
			Name: miniov1.LogPgPassKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.LogSecretName(),
					},
					Key: miniov1.LogPgPassKey,
				},
			},
		},
	}

}

func logVolumeMounts(t *miniov1.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.LogStatefulsetName(),
			MountPath: "/var/lib/postgresql/data",
		},
	}
}

// logServerContainer returns a postgresql server container for a Log StatefulSet.
func logServerContainer(t *miniov1.Tenant) corev1.Container {
	return corev1.Container{
		Name:  miniov1.LogPgContainerName,
		Image: miniov1.LogPgImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.LogPgPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    logVolumeMounts(t),
		Env:             logEnvVars(t),
	}
}

const logDefaultVolumeSize = 5 * 1024 * 1024 * 1024 // 5GiB

// NewForLog creates a new Log StatefulSet for Log feature
func NewForLog(t *miniov1.Tenant, serviceName string) *appsv1.StatefulSet {
	var replicas int32 = 1
	logMeta := metav1.ObjectMeta{
		Name:            t.LogStatefulsetName(),
		Namespace:       t.Namespace,
		OwnerReferences: t.OwnerRef(),
	}
	// Create a PVC to store log data
	volumeReq := corev1.ResourceList{}
	volumeSize := int64(logDefaultVolumeSize)
	if t.Spec.Log.Audit.DiskCapacityGB != nil && *t.Spec.Log.Audit.DiskCapacityGB > 0 {
		volumeSize = int64(*t.Spec.Log.Audit.DiskCapacityGB) * 1024 * 1024 * 1024
	}
	volumeReq[corev1.ResourceStorage] = *resource.NewQuantity(volumeSize, resource.BinarySI)
	volumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: logMeta,
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:   corev1.ResourceRequirements{Requests: volumeReq},
		},
	}

	containers := []corev1.Container{logServerContainer(t)}
	ss := &appsv1.StatefulSet{
		ObjectMeta: logMeta,
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov1.DefaultUpdateStrategy,
			},
			PodManagementPolicy:  t.Spec.PodManagementPolicy,
			Selector:             logSelector(t),
			ServiceName:          serviceName,
			Replicas:             &replicas,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{volumeClaim},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: logMetadata(t),
				Spec: corev1.PodSpec{
					ServiceAccountName: t.Spec.ServiceAccountName,
					Containers:         containers,
					RestartPolicy:      corev1.RestartPolicyAlways,
					SchedulerName:      t.Scheduler.Name,
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
