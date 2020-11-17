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

	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_zoneSSMatchesSpec(t *testing.T) {
	type args struct {
		tenant *miniov1.Tenant
		zone   *miniov1.Zone
		ss     *appsv1.StatefulSet
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
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			name: "Sidecar added",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						SideCars: &miniov1.SideCars{
							Containers: []corev1.Container{
								{
									Name: "warp",
								},
							},
						},
					},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			name: "Sidecar removed",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						Zones: []miniov1.Zone{
							{
								Name: "zone-0",
							},
						},
					},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			want:    true,
			wantErr: false,
		},
		{
			name: "Sidecar image changed",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						SideCars: &miniov1.SideCars{
							Containers: []corev1.Container{
								{
									Name:  "warp",
									Image: "minio/warp:v0.3.21",
								},
							},
						},
					},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			want:    true,
			wantErr: false,
		},
		{
			name: "Tenant Resource Change",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						Zones: []miniov1.Zone{
							{
								Name: "zone-0",
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
				zone: &miniov1.Zone{
					Name: "zone-0",
					Resources: corev1.ResourceRequirements{
						Limits: nil,
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("16"),
						},
					},
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			want:    true,
			wantErr: false,
		},
		{
			name: "Tenant Resource Removed",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						Zones: []miniov1.Zone{
							{
								Name: "zone-0",
							},
						},
					},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
			want:    true,
			wantErr: false,
		},
		{
			name: "Tenant Affinity Change",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						Zones: []miniov1.Zone{
							{
								Name: "zone-0",
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      miniov1.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov1.ZoneLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"zone-0"},
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
				zone: &miniov1.Zone{
					Name: "zone-0",
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      miniov1.TenantLabel,
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"tenant-a"},
											}, {
												Key:      miniov1.ZoneLabel,
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"zone-0"},
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
						Name: "tenant-a-zone-0",
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
			name: "Tenant Affinity Removed",
			args: args{
				tenant: &miniov1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a",
					},
					Spec: miniov1.TenantSpec{
						Zones: []miniov1.Zone{
							{
								Name: "zone-0",
							},
						},
					},
				},
				zone: &miniov1.Zone{
					Name: "zone-0",
				},
				ss: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-a-zone-0",
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
															Key:      miniov1.TenantLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"tenant-a"},
														}, {
															Key:      miniov1.ZoneLabel,
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{"zone-0"},
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
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := zoneSSMatchesSpec(tt.args.tenant, tt.args.zone, tt.args.ss)
			if (err != nil) != tt.wantErr {
				t.Errorf("zoneSSMatchesSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("zoneSSMatchesSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}
