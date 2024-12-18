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
	"flag"
	"strconv"

	"github.com/minio/cli"
	"github.com/minio/operator/pkg/controller"
)

// starts the controller
var controllerCmd = cli.Command{
	Name:    "controller",
	Aliases: []string{"ctl"},
	Usage:   "Start MinIO Operator Controller",
	Action:  startController,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Load configuration from `KUBECONFIG`",
		},
		cli.IntFlag{
			Name:  "v",
			Usage: "logging level",
		},
	},
}

func startController(ctx *cli.Context) {
	if ctx.Int("v") > 0 {
		flag.Set("v", strconv.Itoa(ctx.Int("v")))
		flag.Parse()
	}
	controller.StartOperator(ctx.String("kubeconfig"))
}
