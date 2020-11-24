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
	"io"
	"os"

	"github.com/minio/kubectl-minio/cmd/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	infoDesc = `
'info' command lists pools from a MinIO tenant`
	infoExample = `  kubectl minio tenant info --name tenant1 --namespace tenant1-ns`
)

type infoCmd struct {
	out    io.Writer
	errOut io.Writer
	name   string
	ns     string
}

func newTenantInfoCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &infoCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "info",
		Short: "List all volumes in existing tenant",
		Long:  infoDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.name, "name", "", "name of the MinIO tenant to list volumes")
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	return cmd
}

func (d *infoCmd) validate() error {
	if d.name == "" {
		return errors.New("--name flag is required for adding volumes to tenant")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *infoCmd) run() error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = listPoolsTenant(oclient, d)
	if err != nil {
		return err
	}
	return nil
}

func listPoolsTenant(client *operatorv1.Clientset, d *infoCmd) error {
	tenant, err := client.MinioV1().Tenants(d.ns).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Pool", "Servers", "Volumes Per Server", "Capacity Per Volume", "Version"})
	for i, z := range tenant.Spec.Pools {
		t.AppendRow(table.Row{i, z.Servers, z.VolumesPerServer, z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().String()})
	}
	t.AppendFooter(table.Row{"Version", tenant.Spec.Image})
	t.SetAlign([]text.Align{text.AlignRight, text.AlignRight, text.AlignRight, text.AlignRight})
	t.Render()
	return nil
}
