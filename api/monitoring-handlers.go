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

package api

import (
	"context"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerMonitoringHandlers(api *operations.OperatorAPI) {
	// Get tenant monitoring info
	api.OperatorAPIGetTenantMonitoringHandler = operator_api.GetTenantMonitoringHandlerFunc(func(params operator_api.GetTenantMonitoringParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantMonitoringResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantMonitoringDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantMonitoringOK().WithPayload(payload)
	})
	// Set configuration fields for Prometheus monitoring on a tenant
	api.OperatorAPISetTenantMonitoringHandler = operator_api.SetTenantMonitoringHandlerFunc(func(params operator_api.SetTenantMonitoringParams, session *models.Principal) middleware.Responder {
		_, err := setTenantMonitoringResponse(session, params)
		if err != nil {
			return operator_api.NewSetTenantMonitoringDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewSetTenantMonitoringCreated()
	})
}

// sets tenant Prometheus monitoring cofiguration fields to values provided
func setTenantMonitoringResponse(session *models.Principal, params operator_api.SetTenantMonitoringParams) (bool, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return false, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return false, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
	return setTenantMonitoring(ctx, minTenant, opClient, params)
}

func setTenantMonitoring(ctx context.Context, minTenant *miniov2.Tenant, opClient OperatorClientI, params operator_api.SetTenantMonitoringParams) (bool, *models.Error) {
	if params.Data.Toggle {
		if params.Data.PrometheusEnabled {
			minTenant.Spec.Prometheus = nil
		} else {
			promDiskSpaceGB := 5
			promImage := ""
			minTenant.Spec.Prometheus = &miniov2.PrometheusConfig{
				DiskCapacityDB: swag.Int(promDiskSpaceGB),
				Image:          promImage,
			}
		}
		_, err := opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		return true, nil
	}

	labels := make(map[string]string)
	for i := 0; i < len(params.Data.Labels); i++ {
		if params.Data.Labels[i] != nil {
			labels[params.Data.Labels[i].Key] = params.Data.Labels[i].Value
		}
	}
	annotations := make(map[string]string)
	for i := 0; i < len(params.Data.Annotations); i++ {
		if params.Data.Annotations[i] != nil {
			annotations[params.Data.Annotations[i].Key] = params.Data.Annotations[i].Value
		}
	}
	nodeSelector := make(map[string]string)
	for i := 0; i < len(params.Data.NodeSelector); i++ {
		if params.Data.NodeSelector[i] != nil {
			nodeSelector[params.Data.NodeSelector[i].Key] = params.Data.NodeSelector[i].Value
		}
	}

	monitoringResourceRequest := make(corev1.ResourceList)
	if params.Data.MonitoringCPURequest != "" {
		cpuQuantity, err := resource.ParseQuantity(params.Data.MonitoringCPURequest)
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		monitoringResourceRequest["cpu"] = cpuQuantity
	}

	if params.Data.MonitoringMemRequest != "" {
		memQuantity, err := resource.ParseQuantity(params.Data.MonitoringMemRequest)
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
		monitoringResourceRequest["memory"] = memQuantity
	}

	minTenant.Spec.Prometheus.Resources.Requests = monitoringResourceRequest
	minTenant.Spec.Prometheus.Labels = labels
	minTenant.Spec.Prometheus.Annotations = annotations
	minTenant.Spec.Prometheus.NodeSelector = nodeSelector
	minTenant.Spec.Prometheus.Image = params.Data.Image
	minTenant.Spec.Prometheus.SideCarImage = params.Data.SidecarImage
	minTenant.Spec.Prometheus.InitImage = params.Data.InitImage
	if params.Data.StorageClassName == "" {
		minTenant.Spec.Prometheus.StorageClassName = nil
	} else {
		minTenant.Spec.Prometheus.StorageClassName = &params.Data.StorageClassName
	}

	diskCapacityGB, err := strconv.Atoi(params.Data.DiskCapacityGB)
	if err == nil {
		*minTenant.Spec.Prometheus.DiskCapacityDB = diskCapacityGB
	}

	minTenant.Spec.Prometheus.ServiceAccountName = params.Data.ServiceAccountName
	minTenant.Spec.Prometheus.SecurityContext, err = convertModelSCToK8sSC(params.Data.SecurityContext)
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	_, err = opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	return true, nil
}

// get values for prometheus metrics
func getTenantMonitoringResponse(session *models.Principal, params operator_api.GetTenantMonitoringParams) (*models.TenantMonitoringInfo, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := GetOperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minInst, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return getTenantMonitoring(minInst), nil
}

func getTenantMonitoring(minInst *miniov2.Tenant) *models.TenantMonitoringInfo {
	monitoringInfo := &models.TenantMonitoringInfo{}

	if minInst.Spec.Prometheus != nil {
		monitoringInfo.PrometheusEnabled = true
	} else {
		monitoringInfo.PrometheusEnabled = false
		return monitoringInfo
	}

	var storageClassName string
	if minInst.Spec.Prometheus.StorageClassName != nil {
		storageClassName = *minInst.Spec.Prometheus.StorageClassName
		monitoringInfo.StorageClassName = storageClassName
	}

	var requestedCPU string
	var requestedMem string

	if minInst.Spec.Prometheus.Resources.Requests != nil {
		// Parse cpu request
		if requestedCPUQ, ok := minInst.Spec.Prometheus.Resources.Requests["cpu"]; ok && requestedCPUQ.Value() != 0 {
			requestedCPU = strconv.FormatInt(requestedCPUQ.Value(), 10)
			monitoringInfo.MonitoringCPURequest = requestedCPU
		}
		// Parse memory request
		if requestedMemQ, ok := minInst.Spec.Prometheus.Resources.Requests["memory"]; ok && requestedMemQ.Value() != 0 {
			requestedMem = strconv.FormatInt(requestedMemQ.Value(), 10)
			monitoringInfo.MonitoringMemRequest = requestedMem
		}
	}

	if len(minInst.Spec.Prometheus.Labels) != 0 && minInst.Spec.Prometheus.Labels != nil {
		mLabels := []*models.Label{}
		for k, v := range minInst.Spec.Prometheus.Labels {
			mLabels = append(mLabels, &models.Label{Key: k, Value: v})
		}
		monitoringInfo.Labels = mLabels
	}

	if len(minInst.Spec.Prometheus.Annotations) != 0 && minInst.Spec.Prometheus.Annotations != nil {
		mAnnotations := []*models.Annotation{}
		for k, v := range minInst.Spec.Prometheus.Annotations {
			mAnnotations = append(mAnnotations, &models.Annotation{Key: k, Value: v})
		}
		monitoringInfo.Annotations = mAnnotations
	}

	if len(minInst.Spec.Prometheus.NodeSelector) != 0 && minInst.Spec.Prometheus.NodeSelector != nil {
		mNodeSelector := []*models.NodeSelector{}
		for k, v := range minInst.Spec.Prometheus.NodeSelector {
			mNodeSelector = append(mNodeSelector, &models.NodeSelector{Key: k, Value: v})
		}
		monitoringInfo.NodeSelector = mNodeSelector
	}

	if *minInst.Spec.Prometheus.DiskCapacityDB != 0 {
		monitoringInfo.DiskCapacityGB = strconv.Itoa(*minInst.Spec.Prometheus.DiskCapacityDB)
	}
	if len(minInst.Spec.Prometheus.Image) != 0 {
		monitoringInfo.Image = minInst.Spec.Prometheus.Image
	}
	if len(minInst.Spec.Prometheus.InitImage) != 0 {
		monitoringInfo.InitImage = minInst.Spec.Prometheus.InitImage
	}
	if len(minInst.Spec.Prometheus.ServiceAccountName) != 0 {
		monitoringInfo.ServiceAccountName = minInst.Spec.Prometheus.ServiceAccountName
	}
	if len(minInst.Spec.Prometheus.SideCarImage) != 0 {
		monitoringInfo.SidecarImage = minInst.Spec.Prometheus.SideCarImage
	}
	if minInst.Spec.Prometheus.SecurityContext != nil {
		monitoringInfo.SecurityContext = convertK8sSCToModelSC(minInst.Spec.Prometheus.SecurityContext)
	}
	return monitoringInfo
}
