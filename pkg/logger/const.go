// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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

package logger

import (
	"context"

	"github.com/minio/operator/pkg/logger/target/http"
)

// Audit/Logger constants
const (
	EnvLoggerJSONEnable      = "CONSOLE_LOGGER_JSON_ENABLE"
	EnvLoggerAnonymousEnable = "CONSOLE_LOGGER_ANONYMOUS_ENABLE"
	EnvLoggerQuietEnable     = "CONSOLE_LOGGER_QUIET_ENABLE"

	EnvGlobalDeploymentID      = "CONSOLE_GLOBAL_DEPLOYMENT_ID"
	EnvLoggerWebhookEnable     = "CONSOLE_LOGGER_WEBHOOK_ENABLE"
	EnvLoggerWebhookEndpoint   = "CONSOLE_LOGGER_WEBHOOK_ENDPOINT"
	EnvLoggerWebhookAuthToken  = "CONSOLE_LOGGER_WEBHOOK_AUTH_TOKEN"
	EnvLoggerWebhookClientCert = "CONSOLE_LOGGER_WEBHOOK_CLIENT_CERT"
	EnvLoggerWebhookClientKey  = "CONSOLE_LOGGER_WEBHOOK_CLIENT_KEY"
	EnvLoggerWebhookQueueSize  = "CONSOLE_LOGGER_WEBHOOK_QUEUE_SIZE"

	EnvAuditWebhookEnable     = "CONSOLE_AUDIT_WEBHOOK_ENABLE"
	EnvAuditWebhookEndpoint   = "CONSOLE_AUDIT_WEBHOOK_ENDPOINT"
	EnvAuditWebhookAuthToken  = "CONSOLE_AUDIT_WEBHOOK_AUTH_TOKEN"
	EnvAuditWebhookClientCert = "CONSOLE_AUDIT_WEBHOOK_CLIENT_CERT"
	EnvAuditWebhookClientKey  = "CONSOLE_AUDIT_WEBHOOK_CLIENT_KEY"
	EnvAuditWebhookQueueSize  = "CONSOLE_AUDIT_WEBHOOK_QUEUE_SIZE"
)

// Config console and http logger targets
type Config struct {
	HTTP         map[string]http.Config `json:"http"`
	AuditWebhook map[string]http.Config `json:"audit"`
}

var (
	globalDeploymentID string
	// GlobalContext context
	GlobalContext context.Context
)
