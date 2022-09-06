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

package helpers

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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

var (
	validTenantName = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	ipAddress       = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)

// CheckValidTenantName validates if input tenantname complies with expected restrictions.
func CheckValidTenantName(tenantName string) error {
	if strings.TrimSpace(tenantName) == "" {
		return errors.New("Tenant name cannot be empty")
	}
	if len(tenantName) > 63 {
		return errors.New("Tenant name cannot be longer than 63 characters")
	}
	if ipAddress.MatchString(tenantName) {
		return errors.New("Tenant name cannot be an ip address")
	}
	if strings.Contains(tenantName, "..") || strings.Contains(tenantName, ".-") || strings.Contains(tenantName, "-.") {
		return errors.New("Tenant name contains invalid characters")
	}
	if !validTenantName.MatchString(tenantName) {
		return errors.New("Tenant name contains invalid characters")
	}
	return nil
}

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
func GetKubeExtensionClient(path string) (*apiextension.Clientset, error) {
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

	extClient, err := apiextension.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return extClient, nil
}

// GetKubeDynamicClient provides k8s client for CRDs
func GetKubeDynamicClient(path string) (dynamic.Interface, error) {
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

	return dynamic.NewForConfig(config)
}

// GetKubeOperatorClient provides k8s client for operator
func GetKubeOperatorClient(path string) (*operatorv1.Clientset, error) {
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
	t.SetAutoFormatHeaders(true)
	t.SetHeaderAlignment(table.ALIGN_LEFT)
	t.SetAlignment(table.ALIGN_LEFT)
	t.SetCenterSeparator("")
	t.SetColumnSeparator("")
	t.SetRowSeparator("")
	t.SetHeaderLine(false)
	t.SetBorder(false)
	t.SetTablePadding("\t")
	t.SetNoWhiteSpace(true)
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
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   "n",
	}
	_, err := prompt.Run()
	return err == nil
}

// AskNumber prompt user for number input
func AskNumber(label string, validate func(int) error) int {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			value, err := strconv.Atoi(input)
			if err != nil {
				return err
			}
			if validate != nil {
				return validate(value)
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}
	r, _ := strconv.Atoi(result)
	return r
}

// AskQuestion ask user for generic input
func AskQuestion(label string, validate func(string) error) string {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if validate != nil {
				return validate(input)
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}
	return result
}
