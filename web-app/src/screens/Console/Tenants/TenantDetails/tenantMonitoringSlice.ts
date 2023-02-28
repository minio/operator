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
import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { IKeyValue } from "../ListTenants/types";

export interface IEditTenantMonitoring {
  prometheusEnabled: boolean;
  image: string;
  sidecarImage: string;
  initImage: string;
  storageClassName: string;
  labels: IKeyValue[];
  annotations: IKeyValue[];
  nodeSelector: IKeyValue[];
  diskCapacityGB: string;
  serviceAccountName: string;
  monitoringCPURequest: string;
  monitoringMemRequest: string;
  runAsUser: string;
  runAsGroup: string;
  fsGroup: string;
  runAsNonRoot: boolean;
}

const initialState: IEditTenantMonitoring = {
  prometheusEnabled: false,
  image: "",
  sidecarImage: "",
  initImage: "",
  storageClassName: "",
  labels: [{ key: " ", value: " " }],
  annotations: [{ key: " ", value: " " }],
  nodeSelector: [{ key: " ", value: " " }],
  diskCapacityGB: "0",
  serviceAccountName: "",
  monitoringCPURequest: "",
  monitoringMemRequest: "",
  runAsUser: "1000",
  runAsGroup: "1000",
  fsGroup: "1000",
  runAsNonRoot: true,
};

export const editTenantMonitoringSlice = createSlice({
  name: "editTenantMonitoring",
  initialState,
  reducers: {
    setPrometheusEnabled: (state, action: PayloadAction<boolean>) => {
      state.prometheusEnabled = action.payload;
    },
    setImage: (state, action: PayloadAction<string>) => {
      state.image = action.payload;
    },
    setSidecarImage: (state, action: PayloadAction<string>) => {
      state.sidecarImage = action.payload;
    },
    setInitImage: (state, action: PayloadAction<string>) => {
      state.initImage = action.payload;
    },
    setStorageClassName: (state, action: PayloadAction<string>) => {
      state.storageClassName = action.payload;
    },
    setLabels: (state, action: PayloadAction<IKeyValue[]>) => {
      state.labels = action.payload;
    },
    setAnnotations: (state, action: PayloadAction<IKeyValue[]>) => {
      state.annotations = action.payload;
    },
    setNodeSelector: (state, action: PayloadAction<IKeyValue[]>) => {
      state.nodeSelector = action.payload;
    },
    setDiskCapacityGB: (state, action: PayloadAction<string>) => {
      state.diskCapacityGB = action.payload;
    },
    setServiceAccountName: (state, action: PayloadAction<string>) => {
      state.serviceAccountName = action.payload;
    },
    setCPURequest: (state, action: PayloadAction<string>) => {
      state.monitoringCPURequest = action.payload;
    },
    setMemRequest: (state, action: PayloadAction<string>) => {
      state.monitoringMemRequest = action.payload;
    },
    setRunAsUser: (state, action: PayloadAction<string>) => {
      state.runAsUser = action.payload;
    },
    setRunAsGroup: (state, action: PayloadAction<string>) => {
      state.runAsGroup = action.payload;
    },
    setFSGroup: (state, action: PayloadAction<string>) => {
      state.fsGroup = action.payload;
    },
    setRunAsNonRoot: (state, action: PayloadAction<boolean>) => {
      state.runAsNonRoot = action.payload;
    },
  },
});

export const {
  setPrometheusEnabled,
  setImage,
  setSidecarImage,
  setInitImage,
  setStorageClassName,
  setLabels,
  setAnnotations,
  setNodeSelector,
  setDiskCapacityGB,
  setServiceAccountName,
  setCPURequest,
  setMemRequest,
  setRunAsUser,
  setRunAsGroup,
  setFSGroup,
  setRunAsNonRoot,
} = editTenantMonitoringSlice.actions;

export default editTenantMonitoringSlice.reducer;
