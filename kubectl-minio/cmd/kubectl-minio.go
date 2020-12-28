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
	"log"

	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/scale/scheme"

	// Statik CRD assets for our plugin
	"github.com/minio/kubectl-minio/cmd/helpers"
	_ "github.com/minio/kubectl-minio/statik"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
Deploy and manage the multi tenant, S3 API compatible object storage on Kubernetes`
	kubeconfig = "kubeconfig"
)

var confPath string
var rootCmd = &cobra.Command{
	Use:          "minio",
	Short:        "manage MinIO operator CRDs",
	Long:         minioDesc,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&confPath, kubeconfig, "", "Custom kubeconfig path")

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	emfs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	sch := runtime.NewScheme()
	scheme.AddToScheme(sch)
	apiextensionv1.AddToScheme(sch)
	rbacv1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	contents, err := fs.ReadFile(emfs, "/crds/minio.min.io_tenants.yaml")
	if err != nil {
		log.Fatal(err)
	}

	obj, _, err := decode(contents, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	var ok bool
	crdObj, ok = obj.(*apiextensionv1.CustomResourceDefinition)
	if !ok {
		log.Fatal("Unable to locate CustomResourceDefinition object")
	}

	contents, err = fs.ReadFile(emfs, "/cluster-role.yaml")
	if err != nil {
		log.Fatal(err)
	}

	obj, _, err = decode(contents, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	crObj, ok = obj.(*rbacv1.ClusterRole)
	if !ok {
		log.Fatal("Unable to locate ClusterRole object")
	}

}

// NewCmdMinIO creates a new root command for kubectl-minio
func NewCmdMinIO(streams genericclioptions.IOStreams) *cobra.Command {
	rootCmd = helpers.DisableHelp(rootCmd)
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(newInitCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	rootCmd.AddCommand(newTenantCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	rootCmd.AddCommand(newDeleteCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	return rootCmd
}
