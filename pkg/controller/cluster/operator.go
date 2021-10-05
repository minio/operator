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
	"io/ioutil"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/version"

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
	// OperatorTLSSecretName is the name of secret created with Operator TLS certs
	OperatorTLSSecretName = "operator-tls"
	// DefaultDeploymentName is the default name of the operator deployment
	DefaultDeploymentName = "minio-operator"
)

var (
	errOperatorWaitForTLS = errors.New("waiting for Operator cert")
)

func getOperatorDeploymentName() string {
	return env.Get("MINIO_OPERATOR_DEPLOYMENT_NAME", DefaultDeploymentName)
}

func isOperatorTLS() bool {
	value, set := os.LookupEnv(OperatorTLS)
	// By default Operator TLS is used.
	return (set && value == "on") || !set
}

var kubeAPIServerVersion *version.Info
var useCertificatesV1API bool

func (c *Controller) getKubeAPIServerVersion() {
	var err error
	kubeAPIServerVersion, err = c.kubeClientSet.Discovery().ServerVersion()
	if err != nil {
		panic(err)
	}
	useCertificatesV1API = versionCompare(kubeAPIServerVersion.String(), "v1.22.0") >= 0
}

func (c *Controller) generateTLSCert() (string, string) {
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
					if useCertificatesV1API {
						if err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Delete(ctx, c.operatorCSRName(), metav1.DeleteOptions{}); err != nil {
							klog.Infof(err.Error())
						}
					} else {
						if err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Delete(ctx, c.operatorCSRName(), metav1.DeleteOptions{}); err != nil {
							klog.Infof(err.Error())
						}
					}
				}
			}
		} else {
			if val, ok := operatorTLSCert.Data["public.crt"]; ok {
				err := ioutil.WriteFile(publicCertPath, val, 0644)
				if err != nil {
					panic(err)
				}
			} else {
				panic(errors.New("operator TLS 'public.crt' missing"))
			}

			if val, ok := operatorTLSCert.Data["private.key"]; ok {
				err := ioutil.WriteFile(publicKeyPath, val, 0644)
				if err != nil {
					panic(err)
				}
			} else {
				panic(errors.New("operator TLS 'private.key' missing"))
			}
			break
		}
	}

	// validate certificates if they are valid, if not panic right here.
	if _, err = tls.LoadX509KeyPair(publicCertPath, publicKeyPath); err != nil {
		panic(err)
	}

	return publicCertPath, publicKeyPath
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
	err = c.createCertificateSigningRequest(ctx, map[string]string{}, c.operatorCSRName(), namespace, csrBytes, "server")
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
	err = c.createOperatorSecret(ctx, operator, map[string]string{}, "operator-tls", encodedPrivateKey, certBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", "operator-tls", err)
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateOperatorCSR(ctx context.Context, operator metav1.Object) error {
	var err error
	if useCertificatesV1API {
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

func (c *Controller) createUsers(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) error {
	var userCredentials []*v1.Secret
	for _, credential := range tenant.Spec.Users {
		credentialSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, credential.Name, metav1.GetOptions{})
		if err == nil && credentialSecret != nil {
			userCredentials = append(userCredentials, credentialSecret)
		}
	}

	// get mc admin info
	adminClnt, err := tenant.NewMinIOAdmin(tenantConfiguration)
	if err != nil {
		// show the error and continue
		klog.Errorf("Error instantiating madmin: %v", err.Error())
	}

	skipCreateUsers := false
	// configuration that means MinIO is running with LDAP enabled
	// and we need to skip the console user creation
	for _, env := range tenant.GetEnvVars() {
		if env.Name == "MINIO_IDENTITY_LDAP_SERVER_ADDR" && env.Value != "" {
			skipCreateUsers = true
			break
		}
	}

	if err := tenant.CreateUsers(adminClnt, userCredentials, skipCreateUsers); err != nil {
		klog.V(2).Infof("Unable to create MinIO users: %v", err)
		return err
	}

	return nil
}

func (c *Controller) operatorCSRName() string {
	namespace := miniov2.GetNSFromFile()
	return fmt.Sprintf("operator-%s-csr", namespace)
}
