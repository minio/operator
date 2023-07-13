// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package utils

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/minio/operator/pkg/common"

	"github.com/google/uuid"
)

// NewUUID - get a random UUID.
func NewUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// DecodeBase64 : decoded base64 input into utf-8 text
func DecodeBase64(s string) (string, error) {
	decodedInput, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decodedInput), nil
}

// Key used for Get/SetReqInfo
type key string

// context keys
const (
	ContextLogKey            = key("console-log")
	ContextRequestID         = key("request-id")
	ContextRequestUserID     = key("request-user-id")
	ContextRequestUserAgent  = key("request-user-agent")
	ContextRequestHost       = key("request-host")
	ContextRequestRemoteAddr = key("request-remote-addr")
	ContextAuditKey          = key("request-audit-entry")
)

// GetOperatorRuntime Retrieves the runtime from env variable
func GetOperatorRuntime() common.Runtime {
	envString := os.Getenv(common.OperatorRuntimeEnv)
	runtimeReturn := common.OperatorRuntimeK8s
	if envString != "" {
		envString = strings.TrimSpace(envString)
		envString = strings.ToUpper(envString)
		if val, ok := common.Runtimes[envString]; ok {
			runtimeReturn = val
		}
	}
	return runtimeReturn
}
