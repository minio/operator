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
import { AppState } from "../../store";
import api from "../../common/api";
import { ErrorResponseHandler } from "../../common/types";
import {
  setErrorSnackMessage,
  showMarketplace,
  userLogged,
} from "../../systemSlice";
import { ILoginDetails, loginStrategyType } from "./types";
import { setNavigateTo } from "./loginSlice";
import {
  getTargetPath,
  loginStrategyEndpoints,
  LoginStrategyPayload,
} from "./LoginPage";

export const doLoginAsync = createAsyncThunk(
  "login/doLoginAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;
    const loginStrategy = state.login.loginStrategy;
    const accessKey = state.login.accessKey;
    const secretKey = state.login.secretKey;
    const jwt = state.login.jwt;
    const sts = state.login.sts;
    const useSTS = state.login.useSTS;

    const isOperator =
      loginStrategy.loginStrategy === loginStrategyType.serviceAccount ||
      loginStrategy.loginStrategy === loginStrategyType.redirectServiceAccount;

    let loginStrategyPayload: LoginStrategyPayload = {
      form: { accessKey, secretKey },
      "service-account": { jwt },
    };
    if (useSTS) {
      loginStrategyPayload = {
        form: { accessKey, secretKey, sts },
      };
    }

    return api
      .invoke(
        "POST",
        loginStrategyEndpoints[loginStrategy.loginStrategy] || "/api/v1/login",
        loginStrategyPayload[loginStrategy.loginStrategy],
      )
      .then((res) => {
        // We set the state in redux
        dispatch(userLogged(true));
        if (loginStrategy.loginStrategy === loginStrategyType.form) {
          localStorage.setItem("userLoggedIn", accessKey);
        }
        // if it's in operator mode, check the Marketplace integration
        if (isOperator) {
          api
            .invoke("GET", "/api/v1/mp-integration/")
            .then((res: any) => {
              dispatch(setNavigateTo(getTargetPath())); // Email already set, continue with normal flow
            })
            .catch((err: ErrorResponseHandler) => {
              if (err.statusCode === 404) {
                dispatch(showMarketplace(true));
                dispatch(setNavigateTo("/marketplace"));
              } else {
                // Unexpected error, continue with normal flow
                dispatch(setNavigateTo(getTargetPath()));
              }
            });
        } else {
          dispatch(setNavigateTo(getTargetPath()));
        }
      })
      .catch((err) => {
        dispatch(setErrorSnackMessage(err));
      });
  },
);
export const getFetchConfigurationAsync = createAsyncThunk(
  "login/getFetchConfigurationAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    return api
      .invoke("GET", "/api/v1/login")
      .then((loginDetails: ILoginDetails) => {
        return loginDetails;
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  },
);

export const getVersionAsync = createAsyncThunk(
  "login/getVersionAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    return api
      .invoke("GET", "/api/v1/check-version")
      .then(
        ({
          current_version,
          latest_version,
        }: {
          current_version: string;
          latest_version: string;
        }) => {
          return latest_version;
        },
      )
      .catch((err: ErrorResponseHandler) => {
        // try the operator version
        api
          .invoke("GET", "/api/v1/check-operator-version")
          .then(
            ({
              current_version,
              latest_version,
            }: {
              current_version: string;
              latest_version: string;
            }) => {
              return latest_version;
            },
          )
          .catch((err: ErrorResponseHandler) => {
            return err;
          });
      });
  },
);
