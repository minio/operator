// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package configuration

import (
	"fmt"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/common"
	"github.com/minio/operator/pkg/resources/statefulsets"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strings"
)

const (
	bucketDNSEnv = "MINIO_DNS_WEBHOOK_ENDPOINT"
)

func GetFullTenantConfig(tenant *miniov2.Tenant, configSecret *corev1.Secret) (string, bool, bool) {
	seededVars := parseConfEnvSecret(configSecret)
	rootUserFound := false
	rootPwdFound := false
	for _, env := range seededVars {
		if env.Name == "MINIO_ROOT_USER" {
			rootUserFound = true
		}
		if env.Name == "MINIO_ACCESS_KEY" {
			rootUserFound = true
		}
		if env.Name == "MINIO_ROOT_PASSWORD" {
			rootPwdFound = true
		}
		if env.Name == "MINIO_SECRET_KEY" {
			rootPwdFound = true
		}
	}
	compiledConfig := buildTenantEnvs(tenant, seededVars)
	configurationFileContent := envVarsToFileContent(compiledConfig)
	return configurationFileContent, rootUserFound, rootPwdFound
}

func parseConfEnvSecret(secret *corev1.Secret) map[string]corev1.EnvVar {
	if secret == nil {
		return nil
	}
	data := secret.Data["config.env"]
	envMap := make(map[string]corev1.EnvVar)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				envVar := corev1.EnvVar{
					Name:  name,
					Value: value,
				}
				envMap[name] = envVar
			}
		}
	}
	return envMap
}

func buildTenantEnvs(tenant *miniov2.Tenant, cfgEnvExisting map[string]corev1.EnvVar) []corev1.EnvVar {
	// Enable `mc admin update` style updates to MinIO binaries
	// within the container, only operator is supposed to perform
	// these operations.
	envVarsMap := map[string]corev1.EnvVar{
		"MINIO_UPDATE": {
			Name:  "MINIO_UPDATE",
			Value: "on",
		},
		"MINIO_UPDATE_MINISIGN_PUBKEY": {
			Name:  "MINIO_UPDATE_MINISIGN_PUBKEY",
			Value: "RWTx5Zr1tiHQLwG9keckT0c45M3AGeHD6IvimQHpyRywVWGbP1aVSGav",
		},
		"MINIO_PROMETHEUS_JOB_ID": {
			Name:  "MINIO_PROMETHEUS_JOB_ID",
			Value: tenant.PrometheusConfigJobName(),
		},
	}
	// Specific case of bug in runtimeClass crun where $HOME is not set
	for _, pool := range tenant.Spec.Pools {
		if pool.RuntimeClassName != nil && *pool.RuntimeClassName == "crun" {
			// Set HOME to /
			envVarsMap["HOME"] = corev1.EnvVar{
				Name:  "HOME",
				Value: "/",
			}
		}
	}
	var domains []string
	// Enable Bucket DNS only if asked for by default turned off
	if tenant.BucketDNS() {
		domains = append(domains, tenant.MinIOBucketBaseDomain())
		sidecarBucketURL := fmt.Sprintf("http://127.0.0.1:%s%s/%s/%s",
			common.WebhookDefaultPort,
			common.WebhookAPIBucketService,
			tenant.Namespace,
			tenant.Name)
		envVarsMap[bucketDNSEnv] = corev1.EnvVar{
			Name:  bucketDNSEnv,
			Value: sidecarBucketURL,
		}
	}
	// Check if any domains are configured
	if tenant.HasMinIODomains() {
		domains = append(domains, tenant.GetDomainHosts()...)
	}
	// tell MinIO about all the domains meant to hit it if they are not passed manually via .spec.env
	if len(domains) > 0 {
		envVarsMap[miniov2.MinIODomain] = corev1.EnvVar{
			Name:  miniov2.MinIODomain,
			Value: strings.Join(domains, ","),
		}
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
	envVarsMap[miniov2.MinIOServerURL] = corev1.EnvVar{
		Name:  miniov2.MinIOServerURL,
		Value: serverURL,
	}

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
		envVarsMap[miniov2.MinIOBrowserRedirectURL] = corev1.EnvVar{
			Name:  miniov2.MinIOBrowserRedirectURL,
			Value: consoleDomain,
		}
	}
	if tenant.HasKESEnabled() {
		envVarsMap["MINIO_KMS_KES_ENDPOINT"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_ENDPOINT",
			Value: tenant.KESServiceEndpoint(),
		}
		envVarsMap["MINIO_KMS_KES_CERT_FILE"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CERT_FILE",
			Value: miniov2.MinIOCertPath + "/client.crt",
		}
		envVarsMap["MINIO_KMS_KES_KEY_FILE"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_FILE",
			Value: miniov2.MinIOCertPath + "/client.key",
		}
		envVarsMap["MINIO_KMS_KES_CA_PATH"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CA_PATH",
			Value: miniov2.MinIOCertPath + "/CAs/kes.crt",
		}
		envVarsMap["MINIO_KMS_KES_CAPATH"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_CAPATH",
			Value: miniov2.MinIOCertPath + "/CAs/kes.crt",
		}
		envVarsMap["MINIO_KMS_KES_KEY_NAME"] = corev1.EnvVar{
			Name:  "MINIO_KMS_KES_KEY_NAME",
			Value: tenant.Spec.KES.KeyName,
		}
	}

	// attach tenant args
	args := strings.Join(statefulsets.GetContainerArgs(tenant, ""), " ")
	envVarsMap["MINIO_ARGS"] = corev1.EnvVar{
		Name:  "MINIO_ARGS",
		Value: args,
	}

	// Add all the tenant.spec.env environment variables
	// User defined environment variables will take precedence over default environment variables
	for _, env := range tenant.GetEnvVars() {
		envVarsMap[env.Name] = env
	}
	var envVars []corev1.EnvVar
	// transform map to array and skip configurations from config.env
	for _, env := range envVarsMap {
		if cfgEnvExisting != nil {
			if _, ok := cfgEnvExisting[env.Name]; !ok {
				envVars = append(envVars, env)
			}
		} else {
			envVars = append(envVars, env)
		}
	}
	// now add everything in the existing config.env
	for _, envVar := range cfgEnvExisting {
		envVars = append(envVars, envVar)
	}
	// sort the array to produce the same result everytime
	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})

	return envVars
}

func envVarsToFileContent(envVars []corev1.EnvVar) string {
	content := ""
	for _, env := range envVars {
		content += fmt.Sprintf("export %s=\"%s\"\n", env.Name, env.Value)
	}
	return content
}
