// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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

package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/minio/operator/pkg/common"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// ErrEmptyRootCredentials is the error returned when we detect missing root credentials
var ErrEmptyRootCredentials = errors.New("empty tenant credentials")

const (
	rootUserEnv      = "MINIO_ROOT_USER"
	rootUserPassword = "MINIO_ROOT_PASSWORD"
	storageClassEnv  = "MINIO_STORAGE_CLASS_STANDARD"
	accessKeyEnv     = "MINIO_ACCESS_KEY"
	secretKeyEnv     = "MINIO_SECRET_KEY"
	minioBrowserEnv  = "MINIO_BROWSER"
)

func (c *Controller) getTenantConfiguration(ctx context.Context, tenant *miniov2.Tenant) (map[string][]byte, error) {
	tenantConfiguration := map[string][]byte{}
	// Load tenant configuration from file
	if tenant.HasConfigurationSecret() {
		minioConfigurationSecretName := tenant.Spec.Configuration.Name
		minioConfigurationSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, minioConfigurationSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		configFromFile := miniov2.ParseRawConfiguration(minioConfigurationSecret.Data["config.env"])
		for key, val := range configFromFile {
			tenantConfiguration[key] = val
		}
	}
	return tenantConfiguration, nil
}

func (c *Controller) saveTenantConfiguration(ctx context.Context, tenant *miniov2.Tenant) error {
	prevConfiguration, err := c.getTenantConfiguration(ctx, tenant)
	if err != nil {
		return fmt.Errorf("secret '%s' reference on tenant.spec.configuration not found: %v", tenant.Spec.Configuration, err)
	}
	tenantConfiguration := map[string]string{}
	if _, ok := prevConfiguration[storageClassEnv]; ok {
		tenantConfiguration[storageClassEnv] = string(prevConfiguration[storageClassEnv])
	}

	if _, ok := prevConfiguration[rootUserEnv]; ok {
		tenantConfiguration[rootUserEnv] = string(prevConfiguration[rootUserEnv])
	}

	if _, ok := prevConfiguration[rootUserPassword]; ok {
		tenantConfiguration[rootUserPassword] = string(prevConfiguration[rootUserPassword])
	}

	if _, ok := prevConfiguration[accessKeyEnv]; ok {
		tenantConfiguration[accessKeyEnv] = string(prevConfiguration[accessKeyEnv])
	}

	if _, ok := prevConfiguration[secretKeyEnv]; ok {
		tenantConfiguration[secretKeyEnv] = string(prevConfiguration[secretKeyEnv])
	}

	if _, ok := prevConfiguration[minioBrowserEnv]; ok {
		tenantConfiguration[minioBrowserEnv] = string(prevConfiguration[minioBrowserEnv])
	}

	// Update credentials based on the existing credsSecret, we will overwrite root credentials only when the fields
	// accesskey and secretkey are not empty
	if tenant.HasCredsSecret() {
		credsSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.Spec.CredsSecret.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		var accessKey string
		var secretKey string
		if _, ok := credsSecret.Data["accesskey"]; ok {
			accessKey = string(credsSecret.Data["accesskey"])
		}
		if _, ok := credsSecret.Data["secretkey"]; ok {
			secretKey = string(credsSecret.Data["secretkey"])
		}
		if accessKey != "" || secretKey != "" {
			tenantConfiguration["MINIO_ROOT_USER"] = accessKey
			tenantConfiguration["MINIO_ROOT_PASSWORD"] = secretKey
		}
	}

	// Enable `mc admin update` style updates to MinIO binaries
	// within the container, only operator is supposed to perform
	// these operations.
	tenantConfiguration["MINIO_UPDATE"] = "on"
	tenantConfiguration["MINIO_UPDATE_MINISIGN_PUBKEY"] = "RWTx5Zr1tiHQLwG9keckT0c45M3AGeHD6IvimQHpyRywVWGbP1aVSGav"
	tenantConfiguration["MINIO_OPERATOR_VERSION"] = c.operatorVersion
	tenantConfiguration["MINIO_PROMETHEUS_JOB_ID"] = tenant.PrometheusConfigJobName()

	var domains []string
	// Set Bucket DNS domain only if enabled
	if tenant.BucketDNS() {
		domains = append(domains, tenant.MinIOBucketBaseDomain())
		sidecarBucketURL := fmt.Sprintf("http://127.0.0.1:%s%s/%s/%s",
			common.WebhookDefaultPort,
			common.WebhookAPIBucketService,
			tenant.Namespace,
			tenant.Name)
		tenantConfiguration[common.BucketDNSEnv] = sidecarBucketURL
	}
	// Check if any domains are configured
	if tenant.HasMinIODomains() {
		domains = append(domains, tenant.GetDomainHosts()...)
	}
	// tell MinIO about all the domains meant to hit it if they are not passed manually via .spec.env
	if len(domains) > 0 {
		tenantConfiguration[miniov2.MinIODomain] = strings.Join(domains, ",")
	}
	// If no specific server URL is specified we will specify the internal k8s url, but if a list of domains was
	// provided we will use the first domain.
	serverURL := tenant.MinIOServerEndpoint()
	if tenant.HasMinIODomains() {
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(tenant.Spec.Features.Domains.Minio[0], "http") {
			useSchema := "http"
			if tenant.TLS() {
				useSchema = "https"
			}
			serverURL = fmt.Sprintf("%s://%s", useSchema, tenant.Spec.Features.Domains.Minio[0])
		} else {
			serverURL = tenant.Spec.Features.Domains.Minio[0]
		}
	}
	tenantConfiguration[miniov2.MinIOServerURL] = serverURL

	// Set the redirect url for console
	if tenant.HasConsoleDomains() {
		consoleDomain := tenant.Spec.Features.Domains.Console
		// Infer schema from tenant TLS, if not explicit
		if !strings.HasPrefix(consoleDomain, "http") {
			useSchema := "http"
			if tenant.TLS() {
				useSchema = "https"
			}
			consoleDomain = fmt.Sprintf("%s://%s", useSchema, consoleDomain)
		}
		tenantConfiguration[miniov2.MinIOBrowserRedirectURL] = consoleDomain
	}

	if tenant.HasKESEnabled() {
		tenantConfiguration["MINIO_KMS_KES_ENDPOINT"] = tenant.KESServiceEndpoint()
		tenantConfiguration["MINIO_KMS_KES_CERT_FILE"] = miniov2.MinIOCertPath + "/client.crt"
		tenantConfiguration["MINIO_KMS_KES_KEY_FILE"] = miniov2.MinIOCertPath + "/client.key"
		tenantConfiguration["MINIO_KMS_KES_CA_PATH"] = miniov2.MinIOCertPath + "/CAs/kes.crt"
		tenantConfiguration["MINIO_KMS_KES_KEY_NAME"] = tenant.Spec.KES.KeyName
	}

	// Set the env variables in the tenant.spec.env field
	// User defined environment variables will take precedence over default environment variables
	envVars := tenant.GetEnvVars()
	for _, ev := range envVars {
		tenantConfiguration[ev.Name] = ev.Value
	}

	configurationSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.Spec.Configuration.Name, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		configurationSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenant.ConfigurationSecretName(),
				Namespace: tenant.Namespace,
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.Version,
			},
			Data: map[string][]byte{
				"config.env": []byte(miniov2.GenerateTenantConfigurationFile(tenantConfiguration)),
			},
		}
		_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, configurationSecret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error updating tenant '%s/%s', could not create tenant.spec.configuration secret: %v", tenant.Namespace, tenant.Name, err)
		}
	} else {
		configurationSecret.Data = map[string][]byte{
			"config.env": []byte(miniov2.GenerateTenantConfigurationFile(tenantConfiguration)),
		}
		_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, configurationSecret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error updating tenant '%s/%s', could not update tenant.spec.configuration secret: %v", tenant.Namespace, tenant.Name, err)
		}
	}

	return nil
}

// getTenantCredentials returns a combination of env, credsSecret and Configuration tenant credentials
func (c *Controller) getTenantCredentials(ctx context.Context, tenant *miniov2.Tenant) (map[string][]byte, error) {
	// Configuration for tenant can be passed using 2 different sources, tenant.spec.env and config.env secret
	// If the user provides duplicated configuration the override order will be:
	// tenant.Spec.Env < config.env file (k8s secret)
	tenantConfiguration := map[string][]byte{}

	for _, config := range tenant.GetEnvVars() {
		tenantConfiguration[config.Name] = []byte(config.Value)
	}

	// Load tenant configuration from file
	config, err := c.getTenantConfiguration(ctx, tenant)
	if err != nil {
		return nil, err
	}
	for key, val := range config {
		tenantConfiguration[key] = val
	}

	var accessKey string
	var secretKey string

	if _, ok := tenantConfiguration["accesskey"]; ok {
		accessKey = string(tenantConfiguration["accesskey"])
	}

	if _, ok := tenantConfiguration["secretkey"]; ok {
		secretKey = string(tenantConfiguration["secretkey"])
	}

	if accessKey == "" || secretKey == "" {
		return tenantConfiguration, ErrEmptyRootCredentials
	}

	return tenantConfiguration, nil
}
