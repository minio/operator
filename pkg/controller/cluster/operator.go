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
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"k8s.io/klog/v2"
)

var (
	errOperatorWaitForTLS = errors.New("waiting for Operator cert")
)

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
			CommonName:   opCommonNoDomain,
			Organization: []string{"Acme Co"},
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

func (c *Controller) createOperatorTLSSecret(ctx context.Context, operator metav1.Object, labels map[string]string, secretName string, pkBytes, certBytes []byte) error {
	secret := &corev1.Secret{
		Type: "Opaque",
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: miniov2.GetNSFromFile(),
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(operator, schema.GroupVersionKind{
					Group:   miniov2.SchemeGroupVersion.Group,
					Version: miniov2.SchemeGroupVersion.Version,
					Kind:    miniov2.OperatorCRDResourceKind,
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

// createOperatorTLSCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that Operator deployment will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createOperatorTLSCSR(ctx context.Context, operator metav1.Object) error {
	privKeysBytes, csrBytes, err := generateOperatorCryptoData()
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}
	namespace := miniov2.GetNSFromFile()
	operatorCSRName := fmt.Sprintf("operator-%s-csr", namespace)
	err = c.createCertificate(ctx, map[string]string{}, operatorCSRName, namespace, csrBytes, operator)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", operatorCSRName, err)
		return err
	}

	// fetch certificate from CSR
	certBytes, err := c.fetchCertificate(ctx, operatorCSRName)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", operatorCSRName, err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for operator to use
	err = c.createOperatorTLSSecret(ctx, operator, map[string]string{}, "operator-tls", encodedPrivKey, certBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", "operator-tls", err)
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateOperatorCSR(ctx context.Context, operator metav1.Object) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, "operator-auto-tls", metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Infof("Creating a new Certificate Signing Request for Operator Server Certs, cluster %q")
			if err = c.createOperatorTLSCSR(ctx, operator); err != nil {
				return err
			}
			return errOperatorWaitForTLS
		}
		return err
	}
	return nil
}
