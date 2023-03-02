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

	"github.com/minio/operator/pkg/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
)

func registerParityHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetParityHandler = operator_api.GetParityHandlerFunc(func(params operator_api.GetParityParams, principal *models.Principal) middleware.Responder {
		resp, err := getParityResponse(params)
		if err != nil {
			return operator_api.NewGetParityDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetParityOK().WithPayload(resp)
	})
}

// GetParityInfo returns the parity for nxd config
func GetParityInfo(nodes int64, disksPerNode int64) (models.ParityResponse, error) {
	parityVals, err := utils.PossibleParityValues(fmt.Sprintf(`http://minio{1...%d}/export/set{1...%d}`, nodes, disksPerNode))
	if err != nil {
		return nil, err
	}

	return parityVals, nil
}

func getParityResponse(params operator_api.GetParityParams) (models.ParityResponse, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	nodes := params.Nodes
	disksPerNode := params.DisksPerNode
	parityValues, err := GetParityInfo(nodes, disksPerNode)
	if err != nil {
		LogError("error getting parity info: %v", err)
		return nil, ErrorWithContext(ctx, err)
	}
	return parityValues, nil
}
