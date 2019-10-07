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

// CertificateConfig is a specification for certificate contents
type CertificateConfig struct {
	CommonName       string   `json:"commonName,omitempty"`
	OrganizationName []string `json:"organizationName,omitempty"`
	DNSNames         []string `json:"dnsNames,omitempty"`
}

// MinIOInstanceSpec is the spec for a MinIOInstance resource
type MinIOInstanceSpec struct {
	// Image defines the MinIOInstance Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// Replicas defines the number of MinIO instances in a MinIOInstance resource
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
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
	// VolumeClaimTemplate allows a user to specify how volumes inside a MinIOInstance
	// +optional
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
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
	ExternalCertSecret *corev1.LocalObjectReference `json:"externalCertSecret,omitempty"`
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
}

// MinIOInstanceStatus is the status for a MinIOInstance resource
type MinIOInstanceStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
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

// Mirror is a backup of a MinIOInstance.
type Mirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   MirrorSpec   `json:"spec"`
	Status MirrorStatus `json:"status"`
}

// MirrorSpec defines the specification for a MinIOInstance backup. This includes the source and
// target where the backup should be stored. Note that both source and target are expected to be
// AWS S3 API compliant services.
type MirrorSpec struct {
	// Version defines the MinIO Client (mc) Docker image version.
	Version string `json:"version"`
	// SourceEndpoint is the endpoint of MinIO instance to backup.
	SourceEndpoint string `json:"srcEndpoint"`
	// SourceCredsSecret as the credentials for source MinIO instance.
	SourceCredsSecret *corev1.LocalObjectReference `json:"srcCredsSecret"`
	// SourceBucket defines the bucket on source MinIO instance
	// +optional
	SourceBucket string `json:"srcBucket,omitempty"`
	// Region in which the source S3 compatible bucket is located.
	// uses "us-east-1" by default
	// +optional
	SourceRegion string `json:"srcRegion"`
	// Endpoint (hostname only or fully qualified URI) of S3 compatible
	// storage service.
	TargetEndpoint string `json:"targetEndpoint"`
	// CredentialsSecret is a reference to the Secret containing the
	// credentials authenticating with the S3 compatible storage service.
	TargetCredsSecret *corev1.LocalObjectReference `json:"targetCredsSecret"`
	// Bucket in which to store the Backup.
	TargetBucket string `json:"targetBucket"`
	// Region in which the Target S3 compatible bucket is located.
	// uses "us-east-1" by default
	// +optional
	TargetRegion string `json:"targetRegion"`
}

// MirrorStatus captures the current status of a Mirror operation.
type MirrorStatus struct {
	// Outcome holds the results of a Mirror operation.
	// +optional
	Outcome string `json:"outcome"`
	// TimeStarted is the time at which the backup was started.
	// +optional
	TimeStarted metav1.Time `json:"timeStarted"`
	// TimeCompleted is the time at which the backup completed.
	// +optional
	TimeCompleted metav1.Time `json:"timeCompleted"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MirrorList is a list of Backups.
type MirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Mirror `json:"items"`
}
