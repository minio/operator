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

import { createAsyncThunk } from "@reduxjs/toolkit";
import { AppState } from "../../../../../../store";
import { IAddPoolRequest } from "../../../ListTenants/types";
import { generatePoolName } from "../../../../../../common/utils";
import { ErrorResponseHandler } from "../../../../../../common/types";
import { setErrorSnackMessage } from "../../../../../../systemSlice";
import { getDefaultAffinity, getNodeSelector } from "../../utils";
import { resetPoolForm } from "./addPoolSlice";
import { getTenantAsync } from "../../../thunks/tenantDetailsAsync";
import api from "../../../../../../common/api";

export const addPoolAsync = createAsyncThunk(
  "addPool/addPoolAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const tenant = state.tenants.tenantInfo;
    const selectedStorageClass = state.addPool.setup.storageClass;
    const numberOfNodes = state.addPool.setup.numberOfNodes;
    const volumeSize = state.addPool.setup.volumeSize;
    const volumesPerServer = state.addPool.setup.volumesPerServer;
    const affinityType = state.addPool.affinity.podAffinity;
    const nodeSelectorLabels = state.addPool.affinity.nodeSelectorLabels;
    const withPodAntiAffinity = state.addPool.affinity.withPodAntiAffinity;
    const tolerations = state.addPool.tolerations;
    const securityContextEnabled =
      state.addPool.configuration.securityContextEnabled;
    const securityContext = state.addPool.configuration.securityContext;
    const customRuntime = state.addPool.configuration.customRuntime;
    const runtimeClassName = state.addPool.configuration.runtimeClassName;

    if (tenant === null) {
      return;
    }

    const poolName = generatePoolName(tenant.pools!);

    let affinityObject = {};

    switch (affinityType) {
      case "default":
        affinityObject = {
          affinity: getDefaultAffinity(tenant.name!, poolName),
        };
        break;
      case "nodeSelector":
        affinityObject = {
          affinity: getNodeSelector(
            nodeSelectorLabels,
            withPodAntiAffinity,
            tenant.name!,
            poolName,
          ),
        };
        break;
    }

    const tolerationValues = tolerations.filter(
      (toleration) => toleration.key.trim() !== "",
    );

    let runtimeClass = {};

    if (customRuntime) {
      runtimeClass = {
        runtimeClassName,
      };
    }

    const data: IAddPoolRequest = {
      name: poolName,
      servers: numberOfNodes,
      volumes_per_server: volumesPerServer,
      volume_configuration: {
        size: volumeSize * 1073741824,
        storage_class_name: selectedStorageClass,
        labels: null,
      },
      tolerations: tolerationValues,
      securityContext: securityContextEnabled ? securityContext : null,
      ...affinityObject,
      ...runtimeClass,
    };
    const poolsURL = `/namespaces/${tenant?.namespace || ""}/tenants/${
      tenant?.name || ""
    }/pools`;
    return api
      .invoke(
        "POST",
        `/api/v1/namespaces/${tenant.namespace}/tenants/${tenant.name}/pools`,
        data,
      )
      .then(() => {
        dispatch(resetPoolForm());
        dispatch(getTenantAsync());
        return poolsURL;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  },
);
