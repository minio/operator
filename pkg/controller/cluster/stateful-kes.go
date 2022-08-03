// Copyright (C) 2022 MinIO, Inc.
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
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/controller/cluster/certificates"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func generateStatefulKESCryptoData(tenant *miniov2.Tenant) ([]byte, []byte, error) {
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

	var csrExtensions []pkix.Extension
	statefulkesHosts := tenant.StatefulKESHosts()
	for _, host := range statefulkesHosts {
		csrExtensions = append(csrExtensions, pkix.Extension{
			Id:       nil,
			Critical: false,
			Value:    []byte(host),
		})
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", tenant.StatefulKESWildCardName()),
			Organization: tenant.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           tenant.StatefulKESHosts(),
		Extensions:         csrExtensions,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createStatefulKESCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that stateful KES Statefulset will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createStatefulKESCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateStatefulKESCryptoData(tenant)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificateSigningRequest(ctx, tenant.StatefulKESPodLabels(), tenant.StatefulKESCSRName(), tenant.Namespace, csrBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.StatefulKESCSRName(), err)
		return err
	}
	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "CSRCreated", "Stateful KES CSR Created")

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.StatefulKESCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.StatefulKESCSRName(), err)
		c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CSRFailed", fmt.Sprintf("Stateful KES CSR Failed to create: %s", err))
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for stateful KES Statefulset to use
	err = c.createSecret(ctx, tenant, tenant.StatefulKESPodLabels(), tenant.StatefulKESTLSSecretName(), encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.StatefulKESTLSSecretName(), err)
		return err
	}

	return nil
}

// statefulKESStatefulSetMatchesSpec checks if the StatefulSet for stateful KES matches what is expected and described from the Tenant
func statefulKESStatefulSetMatchesSpec(tenant *miniov2.Tenant, existingStatefulSet *appsv1.StatefulSet) (bool, error) {
	if existingStatefulSet == nil {
		return false, errors.New("cannot process an empty stateful kes StatefulSet")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}
	// compare image directly
	if !tenant.Spec.StatefulKES.EqualImage(existingStatefulSet.Spec.Template.Spec.Containers[0].Image) {
		klog.V(2).Infof("Tenant %s KES version %s doesn't match: %s", tenant.Name,
			tenant.Spec.StatefulKES.Image, existingStatefulSet.Spec.Template.Spec.Containers[0].Image)
		return false, nil
	}
	// compare any other change from what is specified on the tenant
	expectedStatefulSet := statefulsets.NewForStatefulKES(tenant, tenant.StatefulKESHLServiceName())
	// compare containers environment variables
	if miniov2.IsContainersEnvUpdated(existingStatefulSet.Spec.Template.Spec.Containers, expectedStatefulSet.Spec.Template.Spec.Containers) {
		return false, nil
	}
	if !equality.Semantic.DeepDerivative(expectedStatefulSet.Spec, existingStatefulSet.Spec) {
		// some field set by the operator has changed
		return false, nil
	}
	return true, nil
}

func (c *Controller) checkStatefulKESCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) (err error) {
	if !tenant.ExternalClientCert() {
		if err = c.checkAndCreateMinIOClientCertificates(ctx, nsName, tenant); err != nil {
			return err
		}
	}
	// if KES is enabled and user didn't provide KES server certificates generate them
	if !tenant.StatefulKESExternalCert() {
		_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.StatefulKESTLSSecretName(), metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err = c.checkAndCreateStatefulKESCSR(ctx, nsName, tenant); err != nil {
					return err
				}
				// TLS secret not found, delete CSR if exists and start certificate generation process again
				if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
					if err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Delete(ctx, tenant.StatefulKESCSRName(), metav1.DeleteOptions{}); err != nil {
						return err
					}
				} else {
					if err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Delete(ctx, tenant.StatefulKESCSRName(), metav1.DeleteOptions{}); err != nil {
						return err
					}
				}
			} else {
				return err
			}
		}
	}
	// Check sys admin creds
	_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.StatefulKESSysAdminSecretName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if err = c.checkAndCreateStatefulKESSysAdminSecret(ctx, nsName, tenant); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// check enclave admin creds
	_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.StatefulKESEnclaveAdminSecretName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if err = c.checkAndCreateStatefulKESEnclaveAdminSecret(ctx, nsName, tenant); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (c *Controller) checkAndCreateStatefulKESSysAdminSecret(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	privateKey, err := newPrivateKey(miniov2.DefaultEllipticCurve)
	if err != nil {
		klog.Errorf("Unexpected error during the ECDSA Key generation: %v", err)
		return err
	}
	publicKey := privateKey.Public()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		klog.Errorf("Unexpected error while creating cert serial number: %v", err)
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "kes-sys-admin",
		},
		KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
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
	err = c.createSecret(ctx, tenant, tenant.StatefulKESPodLabels(), tenant.StatefulKESSysAdminSecretName(), keyPem, certPem)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.StatefulKESSysAdminSecretName(), err)
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateStatefulKESEnclaveAdminSecret(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	privateKey, err := newPrivateKey(miniov2.DefaultEllipticCurve)
	if err != nil {
		klog.Errorf("Unexpected error during the ECDSA Key generation: %v", err)
		return err
	}
	publicKey := privateKey.Public()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		klog.Errorf("Unexpected error while creating cert serial number: %v", err)
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "minio-admin",
		},
		KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
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
	err = c.createSecret(ctx, tenant, tenant.StatefulKESPodLabels(), tenant.StatefulKESEnclaveAdminSecretName(), keyPem, certPem)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.MinIOClientTLSSecretName(), err)
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateStatefulKESCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	var err error
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, tenant.StatefulKESCSRName(), metav1.GetOptions{})
	} else {
		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, tenant.StatefulKESCSRName(), metav1.GetOptions{})
	}
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingStatefulKESCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for stateful KES Server Certs, cluster %q", nsName)
			if err = c.createStatefulKESCSR(ctx, tenant); err != nil {
				return err
			}
			return errors.New("waiting for stateful kes cert")
		}
		return err
	}
	return nil
}

func (c *Controller) checkStatefulKESStatus(ctx context.Context, tenant *miniov2.Tenant, totalReplicas int32, cOpts metav1.CreateOptions, uOpts metav1.UpdateOptions, nsName types.NamespacedName) error {
	if tenant.HasStatefulKESEnabled() {
		if err := c.checkStatefulKESCertificatesStatus(ctx, tenant, nsName); err != nil {
			return err
		}
		// Calculate sys-admin identity based on auto generated KES client certificate
		sysAdminIdentity, err := c.getCertIdentity(tenant.Namespace, &miniov2.LocalCertificateReference{Name: tenant.StatefulKESSysAdminSecretName()})
		if err != nil {
			return err
		}
		if !tenant.HasEnv("MINIO_STATEFUL_KES_SYS_ADMIN_IDENTITY") {
			tenant.Spec.StatefulKES.Env = append(tenant.Spec.StatefulKES.Env, corev1.EnvVar{
				Name:  "MINIO_STATEFUL_KES_SYS_ADMIN_IDENTITY",
				Value: sysAdminIdentity,
			})
		}
		// Calculate enclave-admin identity based on auto generated KES client certificate
		enclaveAdminIdentity, err := c.getCertIdentity(tenant.Namespace, &miniov2.LocalCertificateReference{Name: tenant.StatefulKESEnclaveAdminSecretName()})
		if err != nil {
			return err
		}
		if !tenant.HasEnv("MINIO_STATEFUL_KES_ENCLAVE_ADMIN_IDENTITY") {
			tenant.Spec.StatefulKES.Env = append(tenant.Spec.StatefulKES.Env, corev1.EnvVar{
				Name:  "MINIO_STATEFUL_KES_ENCLAVE_ADMIN_IDENTITY",
				Value: enclaveAdminIdentity,
			})
		}
		var tenantIdentity string
		if tenant.ExternalClientCert() {
			tenantIdentity, err = c.getCertIdentity(tenant.Namespace, tenant.Spec.ExternalClientCertSecret)
			if err != nil {
				return err
			}
		} else {
			tenantIdentity, err = c.getCertIdentity(tenant.Namespace, &miniov2.LocalCertificateReference{Name: tenant.MinIOClientTLSSecretName()})
			if err != nil {
				return err
			}
		}
		// pass the identity of the MinIO client certificate
		if !tenant.HasEnv("MINIO_STATEFUL_KES_IDENTITY") {
			tenant.Spec.StatefulKES.Env = append(tenant.Spec.StatefulKES.Env, corev1.EnvVar{
				Name:  "MINIO_STATEFUL_KES_IDENTITY",
				Value: tenantIdentity,
			})
		}
		svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.StatefulKESHLServiceName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.V(2).Infof("Creating a new Headless Service for stateful kes %q", nsName)
				svc = services.NewHeadlessForStatefulKES(tenant)
				if _, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, cOpts); err != nil {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "SvcFailed", fmt.Sprintf("Stateful KES Headless Service failed to create: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "Stateful KES Headless Service created")
			} else {
				return err
			}
		}

		// Get the StatefulSet with the name specified in spec
		if existingStatefulSet, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.StatefulKESStatefulSetName()); err != nil {
			if k8serrors.IsNotFound(err) {
				ks := statefulsets.NewForStatefulKES(tenant, svc.Name)
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningStatefulKESStatefulSet, 0); err != nil {
					return err
				}
				klog.V(2).Infof("Creating a new KES StatefulSet for %q", nsName)
				if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, ks, cOpts); err != nil {
					klog.V(2).Infof(err.Error())
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "StsFailed", fmt.Sprintf("Stateful KES Statefulset failed to create: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "StsCreated", "Stateful KES Statefulset Created")
			} else {
				return err
			}
		} else {
			// Verify if this KES StatefulSet matches the spec on the tenant (resources, affinity, sidecars, etc)
			statefulkesStatefulSetMatchesSpec, err := statefulKESStatefulSetMatchesSpec(tenant, existingStatefulSet)
			if err != nil {
				return err
			}

			// if the KES StatefulSet doesn't match the spec
			if !statefulkesStatefulSetMatchesSpec {
				sks := statefulsets.NewForStatefulKES(tenant, svc.Name)
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingStatefulKES, totalReplicas); err != nil {
					return err
				}
				if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, sks, uOpts); err != nil {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "StsFailed", fmt.Sprintf("Stateful KES Statefulset failed to update: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "StsUpdated", "Stateful KES Statefulset Updated")
			}
		}
	}
	return nil
}
