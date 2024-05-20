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

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/minio/operator/pkg/auth"
	"github.com/minio/operator/pkg/auth/idp/oauth2"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	authApi "github.com/minio/operator/api/operations/auth"
	"github.com/minio/operator/models"
)

func registerLogoutHandlers(api *operations.OperatorAPI) {
	// logout from console
	api.AuthLogoutHandler = authApi.LogoutHandlerFunc(func(params authApi.LogoutParams, session *models.Principal) middleware.Responder {
		// Custom response writer to expire the session cookies
		return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
			if oauth2.IsIDPEnabled() {
				err := logoutIDP(params.HTTPRequest)
				if err != nil {
					api.Logger("IDP logout failed: %v", err.DetailedMessage)
					w.Header().Set("IDP-Logout", fmt.Sprintf("%v", err.DetailedMessage))
				}
			}
			expiredCookie := ExpireSessionCookie()
			expiredIDPCookie := ExpireIDPSessionCookie()
			// this will tell the browser to clear the cookies and invalidate user session
			// additionally we are deleting the cookie from the client side
			http.SetCookie(w, &expiredCookie)
			http.SetCookie(w, &expiredIDPCookie)
			authApi.NewLogoutOK().WriteResponse(w, p)
		})
	})
}

func logoutIDP(r *http.Request) *models.Error {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// initialize new oauth2 client
	oauth2Client, err := oauth2.NewOauth2ProviderClient(nil, r, GetConsoleHTTPClient(""))
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	// initialize new identity provider
	identityProvider := auth.IdentityProvider{
		KeyFunc: oauth2.DefaultDerivedKey,
		Client:  oauth2Client,
	}
	refreshToken, err := r.Cookie("idp-refresh-token")
	if err != nil {
		return ErrorWithContext(ctx, err)
	}

	err = identityProvider.Logout(refreshToken.Value)
	if err != nil {
		return ErrorWithContext(ctx, ErrDefault, nil, err)
	}
	return nil
}
