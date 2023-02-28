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
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	getPVCWithError  = true
	pvcMockName      = "mockName"
	pvcMockNamespace = "mockNamespace"
)

func (c k8sClientMock) getPVC(ctx context.Context, namespace string, pvcName string, opts metav1.GetOptions) (*v1.PersistentVolumeClaim, error) {
	if getPVCWithError {
		return nil, errors.New("Mock error during getPVC")
	}
	return createMockPVC(pvcMockName, pvcMockNamespace), nil
}

var testCasesGetPVCDescribe = []struct {
	name          string
	errorExpected bool
}{
	{
		name:          "Successful getPVCDescribe",
		errorExpected: false,
	},
	{
		name:          "Error getPVCDescribe",
		errorExpected: true,
	},
}

func TestGetPVCDescribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := k8sClientMock{}
	for _, tt := range testCasesGetPVCDescribe {
		t.Run(tt.name, func(t *testing.T) {
			getPVCWithError = tt.errorExpected
			pvc, err := getPVCDescribe(ctx, pvcMockNamespace, pvcMockName, client)
			if err != nil {
				if tt.errorExpected {
					return
				}
				t.Errorf("getPVCDescribe() error = %v, errorExpected %v", err, tt.errorExpected)
			}
			if pvc == nil {
				t.Errorf("getPVCDescribe() expected type: *v1.PersistentVolumeClaim, got: nil")
				return
			}
			if pvc.Name != pvcMockName {
				t.Errorf("Expected pvc name %s got %s", pvc.Name, pvcMockName)
			}
			if pvc.Namespace != pvcMockNamespace {
				t.Errorf("Expected pvc namespace %s got %s", pvc.Namespace, pvcMockNamespace)
			}
		})
	}
}
