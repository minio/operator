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

package subnet

import (
	"errors"
	"log"

	"github.com/minio/pkg/licverifier"
)

// GetLicenseInfoFromJWT will return license metadata from a jwt string license
func GetLicenseInfoFromJWT(license string, publicKeys []string) (*licverifier.LicenseInfo, error) {
	if license == "" {
		return nil, errors.New("license is not present")
	}
	for _, publicKey := range publicKeys {
		lv, err := licverifier.NewLicenseVerifier([]byte(publicKey))
		if err != nil {
			log.Print(err)
			continue
		}
		licInfo, err := lv.Verify(license)
		if err != nil {
			log.Print(err)
			continue
		}
		return &licInfo, nil
	}
	return nil, errors.New("invalid license key")
}

// MfaReq - JSON payload of the SUBNET mfa api
type MfaReq struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
	Token    string `json:"token"`
}

// LoginResp response for MFA login attempt
type LoginResp struct {
	AccessToken string
	MfaToken    string
}
