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
	"strings"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPrometheusJWTExpiry = 100 * 365 * 24 * time.Hour
)

// prometheusMetadata returns the object metadata for Prometheus pods. User
// specified metadata in the spec is also included here.
func prometheusMetadata(t *miniov1.Tenant) metav1.ObjectMeta {
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
func prometheusSelector(t *miniov1.Tenant) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: t.PrometheusPodLabels(),
	}
}

func prometheusEnvVars(t *miniov1.Tenant) []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func prometheusConfigVolumeMount(t *miniov1.Tenant) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      t.PrometheusConfigVolMountName(),
		MountPath: "/etc/prometheus",
	}
}

func prometheusVolumeMounts(t *miniov1.Tenant) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      t.PrometheusStatefulsetName(),
			MountPath: "/prometheus",
		},
		prometheusConfigVolumeMount(t),
	}
}

// prometheusServerContainer returns a container for Prometheus StatefulSet.
func prometheusServerContainer(t *miniov1.Tenant) corev1.Container {
	return corev1.Container{
		Name:  miniov1.PrometheusContainerName,
		Image: miniov1.PrometheusImage,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: miniov1.PrometheusPort,
			},
		},
		ImagePullPolicy: t.Spec.ImagePullPolicy,
		VolumeMounts:    prometheusVolumeMounts(t),
		Env:             prometheusEnvVars(t),
		Resources:       t.Spec.Prometheus.Resources,
	}
}

type globalConfig struct {
	ScrapeInterval     time.Duration `yaml:"scrape_interval"`
	EvaluationInterval time.Duration `yaml:"evaluation_interval"`
}

type staticConfig struct {
	Targets []string `yaml:"targets"`
}

type tlsConfig struct {
	CAFile string `yaml:"ca_file"`
}

type scrapeConfig struct {
	JobName       string         `yaml:"job_name"`
	BearerToken   string         `yaml:"bearer_token"`
	MetricsPath   string         `yaml:"metrics_path"`
	Scheme        string         `yaml:"scheme"`
	TLSConfig     tlsConfig      `yaml:"tls_config"`
	StaticConfigs []staticConfig `yaml:"static_configs"`
}

type prometheusConfig struct {
	Global        globalConfig   `yaml:"global"`
	ScrapeConfigs []scrapeConfig `yaml:"scrape_configs"`
}

func genBearerToken(accessKey, secretKey string) string {
	jwt := jwtgo.NewWithClaims(jwtgo.SigningMethodHS512, jwtgo.StandardClaims{
		ExpiresAt: time.Now().UTC().Add(defaultPrometheusJWTExpiry).Unix(),
		Subject:   accessKey,
		Issuer:    "prometheus",
	})

	token, err := jwt.SignedString([]byte(secretKey))
	if err != nil {
		panic(fmt.Sprintf("jwt key generation: %v", err))
	}

	return token
}

// getMinioPodAddrs returns a list of stable minio pod addresses.
func getMinioPodAddrs(t *miniov1.Tenant) []string {
	targets := []string{}
	for _, pool := range t.Spec.Pools {
		poolName := t.PoolStatefulsetName(&pool)
		for i := 0; i < int(pool.Servers); i++ {
			target := fmt.Sprintf("%s-%d.%s.%s.svc.%s:%d", poolName, i, t.MinIOHLServiceName(), t.Namespace, miniov1.GetClusterDomain(), miniov1.MinIOPort)
			targets = append(targets, target)
		}
	}
	return targets
}

func prometheusInitContainers(t *miniov1.Tenant, accessKey, secretKey string) []corev1.Container {
	bearerToken := genBearerToken(accessKey, secretKey)
	minioTargets := getMinioPodAddrs(t)

	// populate config
	promConfig := prometheusConfig{
		Global: globalConfig{
			ScrapeInterval:     10 * time.Second,
			EvaluationInterval: 30 * time.Second,
		},
		ScrapeConfigs: []scrapeConfig{
			{
				JobName:     "minio",
				BearerToken: bearerToken,
				MetricsPath: "/minio/prometheus/metrics",
				Scheme:      "https",
				TLSConfig: tlsConfig{
					CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				},
				StaticConfigs: []staticConfig{
					{
						Targets: minioTargets,
					},
				},
			},
		},
	}

	d, err := yaml.Marshal(&promConfig)
	if err != nil {
		panic(fmt.Sprintf("error marshaling to yaml: %v", err))
	}

	scriptLines := []string{}
	for i, line := range strings.Split(string(d), "\n") {
		outline := ""
		if i == 0 {
			outline = fmt.Sprintf("echo '%s' > /etc/prometheus/prometheus.yml", line)
		} else {
			outline = fmt.Sprintf("echo '%s' >> /etc/prometheus/prometheus.yml", line)
		}
		scriptLines = append(scriptLines, outline)
	}

	scriptArg := strings.Join(scriptLines, "\n")

	return []corev1.Container{
		{
			Name:            miniov1.PrometheusInitContainerName,
			Image:           miniov1.PrometheusInitContainerImage,
			ImagePullPolicy: t.Spec.ImagePullPolicy,
			VolumeMounts:    []corev1.VolumeMount{prometheusConfigVolumeMount(t)},
			Command:         []string{"/bin/sh"},
			Args:            []string{"-c", scriptArg},
			Resources:       t.Spec.Prometheus.Resources,
		},
	}
}

const prometheusDefaultVolumeSize = 5 * 1024 * 1024 * 1024 // 5GiB

// NewForPrometheus creates a new Prometheus StatefulSet for prometheus metrics
func NewForPrometheus(t *miniov1.Tenant, serviceName string, accessKey, secretKey string) *appsv1.StatefulSet {
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
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	initContainers := prometheusInitContainers(t, accessKey, secretKey)
	containers := []corev1.Container{prometheusServerContainer(t)}
	ss := &appsv1.StatefulSet{
		ObjectMeta: promMeta,
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: miniov1.DefaultUpdateStrategy,
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
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            podVolumes,
					RestartPolicy:      corev1.RestartPolicyAlways,
					SchedulerName:      t.Scheduler.Name,
					NodeSelector:       t.Spec.Prometheus.NodeSelector,
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
