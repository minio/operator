// This file is part of MinIO Operator
// Copyright (C) 2022, MinIO, Inc.
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
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/madmin-go"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/subnet"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

const (
	licenseRegisterDesc = `'register' command registers Operator tenants on SUBNET`
)

type licenseRegisterCmd struct {
	out    io.Writer
	errOut io.Writer
	apiKey string
}

func newLicenseRegisterCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &licenseRegisterCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "register",
		Short:   "Register tenants with MinIO Subscription Network",
		Long:    licenseRegisterDesc,
		Example: `  kubectl minio license register`,
		Args: func(cmd *cobra.Command, args []string) error {
			return c.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.run(args)
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVar(&c.apiKey, "api-key", "", "SUBNET API key")
	return cmd
}

func (d *licenseRegisterCmd) validate(args []string) error {
	if len(args) > 0 {
		return errors.New("register command does not support arguments, e.g. 'kubectl minio license register'")
	}
	return nil
}

func (d *licenseRegisterCmd) run(args []string) error {
	if err := subnet.CheckURLReachable(subnet.BaseURL()); err != nil {
		return err
	}

	if err := d.validateAPIKey(); err != nil {
		return err
	}
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	oclient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}
	kubeClient, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}
	tenants, err := oclient.MinioV2().Tenants("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	return d.registerTenants(tenants.Items, kubeClient)
}

func (d *licenseRegisterCmd) registerTenants(tenants []v2.Tenant, kubeClient *kubernetes.Clientset) error {
	ctx := context.Background()
	for _, tenant := range tenants {
		fmt.Printf("Registering tenant %s\n", tenant.Name)
		if err := d.registerTenant(ctx, tenant, kubeClient); err != nil {
			fmt.Printf("Failed to register tenant %s: %s\n", tenant.Name, err.Error())
			continue
		}
		fmt.Printf("Tenant %s registered\n", tenant.Name)
	}
	return nil
}

func (d *licenseRegisterCmd) registerTenant(ctx context.Context, tenant v2.Tenant, kubeClient *kubernetes.Clientset) error {
	adminClient, err := d.getTenantAdminClient(ctx, tenant, kubeClient)
	if err != nil {
		return err
	}
	serverInfo, err := adminClient.ServerInfo(ctx)
	if err != nil {
		return err
	}
	license, err := subnet.RegisterWithAPIKey(serverInfo, d.apiKey)
	if err != nil {
		return err
	}
	err = d.updateSubnetConfig(ctx, license, adminClient)
	if err != nil {
		return err
	}
	return d.createSubnetAPIKeySecret(ctx, kubeClient)
}

func (d *licenseRegisterCmd) getTenantAdminClient(ctx context.Context, tenant v2.Tenant, kubeClient *kubernetes.Clientset) (*madmin.AdminClient, error) {
	tenantConfiguration, err := v2.GetTenantConfiguration(ctx, &tenant, kubeClient)
	if err != nil {
		return nil, err
	}
	return tenant.NewMinIOAdmin(tenantConfiguration, subnet.PrepareClientTransport(true))
}

func (d *licenseRegisterCmd) validateAPIKey() error {
	if d.apiKey != "" {
		_, err := uuid.Parse(d.apiKey)
		return err
	}
	return d.getAPIKeyFromSubnet()
}

func (d *licenseRegisterCmd) getAPIKeyFromSubnet() error {
	token, err := subnet.Login()
	if err != nil {
		return err
	}
	apiKey, err := subnet.GetAPIKey(token)
	if err != nil {
		return err
	}
	d.apiKey = apiKey
	return nil
}

func (d *licenseRegisterCmd) updateSubnetConfig(ctx context.Context, license *subnet.LicenseTokenConfig, adminClient *madmin.AdminClient) error {
	// Keep existing subnet proxy if exists
	subnetKey, err := subnet.GetSubnetKeyFromMinIOConfig(ctx, adminClient)
	if err != nil {
		return err
	}
	configStr := fmt.Sprintf("subnet license=%s api_key=%s proxy=%s", license.License, license.APIKey, subnetKey.Proxy)
	_, err = adminClient.SetConfigKV(ctx, configStr)
	return err
}

func (d *licenseRegisterCmd) createSubnetAPIKeySecret(ctx context.Context, kubeClient *kubernetes.Clientset) error {
	secretName := "operator-subnet"
	_, err := kubeClient.CoreV1().Secrets("default").Get(ctx, secretName, metav1.GetOptions{})
	if err == nil { // Secret was already created
		return nil
	}
	apiKeySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName},
		Type:       corev1.SecretTypeOpaque,
		Data:       map[string][]byte{"api-key": []byte(d.apiKey)},
	}
	_, err = kubeClient.CoreV1().Secrets("default").Create(ctx, apiKeySecret, metav1.CreateOptions{})
	return err
}
