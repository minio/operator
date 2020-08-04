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
	listDesc = `
'list' command lists zones from a MinIO tenant`
	listExample = `  kubectl minio tenant volume list --name tenant1 --namespace tenant1-ns`
)

type volumeListCmd struct {
	out    io.Writer
	errOut io.Writer
	name   string
	ns     string
}

func newVolumeListCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &volumeListCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all volumes in existing tenant",
		Long:  listDesc,
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

func (d *volumeListCmd) validate() error {
	if d.name == "" {
		return errors.New("--name flag is required for adding volumes to tenant")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *volumeListCmd) run() error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = listZonesTenant(oclient, d)
	if err != nil {
		return err
	}
	return nil
}

func listZonesTenant(client *operatorv1.Clientset, d *volumeListCmd) error {
	tenant, err := client.MinioV1().Tenants(d.ns).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Zone", "Servers", "Volumes Per Server", "Capacity Per Volume"})
	for i, z := range tenant.Spec.Zones {
		t.AppendRow(table.Row{i, z.Servers, z.VolumesPerServer, z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().String()})
	}
	t.SetAlign([]text.Align{text.AlignRight, text.AlignRight, text.AlignRight, text.AlignRight})
	t.Render()
	return nil
}
