/*
 * MinIO-Operator - Manage MinIO clusters in Kubernetes
 *
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"

	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	"github.com/minio/minio-operator/pkg/constants"

	certificates "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	csrType        = "CERTIFICATE REQUEST"
	privateKeyType = "PRIVATE KEY"
)

// newPrivateKey returns randomly generated ECDSA private key.
func newPrivateKey(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(curve, rand.Reader)
}

func generateCryptoData(mi *miniov1beta1.MinIOInstance, serviceName string) ([]byte, []byte, error) {
	glog.V(0).Infof("Generating private key")
	privateKey, err := newPrivateKey(constants.DefaultEllipticCurve)
	if err != nil {
		glog.Errorf("Unexpected error during the ECDSA Key generation: %v", err)
		return nil, nil, err
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		glog.Errorf("Unexpected error during encoding the ECDSA Private Key: %v", err)
		return nil, nil, err
	}

	glog.V(0).Infof("Generating CSR with CN=%s", mi.Spec.CertConfig.CommonName)
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   mi.Spec.CertConfig.CommonName,
			Organization: mi.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           mi.Spec.CertConfig.DNSNames,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		glog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that MinIO statefulset will use to mount private key and certificate for TLS
func (c *Controller) createCSR(mi *miniov1beta1.MinIOInstance, serviceName string) error {
	privKeysBytes, csrBytes, err := generateCryptoData(mi, serviceName)
	if err != nil {
		glog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}
	csrString := string(csrBytes)
	glog.V(2).Infof("Creating csr/%s:\n%s", mi.Name, csrString)
	err = c.submitCSR(mi.Name, csrBytes)
	if err != nil {
		glog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.Name, err)
		return err
	}
	glog.V(0).Infof("Successfully created csr/%s", mi.Name)

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(mi.Name)
	if err != nil {
		glog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.Name, err)
		return err
	}
	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})
	// Create secret for MinIO Statefulset to use
	err = c.createSecret(mi.GetTLSSecretName(), mi.Namespace, encodedPrivKey, certbytes)
	if err != nil {
		glog.Errorf("Unexpected error during the creation of the secret/%s: %v", mi.GetTLSSecretName(), err)
		return err
	}
	return nil
}

// SubmitCSR is equivalent to kubectl create ${CSR}, if the override is configured, it becomes kubectl apply ${CSR}
func (c *Controller) submitCSR(csrName string, csrBytes []byte) error {
	encodedBytes := pem.EncodeToMemory(&pem.Block{Type: csrType, Bytes: csrBytes})

	kubeCSR := &certificates.CertificateSigningRequest{
		TypeMeta: v1.TypeMeta{
			APIVersion: "certificates.k8s.io/v1beta1",
			Kind:       "CertificateSigningRequest",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: csrName,
		},
		Spec: certificates.CertificateSigningRequestSpec{
			Request: encodedBytes,
			Groups:  []string{"system:authenticated"},
			Usages: []certificates.KeyUsage{
				certificates.UsageDigitalSignature,
				certificates.UsageServerAuth,
				certificates.UsageClientAuth,
			},
		},
	}

	_, err := c.certClient.CertificateSigningRequests().Create(kubeCSR)
	if err != nil {
		return err
	}
	return nil
}

// FetchCertificate fetches the generated certificate from the CSR
func (c *Controller) fetchCertificate(csrName string) ([]byte, error) {
	glog.V(0).Infof("Start polling for certificate of csr/%s, every %s, timeout after %s", csrName,
		constants.DefaultQueryInterval, constants.DefaultQueryTimeout)

	tick := time.NewTicker(constants.DefaultQueryInterval)
	defer tick.Stop()

	timeout := time.NewTimer(constants.DefaultQueryTimeout)
	defer tick.Stop()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	for {
		select {
		case s := <-ch:
			glog.Infof("Signal %s received, exiting ...", s.String())
			return nil, fmt.Errorf("%s", s.String())

		case <-tick.C:
			r, err := c.certClient.CertificateSigningRequests().Get(csrName, v1.GetOptions{})
			if err != nil {
				glog.Errorf("Unexpected error during certificate fetching of csr/%s: %s", csrName, err)
				return nil, err
			}
			if r.Status.Certificate != nil {
				glog.V(0).Infof("Certificate successfully fetched, creating secret with Private key and Certificate")
				return r.Status.Certificate, nil
			}
			for _, c := range r.Status.Conditions {
				if c.Type == certificates.CertificateDenied {
					err := fmt.Errorf("csr/%s uid: %s is %q: %s", r.Name, r.UID, c.Type, c.String())
					glog.Errorf("Unexpected error during fetch: %v", err)
					return nil, err
				}
			}
			glog.V(1).Infof("Certificate of csr/%s still not available, next try in %d", csrName, constants.DefaultQueryInterval)

		case <-timeout.C:
			return nil, fmt.Errorf("timeout during certificate fetching of csr/%s", csrName)
		}
	}
}

func (c *Controller) createSecret(secretName, nameSpace string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: nameSpace,
		},
		Data: map[string][]byte{
			"private.key": pkBytes,
			"public.crt":  certBytes,
		},
	}
	_, err := c.kubeClientSet.CoreV1().Secrets(nameSpace).Create(secret)
	if err != nil {
		return err
	}
	return nil
}
