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
	"io/ioutil"
	"log"

	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale/scheme"

	_ "github.com/minio/kubectl-minio/statik"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	kubeConfig string
	namespace  string
	kubeClient *kubernetes.Clientset
	crdObj     *apiextensionv1.CustomResourceDefinition
	crObj      *rbacv1.ClusterRole
)

const (
	minioDesc = `
kubectl plugin to manage MinIO operator CRDs.`
)

func init() {
	fs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	sch := runtime.NewScheme()
	scheme.AddToScheme(sch)
	apiextensionv1.AddToScheme(sch)
	rbacv1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	crd, err := fs.Open("/crd.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contents, err := ioutil.ReadAll(crd)
	if err != nil {
		log.Fatal(err)
	}
	obj, _, err := decode(contents, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	crdObj = obj.(*apiextensionv1.CustomResourceDefinition)

	cr, err := fs.Open("/cluster-role.yaml")
	if err != nil {
		log.Fatal(err)
	}
	contents, err = ioutil.ReadAll(cr)
	if err != nil {
		log.Fatal(err)
	}
	obj, _, err = decode(contents, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	crObj = obj.(*rbacv1.ClusterRole)
}

// NewCmdMinIO creates a new root command for kubectl-minio
func NewCmdMinIO(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "minio",
		Short:        "manage MinIO operator CRDs",
		Long:         minioDesc,
		SilenceUsage: true,
	}

	cmd.AddCommand(newOperatorCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))

	return cmd
}
