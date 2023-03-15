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
	"errors"
	"net/http"

	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/models"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func (suite *TenantTestSuite) TestTenantAddPoolHandlerWithError() {
	params, api := suite.initTenantAddPoolRequest()
	response := api.OperatorAPITenantAddPoolHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.TenantAddPoolDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initTenantAddPoolRequest() (params operator_api.TenantAddPoolParams, api operations.OperatorAPI) {
	registerPoolHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	return params, api
}

func (suite *TenantTestSuite) TestTenantUpdatePoolsHandlerWithError() {
	params, api := suite.initTenantUpdatePoolsRequest()
	response := api.OperatorAPITenantUpdatePoolsHandler.Handle(params, &models.Principal{})
	_, ok := response.(*operator_api.TenantUpdatePoolsDefault)
	suite.assert.True(ok)
}

func (suite *TenantTestSuite) initTenantUpdatePoolsRequest() (params operator_api.TenantUpdatePoolsParams, api operations.OperatorAPI) {
	registerPoolHandlers(&api)
	params.HTTPRequest = &http.Request{}
	params.Namespace = "mock-namespace"
	params.Tenant = "mock-tenant"
	params.Body = &models.PoolUpdateRequest{
		Pools: []*models.Pool{},
	}
	return params, api
}

func (suite *TenantTestSuite) TestUpdateTenantPoolsWithPoolError() {
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{}, nil
	}
	_, err := updateTenantPools(context.Background(), suite.opClient, "mock-namespace", "mock-tenant", []*models.Pool{{}})
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestUpdateTenantPoolsWithPatchError() {
	size := int64(1024)
	seconds := int64(5)
	weight := int32(1024)
	servers := int64(4)
	volumes := int32(4)
	mockString := "mock-string"
	pools := []*models.Pool{{
		VolumeConfiguration: &models.PoolVolumeConfiguration{
			Size: &size,
		},
		Servers:          &servers,
		VolumesPerServer: &volumes,
		Resources: &models.PoolResources{
			Requests: map[string]int64{
				"cpu": 1,
			},
			Limits: map[string]int64{
				"memory": 1,
			},
		},
		Tolerations: models.PoolTolerations{{
			TolerationSeconds: &models.PoolTolerationSeconds{
				Seconds: &seconds,
			},
		}},
		Affinity: &models.PoolAffinity{
			NodeAffinity: &models.PoolAffinityNodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &models.PoolAffinityNodeAffinityRequiredDuringSchedulingIgnoredDuringExecution{
					NodeSelectorTerms: []*models.NodeSelectorTerm{{
						MatchExpressions: []*models.NodeSelectorTermMatchExpressionsItems0{{
							Key:      &mockString,
							Operator: &mockString,
						}},
					}},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []*models.PoolAffinityNodeAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{{
					Weight: &weight,
					Preference: &models.NodeSelectorTerm{
						MatchFields: []*models.NodeSelectorTermMatchFieldsItems0{{
							Key:      &mockString,
							Operator: &mockString,
						}},
					},
				}},
			},
			PodAffinity: &models.PoolAffinityPodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []*models.PodAffinityTerm{{
					LabelSelector: &models.PodAffinityTermLabelSelector{
						MatchExpressions: []*models.PodAffinityTermLabelSelectorMatchExpressionsItems0{{
							Key:      &mockString,
							Operator: &mockString,
						}},
					},
					TopologyKey: &mockString,
				}},
				PreferredDuringSchedulingIgnoredDuringExecution: []*models.PoolAffinityPodAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{{
					PodAffinityTerm: &models.PodAffinityTerm{
						LabelSelector: &models.PodAffinityTermLabelSelector{
							MatchExpressions: []*models.PodAffinityTermLabelSelectorMatchExpressionsItems0{{
								Key:      &mockString,
								Operator: &mockString,
							}},
						},
						TopologyKey: &mockString,
					},
					Weight: &weight,
				}},
			},
			PodAntiAffinity: &models.PoolAffinityPodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []*models.PodAffinityTerm{{
					LabelSelector: &models.PodAffinityTermLabelSelector{
						MatchExpressions: []*models.PodAffinityTermLabelSelectorMatchExpressionsItems0{{
							Key:      &mockString,
							Operator: &mockString,
						}},
					},
					TopologyKey: &mockString,
				}},
				PreferredDuringSchedulingIgnoredDuringExecution: []*models.PoolAffinityPodAntiAffinityPreferredDuringSchedulingIgnoredDuringExecutionItems0{{
					PodAffinityTerm: &models.PodAffinityTerm{
						LabelSelector: &models.PodAffinityTermLabelSelector{
							MatchExpressions: []*models.PodAffinityTermLabelSelectorMatchExpressionsItems0{{
								Key:      &mockString,
								Operator: &mockString,
							}},
						},
						TopologyKey: &mockString,
					},
					Weight: &weight,
				}},
			},
		},
	}}
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{}, nil
	}
	opClientTenantPatchMock = func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
		return nil, errors.New("mock-patch-error")
	}
	_, err := updateTenantPools(context.Background(), suite.opClient, "mock-namespace", "mock-tenant", pools)
	suite.assert.NotNil(err)
}

func (suite *TenantTestSuite) TestUpdateTenantPoolsWithoutError() {
	seconds := int64(10)
	opClientTenantGetMock = func(ctx context.Context, namespace string, tenantName string, options metav1.GetOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{}, nil
	}
	opClientTenantPatchMock = func(ctx context.Context, namespace string, tenantName string, pt types.PatchType, data []byte, options metav1.PatchOptions) (*miniov2.Tenant, error) {
		return &miniov2.Tenant{
			Spec: miniov2.TenantSpec{
				Pools: []miniov2.Pool{{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						Spec: corev1.PersistentVolumeClaimSpec{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("1Gi"),
								},
							},
						},
					},
					SecurityContext: suite.createTenantPodSecurityContext(),
					Tolerations: []corev1.Toleration{{
						TolerationSeconds: &seconds,
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceLimitsMemory: resource.MustParse("1"),
						},
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{{
									MatchExpressions: []corev1.NodeSelectorRequirement{{}},
								}},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
								Preference: corev1.NodeSelectorTerm{
									MatchFields: []corev1.NodeSelectorRequirement{{}},
								},
							}},
						},
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{{}},
								},
							}},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{{}},
									},
								},
							}},
						},
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{{}},
								},
							}},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{{}},
									},
								},
							}},
						},
					},
				}},
			},
		}, nil
	}
	_, err := updateTenantPools(context.Background(), suite.opClient, "mock-namespace", "mock-tenant", []*models.Pool{})
	suite.assert.Nil(err)
}
