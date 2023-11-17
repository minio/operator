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

import React, { Fragment } from "react";
import { useSelector } from "react-redux";
import { ConfirmModalIcon, ProgressBar } from "mds";
import ConfirmDialog from "../../../../Common/ModalWrapper/ConfirmDialog";
import { AppState, useAppDispatch } from "../../../../../../store";
import { closeAddNSModal } from "../../createTenantSlice";
import { createNamespaceAsync } from "../../thunks/namespaceThunks";

const AddNamespaceModal = () => {
  const dispatch = useAppDispatch();

  const namespace = useSelector(
    (state: AppState) => state.createTenant.fields.nameTenant.namespace,
  );
  const addNamespaceLoading = useSelector(
    (state: AppState) => state.createTenant.addNSLoading,
  );
  const addNamespaceOpen = useSelector(
    (state: AppState) => state.createTenant.addNSOpen,
  );

  return (
    <ConfirmDialog
      title={`New namespace`}
      confirmText={"Create"}
      confirmButtonProps={{
        variant: "callAction",
      }}
      isOpen={addNamespaceOpen}
      titleIcon={<ConfirmModalIcon />}
      isLoading={addNamespaceLoading}
      onConfirm={() => {
        dispatch(createNamespaceAsync());
      }}
      onClose={() => {
        dispatch(closeAddNSModal());
      }}
      confirmationContent={
        <Fragment>
          {addNamespaceLoading && <ProgressBar />}
          Are you sure you want to add a namespace called
          <br />
          <b
            style={{
              maxWidth: "200px",
              whiteSpace: "normal",
              wordWrap: "break-word",
            }}
          >
            {namespace}
          </b>
          ?
        </Fragment>
      }
    />
  );
};

export default AddNamespaceModal;
