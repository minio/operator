// This file is part of MinIO Operator
// Copyright (c) 2020 MinIO, Inc.
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

package cluster

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_poolSSMatchesSpec(t *testing.T) {
	type args struct {
		tenant          *miniov2.Tenant
		pool            *miniov2.Pool
		ss              *appsv1.StatefulSet
		operatorVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Tenant Unchanged",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Sidecar added",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						SideCars: &miniov2.SideCars{
							Containers: []corev1.Container{
								{
									Name: "warp",
								},
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Sidecar removed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
									{
										Name:  "warp",
										Image: "minio/warp:v0.3.20",
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Sidecar image changed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						SideCars: &miniov2.SideCars{
							Containers: []corev1.Container{
								{
									Name:  "warp",
									Image: "minio/warp:v0.3.21",
								},
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
									{
										Name:  "warp",
										Image: "minio/warp:v0.3.20",
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Resource Change",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
								Resources: corev1.ResourceRequirements{
									Limits: nil,
									Requests: corev1.ResourceList{
										corev1.ResourceCPU: resource.MustParse("16"),
									},
								},
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
					Resources: corev1.ResourceRequirements{
						Limits: nil,
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("16"),
						},
					},
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
										Resources: corev1.ResourceRequirements{
											Limits: nil,
											Requests: corev1.ResourceList{
												corev1.ResourceCPU: resource.MustParse("14"),
											},
										},
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Resource Removed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
										Resources: corev1.ResourceRequirements{
											Limits: nil,
											Requests: corev1.ResourceList{
												corev1.ResourceCPU: resource.MustParse("14"),
											},
										},
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Affinity Change",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      miniov2.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov2.PoolLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"pool-0"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      miniov2.TenantLabel,
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"tenant-a"},
											}, {
												Key:      miniov2.PoolLabel,
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"pool-0"},
											},
										},
									},
								},
							},
						},
					},
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Affinity Removed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      miniov2.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov2.PoolLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"pool-0"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Annotations Changed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
						Annotations: map[string]string{
							"x": "y",
						},
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
							"x":              "x",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      miniov2.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov2.PoolLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"pool-0"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Label Changed",
			args: args{
				tenant: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
						Labels: map[string]string{
							"x": "y",
						},
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name: "pool-0",
							},
						},
					},
				},
				pool: &miniov2.Pool{
					Name: "pool-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:     "pool-0",
							miniov2.TenantLabel:   "tenant-a",
							miniov2.OperatorLabel: "0.1",
							"x":                   "x",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      miniov2.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov2.PoolLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"pool-0"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				operatorVersion: "0.1",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := poolSSMatchesSpec(tt.args.tenant, tt.args.pool, tt.args.ss, tt.args.operatorVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("poolSSMatchesSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("poolSSMatchesSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}
