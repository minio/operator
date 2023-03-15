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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	opClientMock   struct{}
	httpClientMock struct{}
)

func createMockPVC(pvcMockName, pvcMockNamespace string) *v1.PersistentVolumeClaim {
	var mockVolumeMode v1.PersistentVolumeMode = "mockVolumeMode"
	mockStorage := "mockStorage"

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcMockName,
			Namespace: pvcMockNamespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &mockStorage,
			VolumeMode:       &mockVolumeMode,
		},
	}
}
