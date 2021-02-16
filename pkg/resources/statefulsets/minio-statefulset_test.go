// This file is part of MinIO Console Server
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

package statefulsets

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func TestGetContainerArgs(t *testing.T) {
	type args struct {
		t             *miniov2.Tenant
		hostsTemplate string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Empty Tenant",
			args: args{
				t:             &miniov2.Tenant{},
				hostsTemplate: "",
			},
			want: nil,
		},
		{
			name: "One Pool Tenant",
			args: args{
				t: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "minio",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Servers:             4,
								VolumesPerServer:    4,
								VolumeClaimTemplate: nil,
							},
						},
					},
				},
				hostsTemplate: "",
			},
			want: []string{
				"https://minio-ss-0-{0...3}.minio-hl..svc.cluster.local/export{0...3}",
			},
		},
		{
			name: "One Pool Tenant With Named Pool",
			args: args{
				t: &miniov2.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "minio",
					},
					Spec: miniov2.TenantSpec{
						Pools: []miniov2.Pool{
							{
								Name:                "pool-0",
								Servers:             4,
								VolumesPerServer:    4,
								VolumeClaimTemplate: nil,
							},
						},
					},
				},
				hostsTemplate: "",
			},
			want: []string{
				"https://minio-pool-0-{0...3}.minio-hl..svc.cluster.local/export{0...3}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ensure defaults
			tt.args.t.EnsureDefaults()

			if got := GetContainerArgs(tt.args.t, tt.args.hostsTemplate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContainerArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
