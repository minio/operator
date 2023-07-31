// This file is part of MinIO Operator
// Copyright (C) 2023, MinIO, Inc.
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	getCredentialsDesc = `
'get-credentials' command get credentials from MinIO tenant`
	getCredentialsExample = ` kubectl minio tenant get-credentials tenant1`
)

type getCredentialsCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	namespace  string
	tenantName string
}

func newTenantGetCredentialsCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	v := &getCredentialsCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "get-credentials <TENANTNAME> --namespace <TENANTNS>",
		Short:   "get credentials from existing tenant",
		Long:    getCredentialsDesc,
		Example: getCredentialsExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("provide the name of the tenant, e.g. 'kubectl minio tenant %s tenant1'", cmd)
			}
			v.tenantName = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := v.run()
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&v.namespace, "namespace", "n", "", "k8s namespace for this MinIO tenant")

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (v *getCredentialsCmd) run() error {
	// Create operator client
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	client, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}

	if v.namespace == "" || v.namespace == helpers.DefaultNamespace {
		v.namespace, err = getTenantNamespace(client, v.tenantName)
		if err != nil {
			return err
		}
	}

	t, err := client.MinioV2().Tenants(v.namespace).Get(context.Background(), v.tenantName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if t.Spec.Configuration != nil {
		t, err := client.Discovery().Tenants(v.namespace).Get(context.Background(), t.Spec.Configuration.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
