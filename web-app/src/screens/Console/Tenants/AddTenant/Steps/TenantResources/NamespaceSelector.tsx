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

import React, { Fragment, useEffect, useMemo } from "react";
import debounce from "lodash/debounce";
import { AddIcon, InputBox } from "mds";
import { useSelector } from "react-redux";
import { openAddNSModal, setNamespace } from "../../createTenantSlice";
import { AppState, useAppDispatch } from "../../../../../../store";
import { IMkEnvs } from "./utils";
import { validateNamespaceAsync } from "../../thunks/namespaceThunks";
import AddNamespaceModal from "../helpers/AddNamespaceModal";

const NamespaceSelector = ({ formToRender }: { formToRender?: IMkEnvs }) => {
  const dispatch = useAppDispatch();

  const namespace = useSelector(
    (state: AppState) => state.createTenant.fields.nameTenant.namespace,
  );

  const showNSCreateButton = useSelector(
    (state: AppState) => state.createTenant.showNSCreateButton,
  );

  const namespaceError = useSelector(
    (state: AppState) => state.createTenant.validationErrors["namespace"],
  );
  const openAddNSConfirm = useSelector(
    (state: AppState) => state.createTenant.addNSOpen,
  );

  const debounceNamespace = useMemo(
    () =>
      debounce(() => {
        dispatch(validateNamespaceAsync());
      }, 500),
    [dispatch],
  );

  useEffect(() => {
    if (namespace !== "") {
      debounceNamespace();
      // Cancel previous debounce calls during useEffect cleanup.
      return debounceNamespace.cancel;
    }
  }, [debounceNamespace, namespace]);

  const addNamespace = () => {
    dispatch(openAddNSModal());
  };

  return (
    <Fragment>
      {openAddNSConfirm && <AddNamespaceModal />}
      <InputBox
        id="namespace"
        name="namespace"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          dispatch(setNamespace(e.target.value));
        }}
        label="Namespace"
        value={namespace}
        error={namespaceError || ""}
        overlayIcon={showNSCreateButton ? <AddIcon /> : null}
        overlayAction={addNamespace}
        required
      />
    </Fragment>
  );
};
export default NamespaceSelector;
