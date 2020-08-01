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

	"github.com/markbates/pkger"
	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	operatorCreateDesc = `
'create' command creates MinIO Operator deployment along with all the dependencies.`
)

type operatorCreateCmd struct {
	out            io.Writer
	errOut         io.Writer
	image          string
	ns             string
	nsToWatch      string
	clusterDomain  string
	serviceAccount string
}

func newOperatorCreateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	i := &operatorCreateCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create MinIO Operator deployment",
		Long:  operatorCreateDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return i.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&i.image, "image", "i", helpers.DefaultOperatorImage, "MinIO Operator image")
	f.StringVarP(&i.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request ")
	f.StringVarP(&i.serviceAccount, "service-account", "", helpers.DefaultServiceAccount, "ServiceAccount for MinIO Operator")
	f.StringVarP(&i.clusterDomain, "cluster-domain", "d", helpers.DefaultClusterDomain, "Cluster domain of the Kubernetes cluster")
	f.StringVarP(&i.nsToWatch, "namespace-to-watch", "", "", "Namespace where MinIO Operator looks for MinIO Instances, leave empty for all namespaces")

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (i *operatorCreateCmd) run() error {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = apiextensionv1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	crd, err := pkger.Open("/static/crd.yaml")
	if err != nil {
		return err
	}
	defer crd.Close()
	obj, _, err := decode(helpers.StreamToByte(crd), nil, nil)
	if err != nil {
		return err
	}

	crdObj := obj.(*apiextensionv1.CustomResourceDefinition)

	cr, err := pkger.Open("/static/cluster-role.yaml")
	if err != nil {
		return err
	}
	defer cr.Close()
	obj, _, err = decode(helpers.StreamToByte(cr), nil, nil)
	if err != nil {
		return err
	}

	crObj := obj.(*rbacv1.ClusterRole)
	sa := resources.NewServiceAccountForOperator(i.serviceAccount, i.ns)
	crb := resources.NewCluterRoleBindingForOperator(i.ns, i.serviceAccount)
	d := resources.NewDeploymentForOperator(i.ns, i.serviceAccount, i.image, i.clusterDomain, i.nsToWatch)

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

func createCRD(client *apiextension.Clientset, crd *apiextensionv1.CustomResourceDefinition) error {
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.Background(), crd, v1.CreateOptions{})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			fmt.Printf("CustomResourceDefinition %s: already present, skipped\n", crd.ObjectMeta.Name)
			return nil
		}
		return err
	}
	fmt.Printf("CustomResourceDefinition %s: created\n", crd.ObjectMeta.Name)
	return nil
}

func createCR(client *kubernetes.Clientset, cr *rbacv1.ClusterRole) error {
	_, err := client.RbacV1().ClusterRoles().Create(context.Background(), cr, v1.CreateOptions{})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			fmt.Printf("ClusterRole %s: already present, skipped\n", cr.ObjectMeta.Name)
			return nil
		}
		return err
	}
	fmt.Printf("ClusterRole %s: created\n", cr.ObjectMeta.Name)
	return nil
}

func createSA(client *kubernetes.Clientset, sa *corev1.ServiceAccount) error {
	_, err := client.CoreV1().ServiceAccounts(sa.ObjectMeta.Namespace).Create(context.Background(), sa, v1.CreateOptions{})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			fmt.Printf("ServiceAccount %s: already present, skipped\n", sa.ObjectMeta.Name)
			return nil
		}
		return err
	}
	fmt.Printf("ServiceAccount %s: created\n", sa.ObjectMeta.Name)
	return nil
}

func createClusterRB(client *kubernetes.Clientset, crb *rbacv1.ClusterRoleBinding) error {
	_, err := client.RbacV1().ClusterRoleBindings().Create(context.Background(), crb, v1.CreateOptions{})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			fmt.Printf("ClusterRoleBinding %s: already present, skipped\n", crb.ObjectMeta.Name)
			return nil
		}
		return err
	}
	fmt.Printf("ClusterRoleBinding %s: created\n", crb.ObjectMeta.Name)
	return nil
}

func createDeployment(client *kubernetes.Clientset, d *appsv1.Deployment) error {
	_, err := client.AppsV1().Deployments(d.ObjectMeta.Namespace).Create(context.Background(), d, v1.CreateOptions{})
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			fmt.Printf("MinIO Operator Deployment %s: already present, skipped\n", d.ObjectMeta.Name)
			return nil
		}
		return err
	}
	fmt.Printf("MinIO Operator Deployment %s: created\n", d.ObjectMeta.Name)
	return nil
}
