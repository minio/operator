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

package v1

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
	// ServiceName defines name of the Service that will be created for this instance, if none is specified,
	// it will default to the instance name
	// +optional
	ServiceName string `json:"serviceName,omitempty"`
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
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key. This is
	// used for enabling TLS support on MinIO Pods.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
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
	// Liveness Probe for container liveness. Container will be restarted if the probe fails.
	// +optional
	Liveness *Liveness `json:"liveness,omitempty"`
	// Readiness Probe for container readiness. Container will be removed from service endpoints if the probe fails.
	// +optional
	Readiness *Readiness `json:"readiness,omitempty"`
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
	// Security Context allows user to set entries like runAsUser, privilege escalation etc.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// Definition for Cluster in given MinIO cluster
	// +optional
	Zones []Zone `json:"zones"`
	// MCSConfig is for setting up minio/mcs for graphical user interface
	//+optional
	MCS *MCSConfig `json:"mcs,omitempty"`
	// KES is for setting up minio/kes as MinIO KMS
	//+optional
	KES *KESConfig `json:"kes,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run pods of all MinIO
	// Pods created as a part of this MinIOInstance.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// MinIOInstanceStatus is the status for a MinIOInstance resource
type MinIOInstanceStatus struct {
	CurrentState      string `json:"currentState"`
	AvailableReplicas int32  `json:"availableReplicas"`
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

// Liveness specifies the spec for liveness probe
type Liveness struct {
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	PeriodSeconds       int32 `json:"periodSeconds"`
	TimeoutSeconds      int32 `json:"timeoutSeconds"`
}

// Readiness specifies the spec for liveness probe
type Readiness struct {
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	PeriodSeconds       int32 `json:"periodSeconds"`
	TimeoutSeconds      int32 `json:"timeoutSeconds"`
}

// MCSConfig defines the specifications for MCS Deployment
type MCSConfig struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the MinIOInstance Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// This secret provides all environment variables for KES
	// This is a mandatory field
	MCSSecret *corev1.LocalObjectReference `json:"mcsSecret"`
	Metadata  *metav1.ObjectMeta           `json:"metadata,omitempty"`
}

// KESConfig defines the specifications for KES StatefulSet
type KESConfig struct {
	// Replicas defines number of pods for KES StatefulSet.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image defines the MinIOInstance Docker image.
	// +optional
	Image string `json:"image,omitempty"`
	// This kesSecret serves as the configuration for KES
	// This is a mandatory field
	Configuration *corev1.LocalObjectReference `json:"kesSecret"`
	Metadata      *metav1.ObjectMeta           `json:"metadata,omitempty"`
	// ExternalCertSecret allows a user to specify custom CA certificate, and private key for group replication SSL.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinIOInstanceList is a list of MinIOInstance resources
type MinIOInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MinIOInstance `json:"items"`
}
