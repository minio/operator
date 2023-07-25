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
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/minio/kubectl-minio/cmd/helpers"
	"github.com/minio/kubectl-minio/cmd/resources"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	createDesc = `
'create' command creates a new MinIO tenant`
	createExample = ` kubectl minio tenant create tenant1 --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns`
)

type createCmd struct {
	out        io.Writer
	errOut     io.Writer
	output     bool
	tenantOpts resources.TenantOptions
}

func newTenantCreateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &createCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "create <TENANTNAME> --pool <POOLNAME> --servers <NSERVERS> --volumes <NVOLUMES> --capacity <SIZE> --namespace <TENANTNS>",
		Short:   "Create a MinIO tenant",
		Long:    createDesc,
		Example: createExample,
		Args: func(cmd *cobra.Command, args []string) error {
			// The disable-tls parameter default value is false, we cannot rely on the default value binded to the tenantOpts.DisableTLS variable
			// to identify if the parameter --disable-tls was actually set on the command line.
			// regardless of which value is being set to the flag, if the flag and ONLY if the flag is present, then we disable TLS
			c.tenantOpts.DisableTLS = cmd.Flags().Lookup("disable-tls").Changed
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
	f.StringVarP(&c.tenantOpts.PoolName, "pool", "p", "", "name for this pool")
	f.Int32Var(&c.tenantOpts.Servers, "servers", 0, "total number of pods in MinIO tenant")
	f.Int32Var(&c.tenantOpts.Volumes, "volumes", 0, "total number of volumes in the MinIO tenant")
	f.StringVar(&c.tenantOpts.Capacity, "capacity", "", "total raw capacity of MinIO tenant in this pool, e.g. 16Ti")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", "", "k8s namespace for this MinIO tenant")
	f.StringVarP(&c.tenantOpts.StorageClass, "storage-class", "s", helpers.DefaultStorageclass, "storage class for this MinIO tenant")
	f.StringVarP(&c.tenantOpts.Image, "image", "i", helpers.DefaultTenantImage, "custom MinIO image for this tenant")
	f.StringVarP(&c.tenantOpts.ImagePullSecret, "image-pull-secret", "", "", "image pull secret to be used for pulling MinIO")
	f.BoolVar(&c.tenantOpts.DisableAntiAffinity, "enable-host-sharing", false, "[TESTING-ONLY] disable anti-affinity to allow pods to be co-located on a single node (unsupported in production environment)")
	f.StringVar(&c.tenantOpts.KmsSecret, "kes-config", "", "name of secret for KES KMS setup, refer https://github.com/minio/operator/blob/master/examples/kes-secret.yaml")
	f.BoolVar(&c.tenantOpts.DisableTLS, "disable-tls", false, "Disable TLS")
	f.BoolVarP(&c.output, "output", "o", false, "generate tenant yaml for 'kubectl apply -f tenant.yaml'")
	f.BoolVar(&c.tenantOpts.Interactive, "interactive", false, "Create tenant in interactive mode")
	f.BoolVar(&c.tenantOpts.ExposeMinioService, "expose-minio-service", false, "Enable/Disable expose the Minio Service")
	f.BoolVar(&c.tenantOpts.ExposeConsoleService, "expose-console-service", false, "Enable/Disable expose the Console service")
	f.BoolVar(&c.tenantOpts.EnableSFTP, "enable-sftp", false, "Enable/Disable SFTP access to the tenant")
	return cmd
}

func (c *createCmd) validate(args []string) error {
	if c.tenantOpts.Interactive {
		return nil
	}
	if args == nil {
		return errors.New("create command requires specifying the tenant name as an argument, e.g. 'kubectl minio tenant create tenant1'")
	}
	if len(args) != 1 {
		return errors.New("create command requires specifying the tenant name as an argument, e.g. 'kubectl minio tenant create tenant1'")
	}
	if args[0] == "" {
		return errors.New("create command requires specifying the tenant name as an argument, e.g. 'kubectl minio tenant create tenant1'")
	}
	// Tenant name should have DNS token restrictions
	if err := helpers.CheckValidTenantName(args[0]); err != nil {
		return err
	}
	c.tenantOpts.Name = args[0]
	c.tenantOpts.ConfigurationSecretName = fmt.Sprintf("%s-env-configuration", c.tenantOpts.Name)
	if c.tenantOpts.NS == "" {
		return errors.New("--namespace flag is required")
	}
	return c.tenantOpts.Validate()
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *createCmd) run(_ []string) error {
	// Create operator and kube client
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	operatorClient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}
	kubeClient, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}
	if c.tenantOpts.Interactive {
		if err := c.populateInteractiveTenant(); err != nil {
			return err
		}
	}
	// Generate MinIO user credentials
	tenantUserCredentials, err := resources.NewUserCredentialsSecret(&c.tenantOpts, fmt.Sprintf("%s-user-1", c.tenantOpts.Name))
	if err != nil {
		return err
	}
	// generate tenant configuration
	tenantConfiguration, err := resources.NewTenantConfigurationSecret(&c.tenantOpts)
	if err != nil {
		return err
	}
	// generate tenant resource
	tenant, err := resources.NewTenant(&c.tenantOpts, tenantUserCredentials)
	if err != nil {
		return err
	}
	// create resources
	if !c.output {
		return createTenant(operatorClient, kubeClient, tenant, tenantConfiguration, tenantUserCredentials)
	}
	tenantYAML, err := yaml.Marshal(&tenant)
	if err != nil {
		return err
	}
	tenantConfigurationYAML, err := yaml.Marshal(&tenantConfiguration)
	if err != nil {
		return err
	}
	tenantUserYAML, err := yaml.Marshal(&tenantUserCredentials)
	if err != nil {
		return err
	}
	fmt.Println(string(tenantYAML))
	fmt.Println("---")
	fmt.Println(string(tenantConfigurationYAML))
	fmt.Println("---")
	fmt.Println(string(tenantUserYAML))
	return nil
}

func (c *createCmd) populateInteractiveTenant() error {
	c.tenantOpts.Name = helpers.AskQuestion("Tenant name", helpers.CheckValidTenantName)
	c.tenantOpts.ConfigurationSecretName = fmt.Sprintf("%s-env-configuration", c.tenantOpts.Name)
	c.tenantOpts.Servers = int32(helpers.AskNumber("Total of servers", greaterThanZero))
	c.tenantOpts.Volumes = int32(helpers.AskNumber("Total of volumes", greaterThanZero))
	c.tenantOpts.NS = helpers.AskQuestion("Namespace", validateEmptyInput)
	c.tenantOpts.Capacity = helpers.AskQuestion("Capacity", validateCapacity)
	if err := c.tenantOpts.Validate(); err != nil {
		return err
	}
	c.tenantOpts.DisableTLS = helpers.Ask("Disable TLS")
	c.tenantOpts.ExposeMinioService = helpers.Ask("Expose Minio Service")
	c.tenantOpts.ExposeConsoleService = helpers.Ask("Expose Console Service")
	c.tenantOpts.EnableSFTP = helpers.Ask("Enable SFTP")
	return nil
}

func validateEmptyInput(value string) error {
	if value == "" {
		return errors.New("value can't be empty")
	}
	return nil
}

func validateCapacity(value string) error {
	if err := validateEmptyInput(value); err != nil {
		return err
	}
	_, err := resource.ParseQuantity(value)
	return err
}

func greaterThanZero(value int) error {
	if value <= 0 {
		return errors.New("value needs to be greater than zero")
	}
	return nil
}

func createTenant(operatorClient *operatorv1.Clientset, kubeClient *kubernetes.Clientset, tenant *miniov2.Tenant, tenantConfiguration, console *corev1.Secret) error {
	if _, err := kubeClient.CoreV1().Namespaces().Get(context.Background(), tenant.Namespace, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("namespace %s not found, please create the namespace using 'kubectl create ns %s'", tenant.Namespace, tenant.Namespace)
	}
	if _, err := kubeClient.CoreV1().Secrets(tenant.Namespace).Create(context.Background(), tenantConfiguration, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := kubeClient.CoreV1().Secrets(tenant.Namespace).Create(context.Background(), console, metav1.CreateOptions{}); err != nil {
		return err
	}
	to, err := operatorClient.MinioV2().Tenants(tenant.Namespace).Create(context.Background(), tenant, v1.CreateOptions{})
	if err != nil {
		return err
	}
	// Check MinIO S3 Endpoint Service
	minSvc := services.NewClusterIPForMinIO(to)

	// Check MinIO Console Endpoint Service
	conSvc := services.NewClusterIPForConsole(to)

	if IsTerminal() {
		printBanner(to.ObjectMeta.Name, to.ObjectMeta.Namespace, string(console.Data["CONSOLE_ACCESS_KEY"]), string(console.Data["CONSOLE_SECRET_KEY"]),
			minSvc, conSvc)
	}
	return nil
}

func printBanner(tenantName, ns, user, pwd string, s, c *corev1.Service) {
	fmt.Printf(Bold(fmt.Sprintf("\nTenant '%s' created in '%s' Namespace\n\n", tenantName, ns)))
	fmt.Printf(Blue("  Username: %s \n", user))
	fmt.Printf(Blue("  Password: %s \n", pwd))
	fmt.Printf(Blue("  Note: Copy the credentials to a secure location. MinIO will not display these again.\n\n"))
	var minPorts, consolePorts string
	for _, p := range s.Spec.Ports {
		minPorts = minPorts + strconv.Itoa(int(p.Port)) + ","
	}
	for _, p := range c.Spec.Ports {
		consolePorts = consolePorts + strconv.Itoa(int(p.Port)) + ","
	}
	t := helpers.GetTable()
	t.SetHeader([]string{"Application", "Service Name", "Namespace", "Service Type", "Service Port"})
	t.Append([]string{"MinIO", s.Name, ns, "ClusterIP", strings.TrimSuffix(minPorts, ",")})
	t.Append([]string{"Console", c.Name, ns, "ClusterIP", strings.TrimSuffix(consolePorts, ",")})
	t.Render()
	fmt.Println()
}
