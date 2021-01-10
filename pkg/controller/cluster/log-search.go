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
	"errors"
	"fmt"

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
	if !equality.Semantic.DeepDerivative(expectedDeployment.Spec, actualDeployment.Spec) {
		// some fields set by the operator have changed
		return false, nil
	}
	return true, nil
}
