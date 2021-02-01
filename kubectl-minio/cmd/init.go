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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

const (
	operatorInitDesc = `
'init' command creates MinIO Operator deployment along with all the dependencies.`
	operatorInitExample = `  kubectl minio init`
)

type operatorInitCmd struct {
	out          io.Writer
	errOut       io.Writer
	output       bool
	operatorOpts resources.OperatorOptions
	steps        []runtime.Object
}

func newInitCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorInitCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize MinIO Operator",
		Long:    operatorInitDesc,
		Example: operatorInitExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return o.run()
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Image, "image", "i", helpers.DefaultOperatorImage, "operator image")
	f.StringVarP(&o.operatorOpts.Namespace, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&o.operatorOpts.ClusterDomain, "cluster-domain", "d", helpers.DefaultClusterDomain, "cluster domain of the Kubernetes cluster")
	f.StringVar(&o.operatorOpts.NSToWatch, "namespace-to-watch", "", "namespace where operator looks for MinIO tenants, leave empty for all namespaces")
	f.StringVar(&o.operatorOpts.ImagePullSecret, "image-pull-secret", "", "image pull secret to be used for pulling operator image")
	f.BoolVarP(&o.output, "output", "o", false, "dry run this command and generate requisite yaml")

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorInitCmd) run() error {
	sa := resources.NewServiceAccountForOperator(helpers.DefaultServiceAccount, o.operatorOpts.Namespace)
	crb := resources.NewCluterRoleBindingForOperator(helpers.DefaultServiceAccount, o.operatorOpts.Namespace)
	d := resources.NewDeploymentForOperator(o.operatorOpts)
	svc := resources.NewServiceForOperator(o.operatorOpts)
	// Load Resources
	emfs, decode := resources.GetFSAndDecoder()
	crdObj := resources.LoadTenantCRD(emfs, decode)
	crObj := resources.LoadClusterRole(emfs, decode)
	consoleResources := resources.LoadConsoleUI(emfs, decode, &o.operatorOpts)
	if !o.output {
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
		if err = createCRD(extclient, crdObj); err != nil {
			return err
		}
		if err = createCR(client, crObj); err != nil {
			return err
		}
		if err = createSA(client, sa); err != nil {
			return err
		}
		if err = createClusterRB(client, crb); err != nil {
			return err
		}
		if err = createService(client, svc); err != nil {
			return err
		}
		if err = createDeployment(client, d); err != nil {
			return err
		}
		if err = createConsoleResources(o.operatorOpts, extclient, dynclient, consoleResources); err != nil {
			return err
		}
		// since we did an explicit deployment of resources, let's show a message telling users how to connect to console
		fmt.Println("-----------------")
		fmt.Println("")
		fmt.Println("To open Operator UI, start a port forward using this command:")
		fmt.Println("")
		fmt.Println("kubectl minio proxy")
		fmt.Println("")
		fmt.Println("-----------------")
		return nil
	}
	// build yaml output
	o.steps = append(o.steps, crdObj, crObj, sa, crb, d)
	o.steps = append(o.steps, consoleResources...)
	op, err := helpers.ToYaml(o.steps)
	if err != nil {
		return err
	}
	for _, s := range op {
		fmt.Printf(s)
		fmt.Println("---")
	}
	return nil
}

func createConsoleResources(opts resources.OperatorOptions, clientset *apiextension.Clientset, dynClient dynamic.Interface, consoleResources []runtime.Object) error {
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
		uns := unstructured.Unstructured{Object: unstructuredObj}

		switch obj.(type) {
		case *rbacv1.ClusterRoleBinding:
			if err := clusterScopeCreate(dynClient, mapping, ctx, uns); err != nil {
				return err
			}
		case *rbacv1.ClusterRole:
			if err := clusterScopeCreate(dynClient, mapping, ctx, uns); err != nil {
				return err
			}
		default:
			if err := namespaceScopeCreate(opts, dynClient, mapping, ctx, uns); err != nil {
				return err
			}
		}

	}
	fmt.Println("MinIO Console Deployment: created")

	return nil
}

func clusterScopeCreate(dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, uns unstructured.Unstructured) error {
	if _, err := dynClient.Resource(mapping.Resource).Create(ctx, &uns, metav1.CreateOptions{}); err != nil {
		fmt.Println(err)
		return errors.New("Cannot create console resources")
	}
	return nil
}

func namespaceScopeCreate(opts resources.OperatorOptions, dynClient dynamic.Interface, mapping *meta.RESTMapping, ctx context.Context, uns unstructured.Unstructured) error {
	if _, err := dynClient.Resource(mapping.Resource).Namespace(opts.Namespace).Create(ctx, &uns, metav1.CreateOptions{}); err != nil {
		fmt.Println(err)
		return errors.New("Cannot create console resources")
	}
	return nil
}

func createCRD(client *apiextension.Clientset, crd *apiextensionv1.CustomResourceDefinition) error {
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("CustomResourceDefinition %s: already present, skipped", crd.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("CustomResourceDefinition %s: created\n", crd.ObjectMeta.Name)
	return nil
}

func createCR(client *kubernetes.Clientset, cr *rbacv1.ClusterRole) error {
	_, err := client.RbacV1().ClusterRoles().Create(context.Background(), cr, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("ClusterRole %s: already present, skipped", cr.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("ClusterRole %s: created\n", cr.ObjectMeta.Name)
	return nil
}

func createSA(client *kubernetes.Clientset, sa *corev1.ServiceAccount) error {
	_, err := client.CoreV1().ServiceAccounts(sa.ObjectMeta.Namespace).Create(context.Background(), sa, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("ServiceAccount %s: already present, skipped", sa.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("ServiceAccount %s: created\n", sa.ObjectMeta.Name)
	return nil
}

func createClusterRB(client *kubernetes.Clientset, crb *rbacv1.ClusterRoleBinding) error {
	_, err := client.RbacV1().ClusterRoleBindings().Create(context.Background(), crb, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("ClusterRoleBinding %s: already present, skipped", crb.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("ClusterRoleBinding %s: created\n", crb.ObjectMeta.Name)
	return nil
}

func createDeployment(client *kubernetes.Clientset, d *appsv1.Deployment) error {
	_, err := client.AppsV1().Deployments(d.ObjectMeta.Namespace).Create(context.Background(), d, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("MinIO Operator Deployment %s: already present, skipped", d.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("MinIO Operator Deployment %s: created\n", d.ObjectMeta.Name)
	return nil
}

func createService(client *kubernetes.Clientset, svc *corev1.Service) error {
	_, err := client.CoreV1().Services(svc.ObjectMeta.Namespace).Create(context.Background(), svc, v1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("MinIO Operator Service %s: already present, skipped", svc.ObjectMeta.Name)
		}
		return err
	}
	fmt.Printf("MinIO Operator Service %s: created\n", svc.ObjectMeta.Name)
	return nil
}
