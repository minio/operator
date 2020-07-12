/*
 * This file is part of MinIO Operator
 * Copyright (C) 2020, MinIO, Inc.
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
	"context"
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	updateDesc = `
'update' command updates a MinIO Cluster tenant`
)

type updateCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newUpdateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &updateCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update MinIO tenant version",
		Args:  cobra.MinimumNArgs(2),
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *updateCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = updateTenant(oclient, args, d)
	if err != nil {
		return err
	}
	return nil
}

func updateTenant(client *operatorv1.Clientset, args []string, d *updateCmd) error {
	t, err := client.OperatorV1().MinIOInstances(d.ns).Get(context.TODO(), args[0], v1.GetOptions{})
	if err != nil {
		return err
	}
	// update the image
	t.Spec.Image = args[1]
	if _, err := client.OperatorV1().MinIOInstances(d.ns).Update(context.TODO(), t, v1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
