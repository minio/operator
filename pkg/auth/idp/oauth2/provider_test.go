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

package oauth2

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type Oauth2configMock struct{}

var (
	oauth2ConfigExchangeMock                 func(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	oauth2ConfigAuthCodeURLMock              func(state string, opts ...oauth2.AuthCodeOption) string
	oauth2ConfigPasswordCredentialsTokenMock func(ctx context.Context, username, password string) (*oauth2.Token, error)
	oauth2ConfigClientMock                   func(ctx context.Context, t *oauth2.Token) *http.Client
	oauth2ConfigokenSourceMock               func(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
)

func (ac Oauth2configMock) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return oauth2ConfigExchangeMock(ctx, code, opts...)
}

func (ac Oauth2configMock) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return oauth2ConfigAuthCodeURLMock(state, opts...)
}

func (ac Oauth2configMock) PasswordCredentialsToken(ctx context.Context, username, password string) (*oauth2.Token, error) {
	return oauth2ConfigPasswordCredentialsTokenMock(ctx, username, password)
}

func (ac Oauth2configMock) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2ConfigClientMock(ctx, t)
}

func (ac Oauth2configMock) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return oauth2ConfigokenSourceMock(ctx, t)
}

func TestGenerateLoginURL(t *testing.T) {
	funcAssert := assert.New(t)
	oauth2Provider := Provider{
		oauth2Config: Oauth2configMock{},
	}
	// Test-1 : GenerateLoginURL() generates URL correctly with provided state
	oauth2ConfigAuthCodeURLMock = func(state string, opts ...oauth2.AuthCodeOption) string {
		// Internally we are testing the private method getRandomStateWithHMAC, this function should always returns
		// a non-empty string
		return state
	}
	url := oauth2Provider.GenerateLoginURL(DefaultDerivedKey, "testIDP")
	funcAssert.NotEqual("", url)
}
