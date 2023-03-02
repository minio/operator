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
	"context"

	"github.com/minio/directpv/pkg/apis/direct.csi.min.io/v1beta4"
	directPVClient "github.com/minio/directpv/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DirectPVDrivesClientI interface with all functions to be implemented
// by mock when testing, it should include all DirectPVDrivesClientI respective api calls
// that are used within this project.
type DirectPVDrivesClientI interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1beta4.DirectCSIDriveList, error)
	Update(ctx context.Context, driveItem *v1beta4.DirectCSIDrive, opts metav1.UpdateOptions) (*v1beta4.DirectCSIDriveList, error)
}

// Interface implementation
//
// Define the structure of directpv drive client and define the functions that are actually used
// from the minio operator / directpv interface.

type directPVDrivesClient struct {
	client *directPVClient.DirectCSIDriveInterface
}

// List implements the listing for DirectPV Drives List functionality
func (dpd *directPVDrivesClient) List(ctx context.Context, opts metav1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
	return dpd.List(ctx, opts)
}

// Update implements the listing for DirectPV Drives Update functionality
func (dpd *directPVDrivesClient) Update(ctx context.Context, driveItem *v1beta4.DirectCSIDrive, opts metav1.UpdateOptions) (*v1beta4.DirectCSIDriveList, error) {
	return dpd.Update(ctx, driveItem, opts)
}

// DirectPVVolumesClientI interface with all functions to be implemented
// by mock when testing, it should include all DirectPVVolumesClientI respective api calls
// that are used within this project.
type DirectPVVolumesClientI interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1beta4.DirectCSIVolumeList, error)
}

// Interface implementation
//
// Define the structure of directpv volumes client and define the functions that are actually used
// from the minio operator / directpv interface.

type directPVVolumesClient struct {
	client *directPVClient.DirectCSIVolumeInterface
}

// List implements the listing for DirectPV Volumes List functionality
func (dpv *directPVVolumesClient) List(ctx context.Context, opts metav1.ListOptions) (*v1beta4.DirectCSIVolumeList, error) {
	return dpv.List(ctx, opts)
}
