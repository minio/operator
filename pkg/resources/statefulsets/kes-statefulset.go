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
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/minio/operator/pkg/certs"

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
	args := []string{"server", "--config=" + miniov2.KESConfigMountPath + "/server-config.yaml"}

	kesVersion, _ := getKesConfigVersion(t.Spec.KES.Image)
	// Add `--auth` flag only on config versions that are still compatible with it (v1 and v2).
	// Starting KES 2023-11-09T17-35-47Z (v3) is no longer supported.
	switch kesVersion {
	case KesConfigVersion1, KesConfigVersion2:
		args = append(args, "--auth=off")
	}

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
		SecurityContext: kesContainerSecurityContext(t),
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

// Builds the security context for kes containers
func kesContainerSecurityContext(t *miniov2.Tenant) *corev1.SecurityContext {
	// Default values:
	// By default, values should be totally empty if not provided
	// This is specially needed in OpenShift where Security Context Constraints restrict them
	// if let empty then OCP can pick the values from the constraints defined.
	containerSecurityContext := corev1.SecurityContext{}
	runAsNonRoot := true
	var runAsUser int64 = 1000
	var runAsGroup int64 = 1000
	poolSCSet := false

	// Values from pool.SecurityContext ONLY if provided
	if t.Spec.KES != nil && t.Spec.KES.SecurityContext != nil {
		if t.Spec.KES.SecurityContext.RunAsNonRoot != nil {
			runAsNonRoot = *t.Spec.KES.SecurityContext.RunAsNonRoot
			poolSCSet = true
		}
		if t.Spec.KES.SecurityContext.RunAsUser != nil {
			runAsUser = *t.Spec.KES.SecurityContext.RunAsUser
			poolSCSet = true
		}
		if t.Spec.KES.SecurityContext.RunAsGroup != nil {
			runAsGroup = *t.Spec.KES.SecurityContext.RunAsGroup
			poolSCSet = true
		}
		if poolSCSet {
			// Only set values if one of above is set otherwise let it empty
			containerSecurityContext = corev1.SecurityContext{
				RunAsNonRoot: &runAsNonRoot,
				RunAsUser:    &runAsUser,
				RunAsGroup:   &runAsGroup,
			}
		}
	}

	// Values from kes.ContainerSecurityContext if provided
	if t.Spec.KES != nil && t.Spec.KES.ContainerSecurityContext != nil {
		containerSecurityContext = *t.Spec.KES.ContainerSecurityContext
	}
	return &containerSecurityContext
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
		{Key: certs.PublicCertFile, Path: certPath},
		{Key: certs.PrivateKeyFile, Path: keyPath},
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
				{Key: certs.TLSCertFile, Path: certPath},
				{Key: certs.TLSKeyFile, Path: keyPath},
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

const (
	// imageTagWithArchRegex is a regular expression to identify if a KES tag
	// includes the arch as suffix, ie: 2023-05-02T22-48-10Z-arm64
	kesImageTagWithArchRegexPattern = `(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z)(-.*)`
)

const (
	// KesConfigVersion1 identifier v1
	KesConfigVersion1 = "v1"
	// KesConfigVersion2 identifier v2
	KesConfigVersion2 = "v2"
)

func getKesConfigVersion(image string) (string, error) {
	version := KesConfigVersion2

	imageStrings := strings.Split(image, ":")
	var imageTag string
	if len(imageStrings) > 1 {
		imageTag = imageStrings[1]
	} else {
		return "", fmt.Errorf("%s not a valid KES release tag", image)
	}

	if imageTag == "edge" {
		return KesConfigVersion2, nil
	}

	if imageTag == "latest" {
		return KesConfigVersion2, nil
	}

	// When the image tag is semantic version is config v1
	if semver.IsValid(imageTag) {
		// Admin is required starting version v0.22.0
		if semver.Compare(imageTag, "v0.22.0") < 0 {
			return KesConfigVersion1, nil
		}
		return KesConfigVersion2, nil
	}

	releaseTagNoArch := imageTag

	re := regexp.MustCompile(kesImageTagWithArchRegexPattern)
	// if pattern matches, that means we have a tag with arch
	if matched := re.Match([]byte(imageTag)); matched {
		slicesOfTag := re.FindStringSubmatch(imageTag)
		// here we will remove the arch suffix by assigning the first group in the regex
		releaseTagNoArch = slicesOfTag[1]
	}

	// v0.22.0 is the initial image version for Kes config v2, any time format came after and is v2
	_, err := miniov2.ReleaseTagToReleaseTime(releaseTagNoArch)
	if err != nil {
		// could not parse semversion either, returning error
		return "", fmt.Errorf("could not identify KES version from image TAG: %s", releaseTagNoArch)
	}

	// Leaving this snippet as comment as this will helpful to compare in future config versions
	// kesv2ReleaseTime, _ := miniov2.ReleaseTagToReleaseTime("2023-04-03T16-41-28Z")
	// if imageVersionTime.Before(kesv2ReleaseTime) {
	// 	version = kesConfigVersion2
	// }
	return version, nil
}
