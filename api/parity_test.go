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
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/minio/operator/api/operations/operator_api"
	"github.com/stretchr/testify/assert"

	"github.com/minio/operator/models"
)

func Test_getParityInfo(t *testing.T) {
	tests := []struct {
		description  string
		wantErr      bool
		nodes        int64
		disksPerNode int64
		expectedResp models.ParityResponse
	}{
		{
			description:  "Incorrect Number of endpoints provided",
			wantErr:      true,
			nodes:        1,
			disksPerNode: 1,
			expectedResp: nil,
		},
		{
			description:  "Number of endpoints is valid",
			wantErr:      false,
			nodes:        4,
			disksPerNode: 10,
			expectedResp: models.ParityResponse{"EC:4", "EC:3", "EC:2"},
		},
		{
			description:  "More nodes than disks",
			wantErr:      false,
			nodes:        4,
			disksPerNode: 1,
			expectedResp: models.ParityResponse{"EC:2"},
		},
		{
			description:  "More disks than nodes",
			wantErr:      false,
			nodes:        2,
			disksPerNode: 50,
			expectedResp: models.ParityResponse{"EC:5", "EC:4", "EC:3", "EC:2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			parity, err := GetParityInfo(tt.nodes, tt.disksPerNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetParityInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(parity, tt.expectedResp) {
				ji, _ := json.Marshal(parity)
				vi, _ := json.Marshal(tt.expectedResp)
				t.Errorf("\ngot: %s \nwant: %s", ji, vi)
			}
		})
	}
}

func Test_getParityResponse(t *testing.T) {
	type args struct {
		params operator_api.GetParityParams
	}
	tests := []struct {
		name    string
		args    args
		want    models.ParityResponse
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				params: operator_api.GetParityParams{
					HTTPRequest:  &http.Request{},
					DisksPerNode: 4,
					Nodes:        4,
				},
			},
			want:    models.ParityResponse{"EC:8", "EC:7", "EC:6", "EC:5", "EC:4", "EC:3", "EC:2"},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				params: operator_api.GetParityParams{
					HTTPRequest:  &http.Request{},
					DisksPerNode: -4,
					Nodes:        4,
				},
			},
			want:    models.ParityResponse(nil),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getParityResponse(tt.args.params)
			assert.Equalf(t, tt.want, got, "getParityResponse(%v)", tt.args.params)
			if tt.wantErr {
				assert.NotNilf(t, got1, "getParityResponse(%v)", tt.args.params)
			} else {
				assert.Nilf(t, got1, "getParityResponse(%v)", tt.args.params)
			}
		})
	}
}
