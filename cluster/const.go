// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

const (
	// K8sAPIServer address of the K8s API
	K8sAPIServer = "OPERATOR_K8S_API_SERVER"
	// K8SAPIServerTLSRootCA location of the root CA
	K8SAPIServerTLSRootCA = "OPERATOR_K8S_API_SERVER_TLS_ROOT_CA"
	// MinioImage image used as the default for MinIO
	MinioImage = "OPERATOR_MINIO_IMAGE"
)
