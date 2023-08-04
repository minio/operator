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

import { ICreateTenant } from "./createTenantSlice";
import { Draft } from "@reduxjs/toolkit";

export const flipValidPageInState = (
  state: Draft<ICreateTenant>,
  pageName: string,
  valid: boolean,
) => {
  let originValidPages = state.validPages;
  if (valid) {
    if (!originValidPages.includes(pageName)) {
      originValidPages.push(pageName);

      state.validPages = [...originValidPages];
    }
  } else {
    const newSetOfPages = originValidPages.filter((elm) => elm !== pageName);
    state.validPages = [...newSetOfPages];
  }
};
