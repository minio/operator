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

package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=tenant,singular=tenant
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.currentState"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion

// Tenant is a specification for a MinIO resource
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scheduler TenantScheduler `json:"scheduler,omitempty"`
	Spec      TenantSpec      `json:"spec"`
	// Status provides details of the state of the Tenant
	// +optional
	Status TenantStatus `json:"status"`
}

// TenantScheduler is the spec for a Tenant scheduler
type TenantScheduler struct {
	// SchedulerName defines the name of scheduler to be used to schedule Tenant pods
	Name string `json:"name"`
}

// S3Features list of S3 features to enable/disable.
// Currently only supports BucketDNS
type S3Features struct {
	// BucketDNS if 'true' means Buckets can be accessed using `<bucket>.minio.default.svc.cluster.local`
	BucketDNS bool `json:"bucketDNS"`
}

// TenantSpec is the spec for a Tenant resource
type TenantSpec struct {
	// Definition for Cluster in given MinIO cluster
	Pools []Pool `json:"pools"`
	// Image defines the Tenant Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecret defines the secret to be used for pull image from a private Docker image.
	// +optional
	ImagePullSecret corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// Pod Management Policy for pod created by StatefulSet
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// If provided, use this secret as the credentials for Tenant resource
	// Otherwise MinIO server creates dynamic credentials printed on MinIO server startup banner
	// +optional
	CredsSecret *corev1.LocalObjectReference `json:"credsSecret,omitempty"`
	// If provided, use these environment variables for Tenant resource
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// ExternalCertSecret allows a user to provide one or more TLS certificates and private keys. This is
	// used for enabling TLS with SNI support on MinIO server.
	// +optional
	ExternalCertSecret []*LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// ExternalCaCertSecret allows a user to provide additional CA certificates. This is
	// used for MinIO to verify TLS connections with other applications.
	// +optional
	ExternalCaCertSecret []*LocalCertificateReference `json:"externalCaCertSecret,omitempty"`
	// ExternalClientCertSecret allows a user to specify custom CA client certificate, and private key. This is
	// used for adding client certificates on MinIO Pods --> used for KES authentication.
	// +optional
	ExternalClientCertSecret *LocalCertificateReference `json:"externalClientCertSecret,omitempty"`
	// Mount path for MinIO volume (PV). Defaults to /export
	// +optional
	Mountpath string `json:"mountPath,omitempty"`
	// Subpath inside mount path. This is the directory where MinIO stores data. Default to "" (empty)
	// +optional
	Subpath string `json:"subPath,omitempty"`
	// RequestAutoCert allows user to enable Kubernetes based TLS cert generation and signing as explained here:
	// https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/
	// +optional
	RequestAutoCert *bool `json:"requestAutoCert,omitempty"`
	// S3 related features can be disabled or enabled such as `bucketDNS` etc.
	S3 *S3Features `json:"s3,omitempty"`
	// +optional
	// CertConfig allows users to set entries like CommonName, Organization, etc for the certificate
	// +optional
	CertConfig *CertificateConfig `json:"certConfig,omitempty"`
	// Security Context allows user to set entries like runAsUser, privilege escalation etc.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// ConsoleConfiguration is for setting up minio/console for graphical user interface
	//+optional
	Console *ConsoleConfiguration `json:"console,omitempty"`
	// KES is for setting up minio/kes as MinIO KMS
	//+optional
	KES *KESConfig `json:"kes,omitempty"`
	// Log is for setting up log search
	//+optional
	Log *LogConfig `json:"log,omitempty"`
	// Prometheus is for setting up Prometheus metrics.
	//+optional
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run pods of all MinIO
	// Pods created as a part of this Tenant.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// PriorityClassName indicates the Pod priority and hence importance of a Pod relative to other Pods.
	// This is applied to MinIO pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to MinIO pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// SideCars a list of containers to run as sidecars along every MinIO Pod on every pool
	// +optional
	SideCars *SideCars `json:"sideCars,omitempty"`
	// ExposeServices tells operator whether to expose the MinIO service and/or the Console Service
	// +optional
	ExposeServices *ExposeServices `json:"exposeServices,omitempty"`
	// ServiceMetadata provides a way to bring your own service to be used for MinIO and / or Console
	// +optional
	ServiceMetadata *ServiceMetadata `json:"serviceMetadata,omitempty"`
}

// ServiceMetadata provides a way to bring your own service to be used for MinIO and / or Console
type ServiceMetadata struct {
	// If provided, append these labels to the MinIO service
	// +optional
	MinIOServiceLabels map[string]string `json:"minioServiceLabels,omitempty"`
	// If provided, append these annotations to the MinIO service
	// +optional
	MinIOServiceAnnotations map[string]string `json:"minioServiceAnnotations,omitempty"`
	// If provided, append these labels to the Console service
	// +optional
	ConsoleServiceLabels map[string]string `json:"consoleServiceLabels,omitempty"`
	// If provided, append these annotations to the Console service
	// +optional
	ConsoleServiceAnnotations map[string]string `json:"consoleServiceAnnotations,omitempty"`
}

// LocalCertificateReference defines the spec for a local certificate
type LocalCertificateReference struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// ExposeServices tells operator whether to expose the services for MinIO and Console
type ExposeServices struct {
	// MinIO tells operator whether to expose the MinIO service
	// +optional
	MinIO bool `json:"minio,omitempty"`
	// Console tells operator whether to expose the Console Service
	// +optional
	Console bool `json:"console,omitempty"`
}

// TenantStatus is the status for a Tenant resource
type TenantStatus struct {
	CurrentState      string `json:"currentState"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Revision          int32  `json:"revision"`
}

// CertificateConfig is a specification for certificate contents
type CertificateConfig struct {
	CommonName       string   `json:"commonName,omitempty"`
	OrganizationName []string `json:"organizationName,omitempty"`
	DNSNames         []string `json:"dnsNames,omitempty"`
}

// Pool defines the spec for a MinIO Pool
type Pool struct {
	// Name of the pool
	// +optional
	Name string `json:"name,omitempty"`
	// Number of Servers in the pool
	Servers int32 `json:"servers"`
	// Number of persistent volumes that will be attached per server
	VolumesPerServer int32 `json:"volumesPerServer"`
	// VolumeClaimTemplate allows a user to specify how volumes are configured for the Pool
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations allows users to set entries like effect, key, operator, value.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// ConsoleConfiguration defines the specifications for Console Deployment
type ConsoleConfiguration struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the Tenant Console Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to MinIO Console pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// This secret provides all environment variables for KES
	// This is a mandatory field
	ConsoleSecret *corev1.LocalObjectReference `json:"consoleSecret"`
	// ServiceAccountName is the name of the ServiceAccount to use to run pods of all Console
	// Pods created as a part of this Tenant.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// If provided, use these environment variables for Console resource
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// ExternalCertSecret allows a user to provide an external certificate and private key. This is
	// used for enabling TLS on Console and has priority over AutoCert.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// ExternalCaCertSecret allows a user to provide additional CA certificates. This is
	// used for Console to verify TLS connections with other applications.
	// +optional
	ExternalCaCertSecret []*LocalCertificateReference `json:"externalCaCertSecret,omitempty"`
	// If provided, use these annotations for Console Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// If provided, use these labels for Console Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// If provided, use these nodeSelector for Console Object Meta nodeSelector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// EqualImage returns true if config image and current input image are same
func (c ConsoleConfiguration) EqualImage(currentImage string) bool {
	return c.Image == currentImage
}

// EqualImage returns true if image specified in `LogConfig` is equal to `image`
func (lc *LogConfig) EqualImage(image string) bool {
	if lc == nil {
		return false
	}
	return lc.Image == image
}

// LogConfig defines configuration parameters for Log feature
type LogConfig struct {
	// Image defines the tenant's LogSearchAPI container image.
	// +optional
	Image string `json:"image,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations allows users to set entries like effect, key, operator, value.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// If provided, use these annotations for Console Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// If provided, use these labels for Console Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Db holds configuration for audit logs DB
	// +optional
	Db *LogDbConfig `json:"db,omitempty"`
	// AuditConfig holds configuration for audit logs from MinIO
	// +optional
	Audit *AuditConfig `json:"audit,omitempty"`
}

// AuditConfig defines configuration parameters for Audit (type) logs
type AuditConfig struct {
	// DiskCapacityGB defines the disk capacity in GB available to store audit logs
	// +optional
	DiskCapacityGB *int `json:"diskCapacityGB,omitempty"`
}

// PrometheusConfig defines configuration for Prometheus metrics server
type PrometheusConfig struct {
	// DiskCapacityGB defines the disk capacity in GB available to the
	// Prometheus server
	// +optional
	DiskCapacityDB *int `json:"diskCapacityGB,omitempty"`
	// If provided, use these annotations for Prometheus Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// If provided, use these labels for Prometheus Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// If provided, use these nodeSelector for Prometheus Object Meta nodeSelector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// LogDbConfig Holds all the configurations regarding the Log DB (Postgres) StatefulSet
type LogDbConfig struct {
	// Image defines postgres DB container image.
	// +optional
	Image string `json:"image,omitempty"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a Tenant
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations allows users to set entries like effect, key, operator, value.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// If provided, use these annotations for Console Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// If provided, use these labels for Console Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// KESConfig defines the specifications for KES StatefulSet
type KESConfig struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the Tenant KES Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// This is applied to KES pods only.
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run pods of all KES
	// Pods created as a part of this Tenant.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// This kesSecret serves as the configuration for KES
	// This is a mandatory field
	Configuration *corev1.LocalObjectReference `json:"kesSecret"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key for group replication SSL.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// ClientCertSecret allows a user to specify a custom root certificate, client certificate and client private key. This is
	// used for adding client certificates on KES --> used for KES authentication against Vault or other KMS that supports mTLS.
	// +optional
	ClientCertSecret *LocalCertificateReference `json:"clientCertSecret,omitempty"`
	// If provided, use these annotations for KES Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// If provided, use these labels for KES Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// If provided, use these nodeSelector for KES Object Meta nodeSelector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TenantList is a list of Tenant resources
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Tenant `json:"items"`
}

// SideCars represents a list of containers that will be attached to the MinIO pods on each pool
type SideCars struct {
	// List of containers to run inside the Pod
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []corev1.Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// TODO: Define the behavior if a claim already exists with the same name.
	// +optional
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
}
