// Copyright (C) 2020, MinIO, Inc.
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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/minio/madmin-go"
	"github.com/minio/operator/pkg/resources/secrets"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/minio/operator/pkg/resources/deployments"
	"github.com/minio/operator/pkg/resources/services"
	"github.com/minio/operator/pkg/resources/statefulsets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/klog/v2"
)

type auditWebhookConfig struct {
	target string
	args   string
}

func newAuditWebhookConfig(tenant *miniov2.Tenant, secret *corev1.Secret) auditWebhookConfig {
	auditToken := string(secret.Data[miniov2.LogAuditTokenKey])
	whTarget := fmt.Sprintf("audit_webhook:%s", tenant.LogSearchAPIDeploymentName())

	logIngestEndpoint := fmt.Sprintf("%s/%s?token=%s", services.GetLogSearchAPIAddr(tenant), "api/ingest", auditToken)
	whArgs := fmt.Sprintf("%s endpoint=\"%s\"", whTarget, logIngestEndpoint)
	return auditWebhookConfig{
		target: whTarget,
		args:   whArgs,
	}
}

// logDBStatefulsetMatchesSpec checks if the log DB statefulset `actualSS`
// matches the desired spec provided by `tenant`
func logDBStatefulsetMatchesSpec(tenant *miniov2.Tenant, actualSS *appsv1.StatefulSet) (bool, error) {
	if actualSS == nil {
		return false, errors.New("cannot process an empty Log DB statefulset")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}

	// Note: Since tenant's Headless service name for postgres server in log
	// search feature is internal to operator and is unaffectd by tenant
	// spec changes we use `actualSS`'s service name to create `expectedSS`.
	expectedSS := statefulsets.NewForLogDb(tenant, actualSS.Spec.ServiceName)
	// compare containers environment variables
	if miniov2.IsContainersEnvUpdated(expectedSS.Spec.Template.Spec.Containers, actualSS.Spec.Template.Spec.Containers) {
		return false, nil
	}
	if !equality.Semantic.DeepDerivative(expectedSS.Spec, actualSS.Spec) {
		// some fields set by the operator have changed
		return false, nil
	}
	return true, nil
}

// logSearchAPIDeploymentMatchesSpec checks if the log DB statefulset `actualSS`
// matches the desired spec provided by `tenant`
func logSearchAPIDeploymentMatchesSpec(tenant *miniov2.Tenant, actualDeployment *appsv1.Deployment) (bool, error) {
	if actualDeployment == nil {
		return false, errors.New("cannot process an empty Logsearch API deployment")
	}
	if tenant == nil {
		return false, errors.New("cannot process an empty tenant")
	}

	// compare container image version directly
	if !tenant.Spec.Log.EqualImage(actualDeployment.Spec.Template.Spec.Containers[0].Image) {
		klog.V(2).Infof("Tenant %s's Logsearch API server version %s doesn't match: %s", tenant.Name,
			tenant.Spec.Log.Image, actualDeployment.Spec.Template.Spec.Containers[0].Image)
		return false, nil
	}

	expectedDeployment := deployments.NewForLogSearchAPI(tenant)
	// compare containers environment variables
	if miniov2.IsContainersEnvUpdated(expectedDeployment.Spec.Template.Spec.Containers, actualDeployment.Spec.Template.Spec.Containers) {
		return false, nil
	}
	if !equality.Semantic.DeepDerivative(expectedDeployment.Spec, actualDeployment.Spec) {
		// some fields set by the operator have changed
		return false, nil
	}
	return true, nil
}

func (c *Controller) deleteLogHeadlessService(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.LogHLServiceName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Deleting Log Headless Service for %s", tenant.Namespace)
	err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Delete(ctx, tenant.LogHLServiceName(), metav1.DeleteOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "Deleted", "Log search headless service deleted")
	return err
}

func (c *Controller) deleteLogStatefulSet(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.LogStatefulsetName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Deleting Log StatefulSet for %s", tenant.Namespace)
	err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Delete(ctx, tenant.LogStatefulsetName(), metav1.DeleteOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "Deleted", "Log search statefulset deleted")
	return err
}

func (c *Controller) deleteLogSearchAPIDeployment(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.LogSearchAPIDeploymentName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Deleting Log Search API deployment for %s", tenant.Name)
	err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Delete(ctx, tenant.LogSearchAPIDeploymentName(), metav1.DeleteOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "Deleted", "Log search deployment deleted")
	return err
}

func (c *Controller) deleteLogSearchAPIService(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.LogSearchAPIServiceName())
	if k8serrors.IsNotFound(err) {
		return nil
	}
	klog.V(2).Infof("Delete Log Search API Service for %s", tenant.Namespace)
	err = c.kubeClientSet.CoreV1().Services(tenant.Namespace).Delete(ctx, tenant.LogSearchAPIServiceName(), metav1.DeleteOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeWarning, "Deleted", "Log search service deleted")
	return err
}

func (c *Controller) checkAndCreateLogHeadless(ctx context.Context, tenant *miniov2.Tenant) (*corev1.Service, error) {
	svc, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.LogHLServiceName())
	if err == nil || !k8serrors.IsNotFound(err) {
		return svc, err
	}

	klog.V(2).Infof("Creating a new Log Headless Service for %s", tenant.Namespace)
	svc = services.NewHeadlessForLog(tenant)
	_, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "Log headless service created")
	return svc, err
}

func (c *Controller) checkAndCreateLogStatefulSet(ctx context.Context, tenant *miniov2.Tenant, svcName string) error {
	logPgSS, err := c.statefulSetLister.StatefulSets(tenant.Namespace).Get(tenant.LogStatefulsetName())
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningLogPGStatefulSet, 0); err != nil {
			return err
		}

		klog.V(2).Infof("Creating a new Log StatefulSet for %s", tenant.Namespace)
		searchSS := statefulsets.NewForLogDb(tenant, svcName)
		_, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Create(ctx, searchSS, metav1.CreateOptions{})
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Created", "Log statefulset created")
		return err

	}

	// check if expected and actual values of Log DB spec match
	dbSpecMatches, err := logDBStatefulsetMatchesSpec(tenant, logPgSS)
	if err != nil {
		return err
	}
	if !dbSpecMatches {
		// Note: using current spec replica count works as long as we don't expose replicas via tenant spec.
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingLogPGStatefulSet, *logPgSS.Spec.Replicas); err != nil {
			return err
		}
		logPgSS = statefulsets.NewForLogDb(tenant, svcName)
		if _, err = c.kubeClientSet.AppsV1().StatefulSets(tenant.Namespace).Update(ctx, logPgSS, metav1.UpdateOptions{}); err != nil {
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Updated", "Log statefulset updated")
	}

	return nil
}

func (c *Controller) checkAndCreateLogSearchAPIService(ctx context.Context, tenant *miniov2.Tenant) error {
	_, err := c.serviceLister.Services(tenant.Namespace).Get(tenant.LogSearchAPIServiceName())
	if err == nil || !k8serrors.IsNotFound(err) {
		return err
	}

	klog.V(2).Infof("Creating a new Log Search API Service for %s", tenant.Namespace)
	svc := services.NewClusterIPForLogSearchAPI(tenant)
	_, err = c.kubeClientSet.CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "SvcCreated", "Log search service created")
	return err
}

func (c *Controller) checkAndCreateLogSearchAPIDeployment(ctx context.Context, tenant *miniov2.Tenant) error {
	logSearchDeployment, err := c.deploymentLister.Deployments(tenant.Namespace).Get(tenant.LogSearchAPIDeploymentName())
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusProvisioningLogSearchAPIDeployment, 0); err != nil {
			return err
		}

		klog.V(2).Infof("Creating a new Log Search API deployment for %s", tenant.Name)
		_, err = c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Create(ctx, deployments.NewForLogSearchAPI(tenant), metav1.CreateOptions{})
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Created", "Log search service deployment")
		return err
	}

	// check if expected and actual values of Log search API deployment match
	apiDeploymentMatches, err := logSearchAPIDeploymentMatchesSpec(tenant, logSearchDeployment)
	if err != nil {
		return err
	}
	if !apiDeploymentMatches {
		// Note: using current spec replica count works as long as we don't expose replicas via tenant spec.
		if tenant, err = c.updateTenantStatus(ctx, tenant, StatusUpdatingLogSearchAPIServer, 0); err != nil {
			return err
		}
		logSearchDeployment = deployments.NewForLogSearchAPI(tenant)
		if _, err := c.kubeClientSet.AppsV1().Deployments(tenant.Namespace).Update(ctx, logSearchDeployment, metav1.UpdateOptions{}); err != nil {
			return err
		}
		c.RegisterEvent(ctx, tenant, corev1.EventTypeNormal, "Updated", "Log search deployment updated")
	}
	return nil
}

func (c *Controller) checkAndCreateLogSecret(ctx context.Context, tenant *miniov2.Tenant) (*corev1.Secret, error) {
	secret, err := c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Get(ctx, tenant.LogSecretName(), metav1.GetOptions{})
	if err == nil || !k8serrors.IsNotFound(err) {
		return secret, err
	}

	klog.V(2).Infof("Creating a new Log secret for %s", tenant.Name)
	secret, err = c.kubeClientSet.CoreV1().Secrets(tenant.Namespace).Create(ctx, secrets.LogSecret(tenant), metav1.CreateOptions{})
	return secret, err
}

func (c *Controller) checkAndConfigureLogSearchAPI(ctx context.Context, tenant *miniov2.Tenant, secret *corev1.Secret, adminClnt *madmin.AdminClient) error {
	// Check if audit webhook is configured for tenant's MinIO
	auditCfg := newAuditWebhookConfig(tenant, secret)
	_, err := adminClnt.GetConfigKV(ctx, auditCfg.target)
	if err != nil {
		// check if log search is ready
		if err = c.checkLogSearchAPIReady(tenant); err != nil {
			klog.V(2).Info(err)
			if _, err = c.updateTenantStatus(ctx, tenant, StatusWaitingForLogSearchReadyState, 0); err != nil {
				return err
			}
			return ErrLogSearchNotReady
		}
		restart, err := adminClnt.SetConfigKV(ctx, auditCfg.args)
		if err != nil {
			return err
		}
		if restart {
			// Restart MinIO for config update to take effect
			if err = adminClnt.ServiceRestart(ctx); err != nil {
				klog.V(2).Info("error restarting minio")
				klog.V(2).Info(err)
			}
			klog.V(2).Info("done restarting minio")
		}
		return nil
	}
	return err
}

func (c *Controller) checkLogSearchAPIReady(tenant *miniov2.Tenant) error {
	endpoint := fmt.Sprintf("http://%s.%s.svc.%s:8080", tenant.LogSearchAPIServiceName(), tenant.Namespace, miniov2.GetClusterDomain())
	client := http.Client{Timeout: 100 * time.Millisecond}
	resp, err := client.Get(endpoint)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.V(2).Info(err)
		}
	}()

	if resp.StatusCode == 404 {
		return nil
	}

	return errors.New("Log Search API Not Ready")
}
