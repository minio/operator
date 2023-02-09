// This file is part of MinIO Console Server
// Copyright (c) 2020 MinIO, Inc.
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

package cluster

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xcerts "github.com/minio/pkg/certs"
	"github.com/minio/pkg/env"
	"k8s.io/klog/v2"
	k8sscheme "k8s.io/kubectl/pkg/scheme"
)

const (
	// CertPasswordEnv Env variable is used to decrypt the private key in the TLS certificate for operator if need it
	CertPasswordEnv = "OPERATOR_CERT_PASSWD"
	// OperatorDeplymentNameEnv Env variable to specify a custom deployment name for Operator
	OperatorDeplymentNameEnv = "MINIO_OPERATOR_DEPLOYMENT_NAME"
	// OperatorTLSEnv Env variable to turn on / off Operator TLS.
	OperatorTLSEnv = "MINIO_OPERATOR_TLS_ENABLE"
	// OperatorCATLSSecretName is the name of the secret for the operator CA
	OperatorCATLSSecretName = "operator-ca-tls"
	// OperatorTLSSecretName is the name of secret created with TLS certs for Operator
	OperatorTLSSecretName = "operator-tls"
	// DefaultDeploymentName is the default name of the operator deployment
	DefaultDeploymentName = "minio-operator"
	// DefaultOperatorImage is the version fo the operator being used
	DefaultOperatorImage = "minio/operator:v4.5.8"
)

var serverCertsManager *xcerts.Manager

// rolloutRestartDeployment - executes the equivalent to kubectl rollout restart deployment
func (c *Controller) rolloutRestartDeployment(deployName string) error {
	ctx := context.Background()
	namespace := miniov2.GetNSFromFile()
	deployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	data, err := runtime.Encode(k8sscheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), deployment)
	if err != nil {
		return err
	}
	_, err2 := c.kubeClientSet.AppsV1().Deployments(namespace).Patch(ctx, deployName, types.StrategicMergePatchType, data, metav1.PatchOptions{FieldManager: "kubectl-rollout"})
	if err2 != nil {
		return err2
	}
	return nil
}

// generateOperatorTLSCert Issues the Operator TLS Certificate
func (c *Controller) generateOperatorTLSCert() (*string, *string) {
	return c.generateTLSCert("operator", OperatorTLSSecretName, getOperatorDeploymentName())
}

// recreateOperatorCertsIfRequired - Generate Operator TLS certs if not present or if is expired
func (c *Controller) recreateOperatorCertsIfRequired(ctx context.Context) error {
	namespace := miniov2.GetNSFromFile()
	operatorTLSSecret, err := c.getTLSSecret(ctx, namespace, OperatorTLSSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Info("TLS certificate not found. Generating one.")
			// Generate new certificate KeyPair for Operator server
			c.generateOperatorTLSCert()
			// reload in memory certificate for the operator server
			if serverCertsManager != nil {
				serverCertsManager.ReloadCerts()
			}

			return nil
		}
		return err
	}

	needsRenewal, err := c.certNeedsRenewal(operatorTLSSecret)
	if err != nil {
		return err
	}

	if !needsRenewal {
		return nil
	}

	// Expired cert. Delete the secret + CSR and re-create the cert
	err = c.deleteCSR(ctx, operatorCSRName())
	if err != nil {
		return err
	}
	klog.V(2).Info("Deleting the TLS secret of expired cert on operator")
	err = c.kubeClientSet.CoreV1().Secrets(namespace).Delete(ctx, OperatorTLSSecretName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	klog.V(2).Info("Generating a fresh TLS certificate for Operator")
	// Generate new certificate KeyPair for Operator server
	c.generateOperatorTLSCert()

	// reload in memory certificate for the operator server
	if serverCertsManager != nil {
		serverCertsManager.ReloadCerts()
	}

	return nil
}

func (c *Controller) fetchUserCredentials(ctx context.Context, tenant *miniov2.Tenant) []*v1.Secret {
	var userCredentials []*v1.Secret
	for _, credential := range tenant.Spec.Users {
		credentialSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, credential.Name, metav1.GetOptions{})
		if err == nil && credentialSecret != nil {
			userCredentials = append(userCredentials, credentialSecret)
		}
	}
	return userCredentials
}

func (c *Controller) getTransport() *http.Transport {
	if c.transport != nil {
		return c.transport
	}
	rootCAs := miniov2.MustGetSystemCertPool()
	// Default kubernetes CA certificate
	rootCAs.AppendCertsFromPEM(miniov2.GetPodCAFromFile())

	// If ca.crt exists in operator-tls (ie if the cert was issued by cert-manager) load the ca certificate from there
	operatorTLSCert, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(context.Background(), OperatorTLSSecretName, metav1.GetOptions{})
	if err == nil && operatorTLSCert != nil {
		// default secret keys for Opaque k8s secret
		caCertKey := "public.crt"
		// if secret type is k8s tls or cert-manager use the right ca key
		if operatorTLSCert.Type == "kubernetes.io/tls" {
			caCertKey = "tls.crt"
		} else if operatorTLSCert.Type == "cert-manager.io/v1alpha2" || operatorTLSCert.Type == "cert-manager.io/v1" {
			caCertKey = "ca.crt"
		}
		if val, ok := operatorTLSCert.Data[caCertKey]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
	}

	// Custom ca certificate to be used by operator
	operatorCATLSCert, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(context.Background(), OperatorCATLSSecretName, metav1.GetOptions{})
	if err == nil && operatorCATLSCert != nil {
		if val, ok := operatorCATLSCert.Data["tls.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
		if val, ok := operatorCATLSCert.Data["ca.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
		if val, ok := operatorCATLSCert.Data["public.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
	}

	c.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       15 * time.Second,
		ResponseHeaderTimeout: 15 * time.Minute,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 15 * time.Second,
		// Go net/http automatically unzip if content-type is
		// gzip disable this feature, as we are always interested
		// in raw stream.
		DisableCompression: true,
		TLSClientConfig: &tls.Config{
			// Can't use SSLv3 because of POODLE and BEAST
			// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
			// Can't use TLSv1.1 because of RC4 cipher usage
			MinVersion: tls.VersionTLS12,
			RootCAs:    rootCAs,
		},
	}

	return c.transport
}

func (c *Controller) createUsers(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (err error) {
	defer func() {
		if err == nil {
			if _, err = c.updateProvisionedUsersStatus(ctx, tenant, true); err != nil {
				klog.V(2).Infof(err.Error())
			}
		}
	}()

	userCredentials := c.fetchUserCredentials(ctx, tenant)
	if len(userCredentials) == 0 {
		return nil
	}

	if _, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningInitialUsers, 0); err != nil {
		return err
	}

	// get a new admin client
	adminClient, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		klog.Errorf("Error instantiating adminClnt: %v", err)
		return err
	}

	// configuration that means MinIO is running with LDAP enabled
	// and we need to skip the console user creation
	if err = tenant.CreateUsers(adminClient, userCredentials, tenantConfiguration); err != nil {
		klog.V(2).Infof("Unable to create MinIO users: %v", err)
	}

	return err
}

func (c *Controller) createBuckets(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (err error) {
	defer func() {
		if err == nil {
			if _, err = c.updateProvisionedBucketStatus(ctx, tenant, true); err != nil {
				klog.V(2).Infof(err.Error())
			}
		}
	}()

	tenantBuckets := tenant.Spec.Buckets
	if len(tenantBuckets) == 0 {
		return nil
	}

	if _, err := c.updateTenantStatus(ctx, tenant, StatusProvisioningDefaultBuckets, 0); err != nil {
		return err
	}

	// get a new admin client
	minioClient, err := tenant.NewMinIOUser(tenantConfiguration, c.getTransport())
	if err != nil {
		// show the error and continue
		klog.Errorf("Error instantiating minio Client: %v ", err)
		return err
	}

	if err := tenant.CreateBuckets(minioClient, tenantBuckets...); err != nil {
		klog.V(2).Infof("Unable to create MinIO buckets: %v", err)
		return err
	}

	return nil
}

// getOperatorDeploymentName Internal func returns the Operator deployment name from MINIO_OPERATOR_DEPLOYMENT_NAME ENV variable or the default name
func getOperatorDeploymentName() string {
	return env.Get(OperatorDeplymentNameEnv, DefaultDeploymentName)
}

// isOperatorTLS Internal func, reads MINIO_OPERATOR_TLS_ENABLE ENV to identify if Operator TLS is enabled, default "on"
func isOperatorTLS() bool {
	value, set := os.LookupEnv(OperatorTLSEnv)
	// By default, Operator TLS is used.
	return (set && value == "on") || !set
}

// operatorCSRName Internal func returns the given operator CSR name
func operatorCSRName() string {
	return getCSRName("operator")
}
