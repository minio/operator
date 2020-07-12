/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cluster

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"

	certificates "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	csrType        = "CERTIFICATE REQUEST"
	privateKeyType = "PRIVATE KEY"
)

// newPrivateKey returns randomly generated ECDSA private key.
func newPrivateKey(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(curve, rand.Reader)
}

func isEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func generateCryptoData(mi *miniov1.MinIOInstance, hostsTemplate string) ([]byte, []byte, error) {
	var dnsNames []string
	klog.V(0).Infof("Generating private key")
	privateKey, err := newPrivateKey(miniov1.DefaultEllipticCurve)
	if err != nil {
		klog.Errorf("Unexpected error during the ECDSA Key generation: %v", err)
		return nil, nil, err
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during encoding the ECDSA Private Key: %v", err)
		return nil, nil, err
	}

	hosts := mi.AllMinIOHosts()
	if hostsTemplate != "" {
		hosts = mi.TemplatedMinIOHosts(hostsTemplate)
	}

	if isEqual(mi.Spec.CertConfig.DNSNames, hosts) {
		dnsNames = mi.Spec.CertConfig.DNSNames
	} else {
		dnsNames = append(mi.Spec.CertConfig.DNSNames, hosts...)
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   mi.Spec.CertConfig.CommonName,
			Organization: mi.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           dnsNames,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that MinIO statefulset will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createCSR(ctx context.Context, mi *miniov1.MinIOInstance) error {
	privKeysBytes, csrBytes, err := generateCryptoData(mi, c.hostsTemplate)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificate(ctx, mi.MinIOPodLabels(), mi.MinIOCSRName(), mi.Namespace, csrBytes, mi)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, mi.MinIOCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOCSRName(), err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for MinIO Statefulset to use
	err = c.createSecret(ctx, mi, mi.MinIOPodLabels(), mi.MinIOTLSSecretName(), mi.Namespace, encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", mi.MinIOTLSSecretName(), err)
		return err
	}

	return nil
}

// createCertificate is equivalent to kubectl create <csr-name> and kubectl approve csr <csr-name>
func (c *Controller) createCertificate(ctx context.Context, labels map[string]string, name, namespace string, csrBytes []byte, mi *miniov1.MinIOInstance) error {
	encodedBytes := pem.EncodeToMemory(&pem.Block{Type: csrType, Bytes: csrBytes})

	kubeCSR := &certificates.CertificateSigningRequest{
		TypeMeta: v1.TypeMeta{
			APIVersion: "certificates.k8s.io/v1beta1",
			Kind:       "CertificateSigningRequest",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Labels:    labels,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1.SchemeGroupVersion.Group,
					Version: miniov1.SchemeGroupVersion.Version,
					Kind:    miniov1.MinIOCRDResourceKind,
				}),
			},
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

	ks, err := c.certClient.CertificateSigningRequests().Create(ctx, kubeCSR, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Update the CSR to be approved automatically
	ks.Status = certificates.CertificateSigningRequestStatus{
		Conditions: []certificates.CertificateSigningRequestCondition{
			{
				Type:           certificates.CertificateApproved,
				Reason:         "MinIOOperatorAutoApproval",
				Message:        "Automatically approved by MinIO Operator",
				LastUpdateTime: metav1.NewTime(time.Now()),
			},
		},
	}

	_, err = c.certClient.CertificateSigningRequests().UpdateApproval(ctx, ks, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// FetchCertificate fetches the generated certificate from the CSR
func (c *Controller) fetchCertificate(ctx context.Context, csrName string) ([]byte, error) {
	klog.V(0).Infof("Start polling for certificate of csr/%s, every %s, timeout after %s", csrName,
		miniov1.DefaultQueryInterval, miniov1.DefaultQueryTimeout)

	tick := time.NewTicker(miniov1.DefaultQueryInterval)
	defer tick.Stop()

	timeout := time.NewTimer(miniov1.DefaultQueryTimeout)
	defer tick.Stop()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	for {
		select {
		case s := <-ch:
			klog.Infof("Signal %s received, exiting ...", s.String())
			return nil, fmt.Errorf("%s", s.String())

		case <-tick.C:
			r, err := c.certClient.CertificateSigningRequests().Get(ctx, csrName, v1.GetOptions{})
			if err != nil {
				klog.Errorf("Unexpected error during certificate fetching of csr/%s: %s", csrName, err)
				return nil, err
			}
			if r.Status.Certificate != nil {
				klog.V(0).Infof("Certificate successfully fetched, creating secret with Private key and Certificate")
				return r.Status.Certificate, nil
			}
			for _, c := range r.Status.Conditions {
				if c.Type == certificates.CertificateDenied {
					err := fmt.Errorf("csr/%s uid: %s is %q: %s", r.Name, r.UID, c.Type, c.String())
					klog.Errorf("Unexpected error during fetch: %v", err)
					return nil, err
				}
			}
			klog.V(1).Infof("Certificate of csr/%s still not available, next try in %d", csrName, miniov1.DefaultQueryInterval)

		case <-timeout.C:
			return nil, fmt.Errorf("timeout during certificate fetching of csr/%s", csrName)
		}
	}
}

func (c *Controller) createSecret(ctx context.Context, mi *miniov1.MinIOInstance, labels map[string]string, name, namespace string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mi, schema.GroupVersionKind{
					Group:   miniov1.SchemeGroupVersion.Group,
					Version: miniov1.SchemeGroupVersion.Version,
					Kind:    miniov1.MinIOCRDResourceKind,
				}),
			},
		},
		Data: map[string][]byte{
			"private.key": pkBytes,
			"public.crt":  certBytes,
		},
	}
	if _, err := c.kubeClientSet.CoreV1().Secrets(mi.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func parseCertificate(r io.Reader) (*x509.Certificate, error) {
	certPEMBlock, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			return x509.ParseCertificate(certDERBlock.Bytes)
		}
	}
	return nil, errors.New("found no (non-CA) certificate in any PEM block")
}
