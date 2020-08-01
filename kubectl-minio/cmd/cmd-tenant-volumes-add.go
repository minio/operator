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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	addDesc = `
'scale' command adds zones to a MinIO Cluster tenant`
)

type volumeAddCmd struct {
	out          io.Writer
	errOut       io.Writer
	name         string
	servers      int32
	volumes      int32
	capacity     string
	ns           string
	storageClass string
}

func newVolumeAddCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &volumeAddCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "add --name TENANT_NAME --server SERVERS --volumes VOLUMES --capacity --capacity VOLUME_CAPACITY",
		Short: "Add Storage Capacity to existing MinIO Tenant",
		Long:  addDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.name, "name", "", "Name of the MinIO Tenant to be created, e.g. Tenant1")
	f.Int32Var(&c.servers, "servers", 0, "Number of pods in the MinIO Tenant, e.g. 4")
	f.Int32Var(&c.volumes, "volumes", 0, "Number of volumes per pod in the MinIO Tenant, e.g. 4")
	f.StringVar(&c.capacity, "capacity", "", "Capacity for each volume, e.g. 1Ti")
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	f.StringVarP(&c.storageClass, "storage-class", "s", "", "Storage Class to be used while PVC creation")
	return cmd
}

func (d *volumeAddCmd) validate() error {
	if d.name == "" {
		return errors.New("--name flag is required for adding volumes to tenant")
	}
	if d.servers == 0 {
		return errors.New("--servers flag is required for adding volumes to tenant")
	}
	if d.volumes == 0 {
		return errors.New("--volumes flag is required for adding volumes to tenant")
	}
	if d.capacity == "" {
		return errors.New("--capacity flag is required for adding volumes to tenant")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *volumeAddCmd) run() error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = addZoneToTenant(oclient, d)
	if err != nil {
		return err
	}
	return nil
}

func addZoneToTenant(client *operatorv1.Clientset, d *volumeAddCmd) error {
	t, err := client.MinioV1().Tenants(d.ns).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	q, err := resource.ParseQuantity(d.capacity)
	if err != nil {
		return err
	}

	t.Spec.Zones = append(t.Spec.Zones, resources.Zone(d.servers, d.volumes, q, d.storageClass))
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	if _, err := client.MinioV1().Tenants(d.ns).Patch(context.Background(), d.name, types.MergePatchType, data, metav1.PatchOptions{FieldManager: "kubectl"}); err != nil {
		return err
	}
	fmt.Printf("Adding new volumes to MinIO Tenant %s\n", t.ObjectMeta.Name)
	return nil
}
