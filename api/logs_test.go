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
	"flag"
	"fmt"
	"testing"

	"github.com/minio/cli"
	"github.com/stretchr/testify/assert"
)

func TestContext_Load(t *testing.T) {
	type fields struct {
		Host           string
		HTTPPort       int
		HTTPSPort      int
		TLSRedirect    string
		TLSCertificate string
		TLSKey         string
		TLSca          string
	}
	type args struct {
		values map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid args",
			args: args{
				values: map[string]string{
					"tls-redirect": "on",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid args",
			args: args{
				values: map[string]string{
					"tls-redirect": "aaaa",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port http",
			args: args{
				values: map[string]string{
					"tls-redirect": "on",
					"port":         "65536",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port https",
			args: args{
				values: map[string]string{
					"tls-redirect": "on",
					"port":         "65534",
					"tls-port":     "65536",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{}

			fs := flag.NewFlagSet("flags", flag.ContinueOnError)
			for k, v := range tt.args.values {
				fs.String(k, v, "ok")
			}

			ctx := cli.NewContext(nil, fs, &cli.Context{})

			err := c.Load(ctx)
			if tt.wantErr {
				assert.NotNilf(t, err, fmt.Sprintf("Load(%v)", err))
			} else {
				assert.Nilf(t, err, fmt.Sprintf("Load(%v)", err))
			}
		})
	}
}

func Test_logInfo(t *testing.T) {
	logInfo("message", nil)
}
