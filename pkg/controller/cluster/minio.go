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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/minio/operator/pkg/resources/services"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"k8s.io/klog/v2"
)

func (c *Controller) checkAndCreateMinIOCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.MinIOCSRName(), metav1.GetOptions{}); err != nil {
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

// checkMinIOSCertificatesStatus checks for the current status of MinIO and it's service
func (c *Controller) checkMinIOSCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	if tenant.AutoCert() {
		// check if there's already a TLS secret for MinIO
		_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.MinIOTLSSecretName(), metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err := c.checkAndCreateMinIOCSR(ctx, nsName, tenant); err != nil {
					return err
				}
				// TLS secret not found, delete CSR if exists and start certificate generation process again
				if err = c.certClient.CertificateSigningRequests().Delete(ctx, tenant.MinIOCSRName(), metav1.DeleteOptions{}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// Check MinIO S3 Endpoint Service
	var tenantPortNum int32 = miniov2.MinIOPortLoadBalancerSVC
	var tenantPortName string = miniov2.MinIOServiceHTTPPortName
	if tenant.TLS() {
		tenantPortNum = miniov2.MinIOTLSPortLoadBalancerSVC
		tenantPortName = miniov2.MinIOServiceHTTPSPortName
	}
	err := c.checkMinIOSvc(ctx, tenant, nsName, tenant.MinIOCIServiceName(), tenantPortName, tenantPortNum)
	if err != nil {
		klog.V(2).Infof("error consolidating console service: %s", err.Error())
		return err
	}

	// Check MinIO Console Endpoint Service
	var consolePortNum int32 = miniov2.ConsolePort
	var consolePortName string = miniov2.ConsoleServicePortName
	if tenant.TLS() || tenant.ConsoleExternalCert() {
		consolePortNum = miniov2.ConsoleTLSPort
		consolePortName = miniov2.ConsoleServiceTLSPortName
	}
	err = c.checkMinIOSvc(ctx, tenant, nsName, tenant.ConsoleCIServiceName(), consolePortName, consolePortNum)
	if err != nil {
		klog.V(2).Infof("error consolidating console service: %s", err.Error())
		return err
	}
	return nil
}

// checkMinIOSvc validates the existence of the MinIO service and validate it's status against what the specification
// states
func (c *Controller) checkMinIOSvc(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName, svcName, svcPortName string, svcPortNum int32) error {

	// Handle the Internal ClusterIP Service for Tenant
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(svcName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningCIService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Cluster IP Service for cluster %q", nsName)
			// Create the clusterIP service for the Tenant
			svc = services.NewClusterIPForMinIO(tenant, svcPortNum, svcName, svcPortName)
			svc, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, svc, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// check the expose status of the MinIO ClusterIP service
	minioSvcMatchesSpec := true
	// compare any other change from what is specified on the tenant
	expectedSvc := services.NewClusterIPForMinIO(tenant, svcPortNum, svcName, svcPortName)
	if !equality.Semantic.DeepDerivative(expectedSvc.Spec, svc.Spec) {
		// some field set by the operator has changed
		minioSvcMatchesSpec = false
	}

	// check the specification of the MinIO ClusterIP service
	if !minioSvcMatchesSpec {
		svc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		svc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		svc.Spec.Ports = expectedSvc.Spec.Ports
		// we can only expose the service, not un-expose it
		if tenant.Spec.ExposeServices != nil && tenant.Spec.ExposeServices.MinIO && svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			svc.Spec.Type = v1.ServiceTypeLoadBalancer
		}
		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return err
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

	err = c.createCertificateSigningRequest(ctx, tenant.MinIOPodLabels(), tenant.MinIOCSRName(), tenant.Namespace, csrBytes, tenant, "server")
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.MinIOCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOCSRName(), err)
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

// createMinIOClientCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that MinIO will use to authenticate (mTLS) with KES or other services
func (c *Controller) createMinIOClientCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateMinIOCryptoData(tenant, c.hostsTemplate)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificateSigningRequest(ctx, tenant.MinIOPodLabels(), tenant.MinIOClientCSRName(), tenant.Namespace, csrBytes, tenant, "client")
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOClientCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.MinIOClientCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOClientCSRName(), err)
		return err
	}

	// parse the certificate here to generate the identity for this certifcate.
	// This is later used to update the identity in KES Server Config File
	h := sha256.New()
	cert, err := parseCertificate(bytes.NewReader(certbytes))
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOClientCSRName(), err)
		return err
	}

	_, err = h.Write(cert.RawSubjectPublicKeyInfo)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.MinIOClientCSRName(), err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for KES StatefulSet to use
	err = c.createSecret(ctx, tenant, tenant.MinIOPodLabels(), tenant.MinIOClientTLSSecretName(), encodedPrivKey, certbytes)
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
