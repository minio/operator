package api

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/cluster"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerTenantLogsHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetTenantLogsHandler = operator_api.GetTenantLogsHandlerFunc(func(params operator_api.GetTenantLogsParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantLogsResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantLogsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantLogsOK().WithPayload(payload)
	})

	api.OperatorAPISetTenantLogsHandler = operator_api.SetTenantLogsHandlerFunc(func(params operator_api.SetTenantLogsParams, session *models.Principal) middleware.Responder {
		payload, err := setTenantLogsResponse(session, params)
		if err != nil {
			return operator_api.NewSetTenantLogsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewSetTenantLogsOK().WithPayload(payload)
	})

	api.OperatorAPIEnableTenantLoggingHandler = operator_api.EnableTenantLoggingHandlerFunc(func(params operator_api.EnableTenantLoggingParams, session *models.Principal) middleware.Responder {
		payload, err := enableTenantLoggingResponse(session, params)
		if err != nil {
			return operator_api.NewEnableTenantLoggingDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewEnableTenantLoggingOK().WithPayload(payload)
	})

	api.OperatorAPIDisableTenantLoggingHandler = operator_api.DisableTenantLoggingHandlerFunc(func(params operator_api.DisableTenantLoggingParams, session *models.Principal) middleware.Responder {
		payload, err := disableTenantLoggingResponse(session, params)
		if err != nil {
			return operator_api.NewDisableTenantLoggingDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDisableTenantLoggingOK().WithPayload(payload)
	})
}

// getTenantLogsResponse returns the Audit Log and Log DB configuration of a tenant
func getTenantLogsResponse(session *models.Principal, params operator_api.GetTenantLogsParams) (*models.TenantLogs, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := cluster.OperatorClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantLogs)
	}

	opClient := &operatorClient{
		client: opClientClientSet,
	}

	minTenant, err := getTenant(ctx, opClient, params.Namespace, params.Tenant)
	if err != nil {
		return nil, ErrorWithContext(ctx, err, ErrUnableToGetTenantLogs)
	}
	return getTenantLogsInfo(minTenant), nil
}

func getTenantLogsInfo(minTenant *miniov2.Tenant) *models.TenantLogs {
	if minTenant.Spec.Log == nil {
		return &models.TenantLogs{
			Disabled: true,
		}
	}
	annotations := []*models.Annotation{}
	for k, v := range minTenant.Spec.Log.Annotations {
		annotations = append(annotations, &models.Annotation{Key: k, Value: v})
	}
	labels := []*models.Label{}
	for k, v := range minTenant.Spec.Log.Labels {
		labels = append(labels, &models.Label{Key: k, Value: v})
	}
	nodeSelector := []*models.NodeSelector{}
	for k, v := range minTenant.Spec.Log.NodeSelector {
		nodeSelector = append(nodeSelector, &models.NodeSelector{Key: k, Value: v})
	}
	if minTenant.Spec.Log.Db == nil {
		minTenant.Spec.Log.Db = &miniov2.LogDbConfig{}
	}
	dbAnnotations := []*models.Annotation{}
	for k, v := range minTenant.Spec.Log.Db.Annotations {
		dbAnnotations = append(dbAnnotations, &models.Annotation{Key: k, Value: v})
	}
	dbLabels := []*models.Label{}
	for k, v := range minTenant.Spec.Log.Db.Labels {
		dbLabels = append(dbLabels, &models.Label{Key: k, Value: v})
	}
	dbNodeSelector := []*models.NodeSelector{}
	for k, v := range minTenant.Spec.Log.Db.NodeSelector {
		dbNodeSelector = append(dbNodeSelector, &models.NodeSelector{Key: k, Value: v})
	}
	var logSecurityContext *models.SecurityContext
	var logDBSecurityContext *models.SecurityContext

	if minTenant.Spec.Log.SecurityContext != nil {
		logSecurityContext = convertK8sSCToModelSC(minTenant.Spec.Log.SecurityContext)
	}
	if minTenant.Spec.Log.Db.SecurityContext != nil {
		logDBSecurityContext = convertK8sSCToModelSC(minTenant.Spec.Log.Db.SecurityContext)
	}

	if minTenant.Spec.Log.Audit == nil || minTenant.Spec.Log.Audit.DiskCapacityGB == nil {
		minTenant.Spec.Log.Audit = &miniov2.AuditConfig{DiskCapacityGB: swag.Int(0)}
	}

	tenantLoggingConfiguration := &models.TenantLogs{
		Image:                minTenant.Spec.Log.Image,
		DiskCapacityGB:       fmt.Sprintf("%d", *minTenant.Spec.Log.Audit.DiskCapacityGB),
		Annotations:          annotations,
		Labels:               labels,
		NodeSelector:         nodeSelector,
		ServiceAccountName:   minTenant.Spec.Log.ServiceAccountName,
		SecurityContext:      logSecurityContext,
		DbImage:              minTenant.Spec.Log.Db.Image,
		DbInitImage:          minTenant.Spec.Log.Db.InitImage,
		DbAnnotations:        dbAnnotations,
		DbLabels:             dbLabels,
		DbNodeSelector:       dbNodeSelector,
		DbServiceAccountName: minTenant.Spec.Log.Db.ServiceAccountName,
		DbSecurityContext:    logDBSecurityContext,
		Disabled:             false,
	}

	var requestedCPU string
	var requestedMem string
	var requestedDBCPU string
	var requestedDBMem string

	if minTenant.Spec.Log.Resources.Requests != nil {
		requestedCPUQ := minTenant.Spec.Log.Resources.Requests["cpu"]
		requestedCPU = strconv.FormatInt(requestedCPUQ.Value(), 10)
		requestedMemQ := minTenant.Spec.Log.Resources.Requests["memory"]
		requestedMem = strconv.FormatInt(requestedMemQ.Value(), 10)

		requestedDBCPUQ := minTenant.Spec.Log.Db.Resources.Requests["cpu"]
		requestedDBCPU = strconv.FormatInt(requestedDBCPUQ.Value(), 10)
		requestedDBMemQ := minTenant.Spec.Log.Db.Resources.Requests["memory"]
		requestedDBMem = strconv.FormatInt(requestedDBMemQ.Value(), 10)

		tenantLoggingConfiguration.LogCPURequest = requestedCPU
		tenantLoggingConfiguration.LogMemRequest = requestedMem
		tenantLoggingConfiguration.LogDBCPURequest = requestedDBCPU
		tenantLoggingConfiguration.LogDBMemRequest = requestedDBMem
	}
	return tenantLoggingConfiguration
}

// setTenantLogsResponse updates the Audit Log and Log DB configuration for the tenant
func setTenantLogsResponse(session *models.Principal, params operator_api.SetTenantLogsParams) (bool, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	opClientClientSet, err := cluster.OperatorClient(session.STSSessionToken)
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
	return setTenantLogs(ctx, minTenant, opClient, params)
}

func setTenantLogs(ctx context.Context, minTenant *miniov2.Tenant, opClient OperatorClientI, params operator_api.SetTenantLogsParams) (bool, *models.Error) {
	var err error
	labels := make(map[string]string)
	if params.Data.Labels != nil {
		for i := 0; i < len(params.Data.Labels); i++ {
			if params.Data.Labels[i] != nil {
				labels[params.Data.Labels[i].Key] = params.Data.Labels[i].Value
			}
		}
		minTenant.Spec.Log.Labels = labels
	}

	if params.Data.Annotations != nil {
		annotations := make(map[string]string)
		for i := 0; i < len(params.Data.Annotations); i++ {
			if params.Data.Annotations[i] != nil {
				annotations[params.Data.Annotations[i].Key] = params.Data.Annotations[i].Value
			}
		}
		minTenant.Spec.Log.Annotations = annotations
	}
	if params.Data.NodeSelector != nil {
		nodeSelector := make(map[string]string)
		for i := 0; i < len(params.Data.NodeSelector); i++ {
			if params.Data.NodeSelector[i] != nil {
				nodeSelector[params.Data.NodeSelector[i].Key] = params.Data.NodeSelector[i].Value
			}
		}
		minTenant.Spec.Log.NodeSelector = nodeSelector
	}
	logResourceRequest := make(corev1.ResourceList)
	if len(params.Data.LogCPURequest) > 0 {
		if reflect.TypeOf(params.Data.LogCPURequest).Kind() == reflect.String && params.Data.LogCPURequest != "0Gi" && params.Data.LogCPURequest != "" {
			cpuQuantity, err := resource.ParseQuantity(params.Data.LogCPURequest)
			if err != nil {
				return false, ErrorWithContext(ctx, err)
			}
			logResourceRequest["cpu"] = cpuQuantity
			minTenant.Spec.Log.Resources.Requests = logResourceRequest
		}
	}
	if len(params.Data.LogMemRequest) > 0 {
		if reflect.TypeOf(params.Data.LogMemRequest).Kind() == reflect.String && params.Data.LogMemRequest != "" {
			memQuantity, err := resource.ParseQuantity(params.Data.LogMemRequest)
			if err != nil {
				return false, ErrorWithContext(ctx, err)
			}

			logResourceRequest["memory"] = memQuantity
			minTenant.Spec.Log.Resources.Requests = logResourceRequest
		}
	}

	modified := false
	if minTenant.Spec.Log.Db != nil {
		modified = true
	}
	dbLabels := make(map[string]string)
	if params.Data.DbLabels != nil {
		for i := 0; i < len(params.Data.DbLabels); i++ {
			if params.Data.DbLabels[i] != nil {
				dbLabels[params.Data.DbLabels[i].Key] = params.Data.DbLabels[i].Value
			}
			modified = true
		}
	}
	dbAnnotations := make(map[string]string)
	if params.Data.DbAnnotations != nil {
		for i := 0; i < len(params.Data.DbAnnotations); i++ {
			if params.Data.DbAnnotations[i] != nil {
				dbAnnotations[params.Data.DbAnnotations[i].Key] = params.Data.DbAnnotations[i].Value
			}
			modified = true
		}
	}
	dbNodeSelector := make(map[string]string)
	if params.Data.DbNodeSelector != nil {
		for i := 0; i < len(params.Data.DbNodeSelector); i++ {
			if params.Data.DbNodeSelector[i] != nil {
				dbNodeSelector[params.Data.DbNodeSelector[i].Key] = params.Data.DbNodeSelector[i].Value
			}
			modified = true
		}
	}
	logDBResourceRequest := make(corev1.ResourceList)
	if len(params.Data.LogDBCPURequest) > 0 {
		if reflect.TypeOf(params.Data.LogDBCPURequest).Kind() == reflect.String && params.Data.LogDBCPURequest != "0Gi" && params.Data.LogDBCPURequest != "" {
			dbCPUQuantity, err := resource.ParseQuantity(params.Data.LogDBCPURequest)
			if err != nil {
				return false, ErrorWithContext(ctx, err)
			}
			logDBResourceRequest["cpu"] = dbCPUQuantity
			minTenant.Spec.Log.Db.Resources.Requests = logDBResourceRequest
		}
	}
	if len(params.Data.LogDBMemRequest) > 0 {
		if reflect.TypeOf(params.Data.LogDBMemRequest).Kind() == reflect.String && params.Data.LogDBMemRequest != "" {
			dbMemQuantity, err := resource.ParseQuantity(params.Data.LogDBMemRequest)
			if err != nil {
				return false, ErrorWithContext(ctx, err)
			}
			logDBResourceRequest["memory"] = dbMemQuantity
			minTenant.Spec.Log.Db.Resources.Requests = logDBResourceRequest
		}
	}
	if len(params.Data.Image) > 0 {
		minTenant.Spec.Log.Image = params.Data.Image
	}
	if params.Data.SecurityContext != nil {
		minTenant.Spec.Log.SecurityContext, err = convertModelSCToK8sSC(params.Data.SecurityContext)
		if err != nil {
			return false, ErrorWithContext(ctx, err)
		}
	}
	if len(params.Data.DiskCapacityGB) > 0 {
		diskCapacityGB, err := strconv.Atoi(params.Data.DiskCapacityGB)
		if err == nil {
			if minTenant.Spec.Log.Audit != nil && minTenant.Spec.Log.Audit.DiskCapacityGB != nil {
				*minTenant.Spec.Log.Audit.DiskCapacityGB = diskCapacityGB
			} else {
				minTenant.Spec.Log.Audit = &miniov2.AuditConfig{DiskCapacityGB: swag.Int(diskCapacityGB)}
			}
		}
	}
	if len(params.Data.ServiceAccountName) > 0 {
		minTenant.Spec.Log.ServiceAccountName = params.Data.ServiceAccountName
	}
	if params.Data.DbLabels != nil {

		if params.Data.DbImage != "" || params.Data.DbServiceAccountName != "" {
			modified = true
		}
		if modified {
			if minTenant.Spec.Log.Db == nil {
				// Default class name for Log search
				diskSpaceFromAPI := int64(5) * humanize.GiByte // Default is 5Gi
				logSearchStorageClass := "standard"

				logSearchDiskSpace := resource.NewQuantity(diskSpaceFromAPI, resource.DecimalExponent)
				resources := corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: *logSearchDiskSpace,
					},
				}
				minTenant.Spec.Log.Db = &miniov2.LogDbConfig{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							Name: params.Tenant + "-log",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources:        resources,
							StorageClassName: &logSearchStorageClass,
						},
					},
					Labels:             dbLabels,
					Annotations:        dbAnnotations,
					NodeSelector:       dbNodeSelector,
					Image:              params.Data.DbImage,
					ServiceAccountName: params.Data.DbServiceAccountName,
					Resources: corev1.ResourceRequirements{
						Requests: resources.Requests,
					},
				}
			} else {
				minTenant.Spec.Log.Db.Labels = dbLabels
				minTenant.Spec.Log.Db.Annotations = dbAnnotations
				minTenant.Spec.Log.Db.NodeSelector = dbNodeSelector
				minTenant.Spec.Log.Db.Image = params.Data.DbImage
				minTenant.Spec.Log.Db.InitImage = params.Data.DbInitImage
				minTenant.Spec.Log.Db.ServiceAccountName = params.Data.DbServiceAccountName
				minTenant.Spec.Log.Db.SecurityContext, err = convertModelSCToK8sSC(params.Data.DbSecurityContext)
				if err != nil {
					return false, ErrorWithContext(ctx, err)
				}
			}
		}
	}

	_, err = opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	return true, nil
}

// enableTenantLoggingResponse enables Tenant Logging
func enableTenantLoggingResponse(session *models.Principal, params operator_api.EnableTenantLoggingParams) (bool, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := cluster.OperatorClient(session.STSSessionToken)
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
	return enableTenantLogging(ctx, minTenant, opClient, params.Tenant)
}

func enableTenantLogging(ctx context.Context, minTenant *miniov2.Tenant, opClient OperatorClientI, tenantName string) (bool, *models.Error) {
	minTenant.EnsureDefaults()
	// Default class name for Log search
	diskSpaceFromAPI := int64(5) * humanize.GiByte // Default is 5Gi
	logSearchStorageClass := "standard"

	logSearchDiskSpace := resource.NewQuantity(diskSpaceFromAPI, resource.DecimalExponent)

	auditMaxCap := 10
	if (diskSpaceFromAPI / humanize.GiByte) < int64(auditMaxCap) {
		auditMaxCap = int(diskSpaceFromAPI / humanize.GiByte)
	}

	minTenant.Spec.Log = &miniov2.LogConfig{
		Audit: &miniov2.AuditConfig{DiskCapacityGB: swag.Int(auditMaxCap)},
		Db: &miniov2.LogDbConfig{
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: tenantName + "-log",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: *logSearchDiskSpace,
						},
					},
					StorageClassName: &logSearchStorageClass,
				},
			},
		},
	}

	_, err := opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	return true, nil
}

// disableTenantLoggingResponse disables Tenant Logging
func disableTenantLoggingResponse(session *models.Principal, params operator_api.DisableTenantLoggingParams) (bool, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()

	opClientClientSet, err := cluster.OperatorClient(session.STSSessionToken)
	if err != nil {
		return false, ErrorWithContext(ctx, err, ErrUnableToGetTenantUsage)
	}
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
	minTenant.EnsureDefaults()
	minTenant.Spec.Log = nil

	_, err = opClient.TenantUpdate(ctx, minTenant, metav1.UpdateOptions{})
	if err != nil {
		return false, ErrorWithContext(ctx, err)
	}
	return true, nil
}
