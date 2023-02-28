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
	"sort"
	"strings"

	"github.com/minio/directpv/pkg/utils"

	"github.com/go-openapi/runtime/middleware"
	directcsi "github.com/minio/directpv/pkg/apis/direct.csi.min.io/v1beta4"
	"github.com/minio/directpv/pkg/sys"
	"github.com/minio/operator/api/operations"
	"github.com/minio/operator/api/operations/operator_api"
	"github.com/minio/operator/cluster"
	"github.com/minio/operator/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// XFS filesystem
const XFS = "xfs"

func registerDirectPVHandlers(api *operations.OperatorAPI) {
	api.OperatorAPIGetDirectPVDriveListHandler = operator_api.GetDirectPVDriveListHandlerFunc(func(params operator_api.GetDirectPVDriveListParams, session *models.Principal) middleware.Responder {
		resp, err := getDirectPVDrivesListResponse(session)
		if err != nil {
			return operator_api.NewGetDirectPVDriveListDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetDirectPVDriveListOK().WithPayload(resp)
	})
	api.OperatorAPIGetDirectPVVolumeListHandler = operator_api.GetDirectPVVolumeListHandlerFunc(func(params operator_api.GetDirectPVVolumeListParams, session *models.Principal) middleware.Responder {
		resp, err := getDirectPVVolumesListResponse(session)
		if err != nil {
			return operator_api.NewGetDirectPVVolumeListDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewGetDirectPVVolumeListOK().WithPayload(resp)
	})
	api.OperatorAPIDirectPVFormatDriveHandler = operator_api.DirectPVFormatDriveHandlerFunc(func(params operator_api.DirectPVFormatDriveParams, session *models.Principal) middleware.Responder {
		resp, err := formatVolumesResponse(session, params)
		if err != nil {
			return operator_api.NewDirectPVFormatDriveDefault(int(err.Code)).WithPayload(err)
		}
		return operator_api.NewDirectPVFormatDriveOK().WithPayload(resp)
	})
}

// getDirectPVVolumesList returns directPV drives
func getDirectPVDriveList(ctx context.Context, driveInterface DirectPVDrivesClientI) (*models.GetDirectPVDriveListResponse, error) {
	drivesList, err := driveInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	res := &models.GetDirectPVDriveListResponse{}

	// implementation same as directPV `drives ls` command
	driveName := func(val string) string {
		dr := strings.ReplaceAll(val, sys.DirectCSIDevRoot+"/", "")
		dr = strings.ReplaceAll(dr, sys.HostDevRoot+"/", "")
		return strings.ReplaceAll(dr, sys.DirectCSIPartitionInfix, "")
	}
	drivesSorted := drivesList.Items
	// sort by nodename, path and status
	sort.Slice(drivesSorted, func(i, j int) bool {
		d1 := drivesSorted[i]
		d2 := drivesSorted[j]

		if v := strings.Compare(d1.Status.NodeName, d2.Status.NodeName); v != 0 {
			return v < 0
		}

		if v := strings.Compare(d1.Status.Path, d2.Status.Path); v != 0 {
			return v < 0
		}

		return strings.Compare(string(d1.Status.DriveStatus), string(d2.Status.DriveStatus)) < 0
	})

	for _, d := range drivesSorted {
		volumes := 0

		if len(d.Finalizers) > 1 {
			volumes = len(d.Finalizers) - 1
		}

		dr := driveName(d.Status.Path)
		dr = strings.ReplaceAll("/dev/"+dr, sys.DirectCSIPartitionInfix, "")

		status := d.Status.DriveStatus
		msg := ""
		for _, c := range d.Status.Conditions {
			switch c.Type {
			case string(directcsi.DirectCSIDriveConditionInitialized), string(directcsi.DirectCSIDriveConditionOwned), string(directcsi.DirectCSIDriveConditionReady):
				if c.Status != metav1.ConditionTrue {
					msg = c.Message
					if msg != "" {
						status = d.Status.DriveStatus + "*"
						msg = strings.ReplaceAll(msg, d.Name, "")
						msg = strings.ReplaceAll(msg, sys.DirectCSIDevRoot, "/dev")
						msg = strings.ReplaceAll(msg, sys.DirectCSIPartitionInfix, "")
						msg = strings.Split(msg, "\n")[0]
					}
				}
			}
		}

		var allocatedCapacity int64
		if status == directcsi.DriveStatusInUse {
			allocatedCapacity = d.Status.AllocatedCapacity
		}

		drStatus := d.Status.DriveStatus

		driveInfo := &models.DirectPVDriveInfo{
			Drive:     dr,
			Capacity:  d.Status.TotalCapacity,
			Allocated: allocatedCapacity,
			Node:      d.Status.NodeName,
			Status:    string(drStatus),
			Message:   msg,
			Volumes:   int64(volumes),
		}
		res.Drives = append(res.Drives, driveInfo)
	}

	return res, nil
}

func getDirectPVDrivesListResponse(session *models.Principal) (*models.GetDirectPVDriveListResponse, *models.Error) {
	ctx := context.Background()

	driveInterface, err := cluster.DirectPVDriveInterface(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	directPVDrvClient := &directPVDrivesClient{
		client: driveInterface,
	}

	drives, err := getDirectPVDriveList(ctx, directPVDrvClient)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return drives, nil
}

// getDirectPVVolumesList returns directPV volumes
func getDirectPVVolumesList(ctx context.Context, volumeInterface DirectPVVolumesClientI) (*models.GetDirectPVVolumeListResponse, error) {
	volList, err := volumeInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	driveName := func(val string) string {
		dr := strings.ReplaceAll(val, sys.DirectCSIDevRoot+"/", "")
		return strings.ReplaceAll(dr, sys.HostDevRoot+"/", "")
	}

	getLabelValue := func(obj metav1.Object, key string) string {
		if labels := obj.GetLabels(); labels != nil {
			return labels[key]
		}
		return ""
	}

	var volumes []*models.DirectPVVolumeInfo
	for _, v := range volList.Items {
		vol := &models.DirectPVVolumeInfo{
			Volume:   v.Name,
			Capacity: v.Status.TotalCapacity,
			Drive:    driveName(getLabelValue(&v, string(utils.DrivePathLabelKey))),
			Node:     v.Status.NodeName,
		}

		volumes = append(volumes, vol)
	}

	res := &models.GetDirectPVVolumeListResponse{
		Volumes: volumes,
	}
	return res, nil
}

func getDirectPVVolumesListResponse(session *models.Principal) (*models.GetDirectPVVolumeListResponse, *models.Error) {
	ctx := context.Background()

	volumeInterface, err := cluster.DirectPVVolumeInterface(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	directPVVolClient := &directPVVolumesClient{
		client: volumeInterface,
	}

	volumes, err := getDirectPVVolumesList(ctx, directPVVolClient)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}
	return volumes, nil
}

func formatDrives(ctx context.Context, driveInterface DirectPVDrivesClientI, drives []string, force bool) (*models.FormatDirectPVDrivesResponse, error) {
	if len(drives) == 0 {
		return nil, errors.New("at least one drive needs to be set")
	}

	driveList, err := driveInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	driveName := func(val string) string {
		dr := strings.ReplaceAll(val, sys.DirectCSIDevRoot+"/", "")
		dr = strings.ReplaceAll(dr, sys.HostDevRoot+"/", "")
		return strings.ReplaceAll(dr, sys.DirectCSIPartitionInfix, "")
	}

	drivesArray := map[string]string{}

	for _, driveFromAPI := range drives {
		drivesArray[driveFromAPI] = driveFromAPI
	}

	if len(driveList.Items) == 0 {
		return nil, errors.New("no resources found globally")
	}

	var errorResponses []*models.PvFormatErrorResponse

	for _, driveItem := range driveList.Items {
		drName := "/dev/" + driveName(driveItem.Status.Path)
		driveName := driveItem.Status.NodeName + ":" + drName

		base := &models.PvFormatErrorResponse{
			Node:  driveItem.Status.NodeName,
			Drive: drName,
			Error: "",
		}

		// Element is requested to be formatted
		if _, ok := drivesArray[driveName]; ok {
			if driveItem.Status.DriveStatus == directcsi.DriveStatusUnavailable {
				base.Error = "Status is unavailable"
				errorResponses = append(errorResponses, base)
				continue
			}

			if driveItem.Status.DriveStatus == directcsi.DriveStatusInUse {
				base.Error = "Drive in use. Cannot be formatted"
				errorResponses = append(errorResponses, base)
				continue
			}

			if driveItem.Status.DriveStatus == directcsi.DriveStatusReady {
				base.Error = "Drive already owned and managed."
				errorResponses = append(errorResponses, base)
				continue
			}
			if driveItem.Status.Filesystem != "" && !force {
				base.Error = "Drive already has a fs. Use force to overwrite"
				errorResponses = append(errorResponses, base)
				continue
			}

			if driveItem.Status.DriveStatus == directcsi.DriveStatusReleased {
				base.Error = "Drive is in 'released state'. Please wait until it becomes available"
				errorResponses = append(errorResponses, base)
				continue
			}

			// Validation passes, we request format
			driveItem.Spec.DirectCSIOwned = true
			driveItem.Spec.RequestedFormat = &directcsi.RequestedFormat{
				Filesystem: XFS,
				Force:      force,
			}

			_, err := driveInterface.Update(ctx, &driveItem, metav1.UpdateOptions{})
			if err != nil {
				base.Error = err.Error()
				errorResponses = append(errorResponses, base)
			}
		}
	}

	returnErrors := &models.FormatDirectPVDrivesResponse{
		FormatIssuesList: errorResponses,
	}

	return returnErrors, nil
}

func formatVolumesResponse(session *models.Principal, params operator_api.DirectPVFormatDriveParams) (*models.FormatDirectPVDrivesResponse, *models.Error) {
	ctx := context.Background()

	driveInterface, err := cluster.DirectPVDriveInterface(session.STSSessionToken)
	if err != nil {
		return nil, ErrorWithContext(ctx, err)
	}

	directPVDrvClient := &directPVDrivesClient{
		client: driveInterface,
	}

	formatResult, errFormat := formatDrives(ctx, directPVDrvClient, params.Body.Drives, *params.Body.Force)
	if errFormat != nil {
		return nil, ErrorWithContext(ctx, errFormat)
	}
	return formatResult, nil
}
