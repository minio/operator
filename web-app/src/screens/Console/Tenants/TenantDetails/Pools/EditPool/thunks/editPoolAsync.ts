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
import { AppState } from "../../../../../../../store";
import { ErrorResponseHandler } from "../../../../../../../common/types";
import { setErrorSnackMessage } from "../../../../../../../systemSlice";
import { generatePoolName } from "../../../../../../../common/utils";
import { getDefaultAffinity, getNodeSelector } from "../../../utils";
import { resetEditPoolForm } from "../editPoolSlice";
import { getTenantAsync } from "../../../../thunks/tenantDetailsAsync";
import { api } from "../../../../../../../api";
import {
  Pool,
  PoolUpdateRequest,
  SecurityContext,
} from "../../../../../../../api/operatorApi";

export const editPoolAsync = createAsyncThunk(
  "editPool/editPoolAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const tenant = state.tenants.tenantInfo;
    const selectedPool = state.tenants.selectedPool;
    const selectedStorageClass = state.editPool.fields.setup.storageClass;
    const numberOfNodes = state.editPool.fields.setup.numberOfNodes;
    const volumeSize = state.editPool.fields.setup.volumeSize;
    const volumesPerServer = state.editPool.fields.setup.volumesPerServer;
    const affinityType = state.editPool.fields.affinity.podAffinity;
    const nodeSelectorLabels =
      state.editPool.fields.affinity.nodeSelectorLabels;
    const withPodAntiAffinity =
      state.editPool.fields.affinity.withPodAntiAffinity;
    const tolerations = state.editPool.fields.tolerations;
    const securityContextEnabled =
      state.editPool.fields.configuration.securityContextEnabled;
    const securityContext = state.editPool.fields.configuration.securityContext;
    const customRuntime = state.editPool.fields.configuration.customRuntime;
    const runtimeClassName =
      state.editPool.fields.configuration.runtimeClassName;

    if (!tenant) {
      return;
    }

    const poolName = generatePoolName(tenant!.pools!);

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

    const cleanPools = tenant?.pools
      ?.filter((pool) => pool.name !== selectedPool)
      .map((pool) => {
        let securityContextOption: SecurityContext | null = null;

        if (pool.securityContext) {
          if (
            !!pool.securityContext.runAsUser ||
            !!pool.securityContext.runAsGroup ||
            !!pool.securityContext.fsGroup
          ) {
            securityContextOption = { ...pool.securityContext };
          }
        }

        const request = pool;
        if (securityContextOption) {
          request.securityContext = securityContextOption!;
        }

        return request;
      }) as Pool[];

    let runtimeClass = {};

    if (customRuntime) {
      runtimeClass = {
        runtimeClassName,
      };
    }

    cleanPools.push({
      name: selectedPool || poolName,
      servers: numberOfNodes,
      volumes_per_server: volumesPerServer,
      volume_configuration: {
        size: volumeSize * 1073741824,
        storage_class_name: selectedStorageClass,
        labels: undefined,
      },
      tolerations: tolerationValues,
      securityContext: securityContextEnabled ? securityContext : undefined,
      ...affinityObject,
      ...runtimeClass,
    });

    const data: PoolUpdateRequest = {
      pools: cleanPools,
    };
    const poolsURL: string = `/namespaces/${tenant?.namespace || ""}/tenants/${
      tenant?.name || ""
    }/pools`;

    return api.namespaces
      .tenantUpdatePools(tenant.namespace!, tenant.name!, data)

      .then(() => {
        dispatch(resetEditPoolForm());
        dispatch(getTenantAsync());
        return poolsURL;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  },
);
