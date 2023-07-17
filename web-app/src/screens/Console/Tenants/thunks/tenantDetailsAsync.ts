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
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { api } from "../../../../api";
import { Error, HttpResponse, Tenant } from "../../../../api/operatorApi";

export const getTenantAsync = createAsyncThunk(
  "tenantDetails/getTenantAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const currentNamespace = state.tenants.currentNamespace;
    const currentTenant = state.tenants.currentTenant;
    return api.namespaces
      .tenantDetails(currentNamespace, currentTenant)
      .then((res: HttpResponse<Tenant, Error>) => {
        return res.data;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        return rejectWithValue(err);
      });
  },
);
