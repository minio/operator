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
//

package subnet

import (
	"errors"

	"github.com/minio/madmin-go/v2"
	mc "github.com/minio/mc/cmd"

	"github.com/minio/operator/pkg/http"

	"github.com/tidwall/gjson"
)

// LoginWithMFA starts a login request with MFA
func LoginWithMFA(client http.ClientI, username, mfaToken, otp string) (*LoginResp, error) {
	mfaLoginReq := MfaReq{Username: username, OTP: otp, Token: mfaToken}
	resp, err := subnetPostReq(client, subnetMFAURL(), mfaLoginReq, nil)
	if err != nil {
		return nil, err
	}
	token := gjson.Get(resp, "token_info.access_token")
	if token.Exists() {
		return &LoginResp{AccessToken: token.String(), MfaToken: ""}, nil
	}
	return nil, errors.New("access token not found in response")
}

// Login starts a login request
func Login(client http.ClientI, username, password string) (*LoginResp, error) {
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}
	respStr, err := subnetPostReq(client, subnetLoginURL(), loginReq, nil)
	if err != nil {
		return nil, err
	}
	mfaRequired := gjson.Get(respStr, "mfa_required").Bool()
	if mfaRequired {
		mfaToken := gjson.Get(respStr, "mfa_token").String()
		if mfaToken == "" {
			return nil, errors.New("missing mfa token")
		}
		return &LoginResp{AccessToken: "", MfaToken: mfaToken}, nil
	}
	token := gjson.Get(respStr, "token_info.access_token")
	if token.Exists() {
		return &LoginResp{AccessToken: token.String(), MfaToken: ""}, nil
	}
	return nil, errors.New("access token not found in response")
}

// LicenseTokenConfig holds registration cfg
type LicenseTokenConfig struct {
	APIKey  string
	License string
	Proxy   string
}

// Register starts a registration flow
func Register(client http.ClientI, admInfo madmin.InfoMessage, apiKey, token, accountID string) (*LicenseTokenConfig, error) {
	var headers map[string]string
	regInfo := GetClusterRegInfo(admInfo)
	regURL := subnetRegisterURL()
	if apiKey != "" {
		regURL += "?api_key=" + apiKey
	} else {
		if accountID == "" || token == "" {
			return nil, errors.New("missing accountID or authentication token")
		}
		headers = subnetAuthHeaders(token)
		regURL += "?aid=" + accountID
	}
	regToken, err := GenerateRegToken(regInfo)
	if err != nil {
		return nil, err
	}
	reqPayload := mc.ClusterRegistrationReq{Token: regToken}
	resp, err := subnetPostReq(client, regURL, reqPayload, headers)
	if err != nil {
		return nil, err
	}
	respJSON := gjson.Parse(resp)
	subnetAPIKey := respJSON.Get("api_key").String()
	licenseJwt := respJSON.Get("license").String()

	if subnetAPIKey != "" || licenseJwt != "" {
		return &LicenseTokenConfig{
			APIKey:  subnetAPIKey,
			License: licenseJwt,
		}, nil
	}
	return nil, errors.New("subnet api key not found")
}

// GetAPIKey returns the API for a token
func GetAPIKey(client http.ClientI, token string) (string, error) {
	resp, err := subnetGetReq(client, subnetAPIKeyURL(), subnetAuthHeaders(token))
	if err != nil {
		return "", err
	}
	respJSON := gjson.Parse(resp)
	apiKey := respJSON.Get("api_key").String()
	return apiKey, nil
}
