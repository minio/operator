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
	"k8s.io/client-go/kubernetes"

	operatorv1 "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	deleteDesc = `
'delete' command deletes a MinIO Cluster tenant`
)

type deleteCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &deleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete MinIO Cluster tenant",
		Args:  cobra.MinimumNArgs(1),
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
func (d *deleteCmd) run(args []string) error {
	client, err := helpers.GetKubeClient()
	if err != nil {
		return err
	}
	err = deleteSecret(client, d.ns, args)
	if err != nil {
		return err
	}
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = deleteTenant(oclient, args, d)
	if err != nil {
		return err
	}
	return nil
}

func deleteSecret(client *kubernetes.Clientset, ns string, args []string) error {
	if err := client.CoreV1().Secrets(ns).Delete(context.TODO(), helpers.SecretName(args[0]), v1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func deleteTenant(client *operatorv1.Clientset, args []string, d *deleteCmd) error {
	if err := client.OperatorV1().MinIOInstances(d.ns).Delete(context.TODO(), args[0], v1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
