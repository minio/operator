// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.
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
	"log"
	"os"

	"github.com/minio/cli"
	"github.com/minio/operator/sidecar/pkg/sidecar"
)

// starts the controller
var sidecarCmd = cli.Command{
	Name:    "sidecar",
	Aliases: []string{"s"},
	Usage:   "Start MinIO Operator Sidecar",
	Action:  startSideCar,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "name of tenant being validated",
		},
	},
}

func startSideCar(ctx *cli.Context) {
	tenantName := ctx.String("tenant")
	if tenantName == "" {
		log.Println("Must pass --tenant flag")
		os.Exit(1)
	}
	sidecar.StartSideCar(tenantName)
}
