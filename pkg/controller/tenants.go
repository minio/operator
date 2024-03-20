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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// ErrEmptyRootCredentials is the error returned when we detect missing root credentials
var ErrEmptyRootCredentials = errors.New("empty tenant credentials")

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

// renewCert will renew one certificate at a time
func (c *Controller) renewCert(secret corev1.Secret, index int, tenant *miniov2.Tenant) error {
	// Check if secret starts with "operator-ca-tls-"
	secretName := OperatorCATLSSecretName + "-"
	// If the secret does not start with "operator-ca-tls-" then no need to continue
	if !strings.HasPrefix(secret.Name, secretName) {
		klog.Info("No secret found for multi-tenancy architecture of external certificates")
		return nil
	}
	klog.Infof("%d external secret found: %s", index, secret.Name)
	klog.Info("We are going to renew the external certificate for the tenant...")
	// Get the new certificate generated by cert-manager
	tenantSecretName := tenant.Spec.ExternalCertSecret[0].Name
	data, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(context.Background(), tenantSecretName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Couldn't get the certificate due to error %s", err)
		return err
	}
	if data == nil || len(data.Data) <= 0 {
		klog.Errorf("certificate's data can't be empty: %s", data)
		return errors.New("empty cert data")
	}
	CACertificate := data.Data["ca.crt"]
	if CACertificate == nil || len(CACertificate) <= 0 {
		klog.Errorf("ca.crt certificate data can't be empty: %s", CACertificate)
		return errors.New("empty cert ca data")
	}
	klog.Info("certificate data is not empty, proceed with renewal")
	// Delete the secret that starts with operator-ca-tls- because it is expired
	err = c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Delete(context.Background(), secret.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Infof("There was an error when deleting the secret: %s", err)
		return err
	}
	// Create the new secret that contains the new certificate
	newSecret := &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: miniov2.GetNSFromFile(),
		},
		Data: map[string][]byte{
			"ca.crt": CACertificate,
		},
	}
	_, err = c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Create(context.Background(), newSecret, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Secret not created %s", err)
		return err
	}
	// Append it
	c.fetchTransportCACertificates()
	// Reload CA certificates
	c.createTransport()
	// Rollout the Operator Deployment to use new certificate and trust the tenant.
	operatorDeployment, err := c.kubeClientSet.AppsV1().Deployments(miniov2.GetNSFromFile()).Get(context.Background(), miniov2.GetNSFromFile(), metav1.GetOptions{})
	if err != nil || operatorDeployment == nil {
		klog.Errorf("Couldn't retrieve the deployment %s", err)
		return err
	}
	operatorDeployment.Spec.Template.ObjectMeta.Name = miniov2.GetNSFromFile()
	operatorDeployment, err = c.kubeClientSet.AppsV1().Deployments(miniov2.GetNSFromFile()).Update(context.Background(), operatorDeployment, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("There was an error on deployment update %s", err)
		return err
	}
	klog.Info("external certificate successfully renewed for the tenant")
	return nil
}

// renewExternalCerts renews external certificates when they expire, ensuring that the Operator trusts its tenants.
func (c *Controller) renewExternalCerts(ctx context.Context, tenant *miniov2.Tenant, err error) error {
	if strings.Contains(err.Error(), "failed to verify certificate") {
		externalCertSecret := tenant.Spec.ExternalCertSecret
		klog.Info("Let's check if there is an external cert for the tenant...")
		if externalCertSecret != nil {
			// Check that there is a secret that starts with "operator-ca-tls-" to proceed with the renewal
			secretsAvailableAtOperatorNS, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				klog.Info("No external certificates are found under the multi-tenancy architecture to handle.")
				return nil
			}
			klog.Info("there are secret(s) for the operator")
			for index, secret := range secretsAvailableAtOperatorNS.Items {
				err = c.renewCert(secret, index, tenant)
				if err != nil {
					klog.Errorf("There was an error while renewing the cert: %s", err)
					return err
				}
			}
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
		// Allow root credentials via Tenant.Spec.Env
		if config.Name == "MINIO_ROOT_USER" || config.Name == "MINIO_ACCESS_KEY" {
			tenantConfiguration["accesskey"] = tenantConfiguration["MINIO_ROOT_USER"]
		} else if config.Name == "MINIO_ROOT_PASSWORD" || config.Name == "MINIO_SECRET_KEY" {
			tenantConfiguration["secretkey"] = tenantConfiguration["MINIO_ROOT_PASSWORD"]
		}
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
