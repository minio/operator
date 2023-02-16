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

package main

import (
	"github.com/minio/cli"
	"github.com/minio/operator/pkg/validator"
)

// starts the controller
var validateCmd = cli.Command{
	Name:    "validate",
	Aliases: []string{"v"},
	Usage:   "Start MinIO Operator Config Validator",
	Action:  startValidator,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "name of tenant being validated",
		},
	},
}

func startValidator(ctx *cli.Context) {
	tenantName := ctx.String("tenant")
	validator.Validate(tenantName)
}
