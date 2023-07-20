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
import { addPoolAsync } from "./addPoolThunks";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../../store";

const AddPoolCreateButton = () => {
  const dispatch = useAppDispatch();

  const selectedStorageClass = useSelector(
    (state: AppState) => state.addPool.setup.storageClass,
  );
  const validPages = useSelector((state: AppState) => state.addPool.validPages);

  const sending = useSelector((state: AppState) => state.addPool.sending);
  const requiredPages = ["setup", "affinity", "configure"];
  const enabled =
    !sending &&
    selectedStorageClass !== "" &&
    requiredPages.every((v) => validPages.includes(v));
  return (
    <Button
      id={"wizard-button-Create"}
      variant="callAction"
      onClick={() => {
        dispatch(addPoolAsync());
      }}
      disabled={!enabled}
      key={`button-AddTenant-Create`}
      label={"Create"}
    />
  );
};

export default AddPoolCreateButton;
