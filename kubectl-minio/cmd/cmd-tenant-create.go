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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	createDesc = `
'create' command creates a new MinIO Tenant`
)

type createCmd struct {
	out        io.Writer
	errOut     io.Writer
	tenantOpts resources.TenantOptions
}

func newCreateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &createCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "create --name TENANT_NAME --secret SECRET_NAME --server SERVERS --volumes VOLUMES --capacity VOLUME_CAPACITY",
		Short: "Create a MinIO Tenant",
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.tenantOpts.Name, "name", "", "Name of the MinIO Tenant to be created, e.g. Tenant1")
	f.StringVar(&c.tenantOpts.SecretName, "secret", "", "Name of the Secret object to be used as credentials for MinIO Tenant, e.g. minio-secret")
	f.Int32Var(&c.tenantOpts.Servers, "servers", 0, "Number of pods in the MinIO Tenant, e.g. 4")
	f.Int32Var(&c.tenantOpts.Volumes, "volumes", 0, "Number of volumes per pod in the MinIO Tenant, e.g. 4")
	f.StringVar(&c.tenantOpts.Capacity, "capacity", "", "Capacity for each volume, e.g. 1Ti")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	f.StringVarP(&c.tenantOpts.Image, "image", "i", helpers.DefaultTenantImage, "MinIO Server image")
	f.StringVarP(&c.tenantOpts.StorageClass, "storage-class", "s", "", "Storage Class to be used while PVC creation")
	f.StringVar(&c.tenantOpts.KmsSecret, "kms-secret", "", "Secret with details for KES deployment")
	f.StringVar(&c.tenantOpts.ConsoleSecret, "console-secret", "", "Secret with details for MinIO console deployment")
	f.StringVar(&c.tenantOpts.CertSecret, "cert-secret", "", "Secret with external certificates for MinIO deployment")
	f.BoolVar(&c.tenantOpts.DisableTLS, "disable-tls", false, "Disable automatic certificate creation for MinIO peer connection")

	return cmd
}

func (c *createCmd) validate() error {
	if c.tenantOpts.Name == "" {
		return errors.New("--name flag is required for tenant creation")
	}
	if c.tenantOpts.SecretName == "" {
		return errors.New("--secret flag is required for tenant creation")
	}
	if c.tenantOpts.Servers == 0 {
		return errors.New("--servers flag is required for tenant creation")
	}
	if c.tenantOpts.Volumes == 0 {
		return errors.New("--volumes flag is required for tenant creation")
	}
	if c.tenantOpts.Capacity == "" {
		return errors.New("--capacity flag is required for tenant creation")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *createCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	return createTenant(oclient, c)
}

func createTenant(client *operatorv1.Clientset, c *createCmd) error {
	t, err := resources.NewTenant(&c.tenantOpts)
	if err != nil {
		return err
	}
	// set name
	t.ObjectMeta.Name = c.tenantOpts.Name
	_, err = client.MinioV1().Tenants(c.tenantOpts.NS).Create(context.Background(), t, v1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("MinIO Tenant %s: created\n", t.ObjectMeta.Name)
	return nil
}
