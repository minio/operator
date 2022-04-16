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

// +k8s:deepcopy-gen=package,register
// go:generate controller-gen crd:trivialVersions=true paths=. output:dir=.

// Package v1 - This page provides a quick automatically generated reference for the MinIO Operator Policy `policy.min.io/v1` CRD. For more complete documentation on the MinIO Policy CRD, see https://docs.min.io/minio/k8s/reference/minio-operator-reference[MinIO Kubernetes Documentation]. +
//
// The `policy.min.io/v1` API was released with the v5.0.0 MinIO Operator.  +
//
//
// +groupName=policy.min.io
// +versionName=v1
package v1
