// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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
	"reflect"
	"testing"

	"github.com/minio/operator/models"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func NoTestMaxAllocatableMemory(t *testing.T) {
	type args struct {
		ctx      context.Context
		numNodes int32
		objs     []runtime.Object
	}
	tests := []struct {
		name     string
		args     args
		expected *models.MaxAllocatableMemResponse
		wantErr  bool
	}{
		{
			name: "Get Max Ram No Taints",
			args: args{
				ctx:      context.Background(),
				numNodes: 2,
				objs: []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Ki"),
								corev1.ResourceCPU:    resource.MustParse("4Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("512"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
				},
			},
			expected: &models.MaxAllocatableMemResponse{
				MaxMemory: int64(512),
			},
			wantErr: false,
		},
		{
			// Description: if there are more nodes than the amount
			// of nodes we want to use, but one has taints of NoSchedule
			// node should not be considered for max memory
			name: "Get Max Ram on nodes with NoSchedule",
			args: args{
				ctx:      context.Background(),
				numNodes: 2,
				objs: []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Ki"),
								corev1.ResourceCPU:    resource.MustParse("4Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
						Spec: corev1.NodeSpec{
							Taints: []corev1.Taint{
								{
									Key:    "node.kubernetes.io/unreachable",
									Effect: corev1.TaintEffectNoSchedule,
								},
							},
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("6Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node3",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("4Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
				},
			},
			expected: &models.MaxAllocatableMemResponse{
				MaxMemory: int64(2048),
			},
			wantErr: false,
		},
		{
			// Description: if there are more nodes than the amount
			// of nodes we want to use, but one has taints of NoExecute
			// node should not be considered for max memory
			// if one node has PreferNoSchedule that should be considered.
			name: "Get Max Ram on nodes with NoExecute",
			args: args{
				ctx:      context.Background(),
				numNodes: 2,
				objs: []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Ki"),
								corev1.ResourceCPU:    resource.MustParse("4Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
						Spec: corev1.NodeSpec{
							Taints: []corev1.Taint{
								{
									Key:    "node.kubernetes.io/unreachable",
									Effect: corev1.TaintEffectNoExecute,
								},
							},
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("6Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node3",
						},
						Spec: corev1.NodeSpec{
							Taints: []corev1.Taint{
								{
									Key:    "node.kubernetes.io/unreachable",
									Effect: corev1.TaintEffectPreferNoSchedule,
								},
							},
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("4Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
				},
			},
			expected: &models.MaxAllocatableMemResponse{
				MaxMemory: int64(2048),
			},
			wantErr: false,
		},
		{
			// Description: if there are more nodes than the amount
			// of nodes we want to use, max allocatable memory should
			// be the minimum ram on the n nodes requested
			name: "Get Max Ram, several nodes available",
			args: args{
				ctx:      context.Background(),
				numNodes: 2,
				objs: []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Ki"),
								corev1.ResourceCPU:    resource.MustParse("4Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("6Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node3",
						},
						Status: corev1.NodeStatus{
							Allocatable: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("4Ki"),
								corev1.ResourceCPU:    resource.MustParse("1Ki"),
							},
						},
					},
				},
			},
			expected: &models.MaxAllocatableMemResponse{
				MaxMemory: int64(4096),
			},
			wantErr: false,
		},
		{
			// Description: if request has nil as request, expect error
			name: "Nil nodes should be greater than 0",
			args: args{
				ctx:      context.Background(),
				numNodes: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		kubeClient := fake.NewSimpleClientset(tt.args.objs...)
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMaxAllocatableMemory(tt.args.ctx, kubeClient.CoreV1(), tt.args.numNodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMaxAllocatableMemory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("\ngot: %d \nwant: %d", got, tt.expected)
			}
		})
	}
}

func Test_MaxMemoryFunc(t *testing.T) {
	type args struct {
		numNodes         int32
		nodesMemorySizes []int64
	}
	tests := []struct {
		name     string
		args     args
		expected int64
		wantErr  bool
	}{
		{
			name: "Get Max memory",
			args: args{
				numNodes:         int32(4),
				nodesMemorySizes: []int64{4294967296, 8589934592, 8589934592, 17179869184, 17179869184, 17179869184, 25769803776, 25769803776, 68719476736},
			},
			expected: int64(17179869184),
			wantErr:  false,
		},
		{
			// Description, if not enough nodes return 0
			name: "Get Max memory Not enough nodes",
			args: args{
				numNodes:         int32(4),
				nodesMemorySizes: []int64{4294967296, 8589934592, 68719476736},
			},
			expected: int64(0),
			wantErr:  false,
		},
		{
			// Description, if not enough nodes return 0
			name: "Get Max memory no nodes",
			args: args{
				numNodes:         int32(4),
				nodesMemorySizes: []int64{},
			},
			expected: int64(0),
			wantErr:  false,
		},
		{
			// Description, if not enough nodes return 0
			name: "Get Max memory no nodes, no request",
			args: args{
				numNodes:         int32(0),
				nodesMemorySizes: []int64{},
			},
			expected: int64(0),
			wantErr:  false,
		},
		{
			// Description, if there are multiple nodes
			// and required nodes is only 1, max should be equal to max memory
			name: "Get Max memory one node",
			args: args{
				numNodes:         int32(1),
				nodesMemorySizes: []int64{4294967296, 8589934592, 68719476736},
			},
			expected: int64(68719476736),
			wantErr:  false,
		},
		{
			// Description: if more nodes max memory should be the minimum
			// value across pairs of numNodes
			name: "Get Max memory two nodes",
			args: args{
				numNodes:         int32(2),
				nodesMemorySizes: []int64{8589934592, 68719476736, 4294967296},
			},
			expected: int64(8589934592),
			wantErr:  false,
		},
		{
			name: "Get Max Multiple Memory Sizes",
			args: args{
				numNodes:         int32(4),
				nodesMemorySizes: []int64{0, 0, 0, 0, 4294967296, 8589934592, 8589934592, 17179869184, 17179869184, 17179869184, 25769803776, 25769803776, 68719476736, 34359738368, 34359738368, 34359738368, 34359738368},
			},
			expected: int64(34359738368),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMaxClusterMemory(tt.args.numNodes, tt.args.nodesMemorySizes)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("\ngot: %d \nwant: %d", got, tt.expected)
			}
		})
	}
}
