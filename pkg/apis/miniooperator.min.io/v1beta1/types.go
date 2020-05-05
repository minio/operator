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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=minioinstance,singular=minioinstance

// MinIOInstance is a specification for a MinIO resource
type MinIOInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scheduler MinIOInstanceScheduler `json:"scheduler,omitempty"`
	Spec      MinIOInstanceSpec      `json:"spec"`
	Status    MinIOInstanceStatus    `json:"status"`
}

// MinIOInstanceScheduler is the spec for a MinIOInstance scheduler
type MinIOInstanceScheduler struct {
	// SchedulerName defines the name of scheduler to be used to schedule MinIOInstance pods
	Name string `json:"name"`
}

// MinIOInstanceSpec is the spec for a MinIOInstance resource
type MinIOInstanceSpec struct {
	// Image defines the MinIOInstance Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecret defines the secret to be used for pull image from a private Docker image.
	// +optional
	ImagePullSecret corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// Pod Management Policy for pod created by StatefulSet
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// Metadata defines the object metadata passed to each pod that is a part of this MinIOInstance
	Metadata *metav1.ObjectMeta `json:"metadata,omitempty"`
	// If provided, use this secret as the credentials for MinIOInstance resource
	// Otherwise MinIO server creates dynamic credentials printed on MinIO server startup banner
	// +optional
	CredsSecret *corev1.LocalObjectReference `json:"credsSecret,omitempty"`
	// If provided, use these environment variables for MinIOInstance resource
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// VolumesPerServer allows a user to specify how many volumes per MinIO Server/Pod instance
	// +optional
	VolumesPerServer int `json:"volumesPerServer"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a MinIOInstance
	// +optional
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key for group replication SSL.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// Mount path for MinIO volume (PV). Defaults to /export
	// +optional
	Mountpath string `json:"mountPath,omitempty"`
	// Subpath inside mount path. This is the directory where MinIO stores data. Default to "" (empty)
	// +optional
	Subpath string `json:"subPath,omitempty"`
	// Liveness Probe for container liveness. Container will be restarted if the probe fails.
	// +optional
	Liveness *corev1.Probe `json:"liveness,omitempty"`
	// Readiness Probe for container readiness. Container will be removed from service endpoints if the probe fails.
	// +optional
	Readiness *corev1.Probe `json:"readiness,omitempty"`
	// RequestAutoCert allows user to enable Kubernetes based TLS cert generation and signing as explained here:
	// https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/
	// +optional
	RequestAutoCert bool `json:"requestAutoCert,omitempty"`
	// CertConfig allows users to set entries like CommonName, Organization, etc for the certificate
	// +optional
	CertConfig *CertificateConfig `json:"certConfig,omitempty"`
	// Tolerations allows users to set entries like effect, key, operator, value.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Security Context allows user to set entries like runAsUser, privlege escalation etc.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// Definition for Cluster in given MinIO cluster
	// +optional
	Zones []Zone `json:"zones"`
}

// MinIOInstanceStatus is the status for a MinIOInstance resource
type MinIOInstanceStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// CertificateConfig is a specification for certificate contents
type CertificateConfig struct {
	CommonName       string   `json:"commonName,omitempty"`
	OrganizationName []string `json:"organizationName,omitempty"`
	DNSNames         []string `json:"dnsNames,omitempty"`
}

// LocalCertificateReference defines the spec for a local certificate
type LocalCertificateReference struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// Zone defines the spec for a MinIO Zone
type Zone struct {
	Name    string `json:"name"`
	Servers int32  `json:"servers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinIOInstanceList is a list of MinIOInstance resources
type MinIOInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MinIOInstance `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=mirrorinstance,singular=mirrorinstance

// MirrorInstance is an instance of mc mirror
// Refer: https://docs.minio.io/docs/minio-client-complete-guide.html#mirror
type MirrorInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec MirrorInstanceSpec `json:"spec"`
}

// MirrorInstanceSpec defines the specification for a MinIOInstance backup. This includes the source and
// target where the backup should be stored. Note that both source and target are expected to be
// AWS S3 API compliant services.
type MirrorInstanceSpec struct {
	// Image defines the mc Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecret defines the secret to be used for pull image from a private Docker image.
	// +optional
	ImagePullSecret corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// Env is used to add alias (sourceAlias and targetAlias MinIO servers) to mc.
	Env []corev1.EnvVar `json:"env"`
	// Args allows configuring fields
	Args Args `json:"args"`
	// If provided, use these requests and limit for cpu/memory resource allocation
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Metadata defines the object metadata passed to each pod that is a part of this MinIOInstance
	Metadata *metav1.ObjectMeta `json:"metadata,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MirrorInstanceList is a list of MirrorInstance resources
type MirrorInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MirrorInstance `json:"items"`
}

// Args specifies configuration for mirror jobs
type Args struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Flags  []string `json:"flags"`
}
