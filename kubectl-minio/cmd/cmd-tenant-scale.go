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
	"fmt"
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1 "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	scaleDesc = `
'scale' command adds zones to a MinIO Cluster tenant`
)

type scaleCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
}

func newScaleCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &scaleCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Scale MinIO tenant zones",
		Args:  cobra.MinimumNArgs(2),
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *scaleCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = addZoneToTenant(oclient, args, d)
	if err != nil {
		return err
	}
	return nil
}

func addZoneToTenant(client *operatorv1.Clientset, args []string, d *scaleCmd) error {
	t, err := client.OperatorV1().MinIOInstances(d.ns).Get(context.TODO(), args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	// add zones
	z, err := helpers.ParseZones(args[1])
	if err != nil {
		return err
	}

	t.Spec.Zones = append(t.Spec.Zones, z)
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	fmt.Println(data)
	if _, err := client.OperatorV1().MinIOInstances(d.ns).Patch(context.TODO(), args[0], types.MergePatchType, data, v1.PatchOptions{FieldManager: "kubectl"}); err != nil {
		return err
	}
	return nil
}
