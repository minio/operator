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
	"sort"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gcpCredentialVolumeMountName = "gcp-ksa"
	gcpCredentialVolumeMountPath = "/var/run/secrets/tokens/gcp-ksa"
	serviceAccountTokenPath      = "token"
	gcpAppCredentialsPath        = "google-application-credentials.json"
)

var (
	defaultServiceAccountTokenExpiryInSecs int64 = 172800 // 48 hrs
	// gcpCredentialVolumeMount represents the volume mount for GCP creds and service token
	gcpCredentialVolumeMount = corev1.VolumeMount{
		Name:      gcpCredentialVolumeMountName,
		ReadOnly:  true,
		MountPath: gcpCredentialVolumeMountPath,
	}
)

// KESMetadata Returns the KES pods metadata set in configuration.
// If a user specifies metadata in the spec we return that
// metadata.
func KESMetadata(t *miniov2.Tenant) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	meta.Labels = t.Spec.KES.Labels
	meta.Annotations = t.Spec.KES.Annotations

	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range t.KESPodLabels() {
		meta.Labels[k] = v
	}
	return meta
}

// KESSelector Returns the KES pods selector set in configuration.
func KESSelector(t *miniov2.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.KESPodLabels(),
	}
}

// KESVolumeMounts builds the volume mounts for MinIO container.
func KESVolumeMounts(t *miniov2.Tenant) (volumeMounts []corev1.VolumeMount) {
	volumeMounts = []corev1.VolumeMount{
		{
			Name:      t.KESVolMountName(),
			MountPath: miniov2.KESConfigMountPath,
		},
	}
	if t.HasGCPCredentialSecretForKES() {
		volumeMounts = append(volumeMounts, gcpCredentialVolumeMount)
	}
	return
}

// KESEnvironmentVars returns the KES environment variables set in configuration.
func KESEnvironmentVars(t *miniov2.Tenant) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	// Add all the tenant.spec.kes.env environment variables
	// User defined environment variables will take precedence over default environment variables
	envVars = append(envVars, t.GetKESEnvVars()...)
	// sort the array to produce the same result everytime
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	return envVars
}

// KESServerContainer returns the KES container for a KES StatefulSet.
func KESServerContainer(t *miniov2.Tenant) corev1.Container {
	// Args to start KES with config mounted at miniov2.KESConfigMountPath and require but don't verify mTLS authentication
	args := []string{"server", "--config=" + miniov2.KESConfigMountPath + "/server-config.yaml", "--auth=off"}

	return corev1.Container{
		Name:  miniov2.KESContainerName,
		Image: t.Spec.KES.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov2.KESPort,
			},
		},
		ImagePullPolicy: t.Spec.KES.ImagePullPolicy,
		VolumeMounts:    KESVolumeMounts(t),
		Args:            args,
		Env:             KESEnvironmentVars(t),
		Resources:       t.Spec.KES.Resources,
	}
}

// kesSecurityContext builds the security context for KES statefulset pods
func kesSecurityContext(t *miniov2.Tenant) *corev1.PodSecurityContext {
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
	if t.HasKESEnabled() && t.Spec.KES.SecurityContext != nil {
		securityContext = *t.Spec.KES.SecurityContext
	}
	return &securityContext
}

// NewForKES creates a new KES StatefulSet for the given Cluster.
func NewForKES(t *miniov2.Tenant, serviceName string) *appsv1.StatefulSet {
	replicas := t.KESReplicas()
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
		{Key: "server-config.yaml", Path: "server-config.yaml"},
	}

	// External certificates will have priority over AutoCert generated certificates
	if t.KESExternalCert() {
		serverCertSecret = t.Spec.KES.ExternalCertSecret.Name
		// This covers both secrets of type "kubernetes.io/tls" and
		// "cert-manager.io/v1alpha2" because of same keys in both.
		if t.Spec.KES.ExternalCertSecret.Type == "kubernetes.io/tls" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1alpha2" || t.Spec.KES.ExternalCertSecret.Type == "cert-manager.io/v1" {
			serverCertPaths = []corev1.KeyToPath{
				{Key: "tls.crt", Path: certPath},
				{Key: "tls.key", Path: keyPath},
			}
		}
	} else {
		serverCertSecret = t.KESTLSSecretName()
	}

	if t.KESClientCert() {
		clientCertSecret = t.Spec.KES.ClientCertSecret.Name
	}

	if t.Spec.KES.Configuration.Name != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: t.Spec.KES.Configuration.Name,
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
			Name: t.KESVolMountName(),
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: volumeProjections,
				},
			},
		},
	}

	if t.HasGCPCredentialSecretForKES() {
		var gcpVolumeDefaultMode int32 = 420
		var isOptional bool
		podVolumes = append(podVolumes, corev1.Volume{
			Name: gcpCredentialVolumeMountName,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: &gcpVolumeDefaultMode,
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: t.Spec.KES.GCPCredentialSecretName,
								},
								Items: []corev1.KeyToPath{
									{Key: "config", Path: gcpAppCredentialsPath},
								},
								Optional: &isOptional,
							},
						},
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Audience:          t.Spec.KES.GCPWorkloadIdentityPool,
								ExpirationSeconds: &defaultServiceAccountTokenExpiryInSecs,
								Path:              serviceAccountTokenPath,
							},
						},
					},
				},
			},
		})
	}

	containers := []corev1.Container{KESServerContainer(t)}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       t.Namespace,
			Name:            t.KESStatefulSetName(),
			OwnerReferences: t.OwnerRef(),
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov2.DefaultUpdateStrategy,
			},
			PodManagementPolicy: t.Spec.PodManagementPolicy,
			// KES is always matched via Tenant Name + KES prefix
			Selector:    KESSelector(t),
			ServiceName: serviceName,
			Replicas:    &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: KESMetadata(t),
				Spec: corev1.PodSpec{
					ServiceAccountName:        t.Spec.KES.ServiceAccountName,
					Containers:                containers,
					Volumes:                   podVolumes,
					RestartPolicy:             corev1.RestartPolicyAlways,
					SchedulerName:             t.Scheduler.Name,
					NodeSelector:              t.Spec.KES.NodeSelector,
					Tolerations:               t.Spec.KES.Tolerations,
					Affinity:                  t.Spec.KES.Affinity,
					TopologySpreadConstraints: t.Spec.KES.TopologySpreadConstraints,
					SecurityContext:           kesSecurityContext(t),
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
