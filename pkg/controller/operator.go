// This file is part of MinIO Operator
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

package controller

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xcerts "github.com/minio/pkg/certs"
	"github.com/minio/pkg/env"
	"k8s.io/klog/v2"
	k8sscheme "k8s.io/kubectl/pkg/scheme"
)

const (
	// CertPasswordEnv Env variable is used to decrypt the private key in the TLS certificate for operator if need it
	CertPasswordEnv = "OPERATOR_CERT_PASSWD"
	// OperatorDeplymentNameEnv Env variable to specify a custom deployment name for Operator
	OperatorDeplymentNameEnv = "MINIO_OPERATOR_DEPLOYMENT_NAME"
	// OperatorCATLSSecretName is the name of the secret for the operator CA
	OperatorCATLSSecretName = "operator-ca-tls"
	// DefaultDeploymentName is the default name of the operator deployment
	DefaultDeploymentName = "minio-operator"
	// DefaultOperatorImage is the version fo the operator being used
	DefaultOperatorImage = "minio/operator:v5.0.6"
)

var serverCertsManager *xcerts.Manager

// rolloutRestartDeployment - executes the equivalent to kubectl rollout restart deployment
func (c *Controller) rolloutRestartDeployment(deployName string) error {
	ctx := context.Background()
	namespace := miniov2.GetNSFromFile()
	deployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	data, err := runtime.Encode(k8sscheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), deployment)
	if err != nil {
		return err
	}
	_, err2 := c.kubeClientSet.AppsV1().Deployments(namespace).Patch(ctx, deployName, types.StrategicMergePatchType, data, metav1.PatchOptions{FieldManager: "kubectl-rollout"})
	if err2 != nil {
		return err2
	}
	return nil
}

func (c *Controller) fetchUserCredentials(ctx context.Context, tenant *miniov2.Tenant) []*v1.Secret {
	var userCredentials []*v1.Secret
	for _, credential := range tenant.Spec.Users {
		credentialSecret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, credential.Name, metav1.GetOptions{})
		if err == nil && credentialSecret != nil {
			userCredentials = append(userCredentials, credentialSecret)
		}
	}
	return userCredentials
}

func (c *Controller) getTransport() *http.Transport {
	if c.transport != nil {
		return c.transport
	}
	rootCAs := miniov2.MustGetSystemCertPool()
	// Default kubernetes CA certificate
	rootCAs.AppendCertsFromPEM(miniov2.GetPodCAFromFile())

	// Custom ca certificate to be used by operator
	operatorCATLSCert, err := c.kubeClientSet.CoreV1().Secrets(miniov2.GetNSFromFile()).Get(context.Background(), OperatorCATLSSecretName, metav1.GetOptions{})
	if err == nil && operatorCATLSCert != nil {
		if val, ok := operatorCATLSCert.Data["tls.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
		if val, ok := operatorCATLSCert.Data["ca.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
		if val, ok := operatorCATLSCert.Data["public.crt"]; ok {
			rootCAs.AppendCertsFromPEM(val)
		}
	}

	// These chunk of code is intended for OpenShift ONLY and it will help us trust the signer to solve issue:
	// https://github.com/minio/operator/issues/1412
	openShiftCATLSCert, err := c.kubeClientSet.CoreV1().Secrets("openshift-kube-controller-manager-operator").Get(
		context.Background(), "csr-signer", metav1.GetOptions{})
	klog.Info("Checking if this is OpenShift Environment to append the certificates...")
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Do nothing special, because this is maybe k8s vanilla
			klog.Info("csr-signer secret wasn't found, very likely this is not OpenShift but k8s Vanilla or other...")
		} else {
			// Lack of permissions to read the secret
			klog.Errorf("csr-signer secret was found but we failed to get openShiftCATLSCert: %#v", err)
		}
	} else if err == nil && openShiftCATLSCert != nil {
		// When secret was obtained with no errors
		if val, ok := openShiftCATLSCert.Data["tls.crt"]; ok {
			// OpenShift csr-signer secret has tls.crt certificates that we need to append in order
			// to trust the signer. If we append the val, Operator will be able to provisioning the
			// initial users and get Tenant Health, so tenant can be properly initialized and in
			// green status, otherwise if we don't append it, it will get stuck and expose this
			// issue in the log:
			// Failed to get cluster health: Get "https://minio.tenant-lite.svc.cluster.local/minio/health/cluster":
			// x509: certificate signed by unknown authority
			klog.Info("Appending OpenShift csr-signer to trust the Signer")
			rootCAs.AppendCertsFromPEM(val)
		}
	}

	dialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 15 * time.Second,
	}
	c.transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       15 * time.Second,
		ResponseHeaderTimeout: 15 * time.Minute,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 15 * time.Second,
		// Go net/http automatically unzip if content-type is
		// gzip disable this feature, as we are always interested
		// in raw stream.
		DisableCompression: true,
		TLSClientConfig: &tls.Config{
			// Can't use SSLv3 because of POODLE and BEAST
			// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
			// Can't use TLSv1.1 because of RC4 cipher usage
			MinVersion: tls.VersionTLS12,
			RootCAs:    rootCAs,
		},
	}

	return c.transport
}

func (c *Controller) createUsers(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (err error) {
	defer func() {
		if err == nil {
			if _, err = c.updateProvisionedUsersStatus(ctx, tenant, true); err != nil {
				klog.V(2).Infof(err.Error())
			}
		}
	}()

	userCredentials := c.fetchUserCredentials(ctx, tenant)
	if len(userCredentials) == 0 {
		return nil
	}

	if _, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningInitialUsers, 0); err != nil {
		return err
	}

	// get a new admin client
	adminClient, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		klog.Errorf("Error instantiating adminClnt: %v", err)
		return err
	}

	// configuration that means MinIO is running with LDAP enabled
	// and we need to skip the console user creation
	if err = tenant.CreateUsers(adminClient, userCredentials, tenantConfiguration); err != nil {
		klog.V(2).Infof("Unable to create MinIO users: %v", err)
	}

	return err
}

func (c *Controller) createBuckets(ctx context.Context, tenant *miniov2.Tenant, tenantConfiguration map[string][]byte) (err error) {
	defer func() {
		if err == nil {
			if _, err = c.updateProvisionedBucketStatus(ctx, tenant, true); err != nil {
				klog.V(2).Infof(err.Error())
			}
		}
	}()

	tenantBuckets := tenant.Spec.Buckets
	if len(tenantBuckets) == 0 {
		return nil
	}

	if _, err := c.updateTenantStatus(ctx, tenant, StatusProvisioningDefaultBuckets, 0); err != nil {
		return err
	}

	// get a new admin client
	minioClient, err := tenant.NewMinIOUser(tenantConfiguration, c.getTransport())
	if err != nil {
		// show the error and continue
		klog.Errorf("Error instantiating minio Client: %v ", err)
		return err
	}

	if err := tenant.CreateBuckets(minioClient, tenantBuckets...); err != nil {
		klog.V(2).Infof("Unable to create MinIO buckets: %v", err)
		return err
	}

	return nil
}

// getOperatorDeploymentName Internal func returns the Operator deployment name from MINIO_OPERATOR_DEPLOYMENT_NAME ENV variable or the default name
func getOperatorDeploymentName() string {
	return env.Get(OperatorDeplymentNameEnv, DefaultDeploymentName)
}
