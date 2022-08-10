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
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/minio/kubectl-minio/cmd/helpers"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	reportDesc = `'report' command saves pod logs from a MinIO tenant`
)

type reportCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newTenantReportCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &reportCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "report <TENANTNAME>",
		Short:   "Collect pod logs, events, and status for a tenant",
		Long:    reportDesc,
		Example: `  kubectl minio report tenant1`,
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

func (d *reportCmd) validate(args []string) error {
	return validateTenantArgs("report", args)
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *reportCmd) run(args []string) error {
	// Create operator client
	ctx := context.Background()

	path, _ := rootCmd.Flags().GetString(kubeconfig)
	oclient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}
	client, err := helpers.GetKubeClient(path)
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
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, tenant.Name),
	}
	podsSet := client.CoreV1().Pods(tenant.Namespace)
	events := client.CoreV1().Events(tenant.Namespace)
	pods, err := podsSet.List(ctx, listOpts)
	if err != nil {
		return err
	}
	w, err := os.Create(tenant.Name + "-report.zip")
	if err != nil {
		return err
	}
	zipw := zip.NewWriter(w)
	tenantAsYaml, err := yaml.Marshal(tenant)
	if err == nil {
		f, err := zipw.Create(tenant.Name + ".yaml")
		if err == nil {
			f.Write(tenantAsYaml)
		}
	}
	for i := 0; i < len(pods.Items); i++ {
		toWrite, err := podsSet.GetLogs(pods.Items[i].Name, &v1.PodLogOptions{}).DoRaw(ctx)
		if err == nil {
			f, err := zipw.Create(pods.Items[i].Name + ".log")
			if err == nil {
				f.Write(toWrite)
			}
		}
		podEvents, err := events.List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pods.Items[i].UID)})
		if err == nil {
			podEventsJSON, err := json.Marshal(podEvents)
			if err == nil {
				f, err := zipw.Create(pods.Items[i].Name + "-events.txt")
				if err == nil {
					f.Write(podEventsJSON)
				}
			}
		}
		status := pods.Items[i].Status
		statusJSON, err := json.Marshal(status)
		if err == nil {
			f, err := zipw.Create(pods.Items[i].Name + "-status.txt")
			if err == nil {
				f.Write(statusJSON)
			}
		}
	}
	zipw.Close()
	fmt.Println("Data stored in " + tenant.Name + "-report.zip")
	return nil
}
