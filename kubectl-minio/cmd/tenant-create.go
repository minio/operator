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
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	createDesc = `
'create' command creates a new MinIO tenant`
	createExample = `  kubectl minio tenant create --name tenant1 --secret tenant1-creds --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns`
)

type createCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	tenantOpts resources.TenantOptions
}

func newCreateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &createCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a MinIO tenant",
		Long:    createDesc,
		Example: createExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.tenantOpts.Name, "name", "", "name of the MinIO tenant to create")
	f.StringVar(&c.tenantOpts.SecretName, "secret", "", "secret name used for tenant credentials")
	f.Int32Var(&c.tenantOpts.Servers, "servers", 0, "total number of pods in MinIO tenant")
	f.Int32Var(&c.tenantOpts.Volumes, "volumes", 0, "total number of volumes in the MinIO tenant")
	f.StringVar(&c.tenantOpts.Capacity, "capacity", "", "total raw capacity of MinIO tenant in this zone, e.g. 16Ti")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&c.tenantOpts.Image, "image", "i", helpers.DefaultTenantImage, "image to be used for MinIO")
	f.StringVarP(&c.tenantOpts.StorageClass, "storage-class", "s", "", "storage class to be used while PVC creation")
	f.StringVar(&c.tenantOpts.KmsSecret, "kms-secret", "", "secret with details for enabled encryption")
	f.StringVar(&c.tenantOpts.ConsoleSecret, "console-secret", "", "secret with details for MinIO console deployment")
	f.StringVar(&c.tenantOpts.CertSecret, "cert-secret", "", "secret with external certificates for MinIO deployment")
	f.StringVar(&c.tenantOpts.ImagePullSecret, "image-pull-secrets", "", "image pull secret to be used for pulling MinIO image")
	f.BoolVar(&c.tenantOpts.DisableTLS, "disable-tls", false, "disable automatic certificate creation for MinIO peer connection")
	f.BoolVarP(&c.output, "output", "o", false, "dry run this command and generate requisite yaml")
	f.StringVar(&c.tenantOpts.ConsoleAnnotations, "console-annotations", "", "json object that contains k/v to be used as annotations for Console pods")
	f.StringVar(&c.tenantOpts.ConsoleLabels, "console-labels", "", "json object that contains k/v to be used as labels for Console pods")
	f.StringVar(&c.tenantOpts.ConsoleNodeSelector, "console-node-selector", "", "json object that contains k/v to be used as node selector for Console pods")
	f.StringVar(&c.tenantOpts.KesAnnotations, "kes-annotations", "", "json object that contains k/v to be used as annotations for Kes pods")
	f.StringVar(&c.tenantOpts.KesLabels, "kes-labels", "", "json object that contains k/v to be used as labels for Kes pods")
	f.StringVar(&c.tenantOpts.KesNodeSelector, "kes-node-selector", "", "json object that contains k/v to be used as node selector for Kes pods")
	f.StringVar(&c.tenantOpts.ServiceAccountName, "service-account", "", "service account name to be used for minio")
	f.StringVar(&c.tenantOpts.ConsoleServiceAccountName, "console-service-account", "", "service account name to be used for console")
	f.StringVar(&c.tenantOpts.KesServiceAccountName, "kes-service-account", "", "service account name to be used for kes")

	return cmd
}

func (c *createCmd) validate() error {
	if c.tenantOpts.SecretName == "" {
		return errors.New("--secret flag is required for tenant creation")
	}
	return c.tenantOpts.Validate()
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *createCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}

	t, err := resources.NewTenant(&c.tenantOpts)
	if err != nil {
		return err
	}
	t.ObjectMeta.Name = c.tenantOpts.Name
	t.ObjectMeta.Namespace = c.tenantOpts.NS

	if !c.output {
		return createTenant(oclient, t)
	}

	o, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}
	fmt.Println(string(o))
	return nil
}

func createTenant(client *operatorv1.Clientset, t *miniov1.Tenant) error {
	_, err := client.MinioV1().Tenants(t.Namespace).Create(context.Background(), t, v1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("MinIO Tenant %s: created\n", t.ObjectMeta.Name)
	return nil
}
