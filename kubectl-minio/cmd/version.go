// This file is part of MinIO Operator
// Copyright (C) 2021, MinIO, Inc.
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
	"os"
	"strings"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/spf13/cobra"
)

// version provides the version of this plugin
var version = "DEVELOPMENT.GOGET"

const (
	operatorVersionDesc = `
'version' command displays the kubectl plugin version.`
	operatorVersionExample = `  kubectl minio version`
)

type operatorVersionCmd struct {
	out    io.Writer
	errOut io.Writer
}

func newVersionCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	o := &operatorVersionCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Display plugin version",
		Long:    operatorVersionDesc,
		Example: operatorVersionExample,
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := rootCmd.Flags().GetString(kubeconfig)
			if path != "" {
				os.Setenv(clientcmd.RecommendedConfigPathEnvVar, path)
			}
			err := o.run()
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)

	return cmd
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (o *operatorVersionCmd) run() error {
	fmt.Println("Kubectl-Plugin Version:", version)
	cfg := config.GetConfigOrDie()
	// If config is passed as a flag use that instead
	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}
	deployList := &v1.DeploymentList{}
	listOpt := &client.ListOptions{}
	client.MatchingLabels{"app.kubernetes.io/name": "minio-operator"}.ApplyToList(listOpt)
	err = k8sClient.List(context.Background(), deployList, listOpt)
	if err != nil {
		return err
	}
	for _, item := range deployList.Items {
		image := ""
		if len(item.Spec.Template.Spec.Containers) == 1 {
			image = item.Spec.Template.Spec.Containers[0].Image
		} else {
			for _, container := range item.Spec.Template.Spec.Containers {
				if strings.Contains(container.Image, "operator") {
					image = container.Image
					break
				}
			}
		}
		if image != "" {
			if item.Name == "console" {
				fmt.Printf("Minio-Operator Console Version: %s/%s:%s \r\n", item.Namespace, item.Name, strings.SplitN(image, ":", 2)[1])
			}
			if item.Name == "minio-operator" {
				fmt.Printf("Minio-Operator Controller Version: %s/%s:%s \r\n", item.Namespace, item.Name, strings.SplitN(image, ":", 2)[1])
			}
		}
	}
	return nil
}
