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
import { fsGroupChangePolicyType } from "../types";
import { SecurityContext } from "../../../../api/operatorApi";

export interface IEditTenantAuditLogging {
  auditLoggingEnabled: boolean;
  image: string;
  labels: IKeyValue[];
  annotations: IKeyValue[];
  nodeSelector: IKeyValue[];
  diskCapacityGB: number;
  serviceAccountName: string;
  dbImage: string;
  dbInitImage: string;
  dbLabels: IKeyValue[];
  dbAnnotations: IKeyValue[];
  dbNodeSelector: IKeyValue[];
  dbServiceAccountName: string;
  cpuRequest: string;
  memRequest: string;
  dbCPURequest: string;
  dbMemRequest: string;
  securityContext: SecurityContext;
  dbSecurityContext: SecurityContext;
  refreshLoggingInfo: boolean;
}

const initialState: IEditTenantAuditLogging = {
  auditLoggingEnabled: false,
  image: "",
  labels: [{ key: " ", value: " " }],
  annotations: [{ key: " ", value: " " }],
  nodeSelector: [{ key: " ", value: " " }],
  diskCapacityGB: 0,
  serviceAccountName: "",
  dbCPURequest: "",
  dbMemRequest: "",
  dbImage: "",
  dbInitImage: "",
  dbLabels: [{ key: " ", value: " " }],
  dbAnnotations: [{ key: " ", value: " " }],
  dbNodeSelector: [{ key: " ", value: " " }],
  dbServiceAccountName: "",
  cpuRequest: "",
  memRequest: "",
  securityContext: {
    runAsUser: "1000",
    runAsGroup: "1000",
    fsGroup: "1000",
    fsGroupChangePolicy: "Always",
    runAsNonRoot: false,
  },
  dbSecurityContext: {
    runAsUser: "1000",
    runAsGroup: "1000",
    fsGroup: "1000",
    fsGroupChangePolicy: "Always",
    runAsNonRoot: false,
  },
  refreshLoggingInfo: true,
};

export const editTenantAuditLoggingSlice = createSlice({
  name: "editTenantAuditLogging",
  initialState,
  reducers: {
    setAuditLoggingEnabled: (state, action: PayloadAction<boolean>) => {
      state.auditLoggingEnabled = action.payload;
    },
    setImage: (state, action: PayloadAction<string>) => {
      state.image = action.payload;
    },
    setDBImage: (state, action: PayloadAction<string>) => {
      state.dbImage = action.payload;
    },
    setDBInitImage: (state, action: PayloadAction<string>) => {
      state.dbInitImage = action.payload;
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
    setDBLabels: (state, action: PayloadAction<IKeyValue[]>) => {
      state.dbLabels = action.payload;
    },
    setDBAnnotations: (state, action: PayloadAction<IKeyValue[]>) => {
      state.dbAnnotations = action.payload;
    },
    setDBNodeSelector: (state, action: PayloadAction<IKeyValue[]>) => {
      state.dbNodeSelector = action.payload;
    },
    setDiskCapacityGB: (state, action: PayloadAction<number>) => {
      state.diskCapacityGB = action.payload;
    },
    setServiceAccountName: (state, action: PayloadAction<string>) => {
      state.serviceAccountName = action.payload;
    },
    setDBServiceAccountName: (state, action: PayloadAction<string>) => {
      state.dbServiceAccountName = action.payload;
    },
    setCPURequest: (state, action: PayloadAction<string>) => {
      state.cpuRequest = action.payload;
    },
    setMemRequest: (state, action: PayloadAction<string>) => {
      state.memRequest = action.payload;
    },
    setDBCPURequest: (state, action: PayloadAction<string>) => {
      state.dbCPURequest = action.payload;
    },
    setDBMemRequest: (state, action: PayloadAction<string>) => {
      state.dbMemRequest = action.payload;
    },
    setRunAsUser: (state, action: PayloadAction<string>) => {
      state.securityContext.runAsUser = action.payload;
    },
    setRunAsGroup: (state, action: PayloadAction<string>) => {
      state.securityContext.runAsGroup = action.payload;
    },
    setFSGroup: (state, action: PayloadAction<string>) => {
      state.securityContext.fsGroup = action.payload;
    },
    setRunAsNonRoot: (state, action: PayloadAction<boolean>) => {
      state.securityContext.runAsNonRoot = action.payload;
    },
    setDBRunAsUser: (state, action: PayloadAction<string>) => {
      state.dbSecurityContext.runAsUser = action.payload;
    },
    setDBRunAsGroup: (state, action: PayloadAction<string>) => {
      state.dbSecurityContext.runAsGroup = action.payload;
    },
    setDBFSGroup: (state, action: PayloadAction<string>) => {
      state.dbSecurityContext.fsGroup = action.payload;
    },
    setDBFSGroupChangePolicy: (
      state,
      action: PayloadAction<fsGroupChangePolicyType>
    ) => {
      state.dbSecurityContext.fsGroupChangePolicy = action.payload;
    },
    setDBRunAsNonRoot: (state, action: PayloadAction<boolean>) => {
      state.dbSecurityContext.runAsNonRoot = action.payload;
    },
    setRefreshLoggingInfo: (state, action: PayloadAction<boolean>) => {
      state.refreshLoggingInfo = action.payload;
    },
    resetAuditLogForm: () => initialState,
  },
});

export const {
  setAuditLoggingEnabled,
  setImage,
  setDBImage,
  setDBInitImage,
  setLabels,
  setAnnotations,
  setNodeSelector,
  setDBLabels,
  setDBAnnotations,
  setDBNodeSelector,
  setDiskCapacityGB,
  setServiceAccountName,
  setDBServiceAccountName,
  setCPURequest,
  setMemRequest,
  setDBCPURequest,
  setDBMemRequest,
  setRunAsUser,
  setFSGroup,
  setRunAsGroup,
  setRunAsNonRoot,
  setDBRunAsUser,
  setDBFSGroup,
  setDBFSGroupChangePolicy,
  setDBRunAsGroup,
  setDBRunAsNonRoot,
  setRefreshLoggingInfo,
  resetAuditLogForm,
} = editTenantAuditLoggingSlice.actions;

export default editTenantAuditLoggingSlice.reducer;
