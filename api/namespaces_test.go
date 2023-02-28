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
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func Test_CreateNamespace(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Namespace created successfully",
			args: args{
				ctx:       context.Background(),
				namespace: "ns-test",
			},
			wantErr: false,
		},
		{
			// Description: If the namespace is blank, an error should be returned
			name: "Namespace creation failed",
			args: args{
				ctx:       context.Background(),
				namespace: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		kubeClient := fake.NewSimpleClientset()
		t.Run(tt.name, func(t *testing.T) {
			err := getNamespaceCreated(tt.args.ctx, kubeClient.CoreV1(), tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaceCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
