// Copyright (C) 2023, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package main

import (
	"log"
	"os"

	"github.com/minio/cli"
	"github.com/minio/operator/pkg/sidecar"
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
		cli.StringFlag{
			Name:  "config-name",
			Value: "",
			Usage: "secret being watched",
		},
	},
}

func startSideCar(ctx *cli.Context) {
	tenantName := ctx.String("tenant")
	if tenantName == "" {
		log.Println("Must pass --tenant flag")
		os.Exit(1)
	}
	configName := ctx.String("config-name")
	if configName == "" {
		log.Println("Must pass --config-name flag")
		os.Exit(1)
	}
	sidecar.StartSideCar(tenantName, configName)
}
