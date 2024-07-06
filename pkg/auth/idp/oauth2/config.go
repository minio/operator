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

// Package oauth2 contains all the necessary configurations to initialize the
// idp communication using oauth2 protocol
package oauth2

import (
	"crypto/sha1"
	"strings"

	"github.com/minio/pkg/env"
	"golang.org/x/crypto/pbkdf2"
)

// ProviderConfig - OpenID IDP Configuration for console.
type ProviderConfig struct {
	URL                      string
	DisplayName              string // user-provided - can be empty
	ClientID, ClientSecret   string
	HMACSalt, HMACPassphrase string
	Scopes                   string
	Userinfo                 bool
	RedirectCallbackDynamic  bool
	RedirectCallback         string
	EndSessionEndpoint       string
	RoleArn                  string // can be empty
}

// GetStateKeyFunc - return the key function used to generate the authorization
// code flow state parameter.
func (pc ProviderConfig) GetStateKeyFunc() StateKeyFunc {
	return func() []byte {
		return pbkdf2.Key([]byte(pc.HMACPassphrase), []byte(pc.HMACSalt), 4096, 32, sha1.New)
	}
}

// GetARNInf return the ARN
func (pc ProviderConfig) GetARNInf() string {
	return pc.RoleArn
}

// OpenIDPCfg type for OpenID configurations
type OpenIDPCfg map[string]ProviderConfig

// GetSTSEndpoint set location to get the STS session from
func GetSTSEndpoint() string {
	return strings.TrimSpace(env.Get(ConsoleMinIOServer, "http://localhost:9000"))
}

// GetIDPSecret returns environment variable
func GetIDPSecret() string {
	return env.Get(ConsoleIDPSecret, "")
}

// getIDPTokenExpiration return default token expiration for access token (in seconds)
func getIDPTokenExpiration() string {
	return env.Get(ConsoleIDPTokenExpiration, "3600")
}
