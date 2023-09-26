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

import { IAffinityModel, ITolerationModel } from "../../../../common/types";
import { SecurityContext } from "../../../../api/operatorApi";

export interface IEvent {
  namespace: string;
  last_seen: number;
  seen: string;
  message: string;
  event_type: string;
  reason: string;
}

export interface IPodListElement {
  name: string;
  status: string;
  timeCreated: string;
  podIP: string;
  restarts: number;
  node: string;
  time: string;
  namespace?: string;
  tenant?: string;
}

export interface IAddPoolRequest {
  name: string;
  servers: number;
  volumes_per_server: number;
  volume_configuration: IVolumeConfiguration;
  affinity?: IAffinityModel;
  tolerations?: ITolerationModel[];
  securityContext?: SecurityContext | null;
}

export interface IVolumeConfiguration {
  size: number;
  storage_class_name: string;
  labels: { [key: string]: any } | null;
}

export interface IResourcesSize {
  error: string;
  memoryRequest: number;
  memoryLimit: number;
  cpuRequest: number;
  cpuLimit: number;
}

export interface ITenantMonitoringStruct {
  image: string;
  sidecarImage: string;
  initImage: string;
  storageClassName: string;
  labels: IKeyValue[];
  annotations: IKeyValue[];
  nodeSelector: IKeyValue[];
  diskCapacityGB: string;
  serviceAccountName: string;
  prometheusEnabled: boolean;
  monitoringCPURequest: string;
  monitoringMemRequest: string;
  securityContext: SecurityContext;
}

export interface IKeyValue {
  key: string;
  value: string;
}

export interface ITenantMonitoringStruct {
  image: string;
  sidecarImage: string;
  initImage: string;
  storageClassName: string;
  labels: IKeyValue[];
  annotations: IKeyValue[];
  nodeSelector: IKeyValue[];
  diskCapacityGB: string;
  serviceAccountName: string;
  prometheusEnabled: boolean;
}

export interface ValueUnit {
  value: string;
  unit: string;
}

export interface CapacityValues {
  value: number;
  variant: string;
}

export interface CapacityValue {
  value: number;
  label: string;
  color: string;
}
