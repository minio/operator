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

import "testing"

func Test_versionCompare(t *testing.T) {
	type args struct {
		v1 string
		v2 string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "v1 == v2",
			args: args{
				v1: "v1.0.0",
				v2: "v1.0.0",
			},
			want: 0,
		},
		{
			name: "v1 > v2",
			args: args{
				v1: "v1.1.0",
				v2: "v1.0.0",
			},
			want: 1,
		},
		{
			name: "v2 > v1",
			args: args{
				v1: "v1.0.0",
				v2: "v1.1.0",
			},
			want: -1,
		},
		{
			name: "v1 > v2",
			args: args{
				v1: "v1.1.1-123+123-dirty",
				v2: "v1.1.0",
			},
			want: 1,
		},
		{
			name: "-1 if invalid",
			args: args{
				v1: "a-2.c",
				v2: "v1.1.0",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := versionCompare(tt.args.v1, tt.args.v2)
			if got != tt.want {
				t.Errorf("versionCompare() got = %v, want %v", got, tt.want)
			}
		})
	}
}
