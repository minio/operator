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

package statefulsets

import (
	"fmt"
	"sort"
	"strconv"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// logSelector returns the Log pods selector
func logSelector(t *miniov2.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.LogPgPodLabels(),
	}
}

// logDbMetadata returns the object metadata for Log pods
func logDbMetadata(t *miniov2.Tenant) metav1.ObjectMeta {
	labels := make(map[string]string)
	labels[miniov2.LogDbLabel] = t.Name
	for k, v := range t.LogPgPodLabels() {
		labels[k] = v
	}

	meta := metav1.ObjectMeta{
		Labels: labels,
	}

	if t.Spec.Log.Db != nil {
		// attach any labels
		for k, v := range t.Spec.Log.Db.Labels {
			meta.Labels[k] = v
		}
		// attach any annotations
		if len(t.Spec.Log.Db.Annotations) > 0 {
			meta.Annotations = make(map[string]string)
			for k, v := range t.Spec.Log.Db.Annotations {
				meta.Annotations[k] = v
			}
		}
	}

	return meta
}

// logEnvVars returns env with POSTGRES_DB set to log database, POSTGRES_USER and POSTGRES_PASSWORD from Log's k8s secret
func logEnvVars(t *miniov2.Tenant) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	envVarsMap := map[string]corev1.EnvVar{
		miniov2.LogAuditDBKey: {
			Name:  miniov2.LogAuditDBKey,
			Value: miniov2.LogAuditDB,
		},
		miniov2.LogPgUserKey: {
			Name:  miniov2.LogPgUserKey,
			Value: miniov2.LogPgUser,
		},
		miniov2.LogPgPassKey: {
			Name: miniov2.LogPgPassKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: t.LogSecretName(),
					},
					Key: miniov2.LogPgPassKey,
				},
			},
		},
	}
	// Add all the tenant.spec.log.db.env environment variables
	// User defined environment variables will take precedence over default environment variables
	for _, env := range t.GetLogDBEnvVars() {
		envVarsMap[env.Name] = env
	}
	// transform map to array
	for _, env := range envVarsMap {
		envVars = append(envVars, env)
	}
	// sort the array to produce the same result everytime
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	return envVars
}

func logVolumeMounts(t *miniov2.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.LogStatefulsetName(),
			MountPath: "/var/lib/postgresql/data",
			SubPath:   "data",
		},
	}
}

// logDbContainer returns a postgresql server container for a Log StatefulSet.
func logDbContainer(t *miniov2.Tenant) corev1.Container {
	container := corev1.Container{
		Name:  miniov2.LogPgContainerName,
		Image: miniov2.LogPgImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.LogPgPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    logVolumeMounts(t),
		Env:             logEnvVars(t),
	}
	// if we have DB configurations
	if t.Spec.Log.Db != nil {
		// if an image was specified, use it.
		if t.Spec.Log.Db.Image != "" {
			container.Image = t.Spec.Log.Db.Image
		}
		// resources constraints
		container.Resources = t.Spec.Log.Db.Resources
	}
	return container
}

// defaultLogVolumeSize is a fallback value if the volume claim template for the DB is not provided
const defaultLogVolumeSize = 5 * 1024 * 1024 * 1024 // 5GiB

// postgresSecurityContext builds the security context for postgres pods
func postgresSecurityContext(t *miniov2.Tenant) *corev1.PodSecurityContext {
	runAsNonRoot := true
	var runAsUser int64 = 999
	var runAsGroup int64 = 999
	var fsGroup int64 = 999
	fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch

	securityContext := corev1.PodSecurityContext{
		RunAsNonRoot:        &runAsNonRoot,
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &fsGroupChangePolicy,
	}
	if t.HasLogSearchAPIEnabled() && t.Spec.Log.Db != nil && t.Spec.Log.Db.SecurityContext != nil {
		securityContext = *t.Spec.Log.Db.SecurityContext
	}
	return &securityContext
}

// NewForLogDb creates a new Log StatefulSet for Log feature
func NewForLogDb(t *miniov2.Tenant, serviceName string) *appsv1.StatefulSet {
	var replicas int32 = 1
	logMeta := metav1.ObjectMeta{
		Name:            t.LogStatefulsetName(),
		Namespace:       t.Namespace,
		OwnerReferences: t.OwnerRef(),
	}

	// Volume for the Logs Database
	var volumeClaim corev1.PersistentVolumeClaim
	if t.Spec.Log.Db != nil && t.Spec.Log.Db.VolumeClaimTemplate != nil {
		volumeClaim = *t.Spec.Log.Db.VolumeClaimTemplate
		if volumeClaim.ObjectMeta.Name == "" {
			volumeClaim.ObjectMeta = metav1.ObjectMeta{
				Name: t.LogStatefulsetName(),
			}
		}
	} else {
		// Create a PVC to store log data
		volumeReq := corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(defaultLogVolumeSize, resource.BinarySI),
		}
		volumeClaim = corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: t.LogStatefulsetName(),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources:   corev1.ResourceRequirements{Requests: volumeReq},
			},
		}
	}
	dbPod := corev1.PodTemplateSpec{
		ObjectMeta: logDbMetadata(t),
		Spec: corev1.PodSpec{
			ServiceAccountName: t.Spec.ServiceAccountName,
			Containers:         []corev1.Container{logDbContainer(t)},
			RestartPolicy:      corev1.RestartPolicyAlways,
			SchedulerName:      t.Scheduler.Name,
		},
	}
	// if we have DB configurations to honor
	if t.Spec.Log.Db != nil {

		// Attach security Policy
		dbPod.Spec.SecurityContext = postgresSecurityContext(t)

		var initContainerSecurityContext corev1.SecurityContext
		var initContainers []corev1.Container

		// If securityContext is present InitContainer still requires running with elevated privileges
		// and user will have to provide a serviceAccount that allows this
		if dbPod.Spec.SecurityContext != nil &&
			dbPod.Spec.SecurityContext.RunAsUser != nil &&
			dbPod.Spec.SecurityContext.RunAsGroup != nil {

			var runAsUser int64
			runAsNonRoot := false
			allowPrivilegeEscalation := true
			initContainerSecurityContext = corev1.SecurityContext{
				RunAsUser:                &runAsUser,
				RunAsNonRoot:             &runAsNonRoot,
				AllowPrivilegeEscalation: &allowPrivilegeEscalation,
			}

			if t.Spec.Log != nil && t.Spec.Log.Db != nil {
				if t.Spec.Log.Db.InitImage != "" {
					initContainers = []corev1.Container{
						{
							Name:    "postgres-init-chown-data",
							Image:   t.Spec.Log.Db.InitImage,
							Command: []string{"sh"},
							Args: []string{
								"-c",
								fmt.Sprintf(`echo -e '#!/bin/sh\n\nchown -vR %s:%s /var/lib/postgresql/data || true\n' > /tmp/chown.sh && echo "ok" && chmod +x /tmp/chown.sh && /tmp/chown.sh`, strconv.FormatInt(*dbPod.Spec.SecurityContext.RunAsUser, 10), strconv.FormatInt(*dbPod.Spec.SecurityContext.RunAsGroup, 10)),
							},
							SecurityContext: &initContainerSecurityContext,
							VolumeMounts:    logVolumeMounts(t),
						},
					}
				}
			}
		}

		// attach affinity clauses
		if t.Spec.Log.Db.Affinity != nil {
			dbPod.Spec.Affinity = t.Spec.Log.Db.Affinity
		}
		// attach node selector clauses
		dbPod.Spec.NodeSelector = t.Spec.Log.Db.NodeSelector
		// attach tolerations
		dbPod.Spec.Tolerations = t.Spec.Log.Db.Tolerations
		// attach topology spread constraints
		dbPod.Spec.TopologySpreadConstraints = t.Spec.Log.Db.TopologySpreadConstraints
		// attach serviceAccount
		if t.Spec.Log.Db.ServiceAccountName != "" {
			dbPod.Spec.ServiceAccountName = t.Spec.Log.Db.ServiceAccountName
		}
		// attach init containers
		dbPod.Spec.InitContainers = initContainers
	}

	ss := &appsv1.StatefulSet{
		ObjectMeta: logMeta,
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov2.DefaultUpdateStrategy,
			},
			PodManagementPolicy:  t.Spec.PodManagementPolicy,
			Selector:             logSelector(t),
			ServiceName:          serviceName,
			Replicas:             &replicas,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{volumeClaim},
			Template:             dbPod,
		},
	}
	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	return ss
}
