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
	"context"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/operator/pkg/auth/idp/oauth2"
	xoauth2 "golang.org/x/oauth2"
)

// IdentityProviderI interface with all functions to be implemented
// by mock when testing, it should include all IdentityProvider respective api calls
// that are used within this project.
type IdentityProviderI interface {
	VerifyIdentity(ctx context.Context, code, state string) (*credentials.Credentials, error)
	VerifyIdentityForOperator(ctx context.Context, code, state string) (*xoauth2.Token, error)
	GenerateLoginURL() string
}

// IdentityProvider Identity implementation
//
// Define the structure of a IdentityProvider with Client inside and define the functions that are used
// during the authentication flow.
type IdentityProvider struct {
	KeyFunc oauth2.StateKeyFunc
	Client  *oauth2.Provider
	RoleARN string
}

// VerifyIdentity will verify the user identity against the idp using the authorization code flow
func (c IdentityProvider) VerifyIdentity(ctx context.Context, code, state string) (*credentials.Credentials, error) {
	return c.Client.VerifyIdentity(ctx, code, state, c.RoleARN, c.KeyFunc)
}

// VerifyIdentityForOperator will verify the user identity against the idp using the authorization code flow
func (c IdentityProvider) VerifyIdentityForOperator(ctx context.Context, code, state string) (*xoauth2.Token, error) {
	return c.Client.VerifyIdentityForOperator(ctx, code, state, c.KeyFunc)
}

// GenerateLoginURL returns a new URL used by the user to login against the idp
func (c IdentityProvider) GenerateLoginURL() string {
	return c.Client.GenerateLoginURL(c.KeyFunc, c.Client.IDPName)
}
