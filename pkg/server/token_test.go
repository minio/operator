/*
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package server

import "testing"

func Test_generateJWTForTenant(t *testing.T) {
	type args struct {
		tenantFQDN   string
		operatorFQDN string
		key          interface{}
	}
	tests := []struct {
		name            string
		args            args
		want            string
		wantEncodeError bool
		wantDecodeError bool
	}{
		{
			name: "Success 1",
			args: args{
				tenantFQDN:   "minio.default.svc.cluster.local",
				operatorFQDN: "minio-operator.namespace.svc.cluster.local",
				key:          []byte("some-random-key"),
			},
			want:            "minio.default.svc.cluster.local",
			wantEncodeError: false,
			wantDecodeError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWTForTenant(tt.args.tenantFQDN, tt.args.operatorFQDN, tt.args.key)
			if (err != nil) != tt.wantEncodeError {
				t.Errorf("GenerateJWTForTenant() error = %v, wantEncodeErr %v", err, tt.wantEncodeError)
				return
			}
			got, err := DecodeJWTGetTenant(token, tt.args.key)
			if (err != nil) != tt.wantDecodeError {
				t.Errorf("GenerateJWTForTenant() error = %v, wantDecodeErr %v", err, tt.wantDecodeError)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateJWTForTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}
