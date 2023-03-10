// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/controller/certificates"
	xcerts "github.com/minio/pkg/certs"
	"github.com/minio/pkg/env"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// waitForCertSecretReady Function designed to run in a non-leader operator container to wait for the leader to issue a TLS certificate
func (c *Controller) waitForCertSecretReady(serviceName string, secretName string) (*string, *string) {
	ctx := context.Background()
	namespace := miniov2.GetNSFromFile()
	var publicCertPath, publicKeyPath string

	for {
		tlsCertSecret, err := c.getCertificateSecret(ctx, namespace, secretName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.Infof("Waiting for the %s certificates secret to be issued", serviceName)
				time.Sleep(time.Second * 10)
			} else {
				klog.Infof(err.Error())
			}
		} else {
			publicCertPath, publicKeyPath = c.writeCertSecretToFile(tlsCertSecret, serviceName)
			break
		}
	}

	// validate certificates if they are valid, if not panic right here.
	if _, err := tls.LoadX509KeyPair(publicCertPath, publicKeyPath); err != nil {
		panic(err)
	}

	return &publicCertPath, &publicKeyPath
}

// getCertificateSecret gets a TLS Certificate secret
func (c *Controller) getCertificateSecret(ctx context.Context, namespace string, secretName string) (*corev1.Secret, error) {
	return c.kubeClientSet.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
}

// writeCertSecretToFile receives a [corev1.Secret] and save it's contain to the filesystem.
// returns publicCertPath (filesystem path to the public certificate file), publicKeyPath, (filesystem path to the private key file)
func (c *Controller) writeCertSecretToFile(tlsCertSecret *corev1.Secret, serviceName string) (string, string) {
	mkdirerr := os.MkdirAll(fmt.Sprintf("/tmp/%s", serviceName), 0o777)
	if mkdirerr != nil {
		panic(mkdirerr)
	}

	publicCertPath := fmt.Sprintf("/tmp/%s/public.crt", serviceName)
	publicKeyPath := fmt.Sprintf("/tmp/%s/private.key", serviceName)
	publicCertKey, privateKeyKey := c.getKeyNames(tlsCertSecret)

	if val, ok := tlsCertSecret.Data[publicCertKey]; ok {
		err := os.WriteFile(publicCertPath, val, 0o644)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("missing '%s' in %s/%s", publicCertKey, tlsCertSecret.Namespace, tlsCertSecret.Name))
	}
	if val, ok := tlsCertSecret.Data[privateKeyKey]; ok {
		err := os.WriteFile(publicKeyPath, val, 0o644)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("missing '%s' in %s/%s", privateKeyKey, tlsCertSecret.Namespace, tlsCertSecret.Name))
	}
	return publicCertPath, publicKeyPath
}

// generateTLSCert Generic method to generate TLS Certificartes for different services
func (c *Controller) generateTLSCert(serviceName string, secretName string, deploymentName string) (*string, *string) {
	ctx := context.Background()
	namespace := miniov2.GetNSFromFile()
	csrName := getCSRName(serviceName)
	var publicCertPath, publicKeyPath string
	// operator deployment for owner reference
	operatorDeployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for {
		// TLS certificates
		tlsCertSecret, err := c.getCertificateSecret(ctx, namespace, secretName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if k8serrors.IsNotFound(err) {
					klog.Infof("%s TLS secret not found: %v", secretName, err)
				}
				if err = c.checkAndCreateCSR(ctx, operatorDeployment, serviceName, csrName, secretName); err != nil {
					klog.Infof("Waiting for the %s certificates to be issued %v", serviceName, err.Error())
					time.Sleep(time.Second * 10)
				} else {
					err = c.deleteCSR(ctx, csrName)
					if err != nil {
						klog.Infof(err.Error())
					}
				}
			}
		} else {
			publicCertPath, publicKeyPath = c.writeCertSecretToFile(tlsCertSecret, serviceName)
			break
		}
	}

	// validate certificates if they are valid, if not panic right here.
	if _, err = tls.LoadX509KeyPair(publicCertPath, publicKeyPath); err != nil {
		panic(err)
	}

	return &publicCertPath, &publicKeyPath
}

// checkAndCreateCSR Queries for the Certificate signing request
func (c *Controller) checkAndCreateCSR(ctx context.Context, deployment metav1.Object, serviceName string, csrName string, secretName string) error {
	var err error
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, csrName, metav1.GetOptions{})
	} else {
		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, csrName, metav1.GetOptions{})
	}
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Infof("Creating a new Certificate Signing Request for %s Server Certs, cluster %q", serviceName)
			if err = c.createAndStoreCSR(ctx, deployment, serviceName, csrName, secretName); err != nil {
				return err
			}
			return fmt.Errorf("waiting for %s cert", serviceName)
		}
		return err
	}
	return nil
}

// getKeyNames Identify the K8s secret keys containing the public and private TLS certificate keys
func (c *Controller) getKeyNames(tlsCertificateSecret *corev1.Secret) (string, string) {
	// default secret keys for Opaque k8s secret
	publicCertKey := "public.crt"
	privateKeyKey := "private.key"
	// if secret type is k8s tls or cert-manager use the right secret keys
	if tlsCertificateSecret.Type == "kubernetes.io/tls" || tlsCertificateSecret.Type == "cert-manager.io/v1alpha2" || tlsCertificateSecret.Type == "cert-manager.io/v1" {
		publicCertKey = "tls.crt"
		privateKeyKey = "tls.key"
	}
	return publicCertKey, privateKeyKey
}

// createCertificateSecret Stores the private and public keys in a Secret
func (c *Controller) createCertificateSecret(ctx context.Context, deployment metav1.Object, labels map[string]string, secretName string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: miniov2.GetNSFromFile(),
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(deployment, schema.GroupVersionKind{
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

// createAndStoreCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret storing private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createAndStoreCSR(ctx context.Context, deployment metav1.Object, serviceName string, csrName string, secretName string) error {
	privKeysBytes, csrBytes, err := generateCSRCryptoData(serviceName)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}
	namespace := miniov2.GetNSFromFile()
	err = c.createCertificateSigningRequest(ctx, map[string]string{}, csrName, namespace, csrBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", csrName, err)
		return err
	}

	// fetch certificate from CSR
	certBytes, err := c.fetchCertificate(ctx, csrName)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", csrName, err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivateKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret
	err = c.createCertificateSecret(ctx, deployment, map[string]string{}, secretName, encodedPrivateKey, certBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the %s/%s secret: %v", deployment.GetNamespace(), secretName, err)
		return err
	}
	return nil
}

// createTLSConfig defines the tls.Config for the http servers.
// Defines the http Protocols, TLS versions, Cipher suites between other TLS settings
func (c *Controller) createTLSConfig(certsManager *xcerts.Manager) *tls.Config {
	config := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.CurveP256},
		NextProtos:               []string{"h2", "http/1.1"},
		MinVersion:               tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		GetCertificate: certsManager.GetCertificate,
	}
	return config
}

// LoadX509KeyPair Internal func load an X509 key pair (private key , certificate)
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
		password := env.Get(CertPasswordEnv, "")
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

// getCSRName Internal func to return the CSR name
func getCSRName(serviceName string) string {
	namespace := miniov2.GetNSFromFile()
	return fmt.Sprintf("%s-%s-csr", serviceName, namespace)
}

// generateCSRCryptoData Internal func Creates the private Key material
func generateCSRCryptoData(serviceName string) ([]byte, []byte, error) {
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

	opCommon := fmt.Sprintf("%s.%s.svc.%s", serviceName, miniov2.GetNSFromFile(), miniov2.GetClusterDomain())
	opCommonNoDomain := fmt.Sprintf("%s.%s.svc", serviceName, miniov2.GetNSFromFile())

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", opCommonNoDomain),
			Organization: []string{"system:nodes"},
		},
		Extensions: []pkix.Extension{
			{
				Id:       nil,
				Critical: false,
				Value:    []byte(serviceName),
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
		DNSNames:           []string{serviceName, opCommonNoDomain, opCommon},
	}

	// This sections addresses the edge case "remote error: tls: bad certificate" error in  https://github.com/minio/operator/issues/1234
	// Openshift OLM creates an additional service `minio-operator-service`, there is no option to choose the name of the service
	// This additional DNSName is to handle the calls from kube-apiserver through that service, which is the easiest way to have a clean operator fresh install.
	if serviceName == "operator" {
		openshiftWebhookServiceCommon := fmt.Sprintf("minio-operator-service.%s.svc", miniov2.GetNSFromFile())
		csrTemplate.Extensions = append(csrTemplate.Extensions, pkix.Extension{
			Id:       nil,
			Critical: false,
			Value:    []byte(openshiftWebhookServiceCommon),
		})
		csrTemplate.DNSNames = append(csrTemplate.DNSNames, openshiftWebhookServiceCommon)
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}
