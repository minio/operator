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
	"errors"
	"testing"

	"github.com/minio/directpv/pkg/apis/direct.csi.min.io/v1beta4"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	dpdClientListMock   func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error)
	dpdClientUpdateMock func(ctx context.Context, driveItem *v1beta4.DirectCSIDrive, opts v1.UpdateOptions) (*v1beta4.DirectCSIDriveList, error)
)

// mock function for drives List()
func (dpdm directPVDriveMock) List(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
	return dpdClientListMock(ctx, opts)
}

func (dpdm directPVDriveMock) Update(ctx context.Context, driveItem *v1beta4.DirectCSIDrive, opts v1.UpdateOptions) (*v1beta4.DirectCSIDriveList, error) {
	return dpdClientUpdateMock(ctx, driveItem, opts)
}

var dpvClientListMock func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error)

// mock function for volumes List()
func (dpvm directPVVolumeMock) List(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error) {
	return dpvClientListMock(ctx, opts)
}

// DirectPVDrivesList
func Test_GetDirectPVDrives(t *testing.T) {
	directPVDrvMock := directPVDriveMock{}

	type args struct {
		ctx  context.Context
		opts v1.ListOptions
	}
	tests := []struct {
		name           string
		args           args
		client         DirectPVDrivesClientI
		mockListDrives func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error)
		wantErr        bool
	}{
		{
			name: "Can List Drives correctly",
			args: args{
				ctx:  context.Background(),
				opts: v1.ListOptions{},
			},
			client: directPVDrvMock,
			mockListDrives: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
				items := []v1beta4.DirectCSIDrive{}

				returnList := v1beta4.DirectCSIDriveList{
					Items: items,
				}

				return &returnList, nil
			},
			wantErr: false,
		},
		{
			name: "Drives request from DirectPV failed",
			args: args{
				ctx: context.Background(),
			},
			client: directPVDrvMock,
			mockListDrives: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
				return nil, errors.New("some error occurred")
			},
			wantErr: true,
		},
		{
			name: "Drives request from DirectPV has information and doesn't return errors",
			args: args{
				ctx: context.Background(),
			},
			client: directPVDrvMock,
			mockListDrives: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
				items := []v1beta4.DirectCSIDrive{
					{
						Status: v1beta4.DirectCSIDriveStatus{
							Path:              "/var/lib/direct-csi/devices/test-part-dev0-directcsi",
							AllocatedCapacity: 0,
							FreeCapacity:      4772382377372,
							RootPartition:     "/",
							PartitionNum:      0,
							Filesystem:        "nfs",
							Mountpoint:        "",
							NodeName:          "test-dev0-directcsi",
							DriveStatus:       v1beta4.DriveStatusReady,
							ModelNumber:       "testModel",
							SerialNumber:      "testSN",
							TotalCapacity:     4772382377372,
							PhysicalBlockSize: 1024,
							LogicalBlockSize:  1024,
							AccessTier:        "",
							FilesystemUUID:    "",
							PartitionUUID:     "",
							MajorNumber:       0,
							MinorNumber:       0,
							UeventSerial:      "",
							UeventFSUUID:      "",
							WWID:              "",
							Vendor:            "",
							DMName:            "",
							DMUUID:            "",
							MDUUID:            "",
							PartTableUUID:     "",
							PartTableType:     "",
							Virtual:           false,
							ReadOnly:          false,
							Partitioned:       false,
							SwapOn:            false,
							Master:            "",
							OtherMountsInfo:   nil,
							PCIPath:           "",
							SerialNumberLong:  "",
							Conditions: []v1.Condition{{
								Type:               "",
								Status:             "",
								ObservedGeneration: 0,
								LastTransitionTime: v1.Time{},
								Reason:             "",
								Message:            "",
							}},
						},
					},
					{
						Status: v1beta4.DirectCSIDriveStatus{
							Path:              "/var/lib/direct-csi/devices/test-part-dev1-directcsi",
							AllocatedCapacity: 0,
							FreeCapacity:      4772382377372,
							RootPartition:     "/",
							PartitionNum:      0,
							Filesystem:        "nfs",
							Mountpoint:        "",
							NodeName:          "test-dev1-directcsi",
							DriveStatus:       v1beta4.DriveStatus(v1beta4.DirectCSIDriveConditionOwned),
							ModelNumber:       "testModel",
							SerialNumber:      "testSN2",
							TotalCapacity:     4772382377372,
							PhysicalBlockSize: 1024,
							LogicalBlockSize:  1024,
							AccessTier:        "",
							FilesystemUUID:    "",
							PartitionUUID:     "",
							MajorNumber:       0,
							MinorNumber:       0,
							UeventSerial:      "",
							UeventFSUUID:      "",
							WWID:              "",
							Vendor:            "",
							DMName:            "",
							DMUUID:            "",
							MDUUID:            "",
							PartTableUUID:     "",
							PartTableType:     "",
							Virtual:           false,
							ReadOnly:          false,
							Partitioned:       false,
							SwapOn:            false,
							Master:            "",
							OtherMountsInfo:   nil,
							PCIPath:           "",
							SerialNumberLong:  "",
							Conditions:        nil,
						},
					},
				}

				returnList := v1beta4.DirectCSIDriveList{
					Items: items,
				}

				return &returnList, nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpdClientListMock = tt.mockListDrives

			_, err := getDirectPVDriveList(tt.args.ctx, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaceCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// DirectPVVolumesList
func Test_GetDirectPVVolumes(t *testing.T) {
	directPVVolMock := directPVVolumeMock{}

	type args struct {
		ctx  context.Context
		opts v1.ListOptions
	}
	tests := []struct {
		name            string
		args            args
		volumesClient   DirectPVVolumesClientI
		mockListVolumes func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error)
		wantErr         bool
	}{
		{
			name: "Can List Volumes correctly",
			args: args{
				ctx:  context.Background(),
				opts: v1.ListOptions{},
			},
			volumesClient: directPVVolMock,
			mockListVolumes: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error) {
				items := []v1beta4.DirectCSIVolume{}

				returnList := v1beta4.DirectCSIVolumeList{
					Items: items,
				}

				return &returnList, nil
			},
			wantErr: false,
		},
		{
			name: "Drives request from DirectPV is ok but volumes request failed",
			args: args{
				ctx: context.Background(),
			},
			volumesClient: directPVVolMock,
			mockListVolumes: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error) {
				return nil, errors.New("some error occurred")
			},
			wantErr: true,
		},
		{
			name: "Can List Volumes & Drives correctly without any issue",
			args: args{
				ctx:  context.Background(),
				opts: v1.ListOptions{},
			},
			volumesClient: directPVVolMock,
			mockListVolumes: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIVolumeList, error) {
				items := []v1beta4.DirectCSIVolume{{
					Status: v1beta4.DirectCSIVolumeStatus{
						Drive:             "/var/lib/direct-csi/devices/test-part-dev1-directcsi",
						NodeName:          "testNodeName",
						HostPath:          "",
						StagingPath:       "",
						ContainerPath:     "",
						TotalCapacity:     4772382377372,
						AvailableCapacity: 4772382377372,
						UsedCapacity:      0,
						Conditions:        nil,
					},
				}}

				returnList := v1beta4.DirectCSIVolumeList{
					Items: items,
				}

				return &returnList, nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpvClientListMock = tt.mockListVolumes

			_, err := getDirectPVVolumesList(tt.args.ctx, tt.volumesClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaceCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// FormatDrives
func Test_GetDirectPVFormatDrives(t *testing.T) {
	directPVDrvMock := directPVDriveMock{}

	type args struct {
		ctx  context.Context
		opts v1.ListOptions
	}
	tests := []struct {
		name           string
		args           args
		drivesClient   DirectPVDrivesClientI
		mockListDrives func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error)
		drives         []string
		force          bool
		wantErr        bool
	}{
		{
			name: "Format doesn't crash on empty list & returns error",
			args: args{
				ctx:  context.Background(),
				opts: v1.ListOptions{},
			},
			drivesClient: directPVDrvMock,
			mockListDrives: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
				items := []v1beta4.DirectCSIDrive{}

				returnList := v1beta4.DirectCSIDriveList{
					Items: items,
				}

				return &returnList, nil
			},
			drives:  []string{},
			force:   false,
			wantErr: true,
		},
		{
			name: "Can Format Selected drives",
			args: args{
				ctx:  context.Background(),
				opts: v1.ListOptions{},
			},
			drivesClient: directPVDrvMock,
			mockListDrives: func(ctx context.Context, opts v1.ListOptions) (*v1beta4.DirectCSIDriveList, error) {
				items := []v1beta4.DirectCSIDrive{
					{
						Status: v1beta4.DirectCSIDriveStatus{
							Path:              "/var/lib/direct-csi/devices/test-part-dev1-directcsi",
							AllocatedCapacity: 0,
							FreeCapacity:      4772382377372,
							RootPartition:     "/",
							PartitionNum:      0,
							Filesystem:        "nfs",
							Mountpoint:        "",
							NodeName:          "test-dev1-directcsi",
							DriveStatus:       v1beta4.DriveStatusAvailable,
							ModelNumber:       "testModel",
							SerialNumber:      "testSN2",
							TotalCapacity:     4772382377372,
							PhysicalBlockSize: 1024,
							LogicalBlockSize:  1024,
							AccessTier:        "",
							FilesystemUUID:    "",
							PartitionUUID:     "",
							MajorNumber:       0,
							MinorNumber:       0,
							UeventSerial:      "",
							UeventFSUUID:      "",
							WWID:              "",
							Vendor:            "",
							DMName:            "",
							DMUUID:            "",
							MDUUID:            "",
							PartTableUUID:     "",
							PartTableType:     "",
							Virtual:           false,
							ReadOnly:          false,
							Partitioned:       false,
							SwapOn:            false,
							Master:            "",
							OtherMountsInfo:   nil,
							PCIPath:           "",
							SerialNumberLong:  "",
							Conditions:        nil,
						},
					},
				}

				returnList := v1beta4.DirectCSIDriveList{
					Items: items,
				}

				return &returnList, nil
			},
			drives:  []string{"test-dev1-directcsi:/dev/testdev1-directcsi"},
			force:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpdClientListMock = tt.mockListDrives

			_, err := formatDrives(tt.args.ctx, tt.drivesClient, tt.drives, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaceCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
