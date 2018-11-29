/*
 * Minio-Operator - Manage Minio clusters in Kubernetes
 *
 * Minio (C) 2018 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioInstance is a specification for a Minio resource
type MinioInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioInstanceSpec   `json:"spec"`
	Status MinioInstanceStatus `json:"status"`
}

// MinioInstanceSpec is the spec for a MinioInstance resource
type MinioInstanceSpec struct {
	// Version defines the MinioInstance Docker image version.
	Version string `json:"version"`
	// Replicas defines the number of Minio instances in a MinioInstance resource
	Replicas int32 `json:"replicas"`
	// If provided, use this secret as the credentials for MinioInstance resource
	// Otherwise Minio server creates dynamic credentials printed on Minio server startup banner
	// +optional
	CredsSecret *corev1.LocalObjectReference `json:"credsSecret,omitempty"`
	// VolumeClaimTemplate allows a user to specify how volumes inside a MinioInstance
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
	// SSLSecret allows a user to specify custom CA certificate, and private key for group replication SSL.
	// +optional
	SSLSecret *corev1.LocalObjectReference `json:"sslSecret,omitempty"`
	// Mount path for Minio volume (PV). Defaults to /export
	// +optional
	Mountpath string `json:"mountPath,omitempty"`
	// Subpath inside mount path. This is the directory where Minio stores data. Default to "" (empty)
	// +optional
	Subpath string `json:"subPath,omitempty"`
}

// MinioInstanceStatus is the status for a MinioInstance resource
type MinioInstanceStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinioInstanceList is a list of MinioInstance resources
type MinioInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MinioInstance `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mirror is a backup of a MinioInstance.
type Mirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   MirrorSpec   `json:"spec"`
	Status MirrorStatus `json:"status"`
}

// MirrorSpec defines the specification for a MinioInstance backup. This includes the source and
// target where the backup should be stored. Note that both source and target are expected to be
// AWS S3 API compliant services.
type MirrorSpec struct {
	// Version defines the Minio Client (mc) Docker image version.
	Version string `json:"version"`
	// SourceEndpoint is the endpoint of Minio instance to backup.
	SourceEndpoint string `json:"srcEndpoint"`
	// SourceCredsSecret as the credentials for source Minio instance.
	SourceCredsSecret *corev1.LocalObjectReference `json:"srcCredsSecret"`
	// SourceBucket defines the bucket on source Minio instance
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
