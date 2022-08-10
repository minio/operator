// This file is part of MinIO Operator
// Copyright (C) 2020, MinIO, Inc.
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

package cmd

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/minio/kubectl-minio/cmd/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/spf13/cobra"
)

const (
	infoDesc = `'info' command lists pools from a MinIO tenant`
)

type infoCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newTenantInfoCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &infoCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "info <TENANTNAME>",
		Short:   "List all volumes in existing tenant",
		Long:    infoDesc,
		Example: `  kubectl minio info tenant1`,
		Args: func(cmd *cobra.Command, args []string) error {
			return c.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.run(args)
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", "", "namespace scope for this request")
	return cmd
}

func (d *infoCmd) validate(args []string) error {
	return validateTenantArgs("info", args)
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *infoCmd) run(args []string) error {
	// Create operator client
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	oclient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}

	if d.ns == "" || d.ns == helpers.DefaultNamespace {
		d.ns, err = getTenantNamespace(oclient, args[0])
		if err != nil {
			return err
		}
	}

	tenant, err := oclient.MinioV2().Tenants(d.ns).Get(context.Background(), args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	printTenantInfo(*tenant)

	return nil
}

func printTenantInfo(tenant miniov2.Tenant) {
	// Check MinIO S3 Endpoint Service
	minSvc := services.NewClusterIPForMinIO(&tenant)

	// Check MinIO Console Endpoint Service
	conSvc := services.NewClusterIPForConsole(&tenant)

	var minPorts, consolePorts string
	for _, p := range minSvc.Spec.Ports {
		minPorts = minPorts + strconv.Itoa(int(p.Port)) + ","
	}
	for _, p := range conSvc.Spec.Ports {
		consolePorts = consolePorts + strconv.Itoa(int(p.Port)) + ","
	}
	fmt.Printf(Bold(fmt.Sprintf("Tenant '%s', Namespace '%s', Total capacity %s\n\n", tenant.Name, tenant.ObjectMeta.Namespace, helpers.TotalCapacity(tenant))))
	fmt.Printf(Blue("Current status: %s\n", tenant.Status.CurrentState))
	fmt.Printf(Blue("MinIO version: %s\n", tenant.Spec.Image))
	fmt.Printf(Blue("MinIO service: %s/ClusterIP (port %s)\n", minSvc.Name, strings.TrimSuffix(minPorts, ",")))
	fmt.Printf(Blue("Console service: %s/ClusterIP (port %s)\n", conSvc.Name, strings.TrimSuffix(consolePorts, ",")))
	if tenant.Spec.KES != nil && tenant.Spec.KES.Image != "" {
		fmt.Printf(Blue("KES version: %s \n\n", tenant.Spec.KES.Image))
	} else {
		fmt.Println()
	}

	t := helpers.GetTable()
	t.SetHeader([]string{"Pool", "Servers", "Volumes(server)", "Capacity(volume)"})
	for i, z := range tenant.Spec.Pools {
		t.Append([]string{
			strconv.Itoa(i),
			strconv.Itoa(int(z.Servers)),
			strconv.Itoa(int(z.VolumesPerServer)),
			humanize.IBytes(uint64(z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value())),
		})
	}
	t.Render()
}
