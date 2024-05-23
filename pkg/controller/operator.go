// This file is part of MinIO Operator
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

package controller

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/minio/operator/pkg/certs"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/common"
	"github.com/minio/operator/pkg/utils"
	xcerts "github.com/minio/pkg/certs"
	"github.com/minio/pkg/env"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	k8sscheme "k8s.io/kubectl/pkg/scheme"
)

const (
	// CertPasswordEnv Env variable is used to decrypt the private key in the TLS certificate for operator if need it
	CertPasswordEnv = "OPERATOR_CERT_PASSWD"
	// OperatorDeploymentNameEnv Env variable to specify a custom deployment name for Operator
	OperatorDeploymentNameEnv = "MINIO_OPERATOR_DEPLOYMENT_NAME"
	// OperatorCATLSSecretName is the name of the secret for the operator CA
	OperatorCATLSSecretName = "operator-ca-tls"
	// OperatorCATLSSecretPrefix is the name of the multi tenant secret for the operator CA
	OperatorCATLSSecretPrefix = OperatorCATLSSecretName + "-"
	// OperatorCSRSignerCASecretName is the name of the secret for the signer-ca certificate
	// this is a copy of the secret signer-ca in namespace
	OperatorCSRSignerCASecretName = "openshift-csr-signer-ca"
	// OpenshiftKubeControllerNamespace is the namespace of kube controller manager operator in Openshift
	OpenshiftKubeControllerNamespace = "openshift-kube-controller-manager-operator"
	// OpenshiftCATLSSecretName is the secret name of the CRD's signer in kubernetes under  OpenshiftKubeControllerNamespace namespace
	OpenshiftCATLSSecretName = "csr-signer"
	// DefaultDeploymentName is the default name of the operator deployment
	DefaultDeploymentName = "minio-operator"
	// DefaultOperatorImage is the version fo the operator being used
	DefaultOperatorImage = "minio/operator:v5.0.15"
	// DefaultOperatorImageEnv is the default image to minio instance
	DefaultOperatorImageEnv = "MINIO_OPERATOR_IMAGE"
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

// getTransport returns a *http.Transport with the collection of the trusted CA certificates
// returns a cached transport if already available
func (c *Controller) getTransport() *http.Transport {
	if c.transport != nil {
		return c.transport
	}
	c.transport = c.createTransport()
	return c.transport
}

// createTransport returns a *http.Transport with the collection of the trusted CA certificates
func (c *Controller) createTransport() *http.Transport {
	rootCAs := c.fetchTransportCACertificates()
	dialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 15 * time.Second,
	}
	c.transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
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

// fetchTransportCACertificates retrieves a *x509.CertPool with all CA that operator will trust
func (c *Controller) fetchTransportCACertificates() (pool *x509.CertPool) {
	rootCAs := miniov2.MustGetSystemCertPool()
	// Default kubernetes CA certificate
	rootCAs.AppendCertsFromPEM(miniov2.GetPodCAFromFile())

	if utils.GetOperatorRuntime() == common.OperatorRuntimeOpenshift {
		// Openshift Service CA certificate
		if serviceCA := miniov2.GetOpenshiftServiceCAFromFile(); serviceCA != nil {
			rootCAs.AppendCertsFromPEM(serviceCA)
		}
		// Openshift csr-signer CA certificate
		if cert := miniov2.GetOpenshiftCSRSignerCAFromFile(); cert != nil {
			rootCAs.AppendCertsFromPEM(cert)
		}
	}

	// Append all external Certificate Authorities added to Operator, secrets with prefix "operator-ca-tls"
	secretsAvailableAtOperatorNS, _ := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).List(context.Background(), metav1.ListOptions{})
	for _, secret := range secretsAvailableAtOperatorNS.Items {
		// Check if secret starts with "operator-ca-tls-"
		if strings.HasPrefix(secret.Name, OperatorCATLSSecretName) {
			operatorCATLSCert, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(context.Background(), secret.Name, metav1.GetOptions{})
			if err == nil && operatorCATLSCert != nil {
				if newPublicCert, err := getFileFromSecretDataField(operatorCATLSCert.Data, certs.PublicCertFile); err == nil {
					rootCAs.AppendCertsFromPEM(newPublicCert)
				}
				if newTLSCert, err := getFileFromSecretDataField(operatorCATLSCert.Data, certs.TLSCertFile); err == nil {
					rootCAs.AppendCertsFromPEM(newTLSCert)
				}
				if newCACert, err := getFileFromSecretDataField(operatorCATLSCert.Data, certs.CAPublicCertFile); err == nil {
					rootCAs.AppendCertsFromPEM(newCACert)
				}
			}
		}
	}

	return rootCAs
}

// getFileFromSecretDataField Get the value of a secret field
// limiting the field key name to public TLS certificate file names
func getFileFromSecretDataField(secretData map[string][]byte, key string) ([]byte, error) {
	keys := []string{
		certs.TLSCertFile,
		certs.CAPublicCertFile,
		certs.PublicCertFile,
	}
	if slices.Contains(keys, key) {
		data, ok := secretData[key]
		if ok {
			return data, nil
		}
	} else {
		return nil, fmt.Errorf("unknow TLS key '%s'", key)
	}
	return nil, fmt.Errorf("key '%s' not found in secret", key)
}

// TrustTLSCertificatesInSecretIfChanged Compares old and new secret content and trusts TLS certificates if field
// content is different, looks for the fields public.crt, tls.crt and ca.crt
func (c *Controller) TrustTLSCertificatesInSecretIfChanged(newSecret *corev1.Secret, oldSecret *corev1.Secret) bool {
	added := false
	if oldSecret == nil {
		// secret did not exist before, we trust all certs in it
		if c.trustPEMInSecretField(newSecret, certs.PublicCertFile) {
			added = true
		}
		if c.trustPEMInSecretField(newSecret, certs.TLSCertFile) {
			added = true
		}
		if c.trustPEMInSecretField(newSecret, certs.CAPublicCertFile) {
			added = true
		}
	} else {
		// compare to add to trust only certs that changed
		if c.trustIfChanged(newSecret, oldSecret, certs.PublicCertFile) {
			added = true
		}
		if c.trustIfChanged(newSecret, oldSecret, certs.TLSCertFile) {
			added = true
		}
		if c.trustIfChanged(newSecret, oldSecret, certs.CAPublicCertFile) {
			added = true
		}
	}
	return added
}

func (c *Controller) trustIfChanged(newSecret *corev1.Secret, oldSecret *corev1.Secret, fieldToCompare string) bool {
	if newPublicCert, err := getFileFromSecretDataField(newSecret.Data, fieldToCompare); err == nil {
		if oldPublicCert, err := getFileFromSecretDataField(oldSecret.Data, fieldToCompare); err == nil {
			newPublicCert = bytes.TrimSpace(newPublicCert)
			oldPublicCert = bytes.TrimSpace(oldPublicCert)
			// add to trust only if cert changed
			if !bytes.Equal(oldPublicCert, newPublicCert) {
				if err := c.addTLSCertificatesToTrustInTransport(newPublicCert); err == nil {
					klog.Infof("Added certificates in field '%s' of '%s/%s' secret to trusted RootCA's", fieldToCompare, newSecret.Namespace, newSecret.Name)
					return true
				}
				klog.Errorf("Failed adding certs in field '%s' of '%s/%s' secret: %v", fieldToCompare, newSecret.Namespace, newSecret.Name, err)
			}
		} else {
			// If filed was not present in old secret but is in new secret then is an addition, we trust it
			if err := c.addTLSCertificatesToTrustInTransport(newPublicCert); err == nil {
				klog.Infof("Added certificates in field '%s' of '%s/%s' secret to trusted RootCA's", fieldToCompare, newSecret.Namespace, newSecret.Name)
				return true
			}
			klog.Errorf("Failed adding certs in field %s of '%s/%s' secret: %v", fieldToCompare, newSecret.Namespace, newSecret.Name, err)
		}
	}
	return false
}

func (c *Controller) trustPEMInSecretField(secret *corev1.Secret, fieldToCompare string) bool {
	newPublicCert, err := getFileFromSecretDataField(secret.Data, fieldToCompare)
	if err == nil {
		if err := c.addTLSCertificatesToTrustInTransport(newPublicCert); err == nil {
			klog.Infof("Added certificates in field '%s' of '%s/%s' secret to trusted RootCA's", fieldToCompare, secret.Namespace, secret.Name)
			return true
		}
		klog.Errorf("Failed adding certs in field '%s' of '%s/%s' secret: %v", fieldToCompare, secret.Namespace, secret.Name, err)
	}
	return false
}

func (c *Controller) addTLSCertificatesToTrustInTransport(certificateData []byte) error {
	var x509Certs []*x509.Certificate
	current := certificateData
	// A single PEM file could contain more than one certificate, keeping track of the index to help debugging
	certIndex := 1
	for len(current) > 0 {
		var pemBlock *pem.Block
		if pemBlock, current = pem.Decode(current); pemBlock == nil {
			return fmt.Errorf("invalid PEM in file in index %d", certIndex)
		}
		x509Cert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing x509 certificate from PEM in index, %d: %v", certIndex, err)
		}
		x509Certs = append(x509Certs, x509Cert)
		certIndex++
	}
	for _, cert := range x509Certs {
		c.getTransport().TLSClientConfig.RootCAs.AddCert(cert)
	}
	return nil
}

// GetSignerCAFromSecret Retrieves the CA certificate for Openshift CSR signed certificates from
// openshift-kube-controller-manager-operator namespace
func (c *Controller) GetSignerCAFromSecret() ([]byte, error) {
	var caCertificate []byte
	openShiftCATLSCertSecret, err := c.kubeClientSet.CoreV1().Secrets(OpenshiftKubeControllerNamespace).Get(
		context.Background(), OpenshiftCATLSSecretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s secret in %s wasn't found", OpenshiftCATLSSecretName, OpenshiftKubeControllerNamespace)
		}
		return nil, fmt.Errorf("%s secret was found but we failed to load the secret: %#v", OpenshiftCATLSSecretName, err)
	} else if openShiftCATLSCertSecret != nil {
		if val, ok := openShiftCATLSCertSecret.Data[certs.TLSCertFile]; ok {
			caCertificate = val
		}
	}
	return caCertificate, nil
}

// GetOpenshiftCSRSignerCAFromSecret loads the tls certificate in openshift-csr-signer-ca secret in operator namespace
func (c *Controller) GetOpenshiftCSRSignerCAFromSecret() ([]byte, error) {
	var caCertificate []byte
	operatorNamespace := miniov2.GetNSFromFile()
	openShifCSRSignerCATLSCertSecret, err := c.kubeClientSet.CoreV1().Secrets(operatorNamespace).Get(
		context.Background(), OperatorCSRSignerCASecretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s secret wasn't found, Skip load the certificate", OperatorCSRSignerCASecretName)
		}
		// Lack of permissions to read the secret
		return nil, fmt.Errorf("%s secret was found but we failed to get certificate: %#v", OperatorCSRSignerCASecretName, err)
	} else if openShifCSRSignerCATLSCertSecret != nil {
		// When secret was obtained with no errors
		if val, ok := openShifCSRSignerCATLSCertSecret.Data[certs.TLSCertFile]; ok {
			// OpenShift csr-signer secret has tls.crt certificates that we need to append in order
			// to trust the signer. If we append the val, Operator will be able to provisioning the
			// initial users and get Tenant Health, so tenant can be properly initialized and in
			// green status, otherwise if we don't append it, it will get stuck and expose this
			// issue in the log:
			// Failed to get cluster health: Get "https://minio.tenant-lite.svc.cluster.local/minio/health/cluster":
			// x509: certificate signed by unknown authority
			caCertificate = val
		}
	}
	return caCertificate, nil
}

// checkOpenshiftSignerCACertInOperatorNamespace checks if csr-signer secret in openshift changed and updates or create
// a copy of the secret in operator namespace
func (c *Controller) checkOpenshiftSignerCACertInOperatorNamespace(ctx context.Context) error {
	// get the current certificate from openshift
	csrSignerCertificate, err := c.GetSignerCAFromSecret()
	if err != nil {
		return err
	}
	namespace := miniov2.GetNSFromFile()
	// get openshift-csr-signer-ca secret in minio-operator namespace
	csrSignerSecret, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(ctx, OperatorCSRSignerCASecretName, metav1.GetOptions{})
	if err != nil {
		// if csrSignerCa doesnt exists create it
		if k8serrors.IsNotFound(err) {
			klog.Infof("'%s/%s' secret is missing, creating", namespace, OperatorCSRSignerCASecretName)
			operatorDeployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(ctx, getOperatorDeploymentName(), metav1.GetOptions{})
			if err != nil {
				return err
			}

			ownerReference := metav1.OwnerReference{
				APIVersion: appsv1.SchemeGroupVersion.Version,
				Kind:       "Deployment",
				Name:       operatorDeployment.Name,
				UID:        operatorDeployment.UID,
			}

			csrSignerSecret := &v1.Secret{
				Type: "Opaque",
				ObjectMeta: metav1.ObjectMeta{
					Name:            OperatorCSRSignerCASecretName,
					Namespace:       miniov2.GetNSFromFile(),
					OwnerReferences: []metav1.OwnerReference{ownerReference},
				},
				Data: map[string][]byte{
					certs.TLSCertFile: csrSignerCertificate,
				},
			}
			_, err = c.kubeClientSet.CoreV1().Secrets(namespace).Create(ctx, csrSignerSecret, metav1.CreateOptions{})
			// Reload CA certificates
			c.createTransport()
			return err
		}
		return err
	}

	if caCert, ok := csrSignerSecret.Data[certs.TLSCertFile]; ok && !bytes.Equal(caCert, csrSignerCertificate) {
		csrSignerSecret.Data[certs.TLSCertFile] = csrSignerCertificate
		_, err = c.kubeClientSet.CoreV1().Secrets(namespace).Update(ctx, csrSignerSecret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		klog.Infof("'%s/%s' secret changed, updating '%s/%s' secret", OpenshiftKubeControllerNamespace, OpenshiftCATLSSecretName, namespace, OperatorCSRSignerCASecretName)
		// Reload CA certificates
		c.createTransport()
	}
	return nil
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

func (c *Controller) createBuckets(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (created bool, err error) {
	tenantBuckets := tenant.Spec.Buckets
	if len(tenantBuckets) == 0 {
		return false, nil
	}
	// get a new admin client
	minioClient, err := tenant.NewMinIOUser(tenantConfiguration, c.getTransport())
	if err != nil {
		// show the error and continue
		klog.Errorf("Error instantiating minio Client: %v ", err)
		return false, err
	}
	created, err = tenant.CreateBuckets(minioClient, tenantBuckets...)
	if err != nil {
		klog.V(2).Infof("Unable to create MinIO buckets: %v", err)
		if _, terr := c.updateTenantStatus(ctx, tenant, StatusProvisioningDefaultBuckets, 0); terr != nil {
			return false, err
		}
		return false, err
	}
	if created {
		if _, err = c.updateProvisionedBucketStatus(ctx, tenant, true); err != nil {
			klog.V(2).Infof(err.Error())
		}
	}
	return created, err
}

// getOperatorDeploymentName Internal func returns the Operator deployment name from MINIO_OPERATOR_DEPLOYMENT_NAME ENV variable or the default name
func getOperatorDeploymentName() string {
	return env.Get(OperatorDeploymentNameEnv, DefaultDeploymentName)
}
