// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math"
	"time"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var secretTypePublicKeyNameMap = map[string]string{
	"kubernetes.io/tls":        "tls.crt",
	"cert-manager.io/v1":       "tls.crt",
	"cert-manager.io/v1alpha2": "tls.crt",
	// Add newer secretTypes and their corresponding values in future
}

// getCustomCertificates
func (c *Controller) getCustomCertificates(ctx context.Context, tenant *miniov2.Tenant) (customCertificates *miniov2.CustomCertificates, err error) {
	namespace := tenant.Namespace
	secretsMap := map[string][]*miniov2.LocalCertificateReference{
		"Minio":    tenant.Spec.ExternalCertSecret,
		"Client":   tenant.Spec.ExternalClientCertSecrets,
		"MinioCAs": tenant.Spec.ExternalCaCertSecret,
	}
	var certificates []*miniov2.CustomCertificateConfig
	var minioExternalServerCertificates []*miniov2.CustomCertificateConfig
	var minioExternalClientCertificates []*miniov2.CustomCertificateConfig
	var minioExternalCaCertificates []*miniov2.CustomCertificateConfig

	for certType, secrets := range secretsMap {
		certificates = nil
		publicKey := "public.crt"
		// Iterate over TLS secrets and build array of CertificateInfo structure
		// that will be used to display information about certs
		for _, secret := range secrets {
			keyPair, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(ctx, secret.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if v, ok := secretTypePublicKeyNameMap[secret.Type]; ok {
				publicKey = v
			}
			var rawCert []byte
			if _, ok := keyPair.Data[publicKey]; !ok {
				return nil, fmt.Errorf("public key: %v not found inside certificate secret %v", publicKey, secret.Name)
			}
			rawCert = keyPair.Data[publicKey]
			var blocks []byte
			for {
				var block *pem.Block
				block, rawCert = pem.Decode(rawCert)
				if block == nil {
					break
				}
				if block.Type == "CERTIFICATE" {
					blocks = append(blocks, block.Bytes...)
				}
			}
			// parse all certificates we found on this k8s secret
			certs, err := x509.ParseCertificates(blocks)
			if err != nil {
				return nil, err
			}
			for _, cert := range certs {
				var domains []string
				if cert.Subject.CommonName != "" {
					domains = append(domains, cert.Subject.CommonName)
				}
				// append certificate domain names
				if len(cert.DNSNames) > 0 {
					domains = append(domains, cert.DNSNames...)
				}
				// append certificate IPs
				if len(cert.IPAddresses) > 0 {
					for _, ip := range cert.IPAddresses {
						domains = append(domains, ip.String())
					}
				}
				// Register event in case of certificate expiring
				expiresIn := time.Until(cert.NotAfter)
				expiresInDays := int64(expiresIn.Hours() / 24)
				expiresInHours := int64(math.Mod(expiresIn.Hours(), 24))
				expiresInMinutes := int64(math.Mod(expiresIn.Minutes(), 60))
				expiresInSeconds := int64(math.Mod(expiresIn.Seconds(), 60))
				expiresInHuman := fmt.Sprintf("%v days, %v hours, %v minutes, %v seconds", expiresInDays, expiresInHours, expiresInMinutes, expiresInSeconds)

				if expiresInDays >= 10 && expiresInDays < 30 {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CertificateExpiring", fmt.Sprintf("%s certificate '%s' is expiring in %d days", certType, secret.Name, expiresInDays))
				}
				if expiresInDays > 0 && expiresInDays < 10 {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CertificateExpiryImminent", fmt.Sprintf("%s certificate '%s' is expiring in %d days", certType, secret.Name, expiresInDays))
				}
				if expiresInDays > 0 && expiresInDays < 1 {
					expiresInHuman = fmt.Sprintf("%v hours, %v minutes, and %v seconds", expiresInHours, expiresInMinutes, expiresInSeconds)
				}
				if expiresInDays <= 0 {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CertificateExpired", fmt.Sprintf("%s certificate '%s' has expired", certType, secret.Name))
					expiresInHuman = "EXPIRED"
				}

				certificates = append(certificates, &miniov2.CustomCertificateConfig{
					CertName:  secret.Name,
					SerialNo:  cert.SerialNumber.String(),
					Domains:   domains,
					Expiry:    cert.NotAfter.Format(time.RFC3339),
					ExpiresIn: expiresInHuman,
				})
			}
		}
		switch certType {
		case "Minio":
			minioExternalServerCertificates = certificates
		case "Client":
			minioExternalClientCertificates = certificates
		case "MinioCAs":
			minioExternalCaCertificates = certificates
		}
	}
	return &miniov2.CustomCertificates{
		Client:   minioExternalClientCertificates,
		Minio:    minioExternalServerCertificates,
		MinioCAs: minioExternalCaCertificates,
	}, nil
}
