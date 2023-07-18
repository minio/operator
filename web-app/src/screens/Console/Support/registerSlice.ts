// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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
import { SubnetInfo, SubnetOrganization } from "../License/types";

export interface RegisterState {
  license: string;
  subnetPassword: string;
  subnetEmail: string;
  subnetMFAToken: string;
  subnetOTP: string;
  subnetAccessToken: string;
  selectedSubnetOrganization: string;
  subnetRegToken: string;
  subnetOrganizations: SubnetOrganization[];
  showPassword: boolean;
  loading: boolean;
  loadingLicenseInfo: boolean;
  clusterRegistered: boolean;
  licenseInfo: SubnetInfo | undefined;
  curTab: number;
}

const initialState: RegisterState = {
  license: "",
  subnetPassword: "",
  subnetEmail: "",
  subnetMFAToken: "",
  subnetOTP: "",
  subnetAccessToken: "",
  selectedSubnetOrganization: "",
  subnetRegToken: "",
  subnetOrganizations: [],
  showPassword: false,
  loading: false,
  loadingLicenseInfo: false,
  clusterRegistered: false,
  licenseInfo: undefined,
  curTab: 0,
};

export const registerSlice = createSlice({
  name: "register",
  initialState,
  reducers: {
    setLicense: (state, action: PayloadAction<string>) => {
      state.license = action.payload;
    },
    setSubnetPassword: (state, action: PayloadAction<string>) => {
      state.subnetPassword = action.payload;
    },
    setSubnetEmail: (state, action: PayloadAction<string>) => {
      state.subnetEmail = action.payload;
    },
    setSubnetMFAToken: (state, action: PayloadAction<string>) => {
      state.subnetMFAToken = action.payload;
    },
    setSubnetOTP: (state, action: PayloadAction<string>) => {
      state.subnetOTP = action.payload;
    },
    setSubnetAccessToken: (state, action: PayloadAction<string>) => {
      state.subnetAccessToken = action.payload;
    },
    setSelectedSubnetOrganization: (state, action: PayloadAction<string>) => {
      state.selectedSubnetOrganization = action.payload;
    },
    setSubnetRegToken: (state, action: PayloadAction<string>) => {
      state.subnetRegToken = action.payload;
    },
    setSubnetOrganizations: (
      state,
      action: PayloadAction<SubnetOrganization[]>,
    ) => {
      state.subnetOrganizations = action.payload;
    },
    setShowPassword: (state, action: PayloadAction<boolean>) => {
      state.showPassword = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setLoadingLicenseInfo: (state, action: PayloadAction<boolean>) => {
      state.loadingLicenseInfo = action.payload;
    },
    setClusterRegistered: (state, action: PayloadAction<boolean>) => {
      state.clusterRegistered = action.payload;
    },
    setLicenseInfo: (state, action: PayloadAction<SubnetInfo | undefined>) => {
      state.licenseInfo = action.payload;
    },
    setCurTab: (state, action: PayloadAction<number>) => {
      state.curTab = action.payload;
    },
    resetRegisterForm: () => initialState,
  },
});

// Action creators are generated for each case reducer function
export const {
  setLicense,
  setSubnetPassword,
  setSubnetEmail,
  setSubnetMFAToken,
  setSubnetOTP,
  setSubnetAccessToken,
  setSelectedSubnetOrganization,
  setSubnetRegToken,
  setSubnetOrganizations,
  setShowPassword,
  setLoading,
  setLoadingLicenseInfo,
  setClusterRegistered,
  setLicenseInfo,
  setCurTab,
  resetRegisterForm,
} = registerSlice.actions;

export default registerSlice.reducer;
