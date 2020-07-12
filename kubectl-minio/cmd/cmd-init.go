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
	"io"

	"github.com/markbates/pkger"
	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	initDesc = `
'init' creates MinIO Operator deployment alongwith all the dependencies. It discovers Kubernetes clusters by reading $KUBECONFIG (default '~/.kube/config') and using the default context. When installing  MinIO Operator, 'minio init' will attempt to install the latest released version. You can specify an alternative image with '--image' which is the fully qualified image name replacement. To dump a manifest containing the deployment YAML, combine the '--dry-run' and '--o' flags.`
)

type initCmd struct {
	out            io.Writer
	errOut         io.Writer
	image          string
	output         bool
	ns             string
	nsToWatch      string
	clusterDomain  string
	serviceAccount string
}

func newInitCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	i := &initCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create MinIO Operator deployment",
		Long:  initDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			if err := i.validate(cmd.Flags()); err != nil {
				return err
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
	f.BoolVar(&i.output, "output", false, "Output the yaml to be used for this command")

	return cmd
}

// TODO: Add validation for flags
func (i *initCmd) validate(flags *pflag.FlagSet) error {
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (i *initCmd) run() error {
	if !i.output {
		sch := runtime.NewScheme()
		_ = scheme.AddToScheme(sch)
		_ = apiextensionv1.AddToScheme(sch)
		decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

		client, err := helpers.GetKubeClient()
		if err != nil {
			return err
		}
		extclient, err := helpers.GetKubeExtensionClient()
		if err != nil {
			return err
		}

		// Create static components - CRD and ClusterRole
		crd, err := pkger.Open("/static/crd.yaml")
		if err != nil {
			return err
		}
		defer crd.Close()
		obj, _, err := decode(helpers.StreamToByte(crd), nil, nil)
		if err != nil {
			return err
		}
		err = createCRD(extclient, obj)
		if err != nil {
			return err
		}

		cr, err := pkger.Open("/static/cluster-role.yaml")
		if err != nil {
			return err
		}
		defer cr.Close()
		obj, _, err = decode(helpers.StreamToByte(cr), nil, nil)
		if err != nil {
			return err
		}
		err = createCR(client, obj)
		if err != nil {
			return err
		}

		// Create remaining components
		err = createNS(client, i.ns)
		if err != nil {
			return err
		}
		err = createSA(client, i)
		if err != nil {
			return err
		}
		err = createClusterRB(client, i)
		if err != nil {
			return err
		}
		err = createDeployment(client, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCRD(client *apiextension.Clientset, obj runtime.Object) error {
	crd := obj.(*apiextensionv1.CustomResourceDefinition)
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func createCR(client *kubernetes.Clientset, obj runtime.Object) error {
	cr := obj.(*rbacv1.ClusterRole)
	_, err := client.RbacV1().ClusterRoles().Create(context.TODO(), cr, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func createSA(client *kubernetes.Clientset, i *initCmd) error {
	sa := resources.NewServiceAccountForOperator(i.serviceAccount, i.ns)
	_, err := client.CoreV1().ServiceAccounts(i.ns).Create(context.TODO(), sa, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func createClusterRB(client *kubernetes.Clientset, i *initCmd) error {
	crb := resources.NewCluterRoleBindingForOperator(i.ns, i.serviceAccount)
	_, err := client.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func createDeployment(client *kubernetes.Clientset, i *initCmd) error {
	d := resources.NewDeploymentForOperator(i.ns, i.serviceAccount, i.image, i.clusterDomain, i.nsToWatch)
	_, err := client.AppsV1().Deployments(i.ns).Create(context.TODO(), d, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func createNS(client *kubernetes.Clientset, ns string) error {
	n := resources.NewNamespaceForOperator(ns)
	_, err := client.CoreV1().Namespaces().Create(context.TODO(), n, v1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
