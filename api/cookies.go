// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package api

import (
	"net/http"
	"time"

	xjwt "github.com/minio/operator/pkg/auth/token"
)

// NewSessionCookieForConsole creates a cookie for a token
func NewSessionCookieForConsole(token string) http.Cookie {
	sessionDuration := xjwt.GetConsoleSTSDuration()
	return http.Cookie{
		Path:     "/",
		Name:     "token",
		Value:    token,
		MaxAge:   int(sessionDuration.Seconds()), // default 1 hr
		Expires:  time.Now().Add(sessionDuration),
		HttpOnly: true,
		// if len(GlobalPublicCerts) > 0 is true, that means Console is running with TLS enable and the browser
		// should not leak any cookie if we access the site using HTTP
		Secure: len(GlobalPublicCerts) > 0,
		// read more: https://web.dev/samesite-cookies-explained/
		SameSite: http.SameSiteLaxMode,
	}
}

// ExpireSessionCookie expires a cookie
func ExpireSessionCookie() http.Cookie {
	return http.Cookie{
		Path:     "/",
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		// if len(GlobalPublicCerts) > 0 is true, that means Console is running with TLS enable and the browser
		// should not leak any cookie if we access the site using HTTP
		Secure: len(GlobalPublicCerts) > 0,
		// read more: https://web.dev/samesite-cookies-explained/
		SameSite: http.SameSiteLaxMode,
	}
}
