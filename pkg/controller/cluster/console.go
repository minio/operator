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

	"github.com/minio/operator/pkg/resources/deployments"
	"k8s.io/apimachinery/pkg/api/equality"

	appsv1 "k8s.io/api/apps/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	"k8s.io/klog/v2"
)

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
			CommonName:   tenant.ConsoleCommonName(),
			Organization: tenant.Spec.CertConfig.OrganizationName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		DNSNames:           []string{tenant.ConsoleCIServiceName()},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		klog.Errorf("Unexpected error during creating the CSR: %v", err)
		return nil, nil, err
	}
	return privKeyBytes, csrBytes, nil
}

// createConsoleTLSCSR handles all the steps required to create the CSR: from creation of keys, submitting CSR and
// finally creating a secret that Console deployment will use to mount private key and certificate for TLS
// This Method Blocks till the CSR Request is approved via kubectl approve
func (c *Controller) createConsoleTLSCSR(ctx context.Context, tenant *miniov2.Tenant) error {
	privKeysBytes, csrBytes, err := generateConsoleCryptoData(tenant)
	if err != nil {
		klog.Errorf("Private Key and CSR generation failed with error: %v", err)
		return err
	}

	err = c.createCertificate(ctx, tenant.ConsolePodLabels(), tenant.ConsoleCSRName(), tenant.Namespace, csrBytes, tenant)
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
