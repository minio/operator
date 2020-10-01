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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	deleteDesc = `
'delete' command delete MinIO Operator deployment along with all the tenants.`
	deleteExample = `  kubectl minio delete`
)

type deleteCmd struct {
	out          io.Writer
	errOut       io.Writer
	output       bool
	operatorOpts resources.OperatorOptions
	steps        []runtime.Object
}

func newDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &deleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete MinIO Operator deployment",
		Long:    deleteDesc,
		Example: deleteExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Are you sure you want to delete ALL the MinIO Tenants and MinIO Operator? [Y/N]: ")
			if !o.ask() {
				return errors.New("Aborting Operator deletion")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return o.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")

	return cmd
}

func (o *deleteCmd) ask() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, _ := reader.ReadString('\n')
		s = strings.TrimSuffix(s, "\n")
		s = strings.ToLower(s)
		if len(s) > 1 {
			fmt.Fprintln(os.Stderr, "Please enter Y or N")
			continue
		}
		if strings.Compare(s, "n") == 0 {
			return false
		} else if strings.Compare(s, "y") == 0 {
			break
		} else {
			continue
		}
	}
	return true
}

func (o *deleteCmd) run() error {
	client, err := helpers.GetKubeClient()
	if err != nil {
		return err
	}
	extclient, err := helpers.GetKubeExtensionClient()
	if err != nil {
		return err
	}
	if err := extclient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), crdObj.Name, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.RbacV1().ClusterRoles().Delete(context.Background(), crObj.Name, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.CoreV1().ServiceAccounts(o.operatorOpts.NS).Delete(context.Background(), helpers.DefaultServiceAccount, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.RbacV1().ClusterRoleBindings().Delete(context.Background(), helpers.ClusterRoleBindingName, v1.DeleteOptions{}); err != nil {
		return err
	}
	return client.AppsV1().Deployments(o.operatorOpts.NS).Delete(context.Background(), helpers.DeploymentName, v1.DeleteOptions{})
}
