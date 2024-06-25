// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package common

import (
	"fmt"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/statefulsets"
	"log"
	"os"
	"strings"
)

func AttachGeneratedConfig(tenant *miniov2.Tenant, fileContents string) string {
	args, err := GetTenantArgs(tenant)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fileContents = fileContents + fmt.Sprintf("export MINIO_ARGS=\"%s\"\n", args)

	return fileContents
}

// GetTenantArgs returns the arguments for the tenant based on the tenants they have
func GetTenantArgs(tenant *miniov2.Tenant) (string, error) {
	if tenant == nil {
		return "", fmt.Errorf("tenant is nil")
	}
	// Validate the MinIO Tenant
	if err := tenant.Validate(); err != nil {
		log.Println(err)
		return "", err
	}
	args := strings.Join(statefulsets.GetContainerArgs(tenant, ""), " ")
	return args, nil
}
