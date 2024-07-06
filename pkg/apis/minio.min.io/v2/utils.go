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

package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// GenerateTenantConfigurationFile :
func GenerateTenantConfigurationFile(configuration map[string]string) string {
	var rawConfiguration strings.Builder
	for key, val := range configuration {
		rawConfiguration.WriteString(fmt.Sprintf(`export %s="%s"`, key, val) + "\n")
	}
	return rawConfiguration.String()
}

// CompactJSONString removes white spaces, tabs and line return
func CompactJSONString(jsonObject string) (string, error) {
	objectByte := []byte(jsonObject)
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, objectByte); err != nil {
		return jsonObject, err
	}
	return buffer.String(), nil
}
