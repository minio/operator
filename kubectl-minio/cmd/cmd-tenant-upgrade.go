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
	"strings"

	"github.com/minio/kubectl-minio/cmd/helpers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	upgradeDesc = `
'upgrade' command upgrades a MinIO Cluster tenant to the specified MinIO version`
)

type upgradeCmd struct {
	out    io.Writer
	errOut io.Writer
	name   string
	image  string
	ns     string
}

func newUpgradeCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &upgradeCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "upgrade --name TENANT_NAME --image MINIO_IMAGE",
		Short: "Upgrade Container Image for existing MinIO Tenant ",
		Long:  upgradeDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.name, "name", "", "Name of the MinIO Tenant to be upgraded, e.g. Tenant1")
	f.StringVarP(&c.image, "image", "i", helpers.DefaultTenantImage, "Target container image to which MinIO Tenant is to be upgraded")
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "If present, the namespace scope for this request")
	return cmd
}

func (c *upgradeCmd) validate() error {
	if c.name == "" {
		return errors.New("--name flag is required for tenant upgrade")
	}
	if c.image == "" {
		return errors.New("--image flag is required for tenant upgrade")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *upgradeCmd) run(args []string) error {
	// Create operator client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	err = upgradeTenant(oclient, c)
	if err != nil {
		return err
	}
	return nil
}

func upgradeTenant(client *operatorv1.Clientset, c *upgradeCmd) error {
	t, err := client.MinioV1().Tenants(c.ns).Get(context.Background(), c.name, v1.GetOptions{})
	if err != nil {
		return err
	}

	imageSplits := strings.Split(c.image, ":")
	if len(imageSplits) == 1 {
		return fmt.Errorf("MinIO operator does not allow images without RELEASE tags")
	}

	latest, err := miniov1.ReleaseTagToReleaseTime(imageSplits[1])
	if err != nil {
		return fmt.Errorf("Unsupported release tag, unable to apply requested update %w", err)
	}

	currentImageSplits := strings.Split(t.Spec.Image, ":")
	if len(currentImageSplits) == 1 {
		return fmt.Errorf("MinIO operator already deployed container with RELEASE tags, update not allowed please manually fix this using 'kubectl patch --help'")
	}

	current, err := miniov1.ReleaseTagToReleaseTime(currentImageSplits[1])
	if err != nil {
		return fmt.Errorf("Unsupported release tag on current image, non-disruptive update not allowed %w", err)
	}

	// Verify if the new release tag is latest, if its not latest refuse to apply the new config.
	if latest.Before(current) {
		return fmt.Errorf("Refusing to downgrade the tenant %s to version %s, from %s",
			c.name, c.image, t.Spec.Image)
	}

	// update the image
	t.Spec.Image = c.image
	if _, err := client.MinioV1().Tenants(c.ns).Update(context.Background(), t, v1.UpdateOptions{}); err != nil {
		return err
	}
	fmt.Printf("Upgrading MinIO Tenant %s to MinIO Image %s\n", t.ObjectMeta.Name, t.Spec.Image)
	return nil
}
