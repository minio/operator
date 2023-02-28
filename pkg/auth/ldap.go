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

package auth

import (
	"net/http"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

// GetCredentialsFromLDAP authenticates the user against MinIO when the LDAP integration is enabled
// if the authentication succeed *credentials.Login object is returned and we continue with the normal STSAssumeRole flow
func GetCredentialsFromLDAP(client *http.Client, endpoint, ldapUser, ldapPassword string) (*credentials.Credentials, error) {
	creds := credentials.New(&credentials.LDAPIdentity{
		Client:       client,
		STSEndpoint:  endpoint,
		LDAPUsername: ldapUser,
		LDAPPassword: ldapPassword,
	})
	return creds, nil
}
