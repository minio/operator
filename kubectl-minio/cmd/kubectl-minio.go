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

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	// Workaround for auth import issues refer https://github.com/minio/operator/issues/283
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	// Statik CRD assets for our plugin
	"github.com/minio/kubectl-minio/cmd/helpers"
	_ "github.com/minio/kubectl-minio/statik"
)

var (
	kubeConfig string
	namespace  string
	kubeClient *kubernetes.Clientset
)

const (
	minioDesc  = `Deploy and manage the multi tenant, S3 API compatible object storage on Kubernetes`
	kubeconfig = "kubeconfig"
)

var confPath string
var rootCmd = &cobra.Command{
	Use:          "minio",
	Long:         minioDesc,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&confPath, kubeconfig, "", "Custom kubeconfig path")

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

}

// NewCmdMinIO creates a new root command for kubectl-minio
func NewCmdMinIO(streams genericclioptions.IOStreams) *cobra.Command {
	rootCmd = helpers.DisableHelp(rootCmd)
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(newInitCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	rootCmd.AddCommand(newTenantCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	rootCmd.AddCommand(newDeleteCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	rootCmd.AddCommand(newProxyCmd(rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()))
	return rootCmd
}
