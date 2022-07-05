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
	"encoding/json"
	"errors"
	"fmt"
	"io"

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
		return errors.New(fmt.Sprintf("provide the name of the tenant, e.g. 'kubectl minio tenant %s tenant1'", cmd))
	}
	if len(args) != 1 {
		return errors.New(fmt.Sprintf("%s command supports a single argument, e.g. 'kubectl minio %s tenant1'", cmd, cmd))
	}
	if args[0] == "" {
		return errors.New(fmt.Sprintf("provide the name of the tenant, e.g. 'kubectl minio tenant %s tenant1'", cmd))
	}
	return helpers.CheckValidTenantName(args[0])

}

func (d *statusCmd) run(args []string) error {
	oclient, err := helpers.GetKubeOperatorClient()
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
	gb := float32(humanize.GiByte)
	fmt.Printf(Blue("Current status: %s \n", tenant.Status.CurrentState))
	fmt.Printf(Blue("Available replicas: %d \n", tenant.Status.AvailableReplicas))
	fmt.Printf(Blue("Pools: %d \n", len(tenant.Status.Pools)))
	fmt.Printf(Blue("Revision: %d \n", tenant.Status.Revision))
	fmt.Printf(Blue("Sync version: %s \n", tenant.Status.SyncVersion))
	fmt.Printf(Blue("Provisioned users: %t \n", tenant.Status.ProvisionedUsers))
	fmt.Printf(Blue("Write quorum: %d \n", tenant.Status.WriteQuorum))
	fmt.Printf(Blue("Drives online: %d \n", tenant.Status.DrivesOnline))
	fmt.Printf(Blue("Drives offline: %d \n", tenant.Status.DrivesOffline))
	fmt.Printf(Blue("Drives healing: %d \n", tenant.Status.DrivesHealing))
	fmt.Printf(Blue("Health status: %s \n", tenant.Status.HealthStatus))
	fmt.Printf(Blue("Capacity: %.1fGi \n", float32(tenant.Status.Usage.Capacity)/gb))
	fmt.Printf(Blue("Raw capacity: %.1fGi \n", float32(tenant.Status.Usage.RawCapacity)/gb))
	fmt.Printf(Blue("Raw usage: %.1fGi \n", float32(tenant.Status.Usage.RawUsage)/gb))
}

func (d *statusCmd) printJSONTenantStatus(tenant *v2.Tenant) error {
	statusJSON, err := json.MarshalIndent(tenant.Status, "", "   ")
	if err != nil {
		return err
	}
	fmt.Printf(Green(fmt.Sprintf("JSON:\n\n%s\n\n", string(statusJSON))))
	return nil
}

func (d *statusCmd) printYAMLTenantStatus(tenant *v2.Tenant) error {
	statusYAML, err := yaml.Marshal(tenant.Status)
	if err != nil {
		return err
	}
	fmt.Printf(Yellow("YAML:\n\n%s", string(statusYAML)))
	return nil
}
