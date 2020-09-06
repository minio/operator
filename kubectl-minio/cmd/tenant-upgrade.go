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
	"github.com/minio/kubectl-minio/cmd/resources"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	upgradeDesc = `
'upgrade' command upgrades a MinIO tenant to the specified MinIO version`
	upgradeExample = `  kubectl minio tenant upgrade --name tenant1 --image minio/minio:RELEASE.2020-09-05T07-14-49Z --namespace tenant1-ns`
)

type upgradeCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	tenantOpts resources.TenantOptions
}

func newUpgradeCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &upgradeCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade MinIO image for existing tenant",
		Long:  upgradeDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.tenantOpts.Name, "name", "", "name of the MinIO tenant to upgrade")
	f.StringVarP(&c.tenantOpts.Image, "image", "i", helpers.DefaultTenantImage, "image to which tenant is to be upgraded")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVar(&c.tenantOpts.ImagePullSecret, "image-pull-secrets", "", "image pull secret to be used for pulling MinIO image")
	f.BoolVarP(&c.output, "output", "o", false, "dry run this command and generate requisite yaml")

	return cmd
}

func (u *upgradeCmd) validate() error {
	if u.tenantOpts.Name == "" {
		return errors.New("--name flag is required for tenant upgrade")
	}
	if u.tenantOpts.Image == "" {
		return errors.New("--image flag is required for tenant upgrade")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (u *upgradeCmd) run(args []string) error {
	// Create operator client
	client, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}

	t, err := client.MinioV1().Tenants(u.tenantOpts.NS).Get(context.Background(), u.tenantOpts.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	imageSplits := strings.Split(u.tenantOpts.Image, ":")
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
			u.tenantOpts.Name, u.tenantOpts.Image, t.Spec.Image)
	}

	// update the image
	t.Spec.Image = u.tenantOpts.Image
	if u.tenantOpts.ImagePullSecret != "" {
		t.Spec.ImagePullSecret = corev1.LocalObjectReference{Name: u.tenantOpts.ImagePullSecret}
	}

	if !u.output {
		return upgradeTenant(client, t)
	}

	o, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}
	fmt.Println(string(o))
	return nil
}

func upgradeTenant(client *operatorv1.Clientset, t *miniov1.Tenant) error {
	if _, err := client.MinioV1().Tenants(t.Namespace).Update(context.Background(), t, v1.UpdateOptions{}); err != nil {
		return err
	}
	fmt.Printf("Upgrading MinIO Tenant %s to MinIO Image %s\n", t.ObjectMeta.Name, t.Spec.Image)
	return nil
}
