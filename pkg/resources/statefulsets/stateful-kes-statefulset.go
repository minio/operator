// Copyright (C) 2022, MinIO, Inc.
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

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const statefulKESInitMountPath = "/mnt/kes"

// StatefulKESMetadata Returns the stateful KES pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func StatefulKESMetadata(t *miniov2.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	meta.Labels = t.Spec.StatefulKES.Labels
	meta.Annotations = t.Spec.StatefulKES.Annotations

	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range t.StatefulKESPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// StatefulKESSelector Returns the stateful KES pods selector set in configuration.
func StatefulKESSelector(t *miniov2.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.StatefulKESPodLabels(),
	}
}

// StatefulKESVolumeMounts builds the volume mounts for MinIO stateful-kes container.
func StatefulKESVolumeMounts(t *miniov2.Tenant) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      t.StatefulKESVolMountName(),
			MountPath: miniov2.StatefulKESConfigMountPath,
		},
	}

	if t.Spec.StatefulKES.VolumeClaimTemplate != nil {
		name := t.Spec.StatefulKES.VolumeClaimTemplate.Name
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      name,
			MountPath: statefulKESInitMountPath,
		})
	}

	return volumeMounts
}

// StatefulKESEnvironmentVars returns the StatefulKES environment variables set in configuration.
func StatefulKESEnvironmentVars(t *miniov2.Tenant) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	// Add all the tenant.spec.kes.env environment variables
	// User defined environment variables will take precedence over default environment variables
	envVars = append(envVars, t.GetStatefulKESEnvVars()...)
	// sort the array to produce the same result everytime
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	return envVars
}

// - result=$([[ -d /tmp/stateful-kes/data ]]; echo $?); if [ "$result" -eq 1 ]; then /kes init --config=/tmp/stateful-kes/init-config.yaml /mnt/kes/data; fi

// StatefulKESInitContainer returns the KES container for a KES StatefulSet.
func StatefulKESInitContainer(t *miniov2.Tenant) corev1.Container {
	var force bool

	kesInitCommand := fmt.Sprintf("/kes init --config=%s/init-config.yaml %s/data", miniov2.StatefulKESConfigMountPath, statefulKESInitMountPath)
	initCommand := fmt.Sprintf("result=$([[ -d %s/data ]]; echo $?); if [ \"$result\" -eq 1 ]; then %s",
		statefulKESInitMountPath, kesInitCommand)
	if force {
		initCommand = initCommand + fmt.Sprintf(";else %s --force", kesInitCommand)
	}
	initCommand = initCommand + "; fi"

	// Args to start KES with config mounted at miniov2.KESConfigMountPath and require but don't verify mTLS authentication
	args := []string{
		"-c",
		initCommand,
	}

	return corev1.Container{
		Name:    miniov2.StatefulKESInitContainerName,
		Image:   t.Spec.StatefulKES.Image,
		Command: []string{"sh"},
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.StatefulKESPort,
			},
		},
		ImagePullPolicy: t.Spec.StatefulKES.ImagePullPolicy,
		VolumeMounts:    StatefulKESVolumeMounts(t),
		Args:            args,
		Env:             StatefulKESEnvironmentVars(t),
		Resources:       t.Spec.StatefulKES.Resources,
	}
}

// StatefulKESServerContainer returns the StatefulKES container for a StatefulKES StatefulSet.
func StatefulKESServerContainer(t *miniov2.Tenant) corev1.Container {
	// Args to start KES with config mounted at miniov2.KESConfigMountPath and require but don't verify mTLS authentication
	args := []string{"server", statefulKESInitMountPath + "/data"}

	return corev1.Container{
		Name:  miniov2.StatefulKESContainerName,
		Image: t.Spec.StatefulKES.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.KESPort,
			},
		},
		ImagePullPolicy: t.Spec.StatefulKES.ImagePullPolicy,
		VolumeMounts:    StatefulKESVolumeMounts(t),
		Args:            args,
		Env:             StatefulKESEnvironmentVars(t),
		Resources:       t.Spec.StatefulKES.Resources,
	}
}

// statefulkesSecurityContext builds the security context for stateful KES statefulset pods
func statefulkesSecurityContext(t *miniov2.Tenant) *corev1.PodSecurityContext {
	runAsNonRoot := true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	var fsGroup int64 = 1000
	fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch

	securityContext := corev1.PodSecurityContext{
		RunAsNonRoot:        &runAsNonRoot,
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &fsGroupChangePolicy,
	}
	if t.HasStatefulKESEnabled() && t.Spec.StatefulKES.SecurityContext != nil {
		securityContext = *t.Spec.StatefulKES.SecurityContext
	}
	return &securityContext
}

// NewForStatefulKES creates a statefulset for stateful-kes
func NewForStatefulKES(t *miniov2.Tenant, serviceName string) *appsv1.StatefulSet {
	replicas := t.StatefulKESReplicas()
	// certificate files used by the KES server
	certPath := "server.crt"
	keyPath := "server.key"

	var volumeProjections []corev1.VolumeProjection

	var serverCertSecret string
	// clientCertSecret holds certificate files (public.crt, private.key and ca.crt) used by KES
	// in mTLS with a KMS (eg: authentication with Vault)
	var clientCertSecret string

	serverCertPaths := []corev1.KeyToPath{
		{Key: "public.crt", Path: certPath},
		{Key: "private.key", Path: keyPath},
	}

	configPath := []corev1.KeyToPath{
		{Key: "init-config.yaml", Path: "init-config.yaml"},
	}

	// External certificates will have priority over AutoCert generated certificates
	if t.StatefulKESExternalCert() {
		serverCertSecret = t.Spec.StatefulKES.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.StatefulKES.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.StatefulKES.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" || t.Spec.StatefulKES.ExternalCertSecret.Type == "cert-manager.io/v1" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: certPath},
				{Key: "tls.key", Path: keyPath},
			}
		}
	} else {
		serverCertSecret = t.StatefulKESTLSSecretName()
	}

	if t.StatefulKESClientCert() {
		clientCertSecret = t.Spec.StatefulKES.ClientCertSecret.Name
	}

	if t.Spec.StatefulKES.Configuration.Name != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.Spec.StatefulKES.Configuration.Name,
				},
				Items: configPath,
			},
		})
	}

	if serverCertSecret != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: serverCertSecret,
				},
				Items: serverCertPaths,
			},
		})
	}

	if clientCertSecret != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: clientCertSecret,
				},
			},
		})
	}

	podVolumes := []corev1.Volume{
		{
			Name: t.StatefulKESVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: volumeProjections,
				},
			},
		},
	}

	initContainers := []corev1.Container{StatefulKESInitContainer(t)}
	containers := []corev1.Container{StatefulKESServerContainer(t)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.StatefulKESStatefulSetName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov2.DefaultUpdateStrategy,
			},
			PodManagementPolicy: t.Spec.PodManagementPolicy,
			// KES is always matched via Tenant Name + KES prefix
			Selector:    StatefulKESSelector(t),
			ServiceName: serviceName,
			Replicas:    &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: StatefulKESMetadata(t),
				Spec: corev1.PodSpec{
					ServiceAccountName:        t.Spec.StatefulKES.ServiceAccountName,
					InitContainers:            initContainers,
					Containers:                containers,
					Volumes:                   podVolumes,
					RestartPolicy:             corev1.RestartPolicyAlways,
					SchedulerName:             t.Scheduler.Name,
					NodeSelector:              t.Spec.StatefulKES.NodeSelector,
					Tolerations:               t.Spec.StatefulKES.Tolerations,
					Affinity:                  t.Spec.StatefulKES.Affinity,
					TopologySpreadConstraints: t.Spec.StatefulKES.TopologySpreadConstraints,
					SecurityContext:           statefulkesSecurityContext(t),
				},
			},
		},
	}
	// Address issue https://github.com/kubernetes/kubernetes/issues/85332
	if t.Spec.ImagePullSecret.Name != "" {
		ss.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{t.Spec.ImagePullSecret}
	}

	if t.Spec.StatefulKES.VolumeClaimTemplate != nil {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *t.Spec.StatefulKES.VolumeClaimTemplate)
	}

	return ss
}
