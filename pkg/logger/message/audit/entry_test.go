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

package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/minio/operator/pkg/utils"
)

func TestNewEntry(t *testing.T) {
	type args struct {
		deploymentID string
	}
	tests := []struct {
		name string
		args args
		want Entry
	}{
		{
			name: "constructs an audit entry object with some fields filled",
			args: args{
				deploymentID: "1",
			},
			want: Entry{
				Version:      Version,
				DeploymentID: "1",
				Time:         time.Now().UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEntry(tt.args.deploymentID); got.DeploymentID != tt.want.DeploymentID {
				t.Errorf("NewEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToEntry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants?test=xyz", nil)
	req.Header.Set("Authorization", "xyz")
	req.Header.Set("ETag", "\"ABCDE\"")

	// applying context information
	ctx := context.WithValue(req.Context(), utils.ContextRequestUserID, "eyJhbGciOiJSUzI1NiIsImtpZCI6Ing5cS0wSkEwQzFMWDJlRlR3dHo2b0t0NVNnRzJad0llMGVNczMxbjU0b2sifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJtaW5pby1vcGVyYXRvciIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJjb25zb2xlLXNhLXRva2VuLWJrZzZwIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImNvbnNvbGUtc2EiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJhZTE2ZGVkNS01MmM3LTRkZTQtOWUxYS1iNmI4NGU2OGMzM2UiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6bWluaW8tb3BlcmF0b3I6Y29uc29sZS1zYSJ9.AjhzekAPC59SQVBQL5sr-1dqr57-jH8a5LVazpnEr_cC0JqT4jXYjdfbrZSF9yaL4gHRv2l0kOhBlrjRK7y-IpMbxE71Fne_lSzaptSuqgI5I9dFvpVfZWP1yMAqav8mrlUoWkWDq9IAkyH4bvvZrVgQJGgd5t9U_7DQCVwbkQvy0wGS5zoMcZhYenn_Ub1BoxWcviADQ1aY1wQju8OP0IOwKTIMXMQqciOFdJ9T5-tQEGUrikTu_tW-1shUHzOxBcEzGVtBvBy2OmbNnRFYogbhmp-Dze6EAi035bY32bfL7XKBUNCW6_3VbN_h3pQNAuT2NJOSKuhJ3cGldCB2zg")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	w.Header().Set("Authorization", "xyz")
	w.Header().Set("ETag", "\"ABCDE\"")

	type args struct {
		w            http.ResponseWriter
		r            *http.Request
		reqClaims    map[string]interface{}
		deploymentID string
	}
	tests := []struct {
		name     string
		args     args
		want     Entry
		preFunc  func()
		postFunc func()
	}{
		{
			preFunc: func() {
				os.Setenv("CONSOLE_OPERATOR_MODE", "on")
			},
			postFunc: func() {
				os.Unsetenv("CONSOLE_OPERATOR_MODE")
			},
			name: "constructs an audit entry from a http request",
			args: args{
				w:            w,
				r:            req,
				reqClaims:    map[string]interface{}{},
				deploymentID: "1",
			},
			want: Entry{
				Version:      "1",
				DeploymentID: "1",
				SessionID:    "system:serviceaccount:minio-operator:console-sa",
				ReqQuery:     map[string]string{"test": "xyz"},
				ReqHeader:    map[string]string{"test": "xyz"},
				RespHeader:   map[string]string{"test": "xyz", "ETag": "ABCDE"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preFunc != nil {
				tt.preFunc()
			}
			if got := ToEntry(tt.args.w, tt.args.r, tt.args.reqClaims, tt.args.deploymentID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToEntry() = %v, want %v", got, tt.want)
			}
			if tt.postFunc != nil {
				tt.postFunc()
			}
		})
	}
}
