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

export interface IPVCDescribeProps {
  tenant: string;
  namespace: string;
  pvcName: string;
  propLoading: boolean;
}

export interface Annotation {
  key: string;
  value: string;
}

export interface Label {
  key: string;
  value: string;
}

export interface DescribeResponse {
  annotations: Annotation[];
  labels: Label[];
  name: string;
  namespace: string;
  status: string;
  storageClass: string;
  capacity: string;
  accessModes: string[];
  finalizers: string[];
  volume: string;
  volumeMode: string;
}

export interface IPVCDescribeSummaryProps {
  describeInfo: DescribeResponse;
}

export interface IPVCDescribeAnnotationsProps {
  annotations: Annotation[];
}

export interface IPVCDescribeLabelsProps {
  labels: Label[];
}
