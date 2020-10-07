/*
 * This file is part of MinIO Operator
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

package resources

import (
	helpers "github.com/minio/kubectl-minio/cmd/helpers"
	miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tenantStorage(q resource.Quantity) v1.ResourceList {
	m := make(v1.ResourceList, 1)
	m[v1.ResourceStorage] = q
	return m
}

// Zone returns a Zone object from given values
func Zone(servers, volumes int32, q resource.Quantity, sc string) miniov1.Zone {
	return miniov1.Zone{
		Servers:          servers,
		VolumesPerServer: volumes,
		VolumeClaimTemplate: &v1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1.ResourcePersistentVolumeClaims.String(),
				APIVersion: v1.SchemeGroupVersion.Version,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{helpers.MinIOAccessMode},
				Resources: v1.ResourceRequirements{
					Requests: tenantStorage(q),
				},
				StorageClassName: storageClass(sc),
			},
		},
	}
}
