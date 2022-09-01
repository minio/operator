// This file is part of MinIO Operator
// Copyright (C) 2020 MinIO, Inc.
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

	"github.com/minio/kubectl-minio/cmd/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

type tenantDeleteCmd struct {
	out    io.Writer
	errOut io.Writer
	ns     string
	force  bool
}

func newTenantDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &tenantDeleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "delete <TENANTNAME> --namespace <TENANTNS>",
		Short:   "Delete a MinIO tenant",
		Long:    deleteDesc,
		Example: deleteExample,
		Args: func(cmd *cobra.Command, args []string) error {
			return c.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !c.force {
				if !helpers.Ask(fmt.Sprintf("This will delete the Tenant %s and ALL its data. Do you want to proceed", args[0])) {
					return fmt.Errorf(Bold("Aborting Tenant deletion"))
				}
			}
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
	f.BoolVarP(&c.force, "force", "f", false, "force delete the tenant")
	cmd.MarkFlagRequired("namespace")

	return cmd
}

func (d *tenantDeleteCmd) validate(args []string) error {
	return validateTenantArgs("delete", args)
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *tenantDeleteCmd) run(args []string) error {
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	oclient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}
	kclient, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}
	for _, arg := range args {
		if err = deleteTenant(oclient, kclient, d, arg); err != nil {
			return err
		}
	}
	return nil
}

func deleteTenant(client *operatorv1.Clientset, kclient *kubernetes.Clientset, d *tenantDeleteCmd, name string) error {
	tenant, err := client.MinioV2().Tenants(d.ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := client.MinioV2().Tenants(d.ns).Delete(context.Background(), name, v1.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Println("Deleting MinIO Tenant: ", name)

	if tenant.HasConfigurationSecret() {
		kclient.CoreV1().Secrets(d.ns).Delete(context.Background(), tenant.Spec.Configuration.Name,
			metav1.DeleteOptions{})
		fmt.Println("Deleting MinIO Tenant Configuration Secret: ", tenant.Spec.Configuration.Name)
	}

	// Delete all users, ignore any errors.
	for _, user := range tenant.Spec.Users {
		kclient.CoreV1().Secrets(d.ns).Delete(context.Background(), user.Name, metav1.DeleteOptions{})
		fmt.Println("Deleting MinIO Tenant user: ", user.Name)
	}

	return nil
}
