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
	"os"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/pkg/env"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	// ConsoleTLSEnv Env variable to turn on / off Console TLS.
	ConsoleTLSEnv = "MINIO_CONSOLE_TLS_ENABLE"
	// DefaultConsoleDeploymentName is the default name of the console deployment
	DefaultConsoleDeploymentName = "console"
	// OperatorConsoleTLSSecretName is the name of secret created with TLS certs for Operator console
	OperatorConsoleTLSSecretName = "operator-console-tls"
)

// checkConsoleSvc validates the existence of the MinIO service and validate it's status against what the specification
// states
func (c *Controller) checkConsoleSvc(ctx context.Context, tenant *miniov2.Tenant, nsName types.NamespacedName) error {
	// Handle the Internal ClusterIP Service for Tenant
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.ConsoleCIServiceName())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningConsoleService, 0); err != nil {
				return err
			}
			klog.V(2).Infof("Creating a new Console Cluster IP Service for cluster %q", nsName)
			// Create the clusterIP service for the Tenant
			svc = services.NewClusterIPForConsole(tenant)
			svc, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Create(ctx, svc, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "Console Service Created")
		} else {
			return err
		}
	}

	// compare any other change from what is specified on the tenant
	expectedSvc := services.NewClusterIPForConsole(tenant)

	// check the expose status of the Console service
	svcMatchesSpec, err := minioSvcMatchesSpecification(svc, expectedSvc)

	// check the specification of the MinIO ClusterIP service
	if !svcMatchesSpec {
		if err != nil {
			klog.Infof("Console Service don't match: %s. Conciliating", err)
		}

		svc.ObjectMeta.Annotations = expectedSvc.ObjectMeta.Annotations
		svc.ObjectMeta.Labels = expectedSvc.ObjectMeta.Labels
		svc.Spec.Ports = expectedSvc.Spec.Ports
		// Only when ExposeServices is set an explicit value we do modifications to the service type
		if tenant.Spec.ExposeServices != nil {
			if tenant.Spec.ExposeServices.Console {
				svc.Spec.Type = v1.ServiceTypeLoadBalancer
			} else {
				svc.Spec.Type = v1.ServiceTypeClusterIP
			}
		}

		// update the selector
		svc.Spec.Selector = expectedSvc.Spec.Selector

		_, err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Updated", "Console Service Updated")
	}
	return err
}

// generateConsoleTLSCert Issues the Operator Console TLS Certificate
func (c *Controller) generateConsoleTLSCert() (*string, *string) {
	return c.generateTLSCert("console", OperatorConsoleTLSSecretName, getConsoleDeploymentName())
}

func (c *Controller) recreateOperatorConsoleCertsIfRequired(ctx context.Context) error {
	namespace := miniov2.GetNSFromFile()
	operatorConsoleTLSSecret, err := c.getTLSSecret(ctx, namespace, OperatorConsoleTLSSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(2).Info("TLS certificate not found. Generating one.")
			// Generate new certificate KeyPair for Operator Console
			c.generateConsoleTLSCert()
			return nil
		}
		return err
	}

	needsRenewal, err := c.certNeedsRenewal(operatorConsoleTLSSecret)
	if err != nil {
		return err
	}

	if !needsRenewal {
		return nil
	}

	// Expired cert. Delete the secret + CSR and re-create the cert
	err = c.deleteCSR(ctx, consoleCSRName())
	if err != nil {
		return err
	}
	klog.V(2).Info("Deleting the TLS secret of expired console cert")
	err = c.kubeClientSet.CoreV1().Secrets(namespace).Delete(ctx, OperatorConsoleTLSSecretName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	klog.V(2).Info("Generating a fresh TLS certificate for Console")
	// Generate new certificate KeyPair for Operator Console
	c.generateConsoleTLSCert()

	return nil
}

// isOperatorConsoleTLS Internal func, reads MINIO_CONSOLE_TLS_ENABLE ENV to identify if Operator Console TLS is enabled, default "off"
func isOperatorConsoleTLS() bool {
	value, set := os.LookupEnv(ConsoleTLSEnv)
	// By default, Console TLS is NOT used.
	return set && value == "on"
}

func getConsoleDeploymentName() string {
	return env.Get("MINIO_CONSOLE_DEPLOYMENT_NAME", DefaultConsoleDeploymentName)
}

// consoleCSRName Internal func returns the given console CSR name
func consoleCSRName() string {
	return getCSRName("console")
}
