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

package common

// Constants for the webhook endpoints
const (
	WebhookAPIVersion       = "/webhook/v1"
	UpgradeServerPort       = "4221"
	WebhookDefaultPort      = "4222"
	WebhookAPIBucketService = WebhookAPIVersion + "/bucketsrv"
	WebhookAPIUpdate        = WebhookAPIVersion + "/update"
)

const (
	// OperatorRuntimeEnv tells us which runtime we have. (EKS, Rancher, OpenShift, etc...)
	OperatorRuntimeEnv = "MINIO_OPERATOR_RUNTIME"
	// OperatorRuntimeK8s is the default runtime when no specific runtime is set
	OperatorRuntimeK8s Runtime = "K8S"
	// OperatorRuntimeEKS is the EKS runtime flag
	OperatorRuntimeEKS Runtime = "EKS"
	// OperatorRuntimeOpenshift is the Openshift runtime flag
	OperatorRuntimeOpenshift Runtime = "OPENSHIFT"
	// OperatorRuntimeRancher is the Rancher runtime flag
	OperatorRuntimeRancher Runtime = "RANCHER"

	// TLSCRT is  name of the field containing tls certificate in secret
	TLSCRT = "tls.crt"

	// CACRT name of the field containing ca certificate in secret
	CACRT = "ca.crt"

	// PublicCRT name of the field containing public certificate in secret
	PublicCRT = "public.crt"
)

// Runtimes is a map of the supported Kubernetes runtimes
var Runtimes = map[string]Runtime{
	"K8S":       OperatorRuntimeK8s,
	"EKS":       OperatorRuntimeEKS,
	"OPENSHIFT": OperatorRuntimeOpenshift,
	"RANCHER":   OperatorRuntimeRancher,
}

// Runtime type to for Operator runtime
type Runtime string
