// This file is part of MinIO Operator
// Copyright (C) 2020, MinIO, Inc.
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
	"strings"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	upgradeDesc = `'upgrade' command upgrades a MinIO tenant to the specified MinIO version`
)

type upgradeCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	tenantOpts resources.TenantOptions
}

func newTenantUpgradeCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &upgradeCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "upgrade <TENANTNAME> --image <MINIO-IMAGE>",
		Short:   "Upgrade MinIO image for existing tenant",
		Long:    upgradeDesc,
		Example: ` kubectl minio upgrade tenant1 --image quay.io/minio/minio:RELEASE.2023-06-23T20-26-00Z`,
		Args: func(cmd *cobra.Command, args []string) error {
			return c.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c.tenantOpts.Name = args[0]
			err := c.run()
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&c.tenantOpts.Image, "image", "i", "", "image to which tenant is to be upgraded")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", "", "namespace scope for this request")
	f.BoolVarP(&c.output, "output", "o", false, "dry run this command and generate requisite yaml")

	cmd.MarkFlagRequired("image")
	return cmd
}

func (u *upgradeCmd) validate(args []string) error {
	if u.tenantOpts.Image == "" {
		return fmt.Errorf("provide the --image flag, e.g. 'kubectl minio tenant upgrade tenant1 --image %s'", helpers.DefaultTenantImage)
	}
	return validateTenantArgs("upgrade", args)
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (u *upgradeCmd) run() error {
	// Create operator client

	path, _ := rootCmd.Flags().GetString(kubeconfig)
	client, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}

	if u.tenantOpts.NS == "" || u.tenantOpts.NS == helpers.DefaultNamespace {
		u.tenantOpts.NS, err = getTenantNamespace(client, u.tenantOpts.Name)
		if err != nil {
			return err
		}
	}

	imageSplits := strings.Split(u.tenantOpts.Image, ":")
	if len(imageSplits) == 1 {
		return fmt.Errorf("MinIO operator does not allow images without RELEASE tags")
	}

	latest, err := miniov2.ReleaseTagToReleaseTime(imageSplits[1])
	if err != nil {
		return fmt.Errorf("Unsupported release tag, unable to apply requested update %w", err)
	}

	t, err := client.MinioV2().Tenants(u.tenantOpts.NS).Get(context.Background(), u.tenantOpts.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	currentImageSplits := strings.Split(t.Spec.Image, ":")
	if len(currentImageSplits) == 1 {
		return fmt.Errorf("MinIO operator already deployed container with RELEASE tags, update not allowed please manually fix this using 'kubectl patch --help'")
	}
	current, err := miniov2.ReleaseTagToReleaseTime(currentImageSplits[1])
	if err != nil {
		return fmt.Errorf("Unsupported release tag on current image, non-disruptive update not allowed %w", err)
	}
	// Verify if the new release tag is latest, if its not latest refuse to apply the new config.
	if latest.Before(current) {
		return fmt.Errorf("Refusing to downgrade the tenant %s to version %s, from %s",
			u.tenantOpts.Name, u.tenantOpts.Image, t.Spec.Image)
	}

	if u.tenantOpts.ImagePullSecret != "" {
		t.Spec.ImagePullSecret = corev1.LocalObjectReference{Name: u.tenantOpts.ImagePullSecret}
	}

	if !u.output {
		return u.upgradeTenant(client, t, t.Spec.Image, u.tenantOpts.Image)
	}
	// update the image
	t.Spec.Image = u.tenantOpts.Image
	o, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}
	fmt.Println(string(o))
	return nil
}

func (u *upgradeCmd) upgradeTenant(client *operatorv1.Clientset, t *miniov2.Tenant, c, p string) error {
	if helpers.Ask(fmt.Sprintf("Upgrade is a one way process. Are you sure to upgrade Tenant '%s/%s' from version %s to %s", t.ObjectMeta.Name, t.ObjectMeta.Namespace, c, p)) {
		fmt.Printf(Bold(fmt.Sprintf("\nUpgrading Tenant '%s/%s'\n\n", t.ObjectMeta.Name, t.ObjectMeta.Namespace)))
		// update the image
		t.Spec.Image = u.tenantOpts.Image
		if _, err := client.MinioV2().Tenants(t.Namespace).Update(context.Background(), t, v1.UpdateOptions{}); err != nil {
			return err
		}
	} else {
		fmt.Printf(Bold("\nAborting Tenant upgrade\n\n"))
	}
	return nil
}
