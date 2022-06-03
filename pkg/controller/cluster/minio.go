// Copyright (C) 2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package cluster

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/minio/operator/pkg/controller/cluster/certificates"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"k8s.io/klog/v2"
)

func (c *Controller) checkAndCreateMinIOCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	var err error
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{})
	} else {
		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{})
	}
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingMinIOCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for MinIO Server Certs, cluster %q", nsName)
			if err = c.createMinIOCSR(ctx, tenant); err != nil {
				return err
			}
			// we want to re-queue this tenant so we can re-check for the health at a later stage
			return errors.New("waiting for minio cert")
		}
		return err
	}
	return nil
}

func (c *Controller) deleteMinIOCSR(ctx context.Context, csrName string) error {
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		if err := c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Delete(ctx, csrName, metav1.DeleteOptions{}); err != nil {
			return err
		}
	} else {
		if err := c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Delete(ctx, csrName, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// recreateMinIOCertsIfRequired - generate TLS certs if not present, or expired
func (c *Controller) recreateMinIOCertsIfRequired(ctx context.Context) error {
	namespace := miniov2.GetNSFromFile()
	operatorTLSSecret, err := c.getTLSSecret(ctx, namespace, OperatorTLSSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Info("TLS certificate not found. Generating one.")
			c.generateTLSCert()
			return nil
		}
		return err
	}

	needsRenewal, err := c.certNeedsRenewal(ctx, operatorTLSSecret)
	if err != nil {
		return err
	}

	if !needsRenewal {
		return nil
	}

	// Expired cert. Delete the secret + CSR and re-create the cert

	klog.V(2).Info("Deleting the TLS secret of expired cert on operator")
	err = c.kubeClientSet.CoreV1().Secrets(namespace).Delete(ctx, OperatorTLSSecretName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = c.deleteMinIOCSR(ctx, c.operatorCSRName())
	if err != nil {
		return err
	}

	klog.V(2).Info("Generating a fresh TLS certificate")
	c.generateTLSCert()

	return nil
}

func (c *Controller) recreateMinIOCertsOnTenant(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	klog.V(2).Info("Deleting the TLS secret and CSR of expired cert on tenant %s", tenant.Name)

	// First delete the TLS secret of expired cert on the tenant
	err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Delete(ctx, tenant.MinIOTLSSecretName(), metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Then delete the CSR
	err = c.deleteMinIOCSR(ctx, tenant.MinIOCSRName())
	if err != nil {
		return err
	}

	// In case the certs on operator are also expired, re-create them
	if err := c.recreateMinIOCertsIfRequired(ctx); err != nil {
		return err
	}

	// Finally re-create the certs on the tenant
	return c.checkAndCreateMinIOCSR(ctx, nsName, tenant)
}

func (c *Controller) getTLSSecret(ctx context.Context, nsName string, secretName string) (*corev1.Secret, error) {
	return c.kubeClientSet.CoreV1().Secrets(nsName).Get(ctx, secretName, metav1.GetOptions{})
}

// checkOperatorCaForTenant create or updates the operator-ca-tls secret for tenant if need it
func (c *Controller) checkOperatorCaForTenant(ctx context.Context, tenant *miniov2.Tenant) (operatorCATLSExists bool, err error) {
	var operatorCaCert []byte
	var tenantCaCert []byte
	var caCertKey string
	// get operator-ca-tls in minio-operator namespace
	operatorCaSecret, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(ctx, OperatorCATLSSecretName, metav1.GetOptions{})
	if err != nil {
		// if operator-ca-tls doesnt exists continue
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	caCertKey = "public.crt"
	// if secret type is k8s tls or cert-manager use the right secret keys
	if operatorCaSecret.Type == "kubernetes.io/tls" {
		caCertKey = "tls.crt"
	} else if operatorCaSecret.Type == "cert-manager.io/v1alpha2" || operatorCaSecret.Type == "cert-manager.io/v1" {
		caCertKey = "ca.crt"
	}

	if _, ok := operatorCaSecret.Data[caCertKey]; !ok {
		return false, fmt.Errorf("missing '%s' in %s/%s secret", caCertKey, miniov2.GetNSFromFile(), OperatorCATLSSecretName)
	}

	operatorCaCert = operatorCaSecret.Data[caCertKey]

	tenantCaSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, OperatorCATLSSecretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Infof("'%s/%s' secret not found, creating one now", tenant.Namespace, OperatorCATLSSecretName)
			// create tenant operator-ca-tls secret
			opCATLSSecret := &corev1.Secret{
				Type: "Opaque",
				ObjectMeta: metav1.ObjectMeta{
					Name:      OperatorCATLSSecretName,
					Namespace: tenant.Namespace,
					Labels:    tenant.MinIOPodLabels(),
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(tenant, schema.GroupVersionKind{
							Group:   miniov2.SchemeGroupVersion.Group,
							Version: miniov2.SchemeGroupVersion.Version,
							Kind:    miniov2.MinIOCRDResourceKind,
						}),
					},
				},
				Data: map[string][]byte{
					"public.crt": operatorCaCert,
				},
			}
			_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, opCATLSSecret, metav1.CreateOptions{})
			if err != nil {
				return false, err
			}
		}
		return false, err
	}

	if _, ok := tenantCaSecret.Data["public.crt"]; !ok {
		err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Delete(ctx, OperatorCATLSSecretName, metav1.DeleteOptions{})
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("missing 'public.crt' in %s/%s secret, re-creating it", tenant.Namespace, OperatorCATLSSecretName)
	}

	tenantCaCert = tenantCaSecret.Data["public.crt"]

	if string(tenantCaCert) != string(operatorCaCert) {
		tenantCaSecret.Data["public.crt"] = operatorCaCert
		_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, tenantCaSecret, metav1.UpdateOptions{})
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("'%s' in '%s/%s' secret changed, updating '%s/%s' secret", caCertKey, miniov2.GetNSFromFile(), OperatorCATLSSecretName, tenant.Namespace, OperatorCATLSSecretName)
	}

	return true, nil
}

// checkOperatorCertForTenant checks create or updates the operator-tls secret for tenant
func (c *Controller) checkOperatorCertForTenant(ctx context.Context, tenant *miniov2.Tenant) error {
	if !isOperatorTLS() {
		return nil
	}
	// get operator-tls in minio-operator namespace
	operatorCertSecret, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(ctx, OperatorTLSSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// default secret keys for Opaque k8s secret
	publicCertKey := "public.crt"
	// if secret type is k8s tls or cert-manager use the right secret keys
	if operatorCertSecret.Type == "kubernetes.io/tls" || operatorCertSecret.Type == "cert-manager.io/v1alpha2" || operatorCertSecret.Type == "cert-manager.io/v1" {
		publicCertKey = "tls.crt"
	}

	if _, ok := operatorCertSecret.Data[publicCertKey]; !ok {
		return fmt.Errorf("missing '%s' in %s/%s secret", publicCertKey, miniov2.GetNSFromFile(), OperatorTLSSecretName)
	}

	operatorCert := operatorCertSecret.Data[publicCertKey]

	tenantOperatorCertSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, OperatorTLSSecretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Infof("'%s/%s' secret not found, creating one now", tenant.Namespace, OperatorTLSSecretName)
			// create tenant operator-tls secret copying only the public.crt portion from the original operator-tls secret
			opTLSSecret := &corev1.Secret{
				Type: "Opaque",
				ObjectMeta: metav1.ObjectMeta{
					Name:      OperatorTLSSecretName,
					Namespace: tenant.Namespace,
					Labels:    tenant.MinIOPodLabels(),
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(tenant, schema.GroupVersionKind{
							Group:   miniov2.SchemeGroupVersion.Group,
							Version: miniov2.SchemeGroupVersion.Version,
							Kind:    miniov2.MinIOCRDResourceKind,
						}),
					},
				},
				Data: map[string][]byte{
					"public.crt": operatorCert,
				},
			}
			_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, opTLSSecret, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}

	if _, ok := tenantOperatorCertSecret.Data["public.crt"]; !ok {
		err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Delete(ctx, OperatorTLSSecretName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return fmt.Errorf("missing 'public.crt' in %s/%s secret, re-creating it", tenant.Namespace, OperatorTLSSecretName)
	}

	tenantOperatorCert := tenantOperatorCertSecret.Data["public.crt"]

	if string(tenantOperatorCert) != string(operatorCert) {
		tenantOperatorCertSecret.Data["public.crt"] = operatorCert
		_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Update(ctx, tenantOperatorCertSecret, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return fmt.Errorf("'%s' key in %s/%s secret changed, updating '%s/%s' secret", publicCertKey, miniov2.GetNSFromFile(), OperatorTLSSecretName, tenant.Namespace, OperatorTLSSecretName)
	}

	return nil
}

// checkMinIOCertificatesStatus checks for the current status of MinIO and it's service
func (c *Controller) checkMinIOCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	if tenant.AutoCert() {
		// check if there's already a TLS secret for MinIO
		tlsSecret, err := c.getTLSSecret(ctx, tenant.Namespace, tenant.MinIOTLSSecretName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err := c.checkAndCreateMinIOCSR(ctx, nsName, tenant); err != nil {
					return err
				}
				// TLS secret not found, delete CSR if exists and start certificate generation process again
				if err := c.deleteMinIOCSR(ctx, tenant.MinIOCSRName()); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		needsRenewal, err := c.certNeedsRenewal(ctx, tlsSecret)
		if err != nil {
			return err
		}

		if needsRenewal {
			return c.recreateMinIOCertsOnTenant(ctx, tenant, nsName)
		}
	}

	return nil
}

// certNeedsRenewal - returns true if the TLS certificate from given secret has expired or is
// about to expire within the next two days.
func (c *Controller) certNeedsRenewal(ctx context.Context, tlsSecret *corev1.Secret) (bool, error) {
	if pubCert, ok := tlsSecret.Data["public.crt"]; ok {
		if privKey, ok := tlsSecret.Data["private.key"]; ok {
			tlsCert, err := tls.X509KeyPair(pubCert, privKey)
			if err != nil {
				return false, err
			}

			leaf := tlsCert.Leaf
			if leaf == nil {
				leaf, err = x509.ParseCertificate(tlsCert.Certificate[0])
				if err != nil {
					return false, err
				}
			}
			if leaf.NotAfter.Before(time.Now().Add(time.Hour * 48)) {
				klog.V(2).Infof("TLS Certificate expiry on %s", leaf.NotAfter.String())
				return true, nil
			}
		}
	}
	return false, nil
}

func generateMinIOCryptoData(tenant *miniov2.Tenant, hostsTemplate string) ([]byte, []byte, error) {
	var dnsNames []string
	var csrExtensions []pkix.Extension

	klog.V(0).Infof("Generating private key")
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

	klog.V(0).Infof("Generating CSR with CN=%s", tenant.Spec.CertConfig.CommonName)

	hosts := tenant.AllMinIOHosts()
	if hostsTemplate != "" {
		hosts = tenant.TemplatedMinIOHosts(hostsTemplate)
	}

	if isEqual(tenant.Spec.CertConfig.DNSNames, hosts) {
		dnsNames = tenant.Spec.CertConfig.DNSNames
	} else {
		dnsNames = append(tenant.Spec.CertConfig.DNSNames, hosts...)
	}
	dnsNames = append(dnsNames, tenant.MinIOBucketBaseWildcardDomain())

	for _, dnsName := range dnsNames {
		csrExtensions = append(csrExtensions, pkix.Extension{
			Id:       nil,
			Critical: false,
			Value:    []byte(dnsName),
		})
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", tenant.Spec.CertConfig.CommonName),
			Organization: tenant.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           dnsNames,
		Extensions:         csrExtensions,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createMinIOCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that MinIO statefulset will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createMinIOCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateMinIOCryptoData(tenant, c.hostsTemplate)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificateSigningRequest(ctx, tenant.MinIOPodLabels(), tenant.MinIOCSRName(), tenant.Namespace, csrBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOCSRName(), err)
		return err
	}
	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "CSRCreated", "MinIO CSR Created")

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.MinIOCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOCSRName(), err)
		c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CSRFailed", fmt.Sprintf("MinIO CSR Failed to create: %s", err))
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for MinIO Statefulset to use
	err = c.createSecret(ctx, tenant, tenant.MinIOPodLabels(), tenant.MinIOTLSSecretName(), encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.MinIOTLSSecretName(), err)
		return err
	}

	return nil
}

// createMinIOClientCertificates handles all the steps required to create the MinIO <-> KES mTLS certificates
func (c *Controller) createMinIOClientCertificates(ctx context.Context, tenant *miniov2.Tenant) error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: tenant.MinIOFQDNServiceName(),
		},
		NotBefore: time.Now().UTC(),
		NotAfter:  time.Now().UTC().Add(8760 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		DNSNames:              tenant.MinIOHosts(),
		IPAddresses:           []net.IP{},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	// Create secret for KES StatefulSet to use
	err = c.createSecret(ctx, tenant, tenant.MinIOPodLabels(), tenant.MinIOClientTLSSecretName(), keyPem, certPem)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.MinIOClientTLSSecretName(), err)
		return err
	}

	return nil
}

func (c *Controller) deleteOldConsoleDeployment(ctx context.Context, tenant *miniov2.Tenant, consoleDeployment string) error {
	err := c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Delete(ctx, consoleDeployment, metav1.DeleteOptions{})
	if err != nil {
		klog.V(2).Infof(err.Error())
		return err
	}
	err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Delete(ctx, tenant.ConsoleCIServiceName(), metav1.DeleteOptions{})
	if err != nil {
		klog.V(2).Infof(err.Error())
		return err
	}

	return nil
}
