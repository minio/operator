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
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/minio/madmin-go"

	v1 "k8s.io/api/core/v1"

	"github.com/minio/operator/pkg/resources/services"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/minio/operator/pkg/resources/deployments"
	"k8s.io/apimachinery/pkg/api/equality"

	appsv1 "k8s.io/api/apps/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	"k8s.io/klog/v2"
)

func (c *Controller) checkConsoleCertificatesStatus(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	if tenant.AutoCert() {
		// AutoCert will generate Console server certificates if user didn't provide any
		if !tenant.ConsoleExternalCert() {
			// check if there's already a TLS secret for console
			_, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.ConsoleTLSSecretName(), metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					if err := c.checkAndCreateConsoleCSR(ctx, nsName, tenant); err != nil {
						return err
					}
					// TLS secret not found, delete CSR if exists and start certificate generation process again
					if err = c.certClient.CertificateSigningRequests().Delete(ctx, tenant.ConsoleCSRName(), metav1.DeleteOptions{}); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}
	return nil
}

// checkConsoleStatus checks for the current status of console and it's service
func (c *Controller) checkConsoleStatus(ctx context.Context, tenant *miniov2.Tenant, tenantConfig map[string][]byte, totalReplicas int32, adminClnt *madmin.AdminClient, cOpts metav1.CreateOptions, uOpts metav1.UpdateOptions, nsName types.NamespacedName) error {
	var userCredentials []*v1.Secret
	for _, credential := range tenant.Spec.Users {
		credentialSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, credential.Name, metav1.GetOptions{})
		if err == nil && credentialSecret != nil {
			userCredentials = append(userCredentials, credentialSecret)
		}
	}
	if tenant.HasConsoleEnabled() {
		if err := c.checkConsoleCertificatesStatus(ctx, tenant, nsName); err != nil {
			return err
		}
		// Get the Deployment with the name specified in MirrorInstace.spec
		consoleDeployment, err := c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.ConsoleDeploymentName())
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}

			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningConsoleDeployment, totalReplicas); err != nil {
				return err
			}

			if tenant.HasCredsSecret() && tenant.HasConsoleSecret() {
				consoleSecretName := tenant.Spec.Console.ConsoleSecret.Name
				consoleSecret, sErr := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, consoleSecretName, metav1.GetOptions{})
				if sErr == nil && consoleSecret != nil {
					_, accessKeyExist := consoleSecret.Data["CONSOLE_ACCESS_KEY"]
					_, secretKeyExist := consoleSecret.Data["CONSOLE_SECRET_KEY"]
					if accessKeyExist && secretKeyExist {
						userCredentials = append(userCredentials, consoleSecret)
					}
				} else {
					// just log the error
					klog.Info(sErr)
				}
			}

			// If OIDC is configured the user doesnt need to provide any initial console users for the tenant
			// If MinIO is using the internal IDP initial list of users is mandatory
			// If MinIO is configured with LDAP initial list of users is mandatory and usernames must have a similar format to "uid=billy,dc=example,dc=org"
			if len(userCredentials) == 0 && string(tenantConfig["MINIO_IDENTITY_OPENID_CONFIG_URL"]) == "" {
				msg := "Please set the credentials"
				klog.V(2).Infof(msg)
				if _, terr := c.updateTenantStatus(ctx, tenant, msg, totalReplicas); terr != nil {
					return err
				}
				// return nil so we don't re-queue this work item
				return err
			}

			// Create Console Deployment
			consoleDeployment = deployments.NewConsole(tenant)
			_, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Create(ctx, consoleDeployment, cOpts)
			if err != nil {
				klog.V(2).Infof(err.Error())
				return err
			}

			if len(userCredentials) > 0 {
				var err error
				// Make sure that MinIO is up and running to enable MinIO console user.
				if !tenant.MinIOHealthCheck() {
					if _, err = c.updateTenantStatus(ctx, tenant, StatusWaitingForReadyState, totalReplicas); err != nil {
						return err
					}
					return err
				}
				skipCreateUsers := false
				// If MinIO is using an external IDP we need to skip the console user creation
				if string(tenantConfig["MINIO_IDENTITY_OPENID_CONFIG_URL"]) != "" && string(tenantConfig["MINIO_IDENTITY_LDAP_SERVER_ADDR"]) != "" {
					skipCreateUsers = true
				}
				if err := tenant.CreateUsers(adminClnt, userCredentials, skipCreateUsers); err != nil {
					klog.V(2).Infof("Unable to create MinIO users: %v", err)
					return err
				}
			}
		} else {
			// Verify if this console deployment matches the spec on the tenant (resources, affinity, sidecars, etc)
			consoleDeploymentMatchesSpec, err := consoleDeploymentMatchesSpec(tenant, consoleDeployment)
			if err != nil {
				return err
			}

			// if the console deployment doesn't match the spec
			if !consoleDeploymentMatchesSpec {
				if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingConsole, totalReplicas); err != nil {
					return err
				}
				consoleDeployment = deployments.NewConsole(tenant)
				if _, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Update(ctx, consoleDeployment, uOpts); err != nil {
					return err
				}
			}
		}

		err = c.checkConsoleSvc(ctx, tenant, nsName)
		if err != nil {
			klog.V(2).Infof("error consolidating console service: %s", err.Error())
			return err
		}

	} else {
		// disable console and service if they exists
		_, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.ConsoleCIServiceName())
		if err == nil {
			err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Delete(ctx, tenant.ConsoleCIServiceName(), metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		_, err = c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.ConsoleDeploymentName())
		if err == nil {
			err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Delete(ctx, tenant.ConsoleDeploymentName(), metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateConsoleCryptoData(tenant *miniov2.Tenant) ([]byte, []byte, error) {
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

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("system:node:%s", tenant.ConsoleCommonName()),
			Organization: tenant.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           []string{tenant.ConsoleCIServiceName()},
		Extensions: []pkix.Extension{
			{
				Id:       nil,
				Critical: false,
				Value:    []byte(tenant.ConsoleCIServiceName()),
			},
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createConsoleCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that Console deployment will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createConsoleCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateConsoleCryptoData(tenant)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificateSigningRequest(ctx, tenant.ConsolePodLabels(), tenant.ConsoleCSRName(), tenant.Namespace, csrBytes, tenant, "server")
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.ConsoleCSRName(), err)
		return err
	}

	// fetch certificate from CSR
	certbytes, err := c.fetchCertificate(ctx, tenant.ConsoleCSRName())
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the csr/%s: %v", tenant.ConsoleCSRName(), err)
		return err
	}

	// PEM encode private ECDSA key
	encodedPrivKey := pem.EncodeToMemory(&pem.Block{Type: privateKeyType, Bytes: privKeysBytes})

	// Create secret for Console Deployment to use
	err = c.createSecret(ctx, tenant, tenant.ConsolePodLabels(), tenant.ConsoleTLSSecretName(), encodedPrivKey, certbytes)
	if err != nil {
		klog.Errorf("Unexpected error during the creation of the secret/%s: %v", tenant.ConsoleTLSSecretName(), err)
		return err
	}

	return nil
}

// consoleDeploymentMatchesSpec checks if the deployment for console matches what is expected and described from the Tenant
func consoleDeploymentMatchesSpec(tenant *miniov2.Tenant, consoleDeployment *appsv1.Deployment) (bool, error) {
	if consoleDeployment == nil {
		return false, errors.New("cannot process an empty console deployment")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}
	// compare image directly
	if !tenant.Spec.Console.EqualImage(consoleDeployment.Spec.Template.Spec.Containers[0].Image) {
		klog.V(2).Infof("Tenant %s console version %s doesn't match: %s", tenant.Name,
			tenant.Spec.Console.Image, consoleDeployment.Spec.Template.Spec.Containers[0].Image)
		return false, nil
	}
	// compare any other change from what is specified on the tenant
	expectedDeployment := deployments.NewConsole(tenant)
	if !equality.Semantic.DeepDerivative(expectedDeployment.Spec, consoleDeployment.Spec) {
		// some field set by the operator has changed
		return false, nil
	}
	return true, nil
}

func (c *Controller) checkAndCreateConsoleCSR(ctx context.Context, nsName types.NamespacedName, tenant *miniov2.Tenant) error {
	if _, err := c.certClient.CertificateSigningRequests().Get(ctx, tenant.ConsoleCSRName(), metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusWaitingConsoleCert, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Certificate Signing Request for Console Server Certs, cluster %q", nsName)
			if err = c.createConsoleCSR(ctx, tenant); err != nil {
				return err
			}
			// we want to re-queue this tenant so we can re-check for the console certificate
			return errors.New("waiting for console cert")
		}
		return err
	}
	return nil
}

func (c *Controller) checkConsoleSvc(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	// check the status of the console service
	consoleSvc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.ConsoleCIServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Infof("Creating a new Cluster IP Service for console %q", nsName)
			// Create the clusterIP service for the Tenant
			consoleSvc = services.NewClusterIPForConsole(tenant)
			consoleSvc, err = c.kubeClientSet.CoreV1().Services(consoleSvc.Namespace).Create(ctx, consoleSvc, metav1.CreateOptions{})
			if err != nil {
				klog.V(2).Infof(err.Error())
				return err
			}
		} else {
			return err
		}
	}

	consoleSvcMatchesSpec := true
	// compare any other change from what is specified on the tenant
	expectedSvc := services.NewClusterIPForConsole(tenant)
	if !equality.Semantic.DeepDerivative(expectedSvc.Spec, consoleSvc.Spec) {
		// some field set by the operator has changed
		consoleSvcMatchesSpec = false
	}

	// check the specification of the Console ClusterIP service
	if !consoleSvcMatchesSpec {
		consoleSvc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		consoleSvc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		consoleSvc.Spec.Ports = expectedSvc.Spec.Ports
		// we can only expose the service, not un-expose it
		if tenant.Spec.ExposeServices != nil && tenant.Spec.ExposeServices.Console && consoleSvc.Spec.Type != v1.ServiceTypeLoadBalancer {
			consoleSvc.Spec.Type = v1.ServiceTypeLoadBalancer
		}
		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, consoleSvc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
