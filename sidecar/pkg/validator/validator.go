// This file is part of MinIO Operator
// Copyright (c) 2024 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/minio/operator/sidecar/pkg/configuration"
	"k8s.io/client-go/kubernetes"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	operatorClientset "github.com/minio/operator/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// Validate checks the configuration on the seeded configuration and issues a valid one for MinIO to
// start, however if root credentials are missing, it will exit with error
func Validate(tenantName string) {
	namespace := miniov2.GetNSFromFile()

	cfg, err := rest.InClusterConfig()
	// If config is passed as a flag use that instead
	//if kubeconfig != "" {
	//	cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	//}
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building MinIO operatorClientset: %s", err.Error())
	}

	controllerClient, err := operatorClientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building MinIO operatorClientset: %s", err.Error())
	}

	ctx := context.Background()

	// get the only tenant in this namespace
	tenant, err := controllerClient.MinioV2().Tenants(namespace).Get(ctx, tenantName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		panic(err)
	}
	tenant.EnsureDefaults()

	// determine the configmaps and secrets to watch
	configMaps, secrets, err := configuration.TenantResources(context.Background(), tenant, func(ctx context.Context, name string) (*corev1.ConfigMap, error) {
		return kubeClient.CoreV1().ConfigMaps(tenant.Namespace).Get(ctx, name, metav1.GetOptions{})
	}, func(ctx context.Context, name string) (*corev1.Secret, error) {
		return kubeClient.CoreV1().Secrets(tenant.Namespace).Get(ctx, name, metav1.GetOptions{})
	})
	if err != nil {
		log.Println(err)
		panic(err)
	}

	fileContents, rootUserFound, rootPwdFound := configuration.GetFullTenantConfig(tenant, configMaps, secrets)

	if !rootUserFound || !rootPwdFound {
		log.Println("Missing root credentials in the configuration.")
		log.Println("MinIO won't start")
		os.Exit(1)
	}

	err = os.WriteFile(miniov2.CfgFile, []byte(fileContents), 0o644)
	if err != nil {
		log.Println(err)
	}
}

// ReadTmpConfig reads the seeded configuration from a tmp location
func ReadTmpConfig() (bool, bool, string, error) {
	file, err := os.Open("/tmp/minio-config/config.env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	rootUserFound := false
	rootPwdFound := false

	scanner := bufio.NewScanner(file)
	newFile := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "MINIO_ROOT_USER") {
			rootUserFound = true
		}
		if strings.Contains(line, "MINIO_ACCESS_KEY") {
			rootUserFound = true
		}
		if strings.Contains(line, "MINIO_ROOT_PASSWORD") {
			rootPwdFound = true
		}
		if strings.Contains(line, "MINIO_SECRET_KEY") {
			rootPwdFound = true
		}
		// We don't allow users to set MINIO_ARGS
		if strings.Contains(line, "MINIO_ARGS") {
			log.Println("MINIO_ARGS in config file found. It will be ignored.")
			continue
		}
		newFile = newFile + line + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return rootUserFound, rootPwdFound, newFile, nil
}
