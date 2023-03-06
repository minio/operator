// Copyright (C) 2022, MinIO, Inc.
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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=policybinding,singular=policybinding
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.currentState"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion

// PolicyBinding is a https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/[Kubernetes object] describing a MinIO PolicyBinding.
type PolicyBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// *Required* +
	//
	// The root field for the MinIO PolicyBinding object.
	Spec PolicyBindingSpec `json:"spec,omitempty"`

	// Status provides details of the state of the PolicyBinding
	// +optional
	Status PolicyBindingStatus `json:"status,omitempty"`
}

// PolicyBindingStatus is the status for a PolicyBinding resource
type PolicyBindingStatus struct {
	// *Required* +
	CurrentState string `json:"currentState"`

	// Keeps track of the invocations related to the PolicyBinding
	// +nullable
	Usage PolicyBindingUsage `json:"usage"`
}

// PolicyBindingUsage are metrics regarding the usage of the policyBinding
type PolicyBindingUsage struct {
	Authorizations int64 `json:"authotizations,omitempty"`
}

// PolicyBindingSpec (`spec`) defines the configuration of a MinIO PolicyBinding object. +
type PolicyBindingSpec struct {
	// *Required* +
	//
	// The Application Property identifies the namespace and service account that will be authorized
	Application *Application `json:"application"`
	// *Required* +
	Policies []string `json:"policies"`
}

// Application defines the `Namespace` and `ServiceAccount` to authorize the usage of the policies listed
type Application struct {
	// *Required* +
	Namespace string `json:"namespace"`
	// *Required* +
	ServiceAccount string `json:"serviceaccount"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyBindingList is a list of PolicyBinding resources
type PolicyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []PolicyBinding `json:"items"`
}
