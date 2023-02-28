// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

import {
  resetRegisterForm,
  setClusterRegistered,
  setLicenseInfo,
  setLoading,
  setLoadingLicenseInfo,
  setSelectedSubnetOrganization,
  setSubnetAccessToken,
  setSubnetMFAToken,
  setSubnetOrganizations,
  setSubnetOTP,
} from "./registerSlice";
import api from "../../../common/api";
import {
  SubnetInfo,
  SubnetLoginRequest,
  SubnetLoginResponse,
  SubnetLoginWithMFARequest,
  SubnetRegisterRequest,
} from "../License/types";
import { ErrorResponseHandler } from "../../../common/types";
import { setErrorSnackMessage } from "../../../systemSlice";
import { createAsyncThunk } from "@reduxjs/toolkit";
import { AppState } from "../../../store";
import { hasPermission } from "../../../common/SecureComponent";
import {
  CONSOLE_UI_RESOURCE,
  IAM_PAGES,
  IAM_PAGES_PERMISSIONS,
} from "../../../common/SecureComponent/permissions";

export const fetchLicenseInfo = createAsyncThunk(
  "register/fetchLicenseInfo",
  async (_, { getState, dispatch }) => {
    const state = getState() as AppState;

    const getSubnetInfo = hasPermission(
      CONSOLE_UI_RESOURCE,
      IAM_PAGES_PERMISSIONS[IAM_PAGES.LICENSE],
      true
    );

    const loadingLicenseInfo = state.register.loadingLicenseInfo;

    if (loadingLicenseInfo) {
      return;
    }
    if (getSubnetInfo) {
      dispatch(setLoadingLicenseInfo(true));
      api
        .invoke("GET", `/api/v1/subnet/info`)
        .then((res: SubnetInfo) => {
          dispatch(setLicenseInfo(res));
          dispatch(setClusterRegistered(true));
          dispatch(setLoadingLicenseInfo(false));
        })
        .catch((err: ErrorResponseHandler) => {
          if (
            err.detailedError.toLowerCase() !==
              "License is not present".toLowerCase() &&
            err.detailedError.toLowerCase() !==
              "license not found".toLowerCase()
          ) {
            dispatch(setErrorSnackMessage(err));
          }
          dispatch(setClusterRegistered(false));
          dispatch(setLoadingLicenseInfo(false));
        });
    } else {
      dispatch(setLoadingLicenseInfo(false));
    }
  }
);

export interface ClassRegisterArgs {
  token: string;
  account_id: string;
}

export const callRegister = createAsyncThunk(
  "register/callRegister",
  async (args: ClassRegisterArgs, { dispatch }) => {
    const request: SubnetRegisterRequest = {
      token: args.token,
      account_id: args.account_id,
    };
    api
      .invoke("POST", "/api/v1/subnet/register", request)
      .then(() => {
        dispatch(setLoading(false));
        dispatch(resetRegisterForm());
        dispatch(fetchLicenseInfo());
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        dispatch(setLoading(false));
      });
  }
);

export const subnetLoginWithMFA = createAsyncThunk(
  "register/subnetLoginWithMFA",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const subnetEmail = state.register.subnetEmail;
    const subnetMFAToken = state.register.subnetMFAToken;
    const subnetOTP = state.register.subnetOTP;
    const loading = state.register.loading;

    if (loading) {
      return;
    }
    dispatch(setLoading(true));
    const request: SubnetLoginWithMFARequest = {
      username: subnetEmail,
      otp: subnetOTP,
      mfa_token: subnetMFAToken,
    };
    api
      .invoke("POST", "/api/v1/subnet/login/mfa", request)
      .then((resp: SubnetLoginResponse) => {
        dispatch(setLoading(false));
        if (resp && resp.access_token && resp.organizations.length > 0) {
          if (resp.organizations.length === 1) {
            dispatch(
              callRegister({
                token: resp.access_token,
                account_id: resp.organizations[0].accountId.toString(),
              })
            );
          } else {
            dispatch(setSubnetAccessToken(resp.access_token));
            dispatch(setSubnetOrganizations(resp.organizations));
            dispatch(
              setSelectedSubnetOrganization(
                resp.organizations[0].accountId.toString()
              )
            );
          }
        }
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        dispatch(setLoading(false));
        dispatch(setSubnetOTP(""));
      });
  }
);

export const subnetLogin = createAsyncThunk(
  "register/subnetLogin",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    const license = state.register.license;
    const subnetPassword = state.register.subnetPassword;
    const subnetEmail = state.register.subnetEmail;
    const loading = state.register.loading;

    if (loading) {
      return;
    }
    dispatch(setLoading(true));
    let request: SubnetLoginRequest = {
      username: subnetEmail,
      password: subnetPassword,
      apiKey: license,
    };
    api
      .invoke("POST", "/api/v1/subnet/login", request)
      .then((resp: SubnetLoginResponse) => {
        dispatch(setLoading(false));
        if (resp && resp.registered) {
          dispatch(resetRegisterForm());
          dispatch(fetchLicenseInfo());
        } else if (resp && resp.mfa_token) {
          dispatch(setSubnetMFAToken(resp.mfa_token));
        } else if (resp && resp.access_token && resp.organizations.length > 0) {
          dispatch(setSubnetAccessToken(resp.access_token));
          dispatch(setSubnetOrganizations(resp.organizations));
          dispatch(
            setSelectedSubnetOrganization(
              resp.organizations[0].accountId.toString()
            )
          );
        }
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        dispatch(setLoading(false));
        dispatch(resetRegisterForm());
      });
  }
);
