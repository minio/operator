/*
 * This file is part of MinIO Operator
 * Copyright (C) 2021, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cmd

import (
	"errors"
	"fmt"
	"io"

	"k8s.io/klog/v2"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/spf13/cobra"
)

// version provides the version of this plugin
var version = "DEVELOPMENT.GOGET"

const (
	operatorVersionDesc = `
'version' command displays the kubectl plugin version.`
	operatorVersionExample = `  kubectl minio version`
)

type operatorVersionCmd struct {
	out    io.Writer
	errOut io.Writer
}

func newVersionCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorVersionCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Display plugin version",
		Long:    operatorVersionDesc,
		Example: operatorVersionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			err := o.run()
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorVersionCmd) run() error {
	fmt.Println(version)
	return nil
}
