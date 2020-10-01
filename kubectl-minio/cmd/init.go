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

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

const (
	operatorInitDesc = `
'init' command creates MinIO Operator deployment along with all the dependencies.`
	operatorInitExample = `  kubectl minio operator init`
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
		Short:   "Initialize MinIO Operator deployment",
		Long:    operatorInitDesc,
		Example: operatorInitExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return o.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.operatorOpts.Image, "image", "i", helpers.DefaultOperatorImage, "operator image")
	f.StringVarP(&o.operatorOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&o.operatorOpts.ClusterDomain, "cluster-domain", "d", helpers.DefaultClusterDomain, "cluster domain of the Kubernetes cluster")
	f.StringVar(&o.operatorOpts.NSToWatch, "namespace-to-watch", "", "namespace where operator looks for MinIO tenants, leave empty for all namespaces")
	f.StringVar(&o.operatorOpts.ImagePullSecret, "image-pull-secret", "", "image pull secret to be used for pulling operator image")
	f.BoolVarP(&o.output, "output", "o", false, "dry run this command and generate requisite yaml")

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorInitCmd) run() error {
	sa := resources.NewServiceAccountForOperator(helpers.DefaultServiceAccount, o.operatorOpts.NS)
	crb := resources.NewCluterRoleBindingForOperator(helpers.DefaultServiceAccount, o.operatorOpts.NS)
	d := resources.NewDeploymentForOperator(o.operatorOpts)

	if !o.output {
		client, err := helpers.GetKubeClient()
		if err != nil {
			return err
		}
		extclient, err := helpers.GetKubeExtensionClient()
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
		return createDeployment(client, d)
	}

	o.steps = append(o.steps, crdObj, crObj, sa, crb, d)
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

func createCRD(client *apiextension.Clientset, crd *apiextensionv1.CustomResourceDefinition) error {
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.Background(), crd, v1.CreateOptions{})
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
