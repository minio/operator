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
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	kubeConfig string
	namespace  string
	kubeClient *kubernetes.Clientset
)

// NewCmdMinIO creates a new root command for kubectl-minio
func NewCmdMinIO(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "minio",
		Short:        "manage MinIO operator CRDs",
		Long:         `MinIO Operator plugin to manage MinIO operator CRDs`,
		SilenceUsage: true,
		Example: `  # Install MinIO Operator from the official GitHub repo.
	kubectl minio init [flags]`,
	}

	cmd.AddCommand(newOperatorCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))
	cmd.AddCommand(newTenantCmd(cmd.OutOrStdout(), cmd.ErrOrStderr()))

	return cmd
}
