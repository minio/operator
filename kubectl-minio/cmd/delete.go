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
	"errors"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"

	rbacv1 "k8s.io/api/rbac/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/minio/minio/pkg/color"
	"github.com/spf13/cobra"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	deleteDesc = `
'delete' command delete MinIO Operator along with all the tenants.`
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
		Short:   "Delete MinIO Operator",
		Long:    deleteDesc,
		Example: deleteExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !helpers.Ask(fmt.Sprintf("Are you sure you want to delete ALL the MinIO Tenants and MinIO Operator?")) {
				return fmt.Errorf(color.Bold("Aborting Operator deletion\n"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("delete command does not accept arguments")
			}
			return o.run()
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	return cmd
}

func (o *deleteCmd) run() error {
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	client, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}
	extclient, err := helpers.GetKubeExtensionClient()
	if err != nil {
		return err
	}
	dynclient, err := helpers.GetKubeDynamicClient()
	if err != nil {
		return err
	}
	// Load Resources
	emfs, decode := resources.GetFSAndDecoder()
	crdObj := resources.LoadTenantCRD(emfs, decode)
	if err := extclient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), crdObj.Name, v1.DeleteOptions{}); err != nil {
		return err
	}
	crObj := resources.LoadClusterRole(emfs, decode)
	if err := client.RbacV1().ClusterRoles().Delete(context.Background(), crObj.Name, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.CoreV1().ServiceAccounts(o.operatorOpts.Namespace).Delete(context.Background(), helpers.DefaultServiceAccount, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.RbacV1().ClusterRoleBindings().Delete(context.Background(), helpers.ClusterRoleBindingName, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := client.AppsV1().Deployments(o.operatorOpts.Namespace).Delete(context.Background(), helpers.DeploymentName, v1.DeleteOptions{}); err != nil {
		return err
	}
	consoleResources := resources.LoadConsoleUI(emfs, decode, &o.operatorOpts)
	if err := deleteConsoleResources(o.operatorOpts, extclient, dynclient, consoleResources); err != nil {
		return err
	}
	return nil
}

func deleteConsoleResources(opts resources.OperatorOptions, clientset *apiextension.Clientset, dynClient dynamic.Interface, consoleResources []runtime.Object) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		fmt.Println(err)
		return errors.New("Cannot get group resources.")
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	for _, obj := range consoleResources {
		gvk := obj.GetObjectKind().GroupVersionKind()
		gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}

		mapping, err := rm.RESTMapping(gk, gvk.Version)

		// convert the runtime.Object to unstructured.Unstructured
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		var resourceName string
		if metaobj, ok := unstructuredObj["metadata"]; ok {
			mtobj := metaobj.(map[string]interface{})
			if name, ok2 := mtobj["name"]; ok2 {
				resourceName = name.(string)
			}
		}

		switch obj.(type) {
		case *rbacv1.ClusterRoleBinding:
			if err := clusterScopeDelete(dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		case *rbacv1.ClusterRole:
			if err := clusterScopeDelete(dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		default:
			if err := namespaceScopeDelete(opts, dynClient, mapping, ctx, resourceName); err != nil {
				return err
			}
		}
	}
	return nil
}

func clusterScopeDelete(dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, name string) error {
	if err := dynClient.Resource(mapping.Resource).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
		return errors.New("Cannot delete console resources")
	}
	return nil
}

func namespaceScopeDelete(opts resources.OperatorOptions, dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, name string) error {
	if err := dynClient.Resource(mapping.Resource).Namespace(opts.Namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		fmt.Println(err)
		return errors.New("Cannot delete console resources")
	}
	return nil
}
