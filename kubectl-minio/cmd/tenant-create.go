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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	createDesc = `
'create' command creates a new MinIO tenant`
	createExample      = ` kubectl minio tenant create tenant1 --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns`
	tenantSecretSuffix = "-creds-secret"
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
		Use:     "create <TENANTNAME> --servers <NSERVERS> --volumes <NVOLUMES> --capacity <SIZE> --namespace <TENANTNS>",
		Short:   "Create a MinIO tenant",
		Long:    createDesc,
		Example: createExample,
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
	f.Int32Var(&c.tenantOpts.Servers, "servers", 0, "total number of pods in MinIO tenant")
	f.Int32Var(&c.tenantOpts.Volumes, "volumes", 0, "total number of volumes in the MinIO tenant")
	f.StringVar(&c.tenantOpts.Capacity, "capacity", "", "total raw capacity of MinIO tenant in this pool, e.g. 16Ti")
	f.StringVarP(&c.tenantOpts.NS, "namespace", "n", "", "k8s namespace for this MinIO tenant")
	f.StringVarP(&c.tenantOpts.StorageClass, "storage-class", "s", helpers.DefaultStorageclass, "storage class for this MinIO tenant")
	f.StringVarP(&c.tenantOpts.Image, "image", "i", helpers.DefaultTenantImage, "custom MinIO image for this tenant")
	f.StringVarP(&c.tenantOpts.ImagePullSecret, "image-pull-secret", "", "", "image pull secret to be used for pulling MinIO")
	f.BoolVar(&c.tenantOpts.DisableAntiAffinity, "enable-host-sharing", false, "[TESTING-ONLY] disable anti-affinity to allow pods to be co-located on a single node (unsupported in production environment)")
	f.StringVar(&c.tenantOpts.KmsSecret, "kes-config", "", "name of secret for KES KMS setup, refer https://github.com/minio/operator/blob/master/examples/kes-secret.yaml")
	f.BoolVar(&c.tenantOpts.EnableAuditLogs, "enable-audit-logs", true, "Enable/Disable audit logs")
	f.Int32Var(&c.tenantOpts.AuditLogsDiskSpace, "audit-logs-disk-space", 5, "(Only used when enable-audit-logs is on) Disk space for audit logs")
	f.StringVar(&c.tenantOpts.AuditLogsImage, "audit-logs-image", "", "(Only used when enable-audit-logs is on) The Docker image to use for audit logs")
	f.StringVar(&c.tenantOpts.AuditLogsPGImage, "audit-logs-pg-image", "", "(Only used when enable-audit-logs is on) The PostgreSQL Docker image to use for audit logs")
	f.StringVar(&c.tenantOpts.AuditLogsPGInitImage, "audit-logs-pg-init-image", "", "(Only used when enable-audit-logs is on) Defines the Docker image to use as the init container for running the postgres server in audit logs")
	f.StringVar(&c.tenantOpts.AuditLogsStorageClass, "audit-logs-storage-class", "", "(Only used when enable-audit-logs is on) Storage class for audit logs")
	f.BoolVarP(&c.output, "output", "o", false, "generate tenant yaml for 'kubectl apply -f tenant.yaml'")

	cmd.MarkFlagRequired("servers")
	cmd.MarkFlagRequired("volumes")
	cmd.MarkFlagRequired("capacity")
	cmd.MarkFlagRequired("namespace")
	return cmd
}

func (c *createCmd) validate(args []string) error {
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
	c.tenantOpts.SecretName = c.tenantOpts.Name + tenantSecretSuffix
	return c.tenantOpts.Validate()
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (c *createCmd) run(args []string) error {
	// Create operator and kube client
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	kclient, err := helpers.GetKubeClient(path)
	if err != nil {
		return err
	}

	// Generate console suer
	consoleUser := resources.NewSecretForConsole(&c.tenantOpts, fmt.Sprintf("%s-user-1", c.tenantOpts.Name))

	// generate the resources
	t, err := resources.NewTenant(&c.tenantOpts, consoleUser)
	if err != nil {
		return err
	}
	s := resources.NewSecretForTenant(&c.tenantOpts)

	// create resources
	if !c.output {
		return createTenant(oclient, kclient, t, s, consoleUser)
	}
	ot, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}
	os, err := yaml.Marshal(&s)
	if err != nil {
		return err
	}
	oc, err := yaml.Marshal(&consoleUser)
	if err != nil {
		return err
	}
	fmt.Println(string(ot))
	fmt.Println("---")
	fmt.Println(string(os))
	fmt.Println("---")
	fmt.Println(string(oc))
	return nil
}

func createTenant(oclient *operatorv1.Clientset, kclient *kubernetes.Clientset, t *miniov2.Tenant, s, console *corev1.Secret) error {
	if _, err := kclient.CoreV1().Namespaces().Get(context.Background(), t.Namespace, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("Namespace %s not found, please create the namespace using 'kubectl create ns %s'", t.Namespace, t.Namespace)
	}
	if _, err := kclient.CoreV1().Secrets(t.Namespace).Create(context.Background(), s, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := kclient.CoreV1().Secrets(t.Namespace).Create(context.Background(), console, metav1.CreateOptions{}); err != nil {
		return err
	}
	to, err := oclient.MinioV2().Tenants(t.Namespace).Create(context.Background(), t, v1.CreateOptions{})
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
