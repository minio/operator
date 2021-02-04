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

package helpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"k8s.io/client-go/dynamic"

	"github.com/dustin/go-humanize"
	"github.com/manifoldco/promptui"
	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	table "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClient provides k8s client for kubeconfig
func GetKubeClient(path string) (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if path != "" {
		loadingRules.ExplicitPath = path
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeClientset, nil
}

// GetKubeExtensionClient provides k8s client for CRDs
func GetKubeExtensionClient() (*apiextension.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	extClient, err := apiextension.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return extClient, nil
}

// GetKubeDynamicClient provides k8s client for CRDs
func GetKubeDynamicClient() (dynamic.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamic.NewForConfig(config)
	return dynClient, nil
}

// GetKubeOperatorClient provides k8s client for operator
func GetKubeOperatorClient() (*operatorv1.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	kubeClientset, err := operatorv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeClientset, nil
}

// ExecKubectl executes the given command using `kubectl`
func ExecKubectl(ctx context.Context, args ...string) ([]byte, error) {
	var stdout, stderr, combined bytes.Buffer

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = io.MultiWriter(&stdout, &combined)
	cmd.Stderr = io.MultiWriter(&stderr, &combined)
	if err := cmd.Run(); err != nil {
		return nil, errors.Errorf("kubectl command failed (%s). output=%s", err, combined.String())
	}
	return stdout.Bytes(), nil
}

// ServiceName returns the secret name for current tenant
func ServiceName(tenant string) string {
	return tenant + DefaultServiceNameSuffix
}

// VolumesPerServer returns volumes per server
// Volumes has total number of volumes in the tenant.
// volume per server is total volumes / total servers.
// we validate during input to ensure that volumes is a
// multiple of servers.
func VolumesPerServer(volumes, servers int32) int32 {
	return volumes / servers
}

// CapacityPerVolume returns capacity per volume
// capacity has total raw capacity required in MinIO tenant.
// divide total capacity by total drives to extract capacity per
// volume.
func CapacityPerVolume(capacity string, volumes int32) (*resource.Quantity, error) {
	totalQuantity, err := resource.ParseQuantity(capacity)
	if err != nil {
		return nil, err
	}
	return resource.NewQuantity(totalQuantity.Value()/int64(volumes), totalQuantity.Format), nil
}

// TotalCapacity returns total capacity of a given tenant
func TotalCapacity(tenant miniov2.Tenant) string {
	var totalBytes int64
	for _, z := range tenant.Spec.Pools {
		pvcBytes, _ := z.VolumeClaimTemplate.Spec.Resources.Requests.Storage().AsInt64()
		totalBytes = totalBytes + (pvcBytes * int64(z.Servers) * int64(z.VolumesPerServer))
	}
	return humanize.IBytes(uint64(totalBytes))
}

// ToYaml takes a slice of values, and returns corresponding YAML
// representation as a string slice
func ToYaml(objs []runtime.Object) ([]string, error) {
	manifests := make([]string, len(objs))
	for i, obj := range objs {
		o, err := yaml.Marshal(obj)
		if err != nil {
			return []string{}, err
		}
		manifests[i] = string(o)
	}

	return manifests, nil
}

// GetTable returns a formatted instance of the table
func GetTable() *table.Table {
	t := table.NewWriter(os.Stdout)
	t.SetAutoWrapText(false)
	t.SetHeaderAlignment(table.ALIGN_LEFT)
	t.SetAlignment(table.ALIGN_LEFT)
	return t
}

// DisableHelp disables the help command
func DisableHelp(cmd *cobra.Command) *cobra.Command {
	cmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
	return cmd
}

// Ask user for Y/N input. Return true if response is "y"
func Ask(label string) bool {
	validate := func(input string) error {
		s := strings.Trim(input, "\n\r")
		s = strings.ToLower(s)
		if strings.Compare(s, "n") != 0 && strings.Compare(s, "y") != 0 {
			return errors.New("Please enter y/n")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}
	fmt.Println()
	result, err := prompt.Run()
	if err != nil {
		return false
	}
	if strings.Compare(result, "n") == 0 {
		return false
	}
	return true
}
