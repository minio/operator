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

// Package v1alpha1 - The following parameters are specific to the `job.min.io/v1alpha1` MinIOJob CRD API
// MinIOJob is an automated InfrastructureAsCode integrated with Minio Operator STS to configure MinIO Tenants.
// +groupName=job.min.io
// +versionName=v1alpha1
package v1alpha1
