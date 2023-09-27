// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
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

import { ErrorResponseHandler } from "./common/types";
import { SubnetInfo } from "./screens/Console/License/types";

// along with this program.  If not, see <http://www.gnu.org/licenses/>.
export interface snackBarMessage {
  message: string;
  detailedErrorMsg: string;
  type: "message" | "error";
}

export interface SRInfoStateType {
  enabled: boolean;
  curSite: boolean;
  siteName: string;
}

export const USER_LOGGED = "USER_LOGGED";
export const OPERATOR_MODE = "OPERATOR_MODE";
export const MENU_OPEN = "MENU_OPEN";
export const SERVER_NEEDS_RESTART = "SERVER_NEEDS_RESTART";
export const SERVER_IS_LOADING = "SERVER_IS_LOADING";
export const SET_LOADING_PROGRESS = "SET_LOADING_PROGRESS";
export const SET_SNACK_BAR_MESSAGE = "SET_SNACK_BAR_MESSAGE";
export const SET_SERVER_DIAG_STAT = "SET_SERVER_DIAG_STAT";
export const SET_ERROR_SNACK_MESSAGE = "SET_ERROR_SNACK_MESSAGE";
export const SET_SNACK_MODAL_MESSAGE = "SET_SNACK_MODAL_MESSAGE";
export const SET_MODAL_ERROR_MESSAGE = "SET_MODAL_ERROR_MESSAGE";
export const GLOBAL_SET_DISTRIBUTED_SETUP = "GLOBAL/SET_DISTRIBUTED_SETUP";
export const SET_SITE_REPLICATION_INFO = "SET_SITE_REPLICATION_INFO";
export const SET_LICENSE_INFO = "SET_LICENSE_INFO";

interface UserLoggedAction {
  type: typeof USER_LOGGED;
  logged: boolean;
}

interface OperatorModeAction {
  type: typeof OPERATOR_MODE;
  operatorMode: boolean;
}

interface SetMenuOpenAction {
  type: typeof MENU_OPEN;
  open: boolean;
}

interface ServerNeedsRestartAction {
  type: typeof SERVER_NEEDS_RESTART;
  needsRestart: boolean;
}

interface ServerIsLoading {
  type: typeof SERVER_IS_LOADING;
  isLoading: boolean;
}

interface SetLoadingProgress {
  type: typeof SET_LOADING_PROGRESS;
  loadingProgress: number;
}

interface SetServerDiagStat {
  type: typeof SET_SERVER_DIAG_STAT;
  serverDiagnosticStatus: string;
}

interface SetSnackBarMessage {
  type: typeof SET_SNACK_BAR_MESSAGE;
  message: string;
}

interface SetErrorSnackMessage {
  type: typeof SET_ERROR_SNACK_MESSAGE;
  message: ErrorResponseHandler;
}

interface SetModalSnackMessage {
  type: typeof SET_SNACK_MODAL_MESSAGE;
  message: string;
}

interface SetModalErrorMessage {
  type: typeof SET_MODAL_ERROR_MESSAGE;
  message: ErrorResponseHandler;
}

interface SetDistributedSetup {
  type: typeof GLOBAL_SET_DISTRIBUTED_SETUP;
  distributedSetup: boolean;
}

interface SetSiteReplicationInfo {
  type: typeof SET_SITE_REPLICATION_INFO;
  siteReplicationInfo: SRInfoStateType;
}

interface SetLicenseInfo {
  type: typeof SET_LICENSE_INFO;
  licenseInfo: SubnetInfo;
}
