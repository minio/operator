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
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/minio/minio/pkg/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	expandDesc = `
'expand' command adds storage capacity to a MinIO tenant`
	expandExample = `  kubectl minio tenant expand tenant1 --servers 4 --volumes 32 --capacity 32Ti --namespace tenant1-ns`
)

type expandCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	tenantOpts resources.TenantOptions
}

func newTenantExpandCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	v := &expandCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "expand <string> --servers <int> --volumes <int> --capacity <str> --namespace <str>",
		Short:   "Add capacity to existing tenant",
		Long:    expandDesc,
		Example: expandExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := v.validate(args); err != nil {
				return err
			}
			return v.run()
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.Int32Var(&v.tenantOpts.Servers, "servers", 0, "total number of pods to add to tenant")
	f.Int32Var(&v.tenantOpts.Volumes, "volumes", 0, "total number of volumes to add to tenant")
	f.StringVar(&v.tenantOpts.Capacity, "capacity", "", "total raw capacity to add to tenant, e.g. 16Ti")
	f.StringVarP(&v.tenantOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.BoolVarP(&v.output, "output", "o", false, "dry run this command and generate requisite yaml")

	cmd.MarkFlagRequired("servers")
	cmd.MarkFlagRequired("volumes")
	cmd.MarkFlagRequired("capacity")
	return cmd
}

func (v *expandCmd) validate(args []string) error {
	if args == nil {
		return errors.New("provide the name of the tenant, e.g. 'kubectl minio tenant expand tenant1'")
	}
	if len(args) != 1 {
		return errors.New("expand command supports a single argument, e.g. 'kubectl minio tenant expand tenant1'")
	}
	if args[0] == "" {
		return errors.New("provide the name of the tenant, e.g. 'kubectl minio tenant expand tenant1'")
	}
	v.tenantOpts.Name = args[0]
	return v.tenantOpts.Validate()
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (v *expandCmd) run() error {
	// Create operator client
	client, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}

	t, err := client.MinioV2().Tenants(v.tenantOpts.NS).Get(context.Background(), v.tenantOpts.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	currentCapacity := helpers.TotalCapacity(*t)
	volumesPerServer := helpers.VolumesPerServer(v.tenantOpts.Volumes, v.tenantOpts.Servers)
	capacityPerVolume, err := helpers.CapacityPerVolume(v.tenantOpts.Capacity, v.tenantOpts.Volumes)
	if err != nil {
		return err
	}

	t.Spec.Pools = append(t.Spec.Pools, resources.Pool(v.tenantOpts.Servers, volumesPerServer, *capacityPerVolume, v.tenantOpts.StorageClass))
	expandedCapacity := helpers.TotalCapacity(*t)
	if !v.output {
		fmt.Printf(color.Bold(fmt.Sprintf("\nExpanding Tenant '%s/%s' from %s to %s\n\n", t.ObjectMeta.Name, t.ObjectMeta.Namespace, currentCapacity, expandedCapacity)))
		return addPoolToTenant(client, t)
	}

	o, err := yaml.Marshal(t)
	if err != nil {
		return err
	}
	fmt.Println(string(o))
	return nil
}

func addPoolToTenant(client *operatorv1.Clientset, t *miniov2.Tenant) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	if _, err := client.MinioV2().Tenants(t.Namespace).Patch(context.Background(), t.Name, types.MergePatchType, data, metav1.PatchOptions{FieldManager: "kubectl"}); err != nil {
		return err
	}
	return nil
}
