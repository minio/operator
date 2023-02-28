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

	"github.com/minio/operator/models"
)

// UsageInfo usage info
type UsageInfo struct {
	Buckets          int64
	Objects          int64
	Usage            int64
	DisksUsage       int64
	Servers          []*models.ServerProperties
	EndpointNotReady bool
	Backend          *models.BackendProperties
}

// GetAdminInfo invokes admin info and returns a parsed `UsageInfo` structure
func GetAdminInfo(ctx context.Context, client MinioAdmin) (*UsageInfo, error) {
	serverInfo, err := client.serverInfo(ctx)
	if err != nil {
		return nil, err
	}
	// we are trimming uint64 to int64 this will report an incorrect measurement for numbers greater than
	// 9,223,372,036,854,775,807

	var backendType string
	var rrSCParity float64
	var standardSCParity float64

	if v, success := serverInfo.Backend.(map[string]interface{}); success {
		bt, ok := v["backendType"]
		if ok {
			backendType = bt.(string)
		}
		rp, ok := v["rrSCParity"]
		if ok {
			rrSCParity = rp.(float64)
		}
		sp, ok := v["standardSCParity"]
		if ok {
			standardSCParity = sp.(float64)
		}
	}

	var usedSpace int64
	// serverArray contains the serverProperties which describe the servers in the network
	var serverArray []*models.ServerProperties
	for _, serv := range serverInfo.Servers {
		drives := []*models.ServerDrives{}

		for _, drive := range serv.Disks {
			usedSpace += int64(drive.UsedSpace)
			drives = append(drives, &models.ServerDrives{
				State:          drive.State,
				UUID:           drive.UUID,
				Endpoint:       drive.Endpoint,
				RootDisk:       drive.RootDisk,
				DrivePath:      drive.DrivePath,
				Healing:        drive.Healing,
				Model:          drive.Model,
				TotalSpace:     int64(drive.TotalSpace),
				UsedSpace:      int64(drive.UsedSpace),
				AvailableSpace: int64(drive.AvailableSpace),
			})
		}

		newServer := &models.ServerProperties{
			State:      serv.State,
			Endpoint:   serv.Endpoint,
			Uptime:     serv.Uptime,
			Version:    serv.Version,
			CommitID:   serv.CommitID,
			PoolNumber: int64(serv.PoolNumber),
			Network:    serv.Network,
			Drives:     drives,
		}

		serverArray = append(serverArray, newServer)
	}

	backendData := &models.BackendProperties{
		BackendType:      backendType,
		RrSCParity:       int64(rrSCParity),
		StandardSCParity: int64(standardSCParity),
	}

	return &UsageInfo{
		Buckets:    int64(serverInfo.Buckets.Count),
		Objects:    int64(serverInfo.Objects.Count),
		Usage:      int64(serverInfo.Usage.Size),
		DisksUsage: usedSpace,
		Servers:    serverArray,
		Backend:    backendData,
	}, nil
}
