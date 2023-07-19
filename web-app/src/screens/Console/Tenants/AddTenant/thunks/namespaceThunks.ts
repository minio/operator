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
import { AppState } from "../../../../../store";
import {
  setErrorSnackMessage,
  setModalErrorSnackMessage,
} from "../../../../../systemSlice";
import api from "../../../../../common/api";
import { ErrorResponseHandler } from "../../../../../common/types";
import { IQuotas } from "../../ListTenants/utils";
import get from "lodash/get";

export const validateNamespaceAsync = createAsyncThunk(
  "createTenant/validateNamespaceAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;
    const namespace = state.createTenant.fields.nameTenant.namespace;

    return api
      .invoke("GET", `/api/v1/namespaces/${namespace}/tenants`)
      .then((res: any[]): boolean => {
        const tenantsList = get(res, "tenants", []);
        let nsEmpty = true;
        if (tenantsList && tenantsList.length > 0) {
          return false;
        }
        if (nsEmpty) {
          dispatch(namespaceResourcesAsync());
        }
        // it's true it's empty
        return nsEmpty;
      })
      .catch((err) => {
        dispatch(
          setModalErrorSnackMessage({
            errorMessage: "Error validating if namespace already has tenants",
            detailedError: err.detailedError,
          }),
        );
        return rejectWithValue(false);
      });
  },
);

export const namespaceResourcesAsync = createAsyncThunk(
  "createTenant/namespaceResourcesAsync",
  async (_, { getState, rejectWithValue }) => {
    const state = getState() as AppState;

    const namespace = state.createTenant.fields.nameTenant.namespace;

    return api
      .invoke(
        "GET",
        `/api/v1/namespaces/${namespace}/resourcequotas/${namespace}-storagequota`,
      )
      .then((res: IQuotas): IQuotas => {
        return res;
      })
      .catch((err) => {
        console.error("Namespace quota error: ", err);
        return rejectWithValue(null);
      });
  },
);

export const createNamespaceAsync = createAsyncThunk(
  "createTenant/createNamespaceAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;
    const namespace = state.createTenant.fields.nameTenant.namespace;

    return api
      .invoke("POST", "/api/v1/namespace", {
        name: namespace,
      })
      .then((_) => {
        // revalidate the name to have the storage classes populated
        dispatch(validateNamespaceAsync());
        return true;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        rejectWithValue(false);
      });
  },
);
