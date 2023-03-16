// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
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
	"os"
	"testing"
)

func Test_getK8sSAToken(t *testing.T) {
	tests := []struct {
		name string
		want string
		envs map[string]string
	}{
		{
			name: "Missing file, empty",
			want: "",
			envs: nil,
		},
		{
			name: "Missing file, return env",
			want: "x",
			envs: map[string]string{
				OperatorSAToken: "x",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envs != nil {
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}
			}
			if got := getK8sSAToken(); got != tt.want {
				t.Errorf("getK8sSAToken() = %v, want %v", got, tt.want)
			}
			if tt.envs != nil {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}
		})
	}
}

func Test_getMarketplace(t *testing.T) {
	tests := []struct {
		name string
		want string
		envs map[string]string
	}{
		{
			name: "Nothing set",
			want: "",
			envs: nil,
		},
		{
			name: "Value set",
			want: "x",
			envs: map[string]string{
				Marketplace: "x",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envs != nil {
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}
			}
			if got := getMarketplace(); got != tt.want {
				t.Errorf("getMarketplace() = %v, want %v", got, tt.want)
			}
			if tt.envs != nil {
				for k := range tt.envs {
					os.Unsetenv(k)
				}
			}
		})
	}
}
