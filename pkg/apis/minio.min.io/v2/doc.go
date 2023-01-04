// Copyright (C) 2020, MinIO, Inc.
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

// Package v2 - This page provides a quick automatically generated reference for the MinIO Operator `minio.min.io/v2` CRD. For more complete documentation on the MinIO Operator CRD, see https://min.io/docs/minio/kubernetes/upstream/index.html[MinIO Kubernetes Documentation]. +
//
// The `minio.min.io/v2` API was released with the v4.0.0 MinIO Operator. The MinIO Operator automatically converts existing tenants using the `/v1` API to `/v2`. +
//
// +groupName=minio.min.io
// +versionName=v2
package v2
