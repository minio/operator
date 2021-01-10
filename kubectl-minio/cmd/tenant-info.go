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
	"strconv"
	"strings"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/minio/pkg/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/spf13/cobra"
)

const (
	infoDesc = `
'info' command lists pools from a MinIO tenant`
	infoExample = `  kubectl minio tenant info tenant1 --namespace tenant1-ns`
)

type infoCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newTenantInfoCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &infoCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "info",
		Short: "List all volumes in existing tenant",
		Long:  infoDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(args); err != nil {
				return err
			}
			return c.run(args)
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	return cmd
}

func (d *infoCmd) validate(args []string) error {
	if args == nil {
		return errors.New("provide the name of the tenant, e.g. 'kubectl minio tenant info tenant1'")
	}
	if len(args) != 1 {
		return errors.New("info command supports a single argument, e.g. 'kubectl minio tenant info tenant1'")
	}
	if args[0] == "" {
		return errors.New("provide the name of the tenant, e.g. 'kubectl minio tenant info tenant1'")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *infoCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}

	tenant, err := oclient.MinioV2().Tenants(d.ns).Get(context.Background(), args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	printTenantInfo(*tenant)

	return nil
}

func printTenantInfo(tenant miniov2.Tenant) {
	minSvc := services.NewClusterIPForMinIO(&tenant)
	conSvc := services.NewClusterIPForConsole(&tenant)
	var minPorts, consolePorts string
	for _, p := range minSvc.Spec.Ports {
		minPorts = minPorts + strconv.Itoa(int(p.Port)) + ","
	}
	for _, p := range conSvc.Spec.Ports {
		consolePorts = consolePorts + strconv.Itoa(int(p.Port)) + ","
	}
	fmt.Printf(color.Bold(fmt.Sprintf("\nTenant '%s/%s', total capacity %s\n\n", tenant.Name, tenant.ObjectMeta.Namespace, helpers.TotalCapacity(tenant))))
	fmt.Printf(color.Blue("  Current status: %s \n", tenant.Status.CurrentState))
	fmt.Printf(color.Blue("  MinIO version: %s \n", tenant.Spec.Image))
	fmt.Printf(color.Blue("  MinIO service: %s/ClusterIP (port %s)\n\n", minSvc.Name, strings.TrimSuffix(minPorts, ",")))
	fmt.Printf(color.Blue("  Console version: %s \n", tenant.Spec.Console.Image))
	fmt.Printf(color.Blue("  Console service: %s/ClusterIP (port %s)\n\n", conSvc.Name, strings.TrimSuffix(consolePorts, ",")))
	if tenant.Spec.KES != nil && tenant.Spec.KES.Image != "" {
		fmt.Printf(color.Blue("  KES version: %s \n\n", tenant.Spec.KES.Image))
	}

	t := helpers.GetTable()
	t.SetHeader([]string{"Pool", "Servers", "Volumes Per Server", "Capacity Per Volume"})
	for i, z := range tenant.Spec.Pools {
		t.Append([]string{strconv.Itoa(i), strconv.Itoa(int(z.Servers)), strconv.Itoa(int(z.VolumesPerServer)), z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().String()})
	}
	t.Render()
	fmt.Println()
}
