// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package legacy

import (
	"fmt"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

// Legacy Prometheus
const (
	LogSearchAPIContainerName = "log-search-api"
	PrometheusHLSvcNameSuffix = "-prometheus-hl-svc"
	LogHLSvcNameSuffix        = "-log-hl-svc"
)

// LogSearchAPIDeploymentName returns name of Log Search API server deployment
func LogSearchAPIDeploymentName(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s-%s", t.Name, LogSearchAPIContainerName)
}

// LogSecretName returns name of secret shared by Log PG server and log-search-api server
func LogSecretName(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s-%s", t.Name, "log-secret")
}

// LogSearchAPIServiceName returns name of Log Search API service name
func LogSearchAPIServiceName(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s-%s", t.Name, LogSearchAPIContainerName)
}

// PrometheusStatefulsetName returns name of statefulset meant for Prometheus
// metrics.
func PrometheusStatefulsetName(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s-%s", t.Name, "prometheus")
}

// PrometheusHLServiceName returns name of Headless service for the Log
// statefulsets
func PrometheusHLServiceName(t *miniov2.Tenant) string {
	return t.Name + PrometheusHLSvcNameSuffix
}

// LogStatefulsetName returns name of statefulsets meant for Log feature
func LogStatefulsetName(t *miniov2.Tenant) string {
	return fmt.Sprintf("%s-%s", t.Name, "log")
}

// LogHLServiceName returns name of Headless service for the Log statefulsets
func LogHLServiceName(t *miniov2.Tenant) string {
	return t.Name + LogHLSvcNameSuffix
}
