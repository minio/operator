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
import { ILoginDetails, loginStrategyType } from "./types";
import {
  doLoginAsync,
  getFetchConfigurationAsync,
  getVersionAsync,
} from "./loginThunks";

export interface LoginState {
  accessKey: string;
  secretKey: string;
  sts: string;
  useSTS: boolean;

  jwt: string;

  loginStrategy: ILoginDetails;

  loginSending: boolean;
  loadingFetchConfiguration: boolean;

  latestMinIOVersion: string;
  loadingVersion: boolean;
  isDirectPV: boolean;
  isK8S: boolean;

  navigateTo: string;
}

const initialState: LoginState = {
  accessKey: "",
  secretKey: "",
  sts: "",
  useSTS: false,
  jwt: "",
  loginStrategy: {
    loginStrategy: loginStrategyType.unknown,
    redirectRules: [],
  },
  loginSending: false,
  loadingFetchConfiguration: true,
  latestMinIOVersion: "",
  loadingVersion: true,
  isDirectPV: false,
  isK8S: false,

  navigateTo: "",
};

export const loginSlice = createSlice({
  name: "login",
  initialState,
  reducers: {
    setAccessKey: (state, action: PayloadAction<string>) => {
      state.accessKey = action.payload;
    },
    setSecretKey: (state, action: PayloadAction<string>) => {
      state.secretKey = action.payload;
    },
    setUseSTS: (state, action: PayloadAction<boolean>) => {
      state.useSTS = action.payload;
    },
    setSTS: (state, action: PayloadAction<string>) => {
      state.sts = action.payload;
    },
    setJwt: (state, action: PayloadAction<string>) => {
      state.jwt = action.payload;
    },
    setNavigateTo: (state, action: PayloadAction<string>) => {
      state.navigateTo = action.payload;
    },
    resetForm: (state) => initialState,
  },
  extraReducers: (builder) => {
    builder
      .addCase(getVersionAsync.pending, (state, action) => {
        state.loadingVersion = true;
      })
      .addCase(getVersionAsync.rejected, (state, action) => {
        state.loadingVersion = false;
      })
      .addCase(getVersionAsync.fulfilled, (state, action) => {
        state.loadingVersion = false;
        if (action.payload) {
          state.latestMinIOVersion = action.payload;
        }
      })
      .addCase(getFetchConfigurationAsync.pending, (state, action) => {
        state.loadingFetchConfiguration = true;
      })
      .addCase(getFetchConfigurationAsync.rejected, (state, action) => {
        state.loadingFetchConfiguration = false;
      })
      .addCase(getFetchConfigurationAsync.fulfilled, (state, action) => {
        state.loadingFetchConfiguration = false;
        if (action.payload) {
          state.loginStrategy = action.payload;
          state.isDirectPV = !!action.payload.isDirectPV;
          state.isK8S = !!action.payload.isK8S;
        }
      })
      .addCase(doLoginAsync.pending, (state, action) => {
        state.loginSending = true;
      })
      .addCase(doLoginAsync.rejected, (state, action) => {
        state.loginSending = false;
      })
      .addCase(doLoginAsync.fulfilled, (state, action) => {
        state.loginSending = false;
      });
  },
});

// Action creators are generated for each case reducer function
export const {
  setAccessKey,
  setSecretKey,
  setUseSTS,
  setSTS,
  setJwt,
  setNavigateTo,
  resetForm,
} = loginSlice.actions;

export default loginSlice.reducer;
