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

import { Button } from "mds";
import React from "react";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../store";
import { requiredPages } from "./common";
import { createTenantAsync } from "./thunks/createTenantThunk";

const CreateTenantButton = () => {
  const dispatch = useAppDispatch();

  const addSending = useSelector(
    (state: AppState) => state.createTenant.addingTenant,
  );

  const validPages = useSelector(
    (state: AppState) => state.createTenant.validPages,
  );

  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass,
  );

  const enabled =
    !addSending &&
    selectedStorageClass !== "" &&
    requiredPages.every((v) => validPages.includes(v));

  return (
    <Button
      id={"wizard-button-Create"}
      variant="callAction"
      color="primary"
      onClick={() => {
        dispatch(createTenantAsync());
      }}
      disabled={!enabled}
      key={`button-AddTenant-Create`}
      label={"Create"}
    />
  );
};

export default CreateTenantButton;
