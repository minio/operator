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

package api

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/minio/operator/api/operations/operator_api"

	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// convertModelSCToK8sSC validates and converts from models.SecurityContext to corev1.PodSecurityContext
func convertModelSCToK8sSC(sc *models.SecurityContext) (*corev1.PodSecurityContext, error) {
	if sc == nil {
		return nil, errors.New("invalid security context")
	}
	runAsUser, err := strconv.ParseInt(*sc.RunAsUser, 10, 64)
	if err != nil {
		return nil, err
	}
	runAsGroup, err := strconv.ParseInt(*sc.RunAsGroup, 10, 64)
	if err != nil {
		return nil, err
	}
	fsGroup, err := strconv.ParseInt(sc.FsGroup, 10, 64)
	if err != nil {
		return nil, err
	}
	FSGroupChangePolicy := corev1.PodFSGroupChangePolicy("Always")
	if sc.FsGroupChangePolicy != "" {
		FSGroupChangePolicy = corev1.PodFSGroupChangePolicy(sc.FsGroupChangePolicy)
	}
	return &corev1.PodSecurityContext{
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		RunAsNonRoot:        sc.RunAsNonRoot,
		FSGroup:             &fsGroup,
		FSGroupChangePolicy: &FSGroupChangePolicy,
	}, nil
}

// convertK8sSCToModelSC validates and converts from corev1.PodSecurityContext to models.SecurityContext
func convertK8sSCToModelSC(sc *corev1.PodSecurityContext) *models.SecurityContext {
	runAsUser := strconv.FormatInt(*sc.RunAsUser, 10)
	runAsGroup := strconv.FormatInt(*sc.RunAsGroup, 10)
	fsGroup := strconv.FormatInt(*sc.FSGroup, 10)
	fsGroupChangePolicy := "Always"

	if sc.FSGroupChangePolicy != nil {
		fsGroupChangePolicy = string(*sc.FSGroupChangePolicy)
	}

	return &models.SecurityContext{
		RunAsUser:           &runAsUser,
		RunAsGroup:          &runAsGroup,
		RunAsNonRoot:        sc.RunAsNonRoot,
		FsGroup:             fsGroup,
		FsGroupChangePolicy: fsGroupChangePolicy,
	}
}

// tenantUpdateCertificates receives the keyPair certificates (public and private keys) for Minio and Console and will try
// to replace the existing kubernetes secrets with the new values, then will restart the affected pods so the new volumes can be mounted
func tenantUpdateCertificates(ctx context.Context, operatorClient OperatorClientI, clientSet K8sClientI, namespace string, params operator_api.TenantUpdateCertificateParams) error {
	tenantName := params.Tenant
	tenant, err := operatorClient.TenantGet(ctx, namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	body := params.Body
	// check if MinIO is deployed with external certs and user provided new MinIO keypair
	if tenant.ExternalCert() && body.MinioServerCertificates != nil {
		minioCertSecretName := fmt.Sprintf("%s-instance-external-certificates", tenantName)
		// update certificates
		if _, err := createOrReplaceExternalCertSecrets(ctx, clientSet, namespace, body.MinioServerCertificates, minioCertSecretName, tenantName); err != nil {
			return err
		}
	}
	return nil
}

// getTenantUpdateCertificatesResponse wrapper of tenantUpdateCertificates
func getTenantUpdateCertificatesResponse(session *models.Principal, params operator_api.TenantUpdateCertificateParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientSet, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	k8sClient := k8sClient{
		client: clientSet,
	}
	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	if err := tenantUpdateCertificates(ctx, &opClient, &k8sClient, params.Namespace, params); err != nil {
		return ErrorWithContext(ctx, err, ErrUnableToUpdateTenantCertificates)
	}
	return nil
}

type tenantSecret struct {
	Name    string
	Content map[string][]byte
}

// createOrReplaceSecrets receives an array of Tenant Secrets to be stored as k8s secrets
func createOrReplaceSecrets(ctx context.Context, clientSet K8sClientI, ns string, secrets []tenantSecret, tenantName string) ([]*miniov2.LocalCertificateReference, error) {
	var k8sSecrets []*miniov2.LocalCertificateReference
	for _, secret := range secrets {
		if len(secret.Content) > 0 && secret.Name != "" {
			// delete secret with same name if exists
			err := clientSet.deleteSecret(ctx, ns, secret.Name, metav1.DeleteOptions{})
			if err != nil {
				// log the errors if any and continue
				LogError("deleting secret name %s failed: %v, continuing..", secret.Name, err)
			}
			k8sSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secret.Name,
					Labels: map[string]string{
						miniov2.TenantLabel: tenantName,
					},
				},
				Type: corev1.SecretTypeOpaque,
				Data: secret.Content,
			}
			_, err = clientSet.createSecret(ctx, ns, k8sSecret, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
			k8sSecrets = append(k8sSecrets, &miniov2.LocalCertificateReference{
				Name: secret.Name,
				Type: "Opaque",
			})
		}
	}
	return k8sSecrets, nil
}

// createOrReplaceExternalCertSecrets receives an array of KeyPairs (public and private key), encoded in base64, decode it and generate an equivalent number of kubernetes
// secrets to be used by the miniov2 for TLS encryption
func createOrReplaceExternalCertSecrets(ctx context.Context, clientSet K8sClientI, ns string, keyPairs []*models.KeyPairConfiguration, secretName, tenantName string) ([]*miniov2.LocalCertificateReference, error) {
	var keyPairSecrets []*miniov2.LocalCertificateReference
	for i, keyPair := range keyPairs {
		keyPairSecretName := fmt.Sprintf("%s-%d", secretName, i)
		if keyPair == nil || keyPair.Crt == nil || keyPair.Key == nil || *keyPair.Crt == "" || *keyPair.Key == "" {
			return nil, errors.New("certificate files must not be empty")
		}
		// delete secret with same name if exists
		err := clientSet.deleteSecret(ctx, ns, keyPairSecretName, metav1.DeleteOptions{})
		if err != nil {
			// log the errors if any and continue
			LogError("deleting secret name %s failed: %v, continuing..", keyPairSecretName, err)
		}
		imm := true
		tlsCrt, err := base64.StdEncoding.DecodeString(*keyPair.Crt)
		if err != nil {
			return nil, err
		}
		tlsKey, err := base64.StdEncoding.DecodeString(*keyPair.Key)
		if err != nil {
			return nil, err
		}
		// check if the key pair is valid
		if _, err = tls.X509KeyPair(tlsCrt, tlsKey); err != nil {
			return nil, err
		}
		externalTLSCertificateSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: keyPairSecretName,
				Labels: map[string]string{
					miniov2.TenantLabel: tenantName,
				},
			},
			Type:      corev1.SecretTypeTLS,
			Immutable: &imm,
			Data: map[string][]byte{
				"tls.crt": tlsCrt,
				"tls.key": tlsKey,
			},
		}
		_, err = clientSet.createSecret(ctx, ns, externalTLSCertificateSecret, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		// Certificates used by the minio instance
		keyPairSecrets = append(keyPairSecrets, &miniov2.LocalCertificateReference{
			Name: keyPairSecretName,
			Type: "kubernetes.io/tls",
		})
	}
	return keyPairSecrets, nil
}

/* this is what I saved from tenants.go that will be lost in the rebase to new setup

// Get tenant log report
api.OperatorAPIGetTenantLogReportHandler = operator_api.GetTenantLogReportHandlerFunc(func(params operator_api.GetTenantLogReportParams, principal *models.Principal) middleware.Responder {
	payload, err := getTenantLogReportResponse(principal, params)
	if err != nil {
		return operator_api.NewGetTenantLogReportDefault(int(err.Code)).WithPayload(err)
	}
	return operator_api.NewGetTenantLogReportOK().WithPayload(payload)
})


func getTenantLogReportResponse(session *models.Principal, params operator_api.GetTenantLogReportParams) (*models.TenantLogReport, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	payload := &models.TenantLogReport{}

	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return payload, ErrorWithContext(ctx, err)
	}
	operatorCli, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return payload, ErrorWithContext(ctx, err)
	}
	opClient := &operatorClient{
		client: operatorCli,
	}
	if err != nil {
		return payload, ErrorWithContext(ctx, err)
	}
	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return payload, ErrorWithContext(ctx, err)
	}
	reportBytes, reportError := generateTenantLogReport(ctx, clientset.CoreV1(), params.Tenant, params.Namespace, minTenant)
	if reportError != nil {
		return payload, ErrorWithContext(ctx, reportError)
	}
	payload.Filename = params.Tenant + "-report.zip"
	sEnc := base64.StdEncoding.EncodeToString(reportBytes)
	payload.Blob = sEnc

	return payload, nil
}

func generateTenantLogReport(ctx context.Context, coreInterface v1.CoreV1Interface, tenantName string, namespace string, tenant *miniov2.Tenant) ([]byte, *models.Error) {
	if tenantName == "" || namespace == "" {
		return []byte{}, ErrorWithContext(ctx, errors.New("Namespace and Tenant name cannot be empty"))
	}
	podListOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("v1.min.io/tenant=%s", tenantName),
	}
	pods, err := coreInterface.Pods(namespace).List(ctx, podListOpts)
	if err != nil {
		return []byte{}, ErrorWithContext(ctx, err)
	}
	events := coreInterface.Events(namespace)

	var report bytes.Buffer

	zipw := zip.NewWriter(&report)

	tenantAsYaml, err := yaml.Marshal(tenant)
	if err == nil {
		f, err := zipw.Create(tenantName + ".yaml")

		if err == nil {

			_, err := f.Write(tenantAsYaml)
			if err != nil {
				return []byte{}, ErrorWithContext(ctx, err)
			}
		}
	} else {
		return []byte{}, ErrorWithContext(ctx, err)
	}
	for i := 0; i < len(pods.Items); i++ {
		listOpts := &corev1.PodLogOptions{Container: "minio"}
		request := coreInterface.Pods(namespace).GetLogs(pods.Items[i].Name, listOpts)
		buff, err := request.DoRaw(ctx)
		if err == nil {
			f, err := zipw.Create(pods.Items[i].Name + ".log")
			if err == nil {
				f.Write(buff)
			}
		} else {
			return []byte{}, ErrorWithContext(ctx, err)
		}
		podEvents, err := events.List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pods.Items[i].UID)})
		if err == nil {
			podEventsJSON, err := json.Marshal(podEvents)
			if err == nil {
				f, err := zipw.Create(pods.Items[i].Name + "-events.txt")
				if err == nil {
					f.Write(podEventsJSON)
				}
			}
		} else {
			return []byte{}, ErrorWithContext(ctx, err)
		}
		status := pods.Items[i].Status
		statusJSON, err := json.Marshal(status)
		if err == nil {
			f, err := zipw.Create(pods.Items[i].Name + "-status.txt")
			if err == nil {
				f.Write(statusJSON)
			}
		} else {
			return []byte{}, ErrorWithContext(ctx, err)
		}
	}
	zipw.Close()

	return report.Bytes(), nil
}
*/
