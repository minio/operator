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

	"github.com/minio/kubectl-minio/cmd/resources"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tenantDesc = `
'tenant' is the top level command for managing MinIO tenants created via MinIO operator.`
)

func getTenantNamespace(client *operatorv1.Clientset, tenantName string) (string, error) {
	tenants, err := client.MinioV2().Tenants("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, tenant := range tenants.Items {
		if tenant.Name == tenantName {
			return tenant.ObjectMeta.Namespace, nil
		}
	}
	return "", fmt.Errorf("tenant: %s not found on any namespace", tenantName)
}

func newTenantCmd(_ io.Writer, _ io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage MinIO tenant(s)",
		Long:  tenantDesc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			path, _ := rootCmd.Flags().GetString(kubeconfig)
			client, err := helpers.GetKubeExtensionClient(path)
			if err != nil {
				return err
			}
			// Load Resources
			decode := resources.GetSchemeDecoder()
			crdObj := resources.LoadTenantCRD(decode)
			_, err = client.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), crdObj.GetObjectMeta().GetName(), v1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("CustomResourceDefinition %s: not found, please run 'kubectl minio init' before using tenant command", crdObj.ObjectMeta.Name)
				}
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	cmd.AddCommand(newTenantCreateCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantInfoCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantListCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantExpandCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantUpgradeCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantDeleteCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantReportCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantStatusCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantEventsCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))

	return cmd
}
