package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Execution is the MinIO Job level execution policy
type Execution string

const (
	// Parallel Run MC Jobs in parallel
	Parallel Execution = "parallel"
	// Sequential Run MC Jobs in sequential mode
	Sequential Execution = "sequential"
)

// FailureStrategy is the failure strategy at MinIO Job level
type FailureStrategy string

const (
	// ContinueOnFailure indicates to MinIO Job to continue execution of following commands even in the case of the
	// failure of a command
	ContinueOnFailure FailureStrategy = "continueOnFailure"

	// StopOnFailure indicates to MinIO Job to stop execution of following commands even in the case of the failure
	// of a command
	StopOnFailure FailureStrategy = "stopOnFailure"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=miniojob,singular=miniojob
// +kubebuilder:printcolumn:name="Tenant",type=string,JSONPath=`.spec.tenant.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.spec.status.phase`
// +kubebuilder:metadata:annotations=operator.min.io/version=v5.0.15

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
	TenantRef TenantRef `json:"tenant"`

	// Execution order of the jobs, either `parallel` or `sequential`.
	// Defaults to `parallel` if not provided.
	// +optional
	// +kubebuilder:default=parallel
	// +kubebuilder:validation:Enum=parallel;sequential;
	Execution Execution `json:"execution"`

	// FailureStrategy is the forward plan in case of the failure of one or more MinioJob pods
	// Either `stopOnFailure` or `continueOnFailure`, defaults to `continueOnFailure`.
	// +optional
	// +kubebuilder:default=continueOnFailure
	// +kubebuilder:validation:Enum=continueOnFailure;stopOnFailure;
	FailureStrategy FailureStrategy `json:"failureStrategy"`

	// *Required* +
	//
	// Commands List of MinioClient commands
	Commands []CommandSpec `json:"commands"`

	// mc job image
	// +optional
	// +kubebuilder:default="minio/mc:latest"
	MCImage string `json:"mcImage,omitempty"`

	// *Optional* +
	//
	// Specify the https://kubernetes.io/docs/tasks/configure-pod-container/security-context/[Security Context] of pods in the pool. The Operator supports only the following pod security fields: +
	//
	// * `fsGroup` +
	//
	// * `fsGroupChangePolicy` +
	//
	// * `runAsGroup` +
	//
	// * `runAsNonRoot` +
	//
	// * `runAsUser` +
	//
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// Specify the https://kubernetes.io/docs/tasks/configure-pod-container/security-context/[Security Context] of containers in the pool. The Operator supports only the following container security fields: +
	//
	// * `runAsGroup` +
	//
	// * `runAsNonRoot` +
	//
	// * `runAsUser` +
	//
	// +optional
	ContainerSecurityContext *corev1.SecurityContext `json:"containerSecurityContext,omitempty"`
}

// CommandSpec (`spec`) defines the configuration of a MinioClient Command.
type CommandSpec struct {
	// *Required* +
	//
	// Operation is the MinioClient Action
	Operation string `json:"op"`

	// Name is the Command Name, optional, required if want to reference it with `DependsOn`
	// +optional
	Name string `json:"name,omitempty"`

	// Args Arguments to pass to the action
	// +optional
	Args map[string]string `json:"args,omitempty"`

	// Command custom command to run
	// +optional
	Command []string `json:"command,omitempty"`

	// DependsOn List of named `command` in this MinioJob that have to be scheduled and executed before this command runs
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// TenantRef Is the reference to the target tenant of the jobs
type TenantRef struct {
	// *Required* +
	Name string `json:"name"`
	// *Required* +
	Namespace string `json:"namespace"`
}

// MinIOJobStatus Status of MinioJob resource
type MinIOJobStatus struct {
	// +optional
	Phase string `json:"phase"`
	// +optional
	CommandsStatus []CommandStatus `json:"commands"`
	// +optional
	Message string `json:"message"`
}

// CommandStatus Status of MinioJob command execution
type CommandStatus struct {
	// +optional
	Name string `json:"name"`
	// *Required* +
	Result string `json:"result"`
	// +optional
	Message string `json:"message"`
}
