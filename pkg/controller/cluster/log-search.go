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
	"fmt"

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	"github.com/minio/operator/pkg/resources/services"
	corev1 "k8s.io/api/core/v1"
)

type auditWebhookConfig struct {
	target string
	args   string
}

func newAuditWebhookConfig(tenant *miniov1.Tenant, secret *corev1.Secret) auditWebhookConfig {
	auditToken := string(secret.Data[miniov1.LogAuditTokenKey])
	whTarget := fmt.Sprintf("audit_webhook:%s", tenant.LogSearchAPIDeploymentName())

	logIngestEndpoint := fmt.Sprintf("%s/%s?token=%s", services.GetLogSearchAPIAddr(tenant), "api/ingest", auditToken)
	whArgs := fmt.Sprintf("%s endpoint=\"%s\"", whTarget, logIngestEndpoint)
	return auditWebhookConfig{
		target: whTarget,
		args:   whArgs,
	}
}
