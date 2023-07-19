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
import { IEditPool, IEditPoolFields, PageFieldValue } from "./types";
import {
  ITolerationEffect,
  ITolerationModel,
  ITolerationOperator,
} from "../../../../../../common/types";
import { fsGroupChangePolicyType, LabelKeyPair } from "../../../types";
import { has } from "lodash";
import get from "lodash/get";
import { Opts } from "../../../ListTenants/utils";
import { editPoolAsync } from "./thunks/editPoolAsync";
import { Pool } from "../../../../../../api/operatorApi";

const initialState: IEditPool = {
  editPoolLoading: false,
  validPages: ["setup", "affinity", "configure"],
  storageClasses: [],
  limitSize: {},
  fields: {
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
  },
  editSending: false,
  navigateTo: "",
};

export const editPoolSlice = createSlice({
  name: "editPool",
  initialState,
  reducers: {
    setInitialPoolDetails: (state, action: PayloadAction<Pool>) => {
      let podAffinity: "default" | "nodeSelector" | "none" = "none";
      let withPodAntiAffinity = false;
      let nodeSelectorLabels = "";
      let tolerations: ITolerationModel[] = [
        {
          key: "",
          tolerationSeconds: { seconds: 0 },
          value: "",
          effect: ITolerationEffect.NoSchedule,
          operator: ITolerationOperator.Equal,
        },
      ];
      let nodeSelectorPairs: LabelKeyPair[] = [{ key: "", value: "" }];

      if (action.payload.affinity?.nodeAffinity) {
        podAffinity = "nodeSelector";
        if (action.payload.affinity?.podAntiAffinity) {
          withPodAntiAffinity = true;
        }
      } else if (action.payload.affinity?.podAntiAffinity) {
        podAffinity = "default";
      }

      if (action.payload.affinity?.nodeAffinity) {
        let labelItems: string[] = [];
        nodeSelectorPairs = [];

        action.payload.affinity?.nodeAffinity?.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms.forEach(
          (labels) => {
            labels.matchExpressions?.forEach((exp) => {
              labelItems.push(`${exp.key}=${exp.values?.join(",")}`);
              nodeSelectorPairs.push({
                key: exp.key,
                value: exp.values?.join(", ")!,
              });
            });
          },
        );
        nodeSelectorLabels = labelItems.join("&");
      }

      let securityContextOption = false;

      if (action.payload.securityContext) {
        securityContextOption =
          !!action.payload.securityContext.runAsUser ||
          !!action.payload.securityContext.runAsGroup ||
          !!action.payload.securityContext.fsGroup;
      }

      if (action.payload.tolerations) {
        tolerations = action.payload.tolerations?.map((toleration) => {
          const tolerationItem: ITolerationModel = {
            key: toleration.key!,
            tolerationSeconds: toleration.tolerationSeconds,
            value: toleration.value,
            effect: toleration.effect as ITolerationEffect,
            operator: toleration.operator! as ITolerationOperator,
          };
          return tolerationItem;
        });
      }

      const volSizeVars = action.payload.volume_configuration.size / 1073741824;

      const newPoolInfoFields: IEditPoolFields = {
        setup: {
          numberOfNodes: action.payload.servers,
          storageClass: action.payload.volume_configuration.storage_class_name!,
          volumeSize: volSizeVars,
          volumesPerServer: action.payload.volumes_per_server,
        },
        configuration: {
          securityContextEnabled: securityContextOption,
          securityContext: {
            runAsUser: action.payload.securityContext?.runAsUser || "",
            runAsGroup: action.payload.securityContext?.runAsGroup || "",
            fsGroup: action.payload.securityContext?.fsGroup || "",
            fsGroupChangePolicy:
              (action.payload.securityContext
                ?.fsGroupChangePolicy as fsGroupChangePolicyType) || "Always",
            runAsNonRoot: !!action.payload.securityContext?.runAsNonRoot,
          },
          customRuntime: !!action.payload.runtimeClassName,
          runtimeClassName: action.payload.runtimeClassName!,
        },
        affinity: {
          podAffinity,
          withPodAntiAffinity,
          nodeSelectorLabels,
        },
        tolerations,
        nodeSelectorPairs,
      };

      state.fields = {
        ...state.fields,
        ...newPoolInfoFields,
      };
    },
    setEditPoolLoading: (state, action: PayloadAction<boolean>) => {
      state.editPoolLoading = action.payload;
    },
    setEditPoolField: (state, action: PayloadAction<PageFieldValue>) => {
      if (has(state.fields, `${action.payload.page}.${action.payload.field}`)) {
        const originPageNameItems = get(
          state.fields,
          `${action.payload.page}`,
          {},
        );

        let newValue: any = {};
        newValue[action.payload.field] = action.payload.value;

        const joinValue = { ...originPageNameItems, ...newValue };

        state.fields[action.payload.page] = { ...joinValue };
      }
    },
    isEditPoolPageValid: (
      state,
      action: PayloadAction<{
        page: string;
        status: boolean;
      }>,
    ) => {
      const edPoolPV = [...state.validPages];

      if (action.payload.status) {
        if (!edPoolPV.includes(action.payload.page)) {
          edPoolPV.push(action.payload.page);

          state.validPages = [...edPoolPV];
        }
      } else {
        const newSetOfPages = edPoolPV.filter(
          (elm) => elm !== action.payload.page,
        );

        state.validPages = [...newSetOfPages];
      }
    },
    setEditPoolStorageClasses: (state, action: PayloadAction<Opts[]>) => {
      state.storageClasses = action.payload;
    },
    setEditPoolTolerationInfo: (
      state,
      action: PayloadAction<{
        index: number;
        tolerationValue: ITolerationModel;
      }>,
    ) => {
      const editPoolTolerationValue = [...state.fields.tolerations];

      editPoolTolerationValue[action.payload.index] =
        action.payload.tolerationValue;
      state.fields.tolerations = editPoolTolerationValue;
    },
    addNewEditPoolToleration: (state) => {
      state.fields.tolerations.push({
        key: "",
        tolerationSeconds: { seconds: 0 },
        value: "",
        effect: ITolerationEffect.NoSchedule,
        operator: ITolerationOperator.Equal,
      });
    },
    removeEditPoolToleration: (state, action: PayloadAction<number>) => {
      state.fields.tolerations = state.fields.tolerations.filter(
        (_, index) => index !== action.payload,
      );
    },
    setEditPoolKeyValuePairs: (
      state,
      action: PayloadAction<LabelKeyPair[]>,
    ) => {
      state.fields.nodeSelectorPairs = action.payload;
    },
    resetEditPoolForm: () => initialState,
  },
  extraReducers: (builder) => {
    builder
      .addCase(editPoolAsync.pending, (state, action) => {
        state.editSending = true;
      })
      .addCase(editPoolAsync.rejected, (state, action) => {
        state.editSending = false;
      })
      .addCase(editPoolAsync.fulfilled, (state, action) => {
        state.editSending = false;
        if (action.payload) {
          state.navigateTo = action.payload;
        }
      });
  },
});

export const {
  setInitialPoolDetails,
  setEditPoolLoading,
  resetEditPoolForm,
  setEditPoolField,
  isEditPoolPageValid,
  setEditPoolStorageClasses,
  setEditPoolTolerationInfo,
  addNewEditPoolToleration,
  removeEditPoolToleration,
  setEditPoolKeyValuePairs,
} = editPoolSlice.actions;

export default editPoolSlice.reducer;
