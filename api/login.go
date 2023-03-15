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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	xoauth2 "golang.org/x/oauth2"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/pkg/env"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	authApi "github.com/minio/operator/api/operations/auth"
	"github.com/minio/operator/models"

	"github.com/minio/operator/pkg/auth"
	"github.com/minio/operator/pkg/auth/idp/oauth2"
)

func registerLoginHandlers(api *operations.OperatorAPI) {
	// GET login strategy
	api.AuthLoginDetailHandler = authApi.LoginDetailHandlerFunc(func(params authApi.LoginDetailParams) middleware.Responder {
		loginDetails, err := getLoginDetailsResponse(params)
		if err != nil {
			return authApi.NewLoginDetailDefault(int(err.Code)).WithPayload(err)
		}
		return authApi.NewLoginDetailOK().WithPayload(loginDetails)
	})
	// POST login using k8s service account token
	api.AuthLoginOperatorHandler = authApi.LoginOperatorHandlerFunc(func(params authApi.LoginOperatorParams) middleware.Responder {
		loginResponse, err := getLoginOperatorResponse(params)
		if err != nil {
			return authApi.NewLoginOperatorDefault(int(err.Code)).WithPayload(err)
		}
		// Custom response writer to set the session cookies
		return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
			cookie := NewSessionCookieForConsole(loginResponse.SessionID)
			http.SetCookie(w, &cookie)
			authApi.NewLoginOperatorNoContent().WriteResponse(w, p)
		})
	})
	// POST login using external IDP
	api.AuthLoginOauth2AuthHandler = authApi.LoginOauth2AuthHandlerFunc(func(params authApi.LoginOauth2AuthParams) middleware.Responder {
		loginResponse, err := getLoginOauth2AuthResponse(params)
		if err != nil {
			return authApi.NewLoginOauth2AuthDefault(int(err.Code)).WithPayload(err)
		}
		// Custom response writer to set the session cookies
		return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
			cookie := NewSessionCookieForConsole(loginResponse.SessionID)
			http.SetCookie(w, &cookie)
			authApi.NewLoginOauth2AuthNoContent().WriteResponse(w, p)
		})
	})
}

// login performs a check of consoleCredentials against MinIO, generates some claims and returns the jwt
// for subsequent authentication
func login(credentials ConsoleCredentialsI) (*string, error) {
	// try to obtain consoleCredentials,
	tokens, err := credentials.Get()
	if err != nil {
		return nil, err
	}
	// if we made it here, the consoleCredentials work, generate a jwt with claims
	token, err := auth.NewEncryptedTokenForClient(&tokens, credentials.GetAccountAccessKey(), nil)
	if err != nil {
		LogError("error authenticating user: %v", err)
		return nil, ErrInvalidLogin
	}
	return &token, nil
}

// isKubernetes returns true if minio is running in kubernetes.
func isKubernetes() bool {
	// Kubernetes env used to validate if we are
	// indeed running inside a kubernetes pod
	// is KUBERNETES_SERVICE_HOST
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet_pods.go#L541
	return env.Get("KUBERNETES_SERVICE_HOST", "") != ""
}

// getLoginDetailsResponse returns information regarding the Console authentication mechanism.
func getLoginDetailsResponse(params authApi.LoginDetailParams) (*models.LoginDetails, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	r := params.HTTPRequest

	loginStrategy := models.LoginDetailsLoginStrategyServiceDashAccount

	var redirectRules []*models.RedirectRule

	if oauth2.IsIDPEnabled() {
		loginStrategy = models.LoginDetailsLoginStrategyRedirectDashServiceDashAccount
		// initialize new oauth2 client
		oauth2Client, err := oauth2.NewOauth2ProviderClient(nil, r, GetConsoleHTTPClient(""))
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		// Validate user against IDP
		identityProvider := &auth.IdentityProvider{
			KeyFunc: oauth2.DefaultDerivedKey,
			Client:  oauth2Client,
		}

		newRedirectRule := &models.RedirectRule{
			Redirect:    identityProvider.GenerateLoginURL(),
			DisplayName: "Login with SSO",
		}

		redirectRules = append(redirectRules, newRedirectRule)
	}

	loginDetails := &models.LoginDetails{
		LoginStrategy: loginStrategy,
		RedirectRules: redirectRules,
		IsK8S:         isKubernetes(),
	}
	return loginDetails, nil
}

// verifyUserAgainstIDP will verify user identity against the configured IDP and return MinIO credentials
func verifyUserAgainstIDP(ctx context.Context, provider auth.IdentityProviderI, code, state string) (*xoauth2.Token, error) {
	oauth2Token, err := provider.VerifyIdentityForOperator(ctx, code, state)
	if err != nil {
		return nil, err
	}
	return oauth2Token, nil
}

func getLoginOauth2AuthResponse(params authApi.LoginOauth2AuthParams) (*models.LoginResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	r := params.HTTPRequest
	lr := params.Body

	if oauth2.IsIDPEnabled() {
		decodedRState, err := base64.StdEncoding.DecodeString(*lr.State)
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}

		var requestItems oauth2.LoginURLParams
		err = json.Unmarshal(decodedRState, &requestItems)

		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}

		// initialize new oauth2 client
		oauth2Client, err := oauth2.NewOauth2ProviderClient(nil, r, GetConsoleHTTPClient(""))
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		// initialize new identity provider
		identityProvider := auth.IdentityProvider{
			KeyFunc: oauth2.DefaultDerivedKey,
			Client:  oauth2Client,
		}
		// Validate user against IDP
		_, err = verifyUserAgainstIDP(ctx, identityProvider, *lr.Code, requestItems.State)
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		// If we pass here that means the IDP correctly authenticate the user with the operator resource
		// we proceed to use the service account token configured in the operator-console pod
		creds, err := newConsoleCredentials(getK8sSAToken())
		if err != nil {
			return nil, ErrorWithContext(ctx, err)
		}
		token, err := login(ConsoleCredentials{ConsoleCredentials: creds})
		if err != nil {
			return nil, ErrorWithContext(ctx, ErrInvalidLogin, nil, err)
		}
		// serialize output
		loginResponse := &models.LoginResponse{
			SessionID: *token,
		}
		return loginResponse, nil
	}
	return nil, ErrorWithContext(ctx, ErrDefault)
}

func newConsoleCredentials(secretKey string) (*credentials.Credentials, error) {
	creds, err := GetConsoleCredentialsForOperator(secretKey)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// getLoginOperatorResponse validate the provided service account token against k8s api
func getLoginOperatorResponse(params authApi.LoginOperatorParams) (*models.LoginResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	lmr := params.Body

	creds, err := newConsoleCredentials(*lmr.Jwt)
	if err != nil {
		log.Println(err)
		return nil, ErrorWithContext(ctx, err)
	}
	consoleCreds := ConsoleCredentials{ConsoleCredentials: creds}
	// Set a random as access key as session identifier
	consoleCreds.AccountAccessKey = fmt.Sprintf("%d", rand.Intn(100000-10000)+10000)
	token, err := login(consoleCreds)
	if err != nil {
		log.Println(err)
		return nil, ErrorWithContext(ctx, ErrInvalidLogin, nil, err)
	}
	// serialize output
	loginResponse := &models.LoginResponse{
		SessionID: *token,
	}
	return loginResponse, nil
}
