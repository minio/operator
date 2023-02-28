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
import { AppState } from "../../../../store";
import { niceBytes } from "../../../../common/utils";
import { ITenant } from "../ListTenants/types";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";

export const getTenantAsync = createAsyncThunk(
  "tenantDetails/getTenantAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const currentNamespace = state.tenants.currentNamespace;
    const currentTenant = state.tenants.currentTenant;

    return api
      .invoke(
        "GET",
        `/api/v1/namespaces/${currentNamespace}/tenants/${currentTenant}`
      )
      .then((res: ITenant) => {
        // add computed fields
        const resPools = !res.pools ? [] : res.pools;

        let totalInstances = 0;
        let totalVolumes = 0;
        let poolNamedIndex = 0;
        for (let pool of resPools) {
          const cap =
            pool.volumes_per_server *
            pool.servers *
            pool.volume_configuration.size;
          pool.label = `pool-${poolNamedIndex}`;
          if (pool.name === undefined || pool.name === "") {
            pool.name = pool.label;
          }
          pool.capacity = niceBytes(cap + "");
          pool.volumes = pool.servers * pool.volumes_per_server;
          totalInstances += pool.servers;
          totalVolumes += pool.volumes;
          poolNamedIndex += 1;
        }
        res.total_instances = totalInstances;
        res.total_volumes = totalVolumes;

        return res;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        return rejectWithValue(err);
      });
  }
);
