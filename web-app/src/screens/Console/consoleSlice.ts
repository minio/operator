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
import { ISessionResponse } from "./types";
import { AppState } from "../../store";

export interface ConsoleState {
  session: ISessionResponse;
}

const initialState: ConsoleState = {
  session: {
    operator: false,
    status: "",
    features: [],
    distributedMode: false,
    permissions: {},
    allowResources: null,
    customStyles: null,
    envConstants: null,
    serverEndPoint: "",
  },
};

export const consoleSlice = createSlice({
  name: "console",
  initialState,
  reducers: {
    saveSessionResponse: (state, action: PayloadAction<ISessionResponse>) => {
      state.session = action.payload;
    },
    resetSession: (state) => {
      state.session = initialState.session;
    },
  },
});

export const { saveSessionResponse, resetSession } = consoleSlice.actions;
export const selSession = (state: AppState) => state.console.session;
export const selFeatures = (state: AppState) =>
  state.console.session ? state.console.session.features : [];

export default consoleSlice.reducer;
