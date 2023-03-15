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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
import { useDispatch } from "react-redux";
import { combineReducers, configureStore } from "@reduxjs/toolkit";

import systemReducer from "./systemSlice";
import loginReducer from "./screens/LoginPage/loginSlice";
import consoleReducer from "./screens/Console/consoleSlice";
import tenantsReducer from "./screens/Console/Tenants/tenantsSlice";
import createTenantReducer from "./screens/Console/Tenants/AddTenant/createTenantSlice";
import addPoolReducer from "./screens/Console/Tenants/TenantDetails/Pools/AddPool/addPoolSlice";
import editPoolReducer from "./screens/Console/Tenants/TenantDetails/Pools/EditPool/editPoolSlice";
import editTenantSecurityContextReducer from "./screens/Console/Tenants/tenantSecurityContextSlice";
import licenseReducer from "./screens/Console/License/licenseSlice";
import registerReducer from "./screens/Console/Support/registerSlice";

const rootReducer = combineReducers({
  system: systemReducer,
  login: loginReducer,
  console: consoleReducer,
  register: registerReducer,
  // Operator Reducers
  tenants: tenantsReducer,
  createTenant: createTenantReducer,
  addPool: addPoolReducer,
  editPool: editPoolReducer,
  editTenantSecurityContext: editTenantSecurityContextReducer,
  license: licenseReducer,
});

export const store = configureStore({
  reducer: rootReducer,
});

if (process.env.NODE_ENV !== "production" && module.hot) {
  module.hot.accept(() => {
    store.replaceReducer(rootReducer);
  });
}

export type AppState = ReturnType<typeof store.getState>;

export type AppDispatch = typeof store.dispatch;
export const useAppDispatch = () => useDispatch<AppDispatch>();

export default store;
