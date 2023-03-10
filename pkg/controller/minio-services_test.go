// Copyright (C) 2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package controller

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_minioSvcMatchesSpecification(t *testing.T) {
	type args struct {
		svc         *v1.Service
		expectedSvc *v1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Everything matches",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Labels don't matches",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val2",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Annotations don't match",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val2",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Target Ports don't match",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9001),
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Ports don't match ",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       444,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Selector doesn't match",
			args: args{
				svc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "valid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
				expectedSvc: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "minio",
						Namespace: "default",
						Labels: map[string]string{
							"label": "val",
						},
						Annotations: map[string]string{
							"annotation": "val",
						},
					},
					Spec: v1.ServiceSpec{
						Type: "LoadBalancer",
						Selector: map[string]string{
							"selector": "invalid",
						},
						Ports: []v1.ServicePort{
							{
								Name:       "minio-https",
								Port:       443,
								TargetPort: intstr.FromInt(9000),
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := minioSvcMatchesSpecification(tt.args.svc, tt.args.expectedSvc); got != tt.want {
				t.Errorf("minioSvcMatchesSpecification() = %v, want %v", got, tt.want)
			}
		})
	}
}
