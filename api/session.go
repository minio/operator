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

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	authApi "github.com/minio/operator/api/operations/auth"
	"github.com/minio/operator/models"
)

func registerSessionHandlers(api *operations.OperatorAPI) {
	// session check
	api.AuthSessionCheckHandler = authApi.SessionCheckHandlerFunc(func(params authApi.SessionCheckParams, session *models.Principal) middleware.Responder {
		sessionResp, err := getSessionResponse(session, params)
		if err != nil {
			return authApi.NewSessionCheckDefault(int(err.Code)).WithPayload(err)
		}
		return authApi.NewSessionCheckOK().WithPayload(sessionResp)
	})
}

// getSessionResponse parse the token of the current session and returns a list of allowed actions to render in the UI
func getSessionResponse(session *models.Principal, params authApi.SessionCheckParams) (*models.OperatorSessionResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// serialize output
	if session == nil {
		return nil, ErrorWithContext(ctx, ErrInvalidSession)
	}
	sessionResp := &models.OperatorSessionResponse{
		Status:      models.OperatorSessionResponseStatusOk,
		Operator:    true,
		Permissions: map[string][]string{},
		Features:    getListOfOperatorFeatures(),
	}
	return sessionResp, nil
}

// getListOfEnabledFeatures returns a list of features
func getListOfOperatorFeatures() []string {
	features := []string{}
	mpEnabled := getMarketplace()

	if mpEnabled != "" {
		features = append(features, fmt.Sprintf("mp-mode-%s", mpEnabled))
	}

	return features
}
