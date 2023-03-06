// Copyright (C) 2023, MinIO, Inc.
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

// +k8s:deepcopy-gen=package,register
// go:generate controller-gen crd:trivialVersions=true paths=. output:dir=.

// Package v1alpha1 - The following parameters are specific to the `sts.min.io/v1alpha1` MinIO Policy Binding CRD API
// PolicyBinding is an Authorization mechanism managed by the Minio Operator.
// Using Kubernetes ServiceAccount JSON Web Tokens the binding allow a ServiceAccount to assume temporary IAM credentials.
// For more complete documentation on this object, see the https://docs.min.io/minio/k8s/reference/minio-operator-reference.html#minio-operator-yaml-reference[MinIO Kubernetes Documentation].
// PolicyBinding is added as part of the MinIO Operator v5.0.0. +
// +groupName=sts.min.io
// +versionName=v1alpha1
package v1alpha1
