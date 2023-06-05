// Copyright (C) 2019 MinIO, Inc.
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

package controller

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/minio/operator/pkg/controller/certificates"

	corev1 "k8s.io/api/core/v1"

	"github.com/minio/operator/pkg/resources/services"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/minio/operator/pkg/resources/statefulsets"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"k8s.io/klog/v2"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func generateKESCryptoData(tenant *miniov2.Tenant) ([]byte, []byte, error) {
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
	kesHosts := tenant.KESHosts()
	for _, host := range kesHosts {
		csrExtensions = append(csrExtensions, pkix.Extension{
			Id:       nil,
			Critical: false,
			Value:    []byte(host),
		})
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", tenant.KESWildCardName()),
			Organization: tenant.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           tenant.KESHosts(),
		Extensions:         csrExtensions,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createKESCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that KES Statefulset will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createKESCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateKESCryptoData(tenant)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificateSigningRequest(ctx, tenant.KESPodLabels(), tenant.KESCSRName(), tenant.Namespace, csrBytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.KESCSRName(), err)
		return err
	}
	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "CSRCreated", "KES CSR Created")

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.KESCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.KESCSRName(), err)
		c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CSRFailed", fmt.Sprintf("KES CSR Failed to create: %s", err))
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for KES Statefulset to use
	err = c.createSecret(ctx, tenant, tenant.KESPodLabels(), tenant.KESTLSSecretName(), encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.KESTLSSecretName(), err)
		return err
	}

	return nil
}

// kesStatefulSetMatchesSpec checks if the StatefulSet for KES matches what is expected and described from the Tenant
func kesStatefulSetMatchesSpec(tenant *miniov2.Tenant, existingStatefulSet *appsv1.StatefulSet) (bool, error) {
	if existingStatefulSet == nil {
		return false, errors.New("cannot process an empty kes StatefulSet")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}
	// compare image directly
	if !tenant.Spec.KES.EqualImage(existingStatefulSet.Spec.Template.Spec.Containers[0].Image) {
		klog.V(2).Infof("Tenant %s KES version %s doesn't match: %s", tenant.Name,
			tenant.Spec.KES.Image, existingStatefulSet.Spec.Template.Spec.Containers[0].Image)
		return false, nil
	}
	// compare any other change from what is specified on the tenant
	expectedStatefulSet := statefulsets.NewForKES(tenant, tenant.KESHLServiceName())
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

func (c *Controller) checkKESCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) (err error) {
	if !tenant.ExternalClientCert() {
		if err = c.checkAndCreateMinIOClientCertificates(ctx, nsName, tenant); err != nil {
			return err
		}
	}
	// if KES is enabled and user didn't provide KES server certificates generate them
	if !tenant.KESExternalCert() {
		_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.KESTLSSecretName(), metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err = c.checkAndCreateKESCSR(ctx, nsName, tenant); err != nil {
					return err
				}
				// TLS secret not found, delete CSR if exists and start certificate generation process again
				if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
					if err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Delete(ctx, tenant.KESCSRName(), metav1.DeleteOptions{}); err != nil {
						return err
					}
				} else {
					if err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Delete(ctx, tenant.KESCSRName(), metav1.DeleteOptions{}); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (c *Controller) checkKESStatus(ctx context.Context, tenant *miniov2.Tenant, totalAvailableReplicas int32, cOpts metav1.CreateOptions, uOpts metav1.UpdateOptions, nsName types.NamespacedName) error {
	if tenant.HasKESEnabled() {
		if err := c.checkKESCertificatesStatus(ctx, tenant, nsName); err != nil {
			return err
		}
		var err error
		var certificateClientIdentity string
		if tenant.ExternalClientCert() {
			// Since we're using external secret, store the identity for later use
			certificateClientIdentity, err = c.getCertIdentity(tenant.Namespace, tenant.Spec.ExternalClientCertSecret)
			if err != nil {
				return err
			}
		} else {
			// Calculate identity based on auto generated KES client certificate
			certificateClientIdentity, err = c.getCertIdentity(tenant.Namespace, &miniov2.LocalCertificateReference{Name: tenant.MinIOClientTLSSecretName()})
			if err != nil {
				return err
			}
		}
		// pass the identity of the MinIO client certificate
		if !tenant.HasEnv("MINIO_KES_IDENTITY") {
			tenant.Spec.KES.Env = append(tenant.Spec.KES.Env, corev1.EnvVar{
				Name:  "MINIO_KES_IDENTITY",
				Value: certificateClientIdentity,
			})
		}
		svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.KESHLServiceName())
		if err != nil {
			if k8serrors.IsNotFound(err) {
				klog.V(2).Infof("Creating a new Headless Service for cluster %q", nsName)
				svc = services.NewHeadlessForKES(tenant)
				if _, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, cOpts); err != nil {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "SvcFailed", fmt.Sprintf("KES Headless Service failed to create: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "KES Headless Service created")
			} else {
				return err
			}
		}

		if tenant.HasGCPCredentialSecretForKES() {
			kesSA, err := c.kubeClientSet.CoreV1().ServiceAccounts(tenant.Namespace).Get(ctx, tenant.Spec.KES.ServiceAccountName, v1.GetOptions{})
			if err != nil {
				klog.Errorf("unable to get the service account %s/%s: %v", tenant.Namespace, tenant.Spec.KES.ServiceAccountName, err)
				return err
			}
			if *kesSA.AutomountServiceAccountToken {
				return fmt.Errorf("automountServiceAccountToken should be set to false in service account %s/%s to mount the service token", tenant.Namespace, tenant.Spec.KES.ServiceAccountName)
			}
		}

		// Get the StatefulSet with the name specified in spec
		if existingStatefulSet, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.KESStatefulSetName()); err != nil {
			if k8serrors.IsNotFound(err) {
				ks := statefulsets.NewForKES(tenant, svc.Name)
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningKESStatefulSet, 0); err != nil {
					return err
				}
				klog.V(2).Infof("Creating a new KES StatefulSet for %q", nsName)
				if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, ks, cOpts); err != nil {
					klog.V(2).Infof(err.Error())
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "StsFailed", fmt.Sprintf("KES Statefulset failed to create: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "StsCreated", "KES Statefulset Created")
			} else {
				return err
			}
		} else {
			// Verify if this KES StatefulSet matches the spec on the tenant (resources, affinity, sidecars, etc)
			kesStatefulSetMatchesSpec, err := kesStatefulSetMatchesSpec(tenant, existingStatefulSet)
			if err != nil {
				return err
			}

			// if the KES StatefulSet doesn't match the spec
			if !kesStatefulSetMatchesSpec {
				ks := statefulsets.NewForKES(tenant, svc.Name)
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingKES, totalAvailableReplicas); err != nil {
					return err
				}
				if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, ks, uOpts); err != nil {
					c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "StsFailed", fmt.Sprintf("KES Statefulset failed to update: %s", err))
					return err
				}
				c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "StsUpdated", "KES Statefulset Updated")
			}
		}
	}
	return nil
}

func (c *Controller) checkAndCreateMinIOClientCertificates(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	var err error
	_, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.MinIOClientTLSSecretName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingMinIOClientCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Client Certificate for MinIO, cluster %q", nsName)
			if err = c.createMinIOClientCertificates(ctx, tenant); err != nil {
				// we want to re-queue this tenant so we can re-check for the health at a later stage
				c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "CertFailed", fmt.Sprintf("KES MinIO Client Certificate failed to create: %s", err))
				return err
			}
			return errors.New("waiting for minio client cert")
		}
		return err
	}
	return nil
}

func (c *Controller) checkAndCreateKESCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	var err error
	if certificates.GetCertificatesAPIVersion(c.kubeClientSet) == certificates.CSRV1 {
		_, err = c.kubeClientSet.CertificatesV1().CertificateSigningRequests().Get(ctx, tenant.KESCSRName(), metav1.GetOptions{})
	} else {
		_, err = c.kubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, tenant.KESCSRName(), metav1.GetOptions{})
	}
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingKESCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for KES Server Certs, cluster %q", nsName)
			if err = c.createKESCSR(ctx, tenant); err != nil {
				return err
			}
			return errors.New("waiting for kes cert")
		}
		return err
	}
	return nil
}

func (c *Controller) getCertIdentity(ns string, cert *miniov2.LocalCertificateReference) (string, error) {
	var certbytes []byte
	secret, err := c.kubeClientSet.CoreV1().Secrets(ns).Get(context.Background(), cert.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	// Store the Identity to be used later during KES container creation
	if secret.Type == "kubernetes.io/tls" || secret.Type == "cert-manager.io/v1alpha2" || secret.Type == "cert-manager.io/v1" {
		certbytes = secret.Data["tls.crt"]
	} else {
		certbytes = secret.Data["public.crt"]
	}

	// parse the certificate here to generate the identity for this certifcate.
	// This is later used to update the identity in KES Server Config File
	h := sha256.New()
	parsedCert, err := parseCertificate(bytes.NewReader(certbytes))
	if err != nil {
		klog.Errorf("Unexpected error during the parsing the secret/%s: %v", cert.Name, err)
		return "", err
	}

	_, err = h.Write(parsedCert.RawSubjectPublicKeyInfo)
	if err != nil {
		klog.Errorf("Unexpected error during the parsing the secret/%s: %v", cert.Name, err)
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
