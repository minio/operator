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
	"time"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

const (
	getCredentialsDesc = `
'get-credentials' command get credentials from MinIO tenant`
	getCredentialsExample = ` kubectl minio tenant get-credentials`
)

type getCredentialsCmd struct {
	out       io.Writer
	errOut    io.Writer
	output    bool
	namespace string
	name      string
}

func newTenantGetCredentialsCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	v := &getCredentialsCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "get-credentials <TENANTNAME> --namespace <TENANTNS>",
		Short:   "get credentials from existing tenant",
		Long:    getCredentialsDesc,
		Example: getCredentialsExample,
		Args: func(cmd *cobra.Command, args []string) error {
			return v.validate(args)
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

func (v *getCredentialsCmd) validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("provide the name of the tenant, e.g. 'kubectl minio tenant get-credentials'")
	}
	// Tenant name should have DNS token restrictions
	if err := helpers.CheckValidTenantName(args[0]); err != nil {
		return err
	}
	v.name = args[0]
	return nil
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
		v.namespace, err = getTenantNamespace(client, v.name)
		if err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t, err := client.MinioV2().Tenants(v.namespace).Get(ctx, v.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if t.Spec.Configuration != nil {
		result := &v1.Secret{}
		err = client.RESTClient().Get().Namespace(v.namespace).
			Resource("secrets").
			Name(t.Spec.Configuration.Name).
			VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
			Do(ctx).
			Into(result)
		if err != nil {
			return err
		}
		fmt.Println(string(result.Data["config.env"]))
	}

	return nil
}
