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

export interface IDirectPVDrives {
  joinName: string;
  drive: string;
  capacity: string;
  allocated: string;
  volumes: number;
  node: string;
  status: "Available" | "Unavailable" | "InUse" | "Ready" | "Terminating";
}

export interface IDirectPVVolumes {
  volume: string;
  capacity: string;
  node: string;
  drive: string;
}

export interface IDrivesResponse {
  drives: IDirectPVDrives[];
}

export interface IVolumesResponse {
  volumes: IDirectPVVolumes[];
}

export interface IDirectPVFormatResult {
  formatIssuesList: IDirectPVFormatResItem[];
}

export interface IDirectPVFormatResItem {
  node: string;
  drive: string;
  error: string;
}
