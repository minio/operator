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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	xcerts "github.com/minio/pkg/certs"

	"github.com/minio/operator/pkg/controller/cluster/certificates"

	"github.com/minio/pkg/env"

	"k8s.io/apimachinery/pkg/runtime/schema"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"k8s.io/klog/v2"
)

const (
	// OperatorTLS is the ENV var to turn on / off Operator TLS.
	OperatorTLS = "MINIO_OPERATOR_TLS_ENABLE"
	// EnvCertPassword ENV var is used to decrypt the private key in the TLS certificate for operator if need it
	EnvCertPassword = "OPERATOR_CERT_PASSWD"
	// OperatorTLSSecretName is the name of secret created with Operator TLS certs
	OperatorTLSSecretName = "operator-tls"
	// OperatorCATLSSecretName is the name of the secret for the operator CA
	OperatorCATLSSecretName = "operator-ca-tls"
	// DefaultDeploymentName is the default name of the operator deployment
	DefaultDeploymentName = "minio-operator"
)

var errOperatorWaitForTLS = errors.New("waiting for Operator cert")

func getOperatorDeploymentName() string {
	return env.Get("MINIO_OPERATOR_DEPLOYMENT_NAME", DefaultDeploymentName)
}

func isOperatorTLS() bool {
	value, set := os.LookupEnv(OperatorTLS)
	// By default, Operator TLS is used.
	return (set && value == "on") || !set
}

func (c *Controller) generateTLSCert() (*string, *string) {
	ctx := context.Background()
	namespace := miniov2.GetNSFromFile()
	// operator deployment for owner reference
	operatorDeployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(ctx, getOperatorDeploymentName(), metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	publicCertPath := "/tmp/public.crt"
	publicKeyPath := "/tmp/private.key"

	for {
		// operator TLS certificates
		operatorTLSCert, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(ctx, OperatorTLSSecretName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.Infof("operator TLS secret not found: %v", err)
				if err = c.checkAndCreateOperatorCSR(ctx, operatorDeployment); err != nil {
					klog.Infof("Waiting for the operator certificates to be issued %v", err.Error())
					time.Sleep(time.Second * 10)
				} else {
					err = c.deleteCSR(ctx, c.operatorCSRName())
					if err != nil {
						klog.Infof(err.Error())
					}
				}
			}
		} else {
			// default secret keys for Opaque k8s secret
			publicCertKey := "public.crt"
			privateKeyKey := "private.key"
			// if secret type is k8s tls or cert-manager use the right secret keys
			if operatorTLSCert.Type == "kubernetes.io/tls" || operatorTLSCert.Type == "cert-manager.io/v1alpha2" || operatorTLSCert.Type == "cert-manager.io/v1" {
				publicCertKey = "tls.crt"
				privateKeyKey = "tls.key"
			}
			if val, ok := operatorTLSCert.Data[publicCertKey]; ok {
				err := os.WriteFile(publicCertPath, val, 0o644)
				if err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("missing '%s' in %s/%s", publicCertKey, operatorTLSCert.Namespace, operatorTLSCert.Name))
			}
			if val, ok := operatorTLSCert.Data[privateKeyKey]; ok {
				err := os.WriteFile(publicKeyPath, val, 0o644)
				if err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("missing '%s' in %s/%s", publicCertKey, operatorTLSCert.Namespace, operatorTLSCert.Name))
			}
			break
		}
	}

	// validate certificates if they are valid, if not panic right here.
	if _, err = tls.LoadX509KeyPair(publicCertPath, publicKeyPath); err != nil {
		panic(err)
	}

	return &publicCertPath, &publicKeyPath
}

func generateOperatorCryptoData() ([]byte, []byte, error) {
	privateKey, err := newPrivateKey(miniov2.DefaultEllipticCurve)
	if err != nil {
		klog.Errorf("Unexpected error during the ECDSA Key generation: %v", err)
		return nil, nil, err
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during encoding the ECDSA Private Key: %v", err)
		return nil, nil, err
	}

	opCommon := fmt.Sprintf("operator.%s.svc.%s", miniov2.GetNSFromFile(), miniov2.GetClusterDomain())
	opCommonNoDomain := fmt.Sprintf("operator.%s.svc", miniov2.GetNSFromFile())

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", opCommonNoDomain),
			Organization: []string{"system:nodes"},
		},
		Extensions: []pkix.Extension{
			{
				Id:       nil,
				Critical: false,
				Value:    []byte("operator"),
			},
			{
				Id:       nil,
				Critical: false,
				Value:    []byte(opCommonNoDomain),
			},
			{
				Id:       nil,
				Critical: false,
				Value:    []byte(opCommon),
			},
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           []string{"operator", opCommonNoDomain, opCommon},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

func (c *Controller) createOperatorSecret(ctx context.Context, operator metav1.Object, labels map[string]string, secretName string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: miniov2.GetNSFromFile(),
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(operator, schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				}),
			},
		},
		Data: map[string][]byte{
			"private.key": pkBytes,
			"public.crt":  certBytes,
		},
	}
	_, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// createOperatorCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that Operator deployment will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createOperatorCSR(ctx context.Context, operator metav1.Object) error {
	privKeysBytes, csrBytes, err := generateOperatorCryptoData()
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}
	namespace := miniov2.GetNSFromFile()
	err = c.createCertificateSigningRequest(ctx, map[string]string{}, c.operatorCSRName(), namespace, csrBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", c.operatorCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certBytes, err := c.fetchCertificate(ctx, c.operatorCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", c.operatorCSRName(), err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivateKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for operator to use
	err = c.createOperatorSecret(ctx, operator, map[string]string{}, OperatorTLSSecretName, encodedPrivateKey, certBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the %s/%s secret: %v", operator.GetNamespace(), OperatorTLSSecretName, err)
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateOperatorCSR(ctx context.Context, operator metav1.Object) error {
	var err error
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, c.operatorCSRName(), metav1.GetOptions{})
	} else {
		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, c.operatorCSRName(), metav1.GetOptions{})
	}
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Infof("Creating a new Certificate Signing Request for Operator Server Certs, cluster %q")
			if err = c.createOperatorCSR(ctx, operator); err != nil {
				return err
			}
			return errOperatorWaitForTLS
		}
		return err
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

func (c *Controller) operatorCSRName() string {
	namespace := miniov2.GetNSFromFile()
	return fmt.Sprintf("operator-%s-csr", namespace)
}

// recreateOperatorCertsIfRequired - Generate Operator TLS certs if not present or if is expired
func (c *Controller) recreateOperatorCertsIfRequired(ctx context.Context) error {
	namespace := miniov2.GetNSFromFile()
	operatorTLSSecret, err := c.getTLSSecret(ctx, namespace, OperatorTLSSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Info("TLS certificate not found. Generating one.")
			// Generate new certificate KeyPair for Operator server
			c.generateTLSCert()
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
	err = c.deleteCSR(ctx, c.operatorCSRName())
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
	c.generateTLSCert()

	// reload in memory certificate for the operator server
	if serverCertsManager != nil {
		serverCertsManager.ReloadCerts()
	}

	return nil
}

var serverCertsManager *xcerts.Manager

// LoadX509KeyPair - load an X509 key pair (private key , certificate)
// from the provided paths. The private key may be encrypted and is
// decrypted using the ENV_VAR: OPERATOR_CERT_PASSWD.
func LoadX509KeyPair(certFile, keyFile string) (tls.Certificate, error) {
	certPEMBlock, err := os.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEMBlock, err := os.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	key, rest := pem.Decode(keyPEMBlock)
	if len(rest) > 0 {
		return tls.Certificate{}, errors.New("the private key contains additional data")
	}
	if x509.IsEncryptedPEMBlock(key) {
		password := env.Get(EnvCertPassword, "")
		if len(password) == 0 {
			return tls.Certificate{}, errors.New("no password")
		}
		decryptedKey, decErr := x509.DecryptPEMBlock(key, []byte(password))
		if decErr != nil {
			return tls.Certificate{}, decErr
		}
		keyPEMBlock = pem.EncodeToMemory(&pem.Block{Type: key.Type, Bytes: decryptedKey})
	}
	return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}
