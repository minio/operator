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
	"io"
	"os/exec"

	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"gopkg.in/yaml.v2"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClient provides k8s client for kubeconfig
func GetKubeClient() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
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
	totalBytes, ok := totalQuantity.AsInt64()
	if !ok {
		totalBytes = totalQuantity.AsDec().UnscaledBig().Int64()
	}
	bytesPerVolume := totalBytes / int64(volumes)
	return resource.NewQuantity(bytesPerVolume, totalQuantity.Format), nil
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
