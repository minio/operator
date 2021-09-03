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
	"errors"
	"fmt"
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

const (
	listDesc = `'list' command lists all MinIO tenant managed by the Operator`
)

type listCmd struct {
	out    io.Writer
	errOut io.Writer
}

func newTenantListCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &listCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tenants",
		Long:  listDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(args); err != nil {
				return err
			}
			klog.Info("list tenant command started")
			err := c.run(args)
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

func (d *listCmd) validate(args []string) error {
	if len(args) != 0 {
		return errors.New("list command doesn't take any argument, try 'kubectl minio tenant list'")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *listCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}

	tenant, err := oclient.MinioV2().Tenants("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	printTenantList(*tenant)

	return nil
}

func printTenantList(tenants miniov2.TenantList) {
	for _, tenant := range tenants.Items {
		fmt.Printf(Bold(fmt.Sprintf("\nTenant '%s', Namespace '%s', Total capacity %s\n\n", tenant.Name, tenant.ObjectMeta.Namespace, helpers.TotalCapacity(tenant))))
		fmt.Printf(Blue("  Current status: %s \n", tenant.Status.CurrentState))
		fmt.Printf(Blue("  MinIO version: %s \n", tenant.Spec.Image))
		if tenant.Spec.KES != nil && tenant.Spec.KES.Image != "" {
			fmt.Printf(Blue("  KES version: %s \n\n", tenant.Spec.KES.Image))
		}
	}
	fmt.Println()
}
