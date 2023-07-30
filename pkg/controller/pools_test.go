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

package controller

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_poolSSMatchesSpec(t *testing.T) {
	always := corev1.FSGroupChangeAlways
	onRootMismatch := corev1.FSGroupChangeOnRootMismatch
	type args struct {
		expectedSS *appsv1.StatefulSet
		existingSS *appsv1.StatefulSet
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
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:          "pool-0",
							miniov2.TenantLabel:        "tenant-a",
							miniov2.ConsoleTenantLabel: "tenant-a-console",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:          "pool-0",
							miniov2.TenantLabel:        "tenant-a",
							miniov2.ConsoleTenantLabel: "tenant-a-console",
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
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Sidecar added",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
										Name: "warp",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Sidecar removed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "minio image upgrade",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "minio",
										Image: "minio/minio:RELEASE.2022-08-25T07-17-05Z",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "minio",
										Image: "minio/minio:RELEASE.2022-08-22T23-53-06Z",
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "security context change",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SecurityContext: &corev1.PodSecurityContext{
									FSGroupChangePolicy: &always,
								},
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SecurityContext: &corev1.PodSecurityContext{
									FSGroupChangePolicy: &onRootMismatch,
								},
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Sidecar image changed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
										Image: "minio/warp:v0.3.21",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Resource Change",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
												corev1.ResourceCPU: resource.MustParse("16"),
											},
										},
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Resource Removed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Affinity Change",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
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
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Tenant Affinity Removed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Annotations Changed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
						},
						Annotations: map[string]string{
							"x": "y",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Label Changed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							"x": "y",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
							"x":                 "x",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Topology Spread Constraints Changed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
							"x":                 "x",
						},
						Annotations: map[string]string{
							miniov2.Revision: "0",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
									{
										MaxSkew:           1,
										TopologyKey:       "zone",
										WhenUnsatisfiable: "DoNotSchedule",
										LabelSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      miniov2.PoolLabel,
													Operator: metav1.LabelSelectorOpIn,
													Values:   []string{"pool-0"},
												},
											},
										},
									},
								},
								Containers: []corev1.Container{
									{
										Name: "minio",
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
						Labels: map[string]string{
							miniov2.PoolLabel:   "pool-0",
							miniov2.TenantLabel: "tenant-a",
							"x":                 "x",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Environment variable added",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
										Env: []corev1.EnvVar{
											{
												Name:  "MINIO_BROWSER_REDIRECT_URL",
												Value: "http://localhost:9000",
											},
										},
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Environment variable removed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "minio",
										Env: []corev1.EnvVar{
											{
												Name:  "MINIO_BROWSER_REDIRECT_URL",
												Value: "http://localhost:9000",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Certificate mounted",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "tenant-tls",
									},
								},
								Containers: []corev1.Container{
									{
										Name: "minio",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "tenant-tls",
												MountPath: "/tmp/certs",
											},
										},
									},
								},
							},
						},
					},
				},
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
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
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Certificate removed",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
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
				existingSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "tenant-tls",
									},
								},
								Containers: []corev1.Container{
									{
										Name: "minio",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "tenant-tls",
												MountPath: "/tmp/certs",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "provide nil statefulset",
			args: args{
				expectedSS: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-pool-0",
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
				existingSS: nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := poolSSMatchesSpec(tt.args.expectedSS, tt.args.existingSS)
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
