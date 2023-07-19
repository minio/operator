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
import {
  ITolerationEffect,
  ITolerationModel,
  ITolerationOperator,
} from "../../../../../../common/types";
import {
  IAddPoolSetup,
  IPoolConfiguration,
  ITenantAffinity,
  LabelKeyPair,
} from "../../../types";
import { has } from "lodash";
import get from "lodash/get";
import { Opts } from "../../../ListTenants/utils";
import { addPoolAsync } from "./addPoolThunks";

export interface IAddPool {
  addPoolLoading: boolean;
  sending: boolean;
  validPages: string[];
  storageClasses: Opts[];
  limitSize: any;
  setup: IAddPoolSetup;
  affinity: ITenantAffinity;
  configuration: IPoolConfiguration;
  tolerations: ITolerationModel[];
  nodeSelectorPairs: LabelKeyPair[];
  navigateTo: string;
}

const initialState: IAddPool = {
  addPoolLoading: false,
  sending: false,
  validPages: ["affinity", "configure"],
  storageClasses: [],
  limitSize: {},
  navigateTo: "",
  setup: {
    numberOfNodes: 0,
    storageClass: "",
    volumeSize: 0,
    volumesPerServer: 0,
  },
  affinity: {
    nodeSelectorLabels: "",
    podAffinity: "default",
    withPodAntiAffinity: true,
  },
  configuration: {
    securityContextEnabled: false,
    securityContext: {
      runAsUser: "1000",
      runAsGroup: "1000",
      fsGroup: "1000",
      fsGroupChangePolicy: "Always",
      runAsNonRoot: true,
    },
    customRuntime: false,
    runtimeClassName: "",
  },
  nodeSelectorPairs: [{ key: "", value: "" }],
  tolerations: [
    {
      key: "",
      tolerationSeconds: { seconds: 0 },
      value: "",
      effect: ITolerationEffect.NoSchedule,
      operator: ITolerationOperator.Equal,
    },
  ],
};

export const addPoolSlice = createSlice({
  name: "addPool",
  initialState,
  reducers: {
    setPoolLoading: (state, action: PayloadAction<boolean>) => {
      state.addPoolLoading = action.payload;
    },
    setPoolField: (
      state,
      action: PayloadAction<{
        page:
          | "setup"
          | "affinity"
          | "configuration"
          | "tolerations"
          | "nodeSelectorPairs";
        field: string;
        value: any;
      }>,
    ) => {
      if (has(state, `${action.payload.page}.${action.payload.field}`)) {
        const originPageNameItems = get(state, `${action.payload.page}`, {});

        let newValue: any = {};
        newValue[action.payload.field] = action.payload.value;

        state[action.payload.page] = {
          ...originPageNameItems,
          ...newValue,
        };
      }
    },
    isPoolPageValid: (
      state,
      action: PayloadAction<{
        page: string;
        status: boolean;
      }>,
    ) => {
      if (action.payload.status) {
        if (!state.validPages.includes(action.payload.page)) {
          state.validPages.push(action.payload.page);
        }
      } else {
        state.validPages = state.validPages.filter(
          (elm) => elm !== action.payload.page,
        );
      }
    },
    setPoolStorageClasses: (state, action: PayloadAction<Opts[]>) => {
      state.storageClasses = action.payload;
    },
    setPoolTolerationInfo: (
      state,
      action: PayloadAction<{
        index: number;
        tolerationValue: ITolerationModel;
      }>,
    ) => {
      state.tolerations[action.payload.index] = action.payload.tolerationValue;
    },
    addNewPoolToleration: (state) => {
      state.tolerations.push({
        key: "",
        tolerationSeconds: { seconds: 0 },
        value: "",
        effect: ITolerationEffect.NoSchedule,
        operator: ITolerationOperator.Equal,
      });
    },
    removePoolToleration: (state, action: PayloadAction<number>) => {
      state.tolerations = state.tolerations.filter(
        (_, index) => index !== action.payload,
      );
    },
    setPoolKeyValuePairs: (state, action: PayloadAction<LabelKeyPair[]>) => {
      state.nodeSelectorPairs = action.payload;
    },
    resetPoolForm: () => initialState,
  },
  extraReducers: (builder) => {
    builder
      .addCase(addPoolAsync.pending, (state) => {
        state.sending = true;
      })
      .addCase(addPoolAsync.rejected, (state) => {
        state.sending = false;
      })
      .addCase(addPoolAsync.fulfilled, (state, action) => {
        state.sending = false;
        if (action.payload) {
          state.navigateTo = action.payload;
        }
      });
  },
});

export const {
  setPoolLoading,
  resetPoolForm,
  setPoolField,
  isPoolPageValid,
  setPoolStorageClasses,
  setPoolTolerationInfo,
  addNewPoolToleration,
  removePoolToleration,
  setPoolKeyValuePairs,
} = addPoolSlice.actions;

export default addPoolSlice.reducer;
