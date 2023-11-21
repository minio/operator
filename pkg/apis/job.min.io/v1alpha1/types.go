package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=miniojob,singular=miniojob
// +kubebuilder:printcolumn:name="Tenant",type=string,JSONPath=`.spec.tenant.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.spec.status.phase`

// MinIOJob is a top-level type. A client is created for it
type MinIOJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// *Required* +
	//
	// The root field for the MinIOJob object.
	Spec MinIOJobSpec `json:"spec,omitempty"`

	// Status provides details of the state of the MinIOJob steps
	// +optional
	Status MinIOJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MinIOJobList is a top-level list type.
type MinIOJobList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MinIOJob `json:"items"`
}

// MinIOJobSpec (`spec`) defines the configuration of a MinIOJob object. +
type MinIOJobSpec struct {
	// *Required* +
	//
	// Service Account name for the jobs to run
	ServiceAccountName string `json:"serviceAccountName"`

	// *Required* +
	//
	// TenantRef Reference for minio Tenant to eun the jobs against
	TenantRef MinioJobTenantRef `json:"tenant"`

	// Execution order of the jobs, either `parallel` or `sequential`.
	// Defaults to `parallel` if not provided.
	// +optional
	Execution string `json:"execution"`

	// FailureStrategy is the forward plan in case of the failure of one or more MinioJob pods
	// Either `stopOnFailure` or `continueOnFailure`, defaults to `continueOnFailure`.
	// +optional
	FailureStrategy string `json:"failureStrategy"`

	// *Required* +
	//
	// Commands List of MinioClient commands
	Commands []CommandSpec `json:"commands"`
}

type CommandSpec struct {
	// *Required* +
	//
	// Operation is the MinioClient Action
	Operation string `json:"op"`

	// Args Arguments to pass to the action
	// +optional
	Args []string `json:"args,omitempty"`

	// DependsOn List of named `command` in this MinioJob that have to be scheduled and executed before this command runs
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`
}

type MinioJobTenantRef struct {
	// *Required* +
	Name string `json:"name"`
	// *Required* +
	Namespace string `json:"namespace"`
}

type MinIOJobStatus struct {
	// *Required* +
	Phase string `json:"phase"`
	// *Required* +
	Commands []CommandStatus `json:"commands"`
}

type CommandStatus struct {
	// +optional
	Name string `json:"name"`
	// *Required* +
	Result string `json:"result"`
	// +optional
	Message string `json:"message"`
}
