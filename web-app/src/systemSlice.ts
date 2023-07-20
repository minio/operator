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
import { snackBarMessage, SRInfoStateType } from "./types";
import { ErrorResponseHandler, IEmbeddedCustomStyles } from "./common/types";
import { AppState } from "./store";
import { SubnetInfo } from "./screens/Console/License/types";

// determine whether we have the sidebar state stored on localstorage
const initSideBarOpen = localStorage.getItem("sidebarOpen")
  ? JSON.parse(localStorage.getItem("sidebarOpen")!)["open"]
  : true;

export interface SystemState {
  value: number;
  loggedIn: boolean;
  showMarketplace: boolean;
  operatorMode: boolean;
  sidebarOpen: boolean;
  session: string;
  userName: string;
  serverNeedsRestart: boolean;
  serverIsLoading: boolean;
  loadingConfigurations: boolean;
  loadingProgress: number;
  snackBar: snackBarMessage;
  modalSnackBar: snackBarMessage;
  serverDiagnosticStatus: string;
  distributedSetup: boolean;
  siteReplicationInfo: SRInfoStateType;
  licenseInfo: null | SubnetInfo;
  overrideStyles: null | IEmbeddedCustomStyles;
  anonymousMode: boolean;
}

const initialState: SystemState = {
  value: 0,
  loggedIn: false,
  showMarketplace: false,
  operatorMode: false,
  session: "",
  userName: "",
  sidebarOpen: initSideBarOpen,
  siteReplicationInfo: { siteName: "", curSite: false, enabled: false },
  serverNeedsRestart: false,
  serverIsLoading: false,
  loadingConfigurations: true,
  loadingProgress: 100,
  snackBar: {
    message: "",
    detailedErrorMsg: "",
    type: "message",
  },
  modalSnackBar: {
    message: "",
    detailedErrorMsg: "",
    type: "message",
  },
  serverDiagnosticStatus: "",
  distributedSetup: false,
  licenseInfo: null,
  overrideStyles: null,
  anonymousMode: false,
};

export const systemSlice = createSlice({
  name: "system",
  initialState,
  reducers: {
    userLogged: (state, action: PayloadAction<boolean>) => {
      state.loggedIn = action.payload;
    },
    showMarketplace: (state, action: PayloadAction<boolean>) => {
      state.showMarketplace = action.payload;
    },
    operatorMode: (state, action: PayloadAction<boolean>) => {
      state.operatorMode = action.payload;
    },
    menuOpen: (state, action: PayloadAction<boolean>) => {
      // persist preference to local storage
      localStorage.setItem(
        "sidebarOpen",
        JSON.stringify({ open: action.payload }),
      );
      state.sidebarOpen = action.payload;
    },
    setServerNeedsRestart: (state, action: PayloadAction<boolean>) => {
      state.serverNeedsRestart = action.payload;
    },
    serverIsLoading: (state, action: PayloadAction<boolean>) => {
      state.serverIsLoading = action.payload;
    },
    configurationIsLoading: (state, action: PayloadAction<boolean>) => {
      state.loadingConfigurations = action.payload;
    },
    setLoadingProgress: (state, action: PayloadAction<number>) => {
      state.loadingProgress = action.payload;
    },
    setSnackBarMessage: (state, action: PayloadAction<string>) => {
      state.snackBar = {
        message: action.payload,
        detailedErrorMsg: "",
        type: "message",
      };
    },
    setErrorSnackMessage: (
      state,
      action: PayloadAction<ErrorResponseHandler>,
    ) => {
      state.snackBar = {
        message: action.payload.errorMessage,
        detailedErrorMsg: action.payload.detailedError,
        type: "error",
      };
    },
    setModalSnackMessage: (state, action: PayloadAction<string>) => {
      state.modalSnackBar = {
        message: action.payload,
        detailedErrorMsg: "",
        type: "message",
      };
    },
    setModalErrorSnackMessage: (
      state,
      action: PayloadAction<{ errorMessage: string; detailedError: string }>,
    ) => {
      state.modalSnackBar = {
        message: action.payload.errorMessage,
        detailedErrorMsg: action.payload.detailedError,
        type: "error",
      };
    },
    setServerDiagStat: (state, action: PayloadAction<string>) => {
      state.serverDiagnosticStatus = action.payload;
    },
    globalSetDistributedSetup: (state, action: PayloadAction<boolean>) => {
      state.distributedSetup = action.payload;
    },
    setSiteReplicationInfo: (state, action: PayloadAction<SRInfoStateType>) => {
      state.siteReplicationInfo = action.payload;
    },
    setLicenseInfo: (state, action: PayloadAction<SubnetInfo | null>) => {
      state.licenseInfo = action.payload;
    },
    setOverrideStyles: (
      state,
      action: PayloadAction<IEmbeddedCustomStyles>,
    ) => {
      state.overrideStyles = action.payload;
    },
    setAnonymousMode: (state) => {
      state.anonymousMode = true;
      state.loggedIn = true;
    },
    resetSystem: () => {
      return initialState;
    },
  },
});

// Action creators are generated for each case reducer function
export const {
  userLogged,
  showMarketplace,
  operatorMode,
  menuOpen,
  setServerNeedsRestart,
  serverIsLoading,
  setLoadingProgress,
  setSnackBarMessage,
  setErrorSnackMessage,
  setModalErrorSnackMessage,
  setModalSnackMessage,
  setServerDiagStat,
  globalSetDistributedSetup,
  setSiteReplicationInfo,
  setLicenseInfo,
  setOverrideStyles,
  setAnonymousMode,
  resetSystem,
  configurationIsLoading,
} = systemSlice.actions;

export const selOpMode = (state: AppState) => state.system.operatorMode;
export const selShowMarketplace = (state: AppState) =>
  state.system.showMarketplace;

export default systemSlice.reducer;
