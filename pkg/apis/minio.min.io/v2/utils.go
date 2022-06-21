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
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

const (
	// Maximum length for MinIO access key.
	// There is no max length enforcement for access keys
	accessKeyMaxLen = 20

	// Maximum secret key length for MinIO, this
	// is used when auto generating new credentials.
	// There is no max length enforcement for secret keys
	secretKeyMaxLen = 40

	// Alpha numeric table used for generating access keys.
	alphaNumericTable = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Alpha numeric table used for generating secret keys.
	alphaNumericTableFull = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	// Total length of the alphanumeric table.
	alphaNumericTableLen = byte(len(alphaNumericTable))

	// Total length of the full alphanumeric table.
	alphaNumericTableFullLen = byte(len(alphaNumericTableFull))
)

// GenerateCredentials - creates randomly generated credentials of maximum allowed length.
func GenerateCredentials() (accessKey, secretKey string, err error) {
	readBytes := func(size int) (data []byte, err error) {
		data = make([]byte, size)
		var n int
		if n, err = io.ReadFull(rand.Reader, data); err != nil {
			return nil, err
		} else if n != size {
			return nil, fmt.Errorf("Not enough data. Expected to read: %v bytes, got: %v bytes", size, n)
		}
		return data, nil
	}

	// Generate access key.
	keyBytes, err := readBytes(accessKeyMaxLen)
	if err != nil {
		return "", "", err
	}
	for i := 0; i < accessKeyMaxLen; i++ {
		keyBytes[i] = alphaNumericTable[keyBytes[i]%alphaNumericTableLen]
	}
	accessKey = string(keyBytes)

	// Generate secret key.
	keyBytes, err = readBytes(secretKeyMaxLen)
	if err != nil {
		return "", "", err
	}

	for i := 0; i < secretKeyMaxLen; i++ {
		keyBytes[i] = alphaNumericTableFull[keyBytes[i]%alphaNumericTableFullLen]
	}
	secretKey = string(keyBytes)

	return accessKey, secretKey, nil
}

// GenerateTenantConfigurationFile :
func GenerateTenantConfigurationFile(configuration map[string]string) string {
	var rawConfiguration strings.Builder
	for key, val := range configuration {
		rawConfiguration.WriteString(fmt.Sprintf(`export %s="%s"`, key, val) + "\n")
	}
	return rawConfiguration.String()
}
