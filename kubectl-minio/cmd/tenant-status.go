// This file is part of MinIO Operator
// Copyright (C) 2022, MinIO, Inc.
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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/minio/kubectl-minio/cmd/helpers"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	statusDesc = `'status' command displays tenant status information`
)

type statusCmd struct {
	out        io.Writer
	errOut     io.Writer
	ns         string
	yamlOutput bool
	jsonOutput bool
}

func newTenantStatusCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &statusCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "status <TENANTNAME>",
		Short:   "Display tenant status",
		Long:    statusDesc,
		Example: `  kubectl minio status tenant1`,
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
	f.BoolVarP(&c.yamlOutput, "yaml", "y", false, "yaml output")
	f.BoolVarP(&c.jsonOutput, "json", "j", false, "json output")
	return cmd
}

func (d *statusCmd) validate(args []string) error {
	return validateTenantArgs("status", args)
}

func validateTenantArgs(cmd string, args []string) error {
	if args == nil {
		return fmt.Errorf("provide the name of the tenant, e.g. 'kubectl minio tenant %s tenant1'", cmd)
	}
	if len(args) != 1 {
		return fmt.Errorf("%s command supports a single argument, e.g. 'kubectl minio %s tenant1'", cmd, cmd)
	}
	if args[0] == "" {
		return fmt.Errorf("provide the name of the tenant, e.g. 'kubectl minio tenant %s tenant1'", cmd)
	}
	return helpers.CheckValidTenantName(args[0])
}

func (d *statusCmd) run(args []string) error {
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
	return d.printTenantStatus(tenant)
}

func (d *statusCmd) printTenantStatus(tenant *v2.Tenant) error {
	if !d.jsonOutput && !d.yamlOutput {
		d.printRawTenantStatus(tenant)
		return nil
	}
	if d.jsonOutput && d.yamlOutput {
		return fmt.Errorf("Only one output can be used to display status")
	}
	if d.jsonOutput {
		return d.printJSONTenantStatus(tenant)
	}
	if d.yamlOutput {
		return d.printYAMLTenantStatus(tenant)
	}
	return nil
}

func (d *statusCmd) printRawTenantStatus(tenant *v2.Tenant) {
	var s strings.Builder
	s.WriteString("=====================\n")
	s.WriteString(Bold("Pools:              %d \n", len(tenant.Status.Pools)))
	s.WriteString(Bold("Revision:           %d \n", tenant.Status.Revision))
	s.WriteString(Bold("Sync version:       %s \n", tenant.Status.SyncVersion))
	s.WriteString(Bold("Write quorum:       %d \n", tenant.Status.WriteQuorum))
	s.WriteString(Bold("Health status:      %s \n", tenant.Status.HealthStatus))
	s.WriteString(Bold("Drives online:      %d \n", tenant.Status.DrivesOnline))
	s.WriteString(Bold("Drives offline:     %d \n", tenant.Status.DrivesOffline))
	s.WriteString(Bold("Drives healing:     %d \n", tenant.Status.DrivesHealing))
	s.WriteString(Bold("Current status:     %s \n", tenant.Status.CurrentState))
	s.WriteString(Bold("Usable capacity:    %s \n", humanize.IBytes(uint64(tenant.Status.Usage.Capacity))))
	s.WriteString(Bold("Provisioned users:  %t \n", tenant.Status.ProvisionedUsers))
	s.WriteString(Bold("Available replicas: %d \n", tenant.Status.AvailableReplicas))

	fmt.Fprintln(d.out, s.String())
}

func (d *statusCmd) printJSONTenantStatus(tenant *v2.Tenant) error {
	enc := json.NewEncoder(d.out)
	enc.SetIndent("", " ")
	return enc.Encode(tenant.Status)
}

func (d *statusCmd) printYAMLTenantStatus(tenant *v2.Tenant) error {
	statusYAML, err := yaml.Marshal(tenant.Status)
	if err != nil {
		return err
	}
	fmt.Fprintln(d.out, string(statusYAML))
	return nil
}
