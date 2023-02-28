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
	"testing"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

var creds = &credentials.Value{
	AccessKeyID:     "fakeAccessKeyID",
	SecretAccessKey: "fakeSecretAccessKey",
	SessionToken:    "fakeSessionToken",
	SignerType:      0,
}

var (
	goodToken = ""
	badToken  = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoiRDMwYWE0ekQ1bWtFaFRyWm5yOWM3NWh0Yko0MkROOWNDZVQ5RHVHUkg1U25SR3RyTXZNOXBMdnlFSVJAAAE5eWxxekhYMXllck8xUXpzMlZzRVFKeUF2ZmpOaDkrTVdoUURWZ2FhK2R5emxzSjNpK0k1dUdoeW5DNWswUW83WEY0UWszY0RtUTdUQUVROVFEbWRKdjBkdVB5L25hQk5vM3dIdlRDZHFNRDJZN3kycktJbmVUbUlFNmVveW9EWmprcW5tckVoYmMrTlhTRU81WjZqa1kwZ1E2eXZLaWhUZGxBRS9zS1lBNlc4Q1R1cm1MU0E0b0dIcGtldFZWU0VXMHEzNU9TU1VaczRXNkxHdGMxSTFWVFZLWUo3ZTlHR2REQ3hMWGtiZHQwcjl0RDNMWUhWRndra0dSZit5ZHBzS1Y3L1Jtbkp3SHNqNVVGV0w5WGVHUkZVUjJQclJTN2plVzFXeGZuYitVeXoxNVpOMzZsZ01GNnBlWFd1LzJGcEtrb2Z2QzNpY2x5Rmp0SE45ZkxYTVpVSFhnV2lsQWVSa3oiLCJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjkwMDAiLCJleHAiOjE1ODc1MTY1NzEsInN1YiI6ImZmYmY4YzljLTJlMjYtNGMwYS1iMmI0LTYyMmVhM2I1YjZhYiJ9.P392RUwzsrBeJOO3fS1xMZcF-lWiDvWZ5hM7LZOyFMmoG5QLccDU5eAPSm8obzPoznX1b7eCFLeEmKK-vKgjiQ"
)

func TestNewJWTWithClaimsForClient(t *testing.T) {
	funcAssert := assert.New(t)
	// Test-1 : NewEncryptedTokenForClient() is generated correctly without errors
	function := "NewEncryptedTokenForClient()"
	token, err := NewEncryptedTokenForClient(creds, "", nil)
	if err != nil || token == "" {
		t.Errorf("Failed on %s:, error occurred: %s", function, err)
	}
	// saving token for future tests
	goodToken = token
	// Test-2 : NewEncryptedTokenForClient() throws error because of empty credentials
	if _, err = NewEncryptedTokenForClient(nil, "", nil); err != nil {
		funcAssert.Equal("provided credentials are empty", err.Error())
	}
}

func TestJWTAuthenticate(t *testing.T) {
	funcAssert := assert.New(t)
	// Test-1 : SessionTokenAuthenticate() should correctly return the claims
	function := "SessionTokenAuthenticate()"
	claims, err := SessionTokenAuthenticate(goodToken)
	if err != nil || claims == nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err)
	} else {
		funcAssert.Equal(claims.STSAccessKeyID, creds.AccessKeyID)
		funcAssert.Equal(claims.STSSecretAccessKey, creds.SecretAccessKey)
		funcAssert.Equal(claims.STSSessionToken, creds.SessionToken)
	}
	// Test-2 : SessionTokenAuthenticate() return an error because of a tampered token
	if _, err := SessionTokenAuthenticate(badToken); err != nil {
		funcAssert.Equal("session token internal data is malformed", err.Error())
	}
	// Test-3 : SessionTokenAuthenticate() return an error because of an empty token
	if _, err := SessionTokenAuthenticate(""); err != nil {
		funcAssert.Equal("session token missing", err.Error())
	}
}

func TestSessionTokenValid(t *testing.T) {
	funcAssert := assert.New(t)
	// Test-1 : SessionTokenAuthenticate() provided token is valid
	funcAssert.Equal(true, IsSessionTokenValid(goodToken))
	// Test-2 : SessionTokenAuthenticate() provided token is invalid
	funcAssert.Equal(false, IsSessionTokenValid(badToken))
}
