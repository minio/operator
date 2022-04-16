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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=iampolicy,singular=iampolicy

// Policy is a https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/[Kubernetes object] describing a MinIO IAM Policy. +
//
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// *Required* +
	//
	// The root field for the MinIO IAM Policy object.
	Spec PolicySpec `json:"spec"`
}

type PolicySpec struct {
	Statements []Statement `json:"statements"`
}

// StatementEffect - policy statement effect Allow or Deny.
type StatementEffect string

const (
	// EffectAllow - allow effect.
	EffectAllow StatementEffect = "Allow"

	// EffectDeny - deny effect.
	EffectDeny = "Deny"
)

type Action string

// Resource - resource in policy statement.
type StatementResource struct {
	BucketName string `json:"BucketName"`
	Pattern    string `json:"Pattern"`
}

// Function - condition function interface.
type ConditionClauseKey string

type ConditionKeyName string

type Condition map[ConditionKeyName]string

type Statement struct {
	// +optional
	SID     string          `json:"Sid,omitempty"`
	Effect  StatementEffect `json:"Effect"`
	Actions []Action        `json:"Action"`
	//+optional
	Resources []StatementResource `json:"Resource,omitempty"`
	//+optional
	Conditions map[ConditionClauseKey]Condition `json:"Condition,omitempty"`
}

type PolicyStatus struct {
	CurrentState string `json:"currentState"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList is a list of Tenant resources
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Policy `json:"items"`
}
