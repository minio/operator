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

package subnet

import (
	"reflect"
	"testing"

	"github.com/minio/pkg/licverifier"
)

var (
	license = "eyJhbGciOiJFUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJrYW5hZ2FyYWorYzFAbWluaW8uaW8iLCJjYXAiOjUwLCJvcmciOiJHcmluZ290dHMgSW5jLiIsImV4cCI6MS42NDE0NDYxNjkwMDExOTg4OTRlOSwicGxhbiI6IlNUQU5EQVJEIiwiaXNzIjoic3VibmV0QG1pbi5pbyIsImFpZCI6MSwiaWF0IjoxLjYwOTkxMDE2OTAwMTE5ODg5NGU5fQ.EhTL2xwMHnUoLQF4UR-5bjUCja3whseLU5mb9XEj7PvAae6HEIDCOMEF8Hhh20DN_v_LRE283j2ZlA5zulcXSZXS0CLcrKqbVy6QLvZfvvLuerOjJI-NBa9dSJWJ0WoN"

	publicKeys = []string{`-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEbo+e1wpBY4tBq9AONKww3Kq7m6QP/TBQ
mr/cKCUyBL7rcAvg0zNq1vcSrUSGlAmY3SEDCu3GOKnjG/U4E7+p957ocWSV+mQU
9NKlTdQFGF3+aO6jbQ4hX/S5qPyF+a3z
-----END PUBLIC KEY-----`}
)

func TestGetLicenseInfoFromJWT(t *testing.T) {
	mockLicense, _ := GetLicenseInfoFromJWT(license, publicKeys)

	type args struct {
		license    string
		publicKeys []string
	}
	tests := []struct {
		name    string
		args    args
		want    *licverifier.LicenseInfo
		wantErr bool
	}{
		{
			name: "error because missing license",
			args: args{
				license:    "",
				publicKeys: OfflinePublicKeys,
			},
			wantErr: true,
		},
		{
			name: "error because invalid license",
			args: args{
				license:    license,
				publicKeys: []string{"eaeaeae"},
			},
			wantErr: true,
		},
		{
			name: "license successfully verified",
			args: args{
				license:    license,
				publicKeys: publicKeys,
			},
			wantErr: false,
			want:    mockLicense,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLicenseInfoFromJWT(tt.args.license, tt.args.publicKeys)
			if !tt.wantErr {
				t.Skip()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLicenseInfoFromJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLicenseInfoFromJWT() got = %v, want %v", got, tt.want)
			}
		})
	}
}
