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
import { getTenantAsync } from "./thunks/tenantDetailsAsync";
import { Tenant } from "../../../api/operatorApi";

export interface FileValue {
  fileName: string;
  value: string;
}

export interface KeyFileValue {
  key: string;
  fileName: string;
  value: string;
}

export interface CertificateFile {
  id: string;
  key: string;
  fileName: string;
  value: string;
}

export interface ITenantState {
  currentTenant: string;
  currentNamespace: string;
  loadingTenant: boolean;
  tenantInfo: Tenant | null;
  currentTab: string;
  poolDetailsOpen: boolean;
  selectedPool: string | null;
}

const initialState: ITenantState = {
  currentTenant: "",
  currentNamespace: "",
  loadingTenant: false,
  tenantInfo: null,
  currentTab: "summary",
  selectedPool: null,
  poolDetailsOpen: false,
};

export const tenantSlice = createSlice({
  name: "tenant",
  initialState,
  reducers: {
    setTenantDetailsLoad: (state, action: PayloadAction<boolean>) => {
      state.loadingTenant = action.payload;
    },
    setTenantName: (
      state,
      action: PayloadAction<{
        name: string;
        namespace: string;
      }>,
    ) => {
      state.currentTenant = action.payload.name;
      state.currentNamespace = action.payload.namespace;
    },
    setTenantInfo: (state, action: PayloadAction<Tenant | null>) => {
      if (action.payload) {
        state.tenantInfo = action.payload;
      }
    },
    setTenantTab: (state, action: PayloadAction<string>) => {
      state.currentTab = action.payload;
    },

    setSelectedPool: (state, action: PayloadAction<string | null>) => {
      state.selectedPool = action.payload;
    },
    setOpenPoolDetails: (state, action: PayloadAction<boolean>) => {
      state.poolDetailsOpen = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(getTenantAsync.pending, (state) => {
        state.loadingTenant = true;
      })
      .addCase(getTenantAsync.rejected, (state) => {
        state.loadingTenant = false;
      })
      .addCase(getTenantAsync.fulfilled, (state, action) => {
        state.loadingTenant = false;
        state.tenantInfo = action.payload;
      });
  },
});

// Action creators are generated for each case reducer function
export const {
  setTenantDetailsLoad,
  setTenantName,
  setTenantInfo,
  setTenantTab,
  setSelectedPool,
  setOpenPoolDetails,
} = tenantSlice.actions;

export default tenantSlice.reducer;
