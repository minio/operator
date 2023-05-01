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
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/strings/slices"
)

func registerPodHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetTenantPodsHandler = operator_api.GetTenantPodsHandlerFunc(func(params operator_api.GetTenantPodsParams, session *models.Principal) middleware.Responder {
		payload, err := getTenantPodsResponse(session, params)
		if err != nil {
			return operator_api.NewGetTenantPodsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantPodsOK().WithPayload(payload)
	})

	api.OperatorAPIGetPodLogsHandler = operator_api.GetPodLogsHandlerFunc(func(params operator_api.GetPodLogsParams, session *models.Principal) middleware.Responder {
		payload, err := getPodLogsResponse(session, params)
		if err != nil {
			return operator_api.NewGetPodLogsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetPodLogsOK().WithPayload(payload)
	})

	api.OperatorAPIGetPodEventsHandler = operator_api.GetPodEventsHandlerFunc(func(params operator_api.GetPodEventsParams, session *models.Principal) middleware.Responder {
		payload, err := getPodEventsResponse(session, params)
		if err != nil {
			return operator_api.NewGetPodEventsDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetPodEventsOK().WithPayload(payload)
	})

	api.OperatorAPIDescribePodHandler = operator_api.DescribePodHandlerFunc(func(params operator_api.DescribePodParams, session *models.Principal) middleware.Responder {
		payload, err := getDescribePodResponse(session, params)
		if err != nil {
			return operator_api.NewDescribePodDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDescribePodOK().WithPayload(payload)
	})
	// Delete Pod
	api.OperatorAPIDeletePodHandler = operator_api.DeletePodHandlerFunc(func(params operator_api.DeletePodParams, session *models.Principal) middleware.Responder {
		err := getDeletePodResponse(session, params)
		if err != nil {
			return operator_api.NewDeletePodDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDeletePodNoContent()
	})
	// Tenant log report
	api.OperatorAPIGetTenantLogReportHandler = operator_api.GetTenantLogReportHandlerFunc(func(params operator_api.GetTenantLogReportParams, principal *models.Principal) middleware.Responder {
		payload, err := getTenantLogReportResponse(principal, params)
		if err != nil {
			return operator_api.NewGetTenantLogReportDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetTenantLogReportOK().WithPayload(payload)
	})
}

func getPodLogsResponse(session *models.Principal, params operator_api.GetPodLogsParams) (string, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	listOpts := &corev1.PodLogOptions{Container: "minio"}
	logs := clientset.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, listOpts)

	buffLogs, err := logs.Stream(ctx)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}

	defer buffLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, buffLogs)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	return buf.String(), nil
}

func getTenantPodsResponse(session *models.Principal, params operator_api.GetTenantPodsParams) ([]*models.TenantPod, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", miniov2.TenantLabel, params.Tenant),
	}
	pods, err := clientset.CoreV1().Pods(params.Namespace).List(ctx, listOpts)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return getTenantPods(pods), nil
}

func getTenantPods(pods *corev1.PodList) []*models.TenantPod {
	retval := []*models.TenantPod{}
	for _, pod := range pods.Items {
		var restarts int64
		if len(pod.Status.ContainerStatuses) > 0 {
			restarts = int64(pod.Status.ContainerStatuses[0].RestartCount)
		}
		status := string(pod.Status.Phase)
		if pod.DeletionTimestamp != nil {
			status = "Terminating"
		}
		retval = append(retval, &models.TenantPod{
			Name:        swag.String(pod.Name),
			Status:      status,
			TimeCreated: pod.CreationTimestamp.Unix(),
			PodIP:       pod.Status.PodIP,
			Restarts:    restarts,
			Node:        pod.Spec.NodeName,
		})
	}
	return retval
}

// getDeletePodResponse gets the output of deleting a minio instance
func getDeletePodResponse(session *models.Principal, params operator_api.DeletePodParams) *models.Error {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	// get Kubernetes Client
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return ErrorWithContext(ctx, err)
	}
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("v1.min.io/tenant=%s", params.Tenant),
		FieldSelector: fmt.Sprintf("metadata.name=%s%s", params.Tenant, params.PodName[len(params.Tenant):]),
	}
	if err = clientset.CoreV1().Pods(params.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOpts); err != nil {
		return ErrorWithContext(ctx, err)
	}
	return nil
}

func getPodEventsResponse(session *models.Principal, params operator_api.GetPodEventsParams) (models.EventListWrapper, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	pod, err := clientset.CoreV1().Pods(params.Namespace).Get(ctx, params.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	events, err := clientset.CoreV1().Events(params.Namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pod.UID)})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	retval := models.EventListWrapper{}
	for i := 0; i < len(events.Items); i++ {
		retval = append(retval, &models.EventListElement{
			Namespace: events.Items[i].Namespace,
			LastSeen:  events.Items[i].LastTimestamp.Unix(),
			Message:   events.Items[i].Message,
			EventType: events.Items[i].Type,
			Reason:    events.Items[i].Reason,
		})
	}
	sort.SliceStable(retval, func(i int, j int) bool {
		return retval[i].LastSeen < retval[j].LastSeen
	})
	return retval, nil
}

func getDescribePodResponse(session *models.Principal, params operator_api.DescribePodParams) (*models.DescribePodWrapper, *models.Error) {
	ctx := context.Background()
	clientset, err := K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	pod, err := clientset.CoreV1().Pods(params.Namespace).Get(ctx, params.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return getDescribePod(pod)
}

func getDescribePod(pod *corev1.Pod) (*models.DescribePodWrapper, *models.Error) {
	retval := &models.DescribePodWrapper{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		PriorityClassName: pod.Spec.PriorityClassName,
		NodeName:          pod.Spec.NodeName,
	}
	if pod.Spec.Priority != nil {
		retval.Priority = int64(*pod.Spec.Priority)
	}
	if pod.Status.StartTime != nil {
		retval.StartTime = pod.Status.StartTime.Time.String()
	}
	labelArray := make([]*models.Label, len(pod.Labels))
	i := 0
	for key := range pod.Labels {
		labelArray[i] = &models.Label{Key: key, Value: pod.Labels[key]}
		i++
	}
	retval.Labels = labelArray
	annotationArray := make([]*models.Annotation, len(pod.Annotations))
	i = 0
	for key := range pod.Annotations {
		annotationArray[i] = &models.Annotation{Key: key, Value: pod.Annotations[key]}
		i++
	}
	retval.Annotations = annotationArray
	if pod.DeletionTimestamp != nil {
		retval.DeletionTimestamp = translateTimestampSince(*pod.DeletionTimestamp)
	}
	if pod.DeletionGracePeriodSeconds != nil {
		retval.DeletionGracePeriodSeconds = *pod.DeletionGracePeriodSeconds
	}
	retval.Phase = string(pod.Status.Phase)
	retval.Reason = pod.Status.Reason
	retval.Message = pod.Status.Message
	retval.PodIP = pod.Status.PodIP
	retval.ControllerRef = metav1.GetControllerOf(pod).String()
	retval.Containers = make([]*models.Container, len(pod.Spec.Containers))
	statusMap := map[string]corev1.ContainerStatus{}
	statusKeys := make([]string, len(pod.Status.ContainerStatuses))
	for i, status := range pod.Status.ContainerStatuses {
		statusMap[status.Name] = status
		statusKeys[i] = status.Name

	}
	for i := range pod.Spec.Containers {
		container := pod.Spec.Containers[i]
		retval.Containers[i] = &models.Container{
			Name:      container.Name,
			Image:     container.Image,
			Ports:     describeContainerPorts(container.Ports),
			HostPorts: describeContainerHostPorts(container.Ports),
			Args:      container.Args,
		}
		if slices.Contains(statusKeys, container.Name) {
			containerStatus := statusMap[container.Name]
			retval.Containers[i].ContainerID = containerStatus.ContainerID
			retval.Containers[i].ImageID = containerStatus.ImageID
			retval.Containers[i].Ready = containerStatus.Ready
			retval.Containers[i].RestartCount = int64(containerStatus.RestartCount)
			retval.Containers[i].State = describeContainerState(containerStatus.State)
			retval.Containers[i].LastState = describeContainerState(containerStatus.LastTerminationState)
		}
		retval.Containers[i].EnvironmentVariables = make([]*models.EnvironmentVariable, len(container.Env))
		for j := range container.Env {
			retval.Containers[i].EnvironmentVariables[j] = &models.EnvironmentVariable{
				Key:   container.Env[j].Name,
				Value: container.Env[j].Value,
			}
		}
		retval.Containers[i].Mounts = make([]*models.Mount, len(container.VolumeMounts))
		for j := range container.VolumeMounts {
			retval.Containers[i].Mounts[j] = &models.Mount{
				Name:      container.VolumeMounts[j].Name,
				MountPath: container.VolumeMounts[j].MountPath,
				SubPath:   container.VolumeMounts[j].SubPath,
				ReadOnly:  container.VolumeMounts[j].ReadOnly,
			}
		}
	}
	retval.Conditions = make([]*models.Condition, len(pod.Status.Conditions))
	for i := range pod.Status.Conditions {
		retval.Conditions[i] = &models.Condition{
			Type:   string(pod.Status.Conditions[i].Type),
			Status: string(pod.Status.Conditions[i].Status),
		}
	}
	retval.Volumes = make([]*models.Volume, len(pod.Spec.Volumes))
	for i := range pod.Spec.Volumes {
		retval.Volumes[i] = &models.Volume{
			Name: pod.Spec.Volumes[i].Name,
		}
		if pod.Spec.Volumes[i].PersistentVolumeClaim != nil {
			retval.Volumes[i].Pvc = &models.Pvc{
				ReadOnly:  pod.Spec.Volumes[i].PersistentVolumeClaim.ReadOnly,
				ClaimName: pod.Spec.Volumes[i].PersistentVolumeClaim.ClaimName,
			}
		} else if pod.Spec.Volumes[i].Projected != nil {
			retval.Volumes[i].Projected = &models.ProjectedVolume{}
			retval.Volumes[i].Projected.Sources = make([]*models.ProjectedVolumeSource, len(pod.Spec.Volumes[i].Projected.Sources))
			for j := range pod.Spec.Volumes[i].Projected.Sources {
				retval.Volumes[i].Projected.Sources[j] = &models.ProjectedVolumeSource{}
				if pod.Spec.Volumes[i].Projected.Sources[j].Secret != nil {
					retval.Volumes[i].Projected.Sources[j].Secret = &models.Secret{
						Name:     pod.Spec.Volumes[i].Projected.Sources[j].Secret.Name,
						Optional: pod.Spec.Volumes[i].Projected.Sources[j].Secret.Optional != nil,
					}
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].DownwardAPI != nil {
					retval.Volumes[i].Projected.Sources[j].DownwardAPI = true
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap != nil {
					retval.Volumes[i].Projected.Sources[j].ConfigMap = &models.ConfigMap{
						Name:     pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap.Name,
						Optional: pod.Spec.Volumes[i].Projected.Sources[j].ConfigMap.Optional != nil,
					}
				}
				if pod.Spec.Volumes[i].Projected.Sources[j].ServiceAccountToken != nil {
					if pod.Spec.Volumes[i].Projected.Sources[j].ServiceAccountToken.ExpirationSeconds != nil {
						retval.Volumes[i].Projected.Sources[j].ServiceAccountToken = &models.ServiceAccountToken{
							ExpirationSeconds: *pod.Spec.Volumes[i].Projected.Sources[j].ServiceAccountToken.ExpirationSeconds,
						}
					}
				}
			}
		}
	}
	retval.QosClass = string(getPodQOS(pod))
	nodeSelectorArray := make([]*models.NodeSelector, len(pod.Spec.NodeSelector))
	i = 0
	for key := range pod.Spec.NodeSelector {
		nodeSelectorArray[i] = &models.NodeSelector{Key: key, Value: pod.Spec.NodeSelector[key]}
		i++
	}
	retval.NodeSelector = nodeSelectorArray
	retval.Tolerations = make([]*models.Toleration, len(pod.Spec.Tolerations))
	for i := range pod.Spec.Tolerations {
		retval.Tolerations[i] = &models.Toleration{
			Effect:            string(pod.Spec.Tolerations[i].Effect),
			Key:               pod.Spec.Tolerations[i].Key,
			Value:             pod.Spec.Tolerations[i].Value,
			Operator:          string(pod.Spec.Tolerations[i].Operator),
			TolerationSeconds: *pod.Spec.Tolerations[i].TolerationSeconds,
		}
	}
	return retval, nil
}

func describeContainerState(status corev1.ContainerState) *models.State {
	retval := &models.State{}
	switch {
	case status.Running != nil:
		retval.State = "Running"
		retval.Started = status.Running.StartedAt.Time.Format(time.RFC1123Z)
	case status.Waiting != nil:
		retval.State = "Waiting"
		retval.Reason = status.Waiting.Reason
		retval.Message = status.Waiting.Message
	case status.Terminated != nil:
		retval.State = "Terminated"
		retval.Message = status.Terminated.Message
		retval.Reason = status.Terminated.Reason
		retval.ExitCode = int64(status.Terminated.ExitCode)
		retval.Signal = int64(status.Terminated.Signal)
		retval.Started = status.Terminated.StartedAt.Time.Format(time.RFC1123Z)
		retval.Finished = status.Terminated.FinishedAt.Time.Format(time.RFC1123Z)
	default:
		retval.State = "Waiting"
	}
	return retval
}

func describeContainerPorts(cPorts []corev1.ContainerPort) []string {
	ports := make([]string, 0, len(cPorts))
	for _, cPort := range cPorts {
		ports = append(ports, fmt.Sprintf("%d/%s", cPort.ContainerPort, cPort.Protocol))
	}
	return ports
}

func describeContainerHostPorts(cPorts []corev1.ContainerPort) []string {
	ports := make([]string, 0, len(cPorts))
	for _, cPort := range cPorts {
		ports = append(ports, fmt.Sprintf("%d/%s", cPort.HostPort, cPort.Protocol))
	}
	return ports
}

// getPodQOS gets Pod's Quality of Service Class
func getPodQOS(pod *corev1.Pod) corev1.PodQOSClass {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}
	zeroQuantity := resource.MustParse("0")
	isGuaranteed := true
	allContainers := []corev1.Container{}
	allContainers = append(allContainers, pod.Spec.Containers...)
	allContainers = append(allContainers, pod.Spec.InitContainers...)
	for _, container := range allContainers {
		// process requests
		for name, quantity := range container.Resources.Requests {
			if !isSupportedQoSComputeResource(name) {
				continue
			}
			if quantity.Cmp(zeroQuantity) == 1 {
				delta := quantity.DeepCopy()
				if _, exists := requests[name]; !exists {
					requests[name] = delta
				} else {
					delta.Add(requests[name])
					requests[name] = delta
				}
			}
		}
		// process limits
		qosLimitsFound := sets.NewString()
		for name, quantity := range container.Resources.Limits {
			if !isSupportedQoSComputeResource(name) {
				continue
			}
			if quantity.Cmp(zeroQuantity) == 1 {
				qosLimitsFound.Insert(string(name))
				delta := quantity.DeepCopy()
				if _, exists := limits[name]; !exists {
					limits[name] = delta
				} else {
					delta.Add(limits[name])
					limits[name] = delta
				}
			}
		}

		if !qosLimitsFound.HasAll(string(corev1.ResourceMemory), string(corev1.ResourceCPU)) {
			isGuaranteed = false
		}
	}
	if len(requests) == 0 && len(limits) == 0 {
		return corev1.PodQOSBestEffort
	}
	// Check is requests match limits for all resources.
	if isGuaranteed {
		for name, req := range requests {
			if lim, exists := limits[name]; !exists || lim.Cmp(req) != 0 {
				isGuaranteed = false
				break
			}
		}
	}
	if isGuaranteed &&
		len(requests) == len(limits) {
		return corev1.PodQOSGuaranteed
	}
	return corev1.PodQOSBurstable
}

var supportedQoSComputeResources = sets.NewString(string(corev1.ResourceCPU), string(corev1.ResourceMemory))

func isSupportedQoSComputeResource(name corev1.ResourceName) bool {
	return supportedQoSComputeResources.Has(string(name))
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

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
