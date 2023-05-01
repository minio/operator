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
	"net/http"
	"time"

	"github.com/go-openapi/swag"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

func (suite *TenantTestSuite) TestDeletePodHandlerWithoutError() {
	params, api := suite.initDeletePodRequest()
	response := api.OperatorAPIDeletePodHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.DeletePodDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initDeletePodRequest() (params operator_api.DeletePodParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-tenantmock-pod"
	return params, api
}

func (suite *TenantTestSuite) TestGetTenantPodsHandlerWithError() {
	params, api := suite.initGetTenantPodsRequest()
	response := api.OperatorAPIGetTenantPodsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetTenantPodsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetTenantPodsRequest() (params operator_api.GetTenantPodsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}

func (suite *TenantTestSuite) TestGetTenantPodsWithoutError() {
	pods := getTenantPods(&corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{{}},
				},
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
		},
	})
	suite.assert.Equal(1, len(pods))
}

func (suite *TenantTestSuite) TestGetPodLogsHandlerWithError() {
	params, api := suite.initGetPodLogsRequest()
	response := api.OperatorAPIGetPodLogsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetPodLogsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetPodLogsRequest() (params operator_api.GetPodLogsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mokc-pod"
	return params, api
}

func (suite *TenantTestSuite) TestGetPodEventsHandlerWithError() {
	params, api := suite.initGetPodEventsRequest()
	response := api.OperatorAPIGetPodEventsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.GetPodEventsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initGetPodEventsRequest() (params operator_api.GetPodEventsParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-pod"
	return params, api
}

func (suite *TenantTestSuite) TestDescribePodHandlerWithError() {
	params, api := suite.initDescribePodRequest()
	response := api.OperatorAPIDescribePodHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.DescribePodDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initDescribePodRequest() (params operator_api.DescribePodParams, api operations.OperatorAPI) {
	registerPodHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.PodName = "mock-pod"
	return params, api
}

func (suite *TenantTestSuite) TestGetDescribePodBuildsResponseFromPodInfo() {
	mockTime := time.Date(2023, 4, 25, 14, 30, 45, 100, time.UTC)
	mockContainerOne := corev1.Container{
		Name:  "c1",
		Image: "c1-image",
		Ports: []corev1.ContainerPort{
			{
				Name:          "c1-port-1",
				HostPort:      int32(9000),
				ContainerPort: int32(8080),
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "c1-port-2",
				HostPort:      int32(3000),
				ContainerPort: int32(7070),
				Protocol:      corev1.ProtocolUDP,
			},
		},
		Args: []string{"c1-arg1", "c1-arg2"},
		Env: []corev1.EnvVar{
			{
				Name:  "c1-env-var1",
				Value: "c1-env-value1",
			},
			{
				Name:  "c1-env-var2",
				Value: "c1-env-value2",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "c1-mount1",
				MountPath: "c1-mount1-path",
				ReadOnly:  true,
				SubPath:   "c1-mount1-subpath",
			},
			{
				Name:      "c1-mount2",
				MountPath: "c1-mount2-path",
				ReadOnly:  true,
				SubPath:   "c1-mount2-subpath",
			},
		},
	}
	mockContainerTwo := corev1.Container{
		Name:  "c2",
		Image: "c2-image",
		Ports: []corev1.ContainerPort{
			{
				Name:          "c2-port-1",
				HostPort:      int32(9000),
				ContainerPort: int32(8080),
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "c2-port-2",
				HostPort:      int32(3000),
				ContainerPort: int32(7070),
				Protocol:      corev1.ProtocolUDP,
			},
		},
		Args: []string{"c2-arg1", "c2-arg2"},
		Env: []corev1.EnvVar{
			{
				Name:  "c2-env-var1",
				Value: "c2-env-value1",
			},
			{
				Name:  "c2-env-var2",
				Value: "c2-env-value2",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "c2-mount1",
				MountPath: "c2-mount1-path",
				ReadOnly:  true,
				SubPath:   "c2-mount1-subpath",
			},
			{
				Name:      "c2-mount2",
				MountPath: "c2-mount2-path",
				ReadOnly:  true,
				SubPath:   "c2-mount2-subpath",
			},
		},
	}
	mockPodInfo := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "mock-pod",
			Namespace:                  "mock-namespace",
			Labels:                     map[string]string{"Key1": "Val1", "Key2": "Val2"},
			Annotations:                map[string]string{"Annotation1": "Annotation1Val1", "Annotation2": "Annotation1Val2"},
			DeletionTimestamp:          &metav1.Time{Time: mockTime},
			DeletionGracePeriodSeconds: swag.Int64(60),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "ReferenceKind",
					Name:       "ReferenceName",
					Controller: swag.Bool(true),
				},
			},
		},
		Spec: corev1.PodSpec{
			PriorityClassName: "mock-priority-class",
			NodeName:          "mock-node",
			Priority:          swag.Int32(10),
			Containers:        []corev1.Container{mockContainerOne, mockContainerTwo},
			Volumes: []corev1.Volume{
				{
					Name: "v1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "v1-pvc",
							ReadOnly:  true,
						},
					},
				},
				{
					Name: "v1",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "v1-vs-secret1",
										},
									},
									DownwardAPI: &corev1.DownwardAPIProjection{
										Items: []corev1.DownwardAPIVolumeFile{},
									},
									ConfigMap: &corev1.ConfigMapProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "v1-vs-configmap1",
										},
									},
									ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
										ExpirationSeconds: swag.Int64(1000),
									},
								},
							},
							DefaultMode: swag.Int32(511),
						},
					},
				},
			},
			NodeSelector: map[string]string{
				"p1-ns-key1": "p1-ns-val1",
				"p1-ns-key2": "p1-ns-val2",
			},
			Tolerations: []corev1.Toleration{
				{
					Key:               "p1-t1-key",
					Operator:          corev1.TolerationOpExists,
					Value:             "p1-t1-val",
					Effect:            corev1.TaintEffectNoSchedule,
					TolerationSeconds: swag.Int64(60),
				},
				{
					Key:               "p1-t2-key",
					Operator:          corev1.TolerationOpEqual,
					Value:             "p1-t2-val",
					Effect:            corev1.TaintEffectPreferNoSchedule,
					TolerationSeconds: swag.Int64(70),
				},
			},
		},
		Status: corev1.PodStatus{
			StartTime: &metav1.Time{Time: mockTime},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "c1",
					ContainerID:  "c1-id",
					ImageID:      "c1-image-id",
					Ready:        true,
					RestartCount: int32(3),
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{Time: mockTime},
						},
					},
					LastTerminationState: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "c1-some-reason",
							Message: "c1-some-message",
						},
					},
				},
				{
					Name:         "c2",
					ContainerID:  "c2-id",
					ImageID:      "c2-image-id",
					Ready:        true,
					RestartCount: int32(3),
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{Time: mockTime},
						},
					},
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason:      "c2-some-reason",
							Message:     "c2-some-message",
							ExitCode:    int32(4),
							Signal:      int32(1),
							StartedAt:   metav1.Time{Time: mockTime},
							FinishedAt:  metav1.Time{Time: mockTime},
							ContainerID: "c2-id",
						},
					},
				},
			},
			Phase:   corev1.PodPhase("phase"),
			Reason:  "StatusReason",
			Message: "StatusMessage",
			PodIP:   "192.1.2.3",
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.ContainersReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.PodInitialized,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}

	res, err := getDescribePod(mockPodInfo)

	suite.assert.NotNil(res)
	suite.assert.Nil(err)
	suite.assert.Equal(mockPodInfo.Name, res.Name)
	suite.assert.Equal(mockPodInfo.Namespace, res.Namespace)
	suite.assert.Equal(mockPodInfo.Spec.PriorityClassName, res.PriorityClassName)
	suite.assert.Equal(mockPodInfo.Spec.NodeName, res.NodeName)
	suite.assert.Equal(int64(10), res.Priority)
	suite.assert.Equal(mockPodInfo.Status.StartTime.Time.String(), res.StartTime)
	suite.assert.Contains(res.Labels, &models.Label{Key: "Key1", Value: "Val1"})
	suite.assert.Contains(res.Labels, &models.Label{Key: "Key2", Value: "Val2"})
	suite.assert.Contains(res.Annotations, &models.Annotation{Key: "Annotation1", Value: "Annotation1Val1"})
	suite.assert.Contains(res.Annotations, &models.Annotation{Key: "Annotation2", Value: "Annotation1Val2"})
	suite.assert.Equal(duration.HumanDuration(time.Since(mockTime)), res.DeletionTimestamp)
	suite.assert.Equal(int64(60), res.DeletionGracePeriodSeconds)
	suite.assert.Equal("phase", res.Phase)
	suite.assert.Equal("StatusReason", res.Reason)
	suite.assert.Equal("StatusMessage", res.Message)
	suite.assert.Equal("192.1.2.3", res.PodIP)
	suite.assert.Equal("&OwnerReference{Kind:ReferenceKind,Name:ReferenceName,UID:,APIVersion:v1,Controller:*true,BlockOwnerDeletion:nil,}", res.ControllerRef)
	suite.assert.Equal([]*models.Container{
		{
			Name:         "c1",
			Image:        "c1-image",
			Ports:        []string{"8080/TCP", "7070/UDP"},
			HostPorts:    []string{"9000/TCP", "3000/UDP"},
			Args:         []string{"c1-arg1", "c1-arg2"},
			ContainerID:  "c1-id",
			ImageID:      "c1-image-id",
			Ready:        true,
			RestartCount: int64(3),
			State: &models.State{
				ExitCode: int64(0),
				Finished: "",
				Message:  "",
				Reason:   "",
				Signal:   int64(0),
				Started:  "Tue, 25 Apr 2023 14:30:45 +0000",
				State:    "Running",
			},
			LastState: &models.State{
				ExitCode: int64(0),
				Finished: "",
				Message:  "c1-some-message",
				Reason:   "c1-some-reason",
				Signal:   int64(0),
				Started:  "",
				State:    "Waiting",
			},
			EnvironmentVariables: []*models.EnvironmentVariable{
				{
					Key:   "c1-env-var1",
					Value: "c1-env-value1",
				},
				{
					Key:   "c1-env-var2",
					Value: "c1-env-value2",
				},
			},
			Mounts: []*models.Mount{
				{
					Name:      "c1-mount1",
					MountPath: "c1-mount1-path",
					ReadOnly:  true,
					SubPath:   "c1-mount1-subpath",
				},
				{
					Name:      "c1-mount2",
					MountPath: "c1-mount2-path",
					ReadOnly:  true,
					SubPath:   "c1-mount2-subpath",
				},
			},
		},
		{
			Name:         "c2",
			Image:        "c2-image",
			Ports:        []string{"8080/TCP", "7070/UDP"},
			HostPorts:    []string{"9000/TCP", "3000/UDP"},
			Args:         []string{"c2-arg1", "c2-arg2"},
			ContainerID:  "c2-id",
			ImageID:      "c2-image-id",
			Ready:        true,
			RestartCount: int64(3),
			State: &models.State{
				ExitCode: int64(0),
				Finished: "",
				Message:  "",
				Reason:   "",
				Signal:   int64(0),
				Started:  "Tue, 25 Apr 2023 14:30:45 +0000",
				State:    "Running",
			},
			LastState: &models.State{
				ExitCode: int64(4),
				Finished: "Tue, 25 Apr 2023 14:30:45 +0000",
				Message:  "c2-some-message",
				Reason:   "c2-some-reason",
				Signal:   int64(1),
				Started:  "Tue, 25 Apr 2023 14:30:45 +0000",
				State:    "Terminated",
			},
			EnvironmentVariables: []*models.EnvironmentVariable{
				{
					Key:   "c2-env-var1",
					Value: "c2-env-value1",
				},
				{
					Key:   "c2-env-var2",
					Value: "c2-env-value2",
				},
			},
			Mounts: []*models.Mount{
				{
					Name:      "c2-mount1",
					MountPath: "c2-mount1-path",
					ReadOnly:  true,
					SubPath:   "c2-mount1-subpath",
				},
				{
					Name:      "c2-mount2",
					MountPath: "c2-mount2-path",
					ReadOnly:  true,
					SubPath:   "c2-mount2-subpath",
				},
			},
		},
	}, res.Containers)
	suite.assert.Equal([]*models.Condition{
		{
			Type:   "ContainersReady",
			Status: "True",
		},
		{
			Type:   "Initialized",
			Status: "False",
		},
	}, res.Conditions)
	suite.assert.Equal([]*models.Volume{
		{
			Name: "v1",
			Pvc: &models.Pvc{
				ClaimName: "v1-pvc",
				ReadOnly:  true,
			},
		},
		{
			Name: "v1",
			Projected: &models.ProjectedVolume{
				Sources: []*models.ProjectedVolumeSource{
					{
						Secret: &models.Secret{
							Name:     "v1-vs-secret1",
							Optional: false,
						},
						DownwardAPI: true,
						ConfigMap: &models.ConfigMap{
							Name:     "v1-vs-configmap1",
							Optional: false,
						},
						ServiceAccountToken: &models.ServiceAccountToken{
							ExpirationSeconds: int64(1000),
						},
					},
				},
			},
		},
	}, res.Volumes)
	suite.assert.Equal("BestEffort", res.QosClass)
	suite.assert.Contains(res.NodeSelector, &models.NodeSelector{
		Key:   "p1-ns-key1",
		Value: "p1-ns-val1",
	})
	suite.assert.Contains(res.NodeSelector, &models.NodeSelector{
		Key:   "p1-ns-key2",
		Value: "p1-ns-val2",
	})
	suite.assert.Equal([]*models.Toleration{
		{
			Key:               "p1-t1-key",
			Operator:          "Exists",
			Value:             "p1-t1-val",
			Effect:            "NoSchedule",
			TolerationSeconds: int64(60),
		},
		{
			Key:               "p1-t2-key",
			Operator:          "Equal",
			Value:             "p1-t2-val",
			Effect:            "PreferNoSchedule",
			TolerationSeconds: int64(70),
		},
	}, res.Tolerations)
}
