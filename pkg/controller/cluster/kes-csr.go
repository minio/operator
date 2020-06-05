/*
 * Copyright (C) 2019, 2020, MinIO, Inc.
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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"

	"k8s.io/klog"

	miniov1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
)

func generateKESCryptoData(mi *miniov1.MinIOInstance) ([]byte, []byte, error) {
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

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   mi.KESWildCardName(),
			Organization: miniov1.DefaultOrgName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           mi.KESHosts(),
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createKESTLSCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that KES Statefulset will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createKESTLSCSR(ctx context.Context, mi *miniov1.MinIOInstance) error {
	privKeysBytes, csrBytes, err := generateKESCryptoData(mi)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.submitCSR(ctx, mi.KESPodLabels(), mi.KESCSRName(), mi.Namespace, csrBytes, mi)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.KESCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, mi.KESCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.KESCSRName(), err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for KES Statefulset to use
	err = c.createSecret(ctx, mi, mi.KESPodLabels(), mi.KESTLSSecretName(), mi.Namespace, encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", mi.MinIOTLSSecretName(), err)
		return err
	}

	return nil
}

// createMinIOClientTLSCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that KES Statefulset will use for MinIO Client Auth
func (c *Controller) createMinIOClientTLSCSR(ctx context.Context, mi *miniov1.MinIOInstance) error {
	privKeysBytes, csrBytes, err := generateCryptoData(mi, c.hostsTemplate)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.submitCSR(ctx, mi.MinIOPodLabels(), mi.MinIOClientCSRName(), mi.Namespace, csrBytes, mi)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOClientCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, mi.MinIOClientCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOClientCSRName(), err)
		return err
	}

	// parse the certificate here to generate the identity for this certifcate.
	// This is later used to update the identity in KES Server Config File
	h := sha256.New()
	cert, err := parseCertificate(bytes.NewReader(certbytes))
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOClientCSRName(), err)
		return err
	}

	_, err = h.Write(cert.RawSubjectPublicKeyInfo)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", mi.MinIOClientCSRName(), err)
		return err
	}

	// Store the Identity to be used later during KES container creation
	miniov1.Identity = hex.EncodeToString(h.Sum(nil))

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for KES Statefulset to use
	err = c.createSecret(ctx, mi, mi.MinIOPodLabels(), mi.MinIOClientTLSSecretName(), mi.Namespace, encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", mi.MinIOClientTLSSecretName(), err)
		return err
	}

	return nil
}
