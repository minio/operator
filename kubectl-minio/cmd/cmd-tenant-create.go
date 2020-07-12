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
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	operatorv1 "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	createDesc = `
'create' command creates a new MinIO Cluster tenant`
)

type createCmd struct {
	out           io.Writer
	errOut        io.Writer
	ns            string
	image         string
	storageClass  string
	kmsSecret     string
	consoleSecret string
	certSecret    string
	output        bool
	disableTLS    bool
}

func newCreateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &createCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create MinIO Cluster tenant",
		Args:  cobra.MinimumNArgs(6),
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(cmd.Flags()); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	f.StringVarP(&c.image, "image", "i", helpers.DefaultTenantImage, "MinIO Server image")
	f.StringVarP(&c.storageClass, "storage-class", "s", "", "Storage Class to be used while PVC creation")
	f.StringVar(&c.kmsSecret, "kms-secret", "", "Secret with details for KES deployment")
	f.StringVar(&c.consoleSecret, "console-secret", "", "Secret with details for MinIO console deployment")
	f.StringVar(&c.certSecret, "cert-secret", "", "Secret with external certificates for MinIO deployment")
	f.BoolVar(&c.disableTLS, "disable-tls", false, "Disable automatic certificate creation for MinIO peer connection")
	f.BoolVar(&c.output, "output", false, "Output the yaml to be used for this command")

	return cmd
}

// TODO: Add validation for flags
func (c *createCmd) validate(flags *pflag.FlagSet) error {
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *createCmd) run(args []string) error {
	// args[0] --> tenant name
	// args[1] --> access key
	// args[2] --> secret key
	// args[3] --> zone
	// args[4] --> volumes per server
	// args[5] --> capacity per volume
	if !c.output {
		client, err := helpers.GetKubeClient()
		if err != nil {
			return err
		}
		err = createNS(client, c.ns)
		if err != nil {
			return err
		}
		err = createSecret(client, c.ns, args)
		if err != nil {
			return err
		}
		// Create operator client
		oclient, err := helpers.GetKubeOperatorClient()
		if err != nil {
			return err
		}
		err = createTenant(oclient, args, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func createSecret(client *kubernetes.Clientset, ns string, args []string) error {
	secret := resources.NewSecretForMinIOInstance(helpers.SecretName(args[0]), ns, args[1], args[2])
	_, err := client.CoreV1().Secrets(ns).Create(context.TODO(), secret, v1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createTenant(client *operatorv1.Clientset, args []string, c *createCmd) error {
	t, err := resources.NewMinIOInstanceForTenant(args, c.ns, c.image, c.storageClass, c.kmsSecret, c.consoleSecret, c.certSecret, c.disableTLS)
	if err != nil {
		return err
	}
	// set name
	t.ObjectMeta.Name = args[0]
	_, err = client.OperatorV1().MinIOInstances(c.ns).Create(context.TODO(), t, v1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
