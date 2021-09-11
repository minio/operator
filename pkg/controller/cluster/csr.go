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
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	certificatesV1 "k8s.io/api/certificates/v1"

	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	certificatesV1beta1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// createCertificateSigningRequest is equivalent to kubectl create <csr-name> and kubectl approve csr <csr-name>
func (c *Controller) createCertificateSigningRequest(ctx context.Context, labels map[string]string, name, namespace string, csrBytes []byte, usage string) error {
	encodedBytes := pem.EncodeToMemory(&pem.Block{Type: csrType, Bytes: csrBytes})
	// for the right set of csr configurations regarding CSR signers and Key usages please read:
	// https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/#kubernetes-signers
	if useCertificatesV1API {
		csrSignerName := certificatesV1.KubeletServingSignerName
		csrKeyUsage := []certificatesV1.KeyUsage{
			certificatesV1.UsageDigitalSignature,
			certificatesV1.UsageKeyEncipherment,
			certificatesV1.UsageServerAuth,
		}
		if usage == "client" {
			csrSignerName = certificatesV1.KubeAPIServerClientSignerName
			csrKeyUsage = []certificatesV1.KeyUsage{
				certificatesV1.UsageDigitalSignature,
				certificatesV1.UsageKeyEncipherment,
				certificatesV1.UsageClientAuth,
			}
		}
		kubeCSR := &certificatesV1.CertificateSigningRequest{
			TypeMeta: v1.TypeMeta{
				APIVersion: "certificates.k8s.io/v1",
				Kind:       "CertificateSigningRequest",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Labels:    labels,
				Namespace: namespace,
			},
			Spec: certificatesV1.CertificateSigningRequestSpec{
				SignerName: csrSignerName,
				Request:    encodedBytes,
				Groups:     []string{"system:authenticated", "system:nodes"},
				Usages:     csrKeyUsage,
			},
		}
		ks, err := c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Create(ctx, kubeCSR, metav1.CreateOptions{})
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
		// Return if certificate already exists.
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		// Update the CSR to be approved automatically
		ks.Status = certificatesV1.CertificateSigningRequestStatus{
			Conditions: []certificatesV1.CertificateSigningRequestCondition{
				{
					Type:           certificatesV1.CertificateApproved,
					Reason:         "MinIOOperatorAutoApproval",
					Message:        "Automatically approved by MinIO Operator",
					LastUpdateTime: metav1.NewTime(time.Now()),
					Status:         "True",
				},
			},
		}
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, name, ks, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		csrSignerName := certificatesV1beta1.LegacyUnknownSignerName
		csrKeyUsage := []certificatesV1beta1.KeyUsage{
			certificatesV1beta1.UsageDigitalSignature,
			certificatesV1beta1.UsageKeyEncipherment,
			certificatesV1beta1.UsageServerAuth,
			certificatesV1beta1.UsageClientAuth,
		}
		kubeCSR := &certificatesV1beta1.CertificateSigningRequest{
			TypeMeta: v1.TypeMeta{
				APIVersion: "certificates.k8s.io/v1beta1",
				Kind:       "CertificateSigningRequest",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Labels:    labels,
				Namespace: namespace,
			},
			Spec: certificatesV1beta1.CertificateSigningRequestSpec{
				SignerName: &csrSignerName,
				Request:    encodedBytes,
				Usages:     csrKeyUsage,
			},
		}

		ks, err := c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Create(ctx, kubeCSR, metav1.CreateOptions{})
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}

		// Return if certificate already exists.
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}

		// Update the CSR to be approved automatically
		ks.Status = certificatesV1beta1.CertificateSigningRequestStatus{
			Conditions: []certificatesV1beta1.CertificateSigningRequestCondition{
				{
					Type:           certificatesV1beta1.CertificateApproved,
					Reason:         "MinIOOperatorAutoApproval",
					Message:        "Automatically approved by MinIO Operator",
					LastUpdateTime: metav1.NewTime(time.Now()),
					Status:         "True",
				},
			},
		}

		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(ctx, ks, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// FetchCertificate fetches the generated certificate from the CSR
func (c *Controller) fetchCertificate(ctx context.Context, csrName string) ([]byte, error) {
	klog.V(0).Infof("Start polling for certificate of csr/%s, every %s, timeout after %s", csrName,
		miniov2.DefaultQueryInterval, miniov2.DefaultQueryTimeout)

	tick := time.NewTicker(miniov2.DefaultQueryInterval)
	defer tick.Stop()

	timeout := time.NewTimer(miniov2.DefaultQueryTimeout)

	ch := make(chan os.Signal, 1) // should be always un-buffered SA1017
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	for {
		select {
		case s := <-ch:
			klog.Infof("Signal %s received, exiting ...", s.String())
			return nil, fmt.Errorf("%s", s.String())

		case <-tick.C:
			if useCertificatesV1API {
				r, err := c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, csrName, v1.GetOptions{})
				if err != nil {
					klog.Errorf("Unexpected error during certificate fetching of csr/%s V1: %s", csrName, err)
					return nil, err
				}
				if r.Status.Certificate != nil {
					klog.V(0).Infof("Certificate successfully fetched, creating secret with Private key and Certificate")
					return r.Status.Certificate, nil
				}
				for _, c := range r.Status.Conditions {
					if c.Type == certificatesV1.CertificateDenied {
						err := fmt.Errorf("csr/%s V1 uid: %s is %q: %s", r.Name, r.UID, c.Type, c.String())
						klog.Errorf("Unexpected error during fetch: %v", err)
						return nil, err
					}
				}
			} else {
				r, err := c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, csrName, v1.GetOptions{})
				if err != nil {
					klog.Errorf("Unexpected error during certificate fetching of csr/%s V1beta1: %s", csrName, err)
					return nil, err
				}
				if r.Status.Certificate != nil {
					klog.V(0).Infof("Certificate successfully fetched, creating secret with Private key and Certificate")
					return r.Status.Certificate, nil
				}
				for _, c := range r.Status.Conditions {
					if c.Type == certificatesV1beta1.CertificateDenied {
						err := fmt.Errorf("csr/%s V1beta1 uid: %s is %q: %s", r.Name, r.UID, c.Type, c.String())
						klog.Errorf("Unexpected error during fetch: %v", err)
						return nil, err
					}
				}
			}

			klog.V(1).Infof("Certificate of csr/%s still not available, next try in %d", csrName, miniov2.DefaultQueryInterval)

		case <-timeout.C:
			return nil, fmt.Errorf("timeout during certificate fetching of csr/%s", csrName)
		}
	}
}

func (c *Controller) createSecret(ctx context.Context, tenant *miniov2.Tenant, labels map[string]string, secretName string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: tenant.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(tenant, schema.GroupVersionKind{
					Group:   miniov2.SchemeGroupVersion.Group,
					Version: miniov2.SchemeGroupVersion.Version,
					Kind:    miniov2.MinIOCRDResourceKind,
				}),
			},
		},
		Data: map[string][]byte{
			"private.key": pkBytes,
			"public.crt":  certBytes,
		},
	}
	_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
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
