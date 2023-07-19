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

import React from "react";
import { useSelector } from "react-redux";
import { DialogContentText, LinearProgress } from "@mui/material";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  deleteDialogStyles,
  modalBasic,
} from "../../../../Common/FormComponents/common/styleLibrary";
import ConfirmDialog from "../../../../Common/ModalWrapper/ConfirmDialog";
import { ConfirmModalIcon } from "mds";
import { AppState, useAppDispatch } from "../../../../../../store";
import { closeAddNSModal } from "../../createTenantSlice";
import makeStyles from "@mui/styles/makeStyles";
import { createNamespaceAsync } from "../../thunks/namespaceThunks";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    wrapText: {
      maxWidth: "200px",
      whiteSpace: "normal",
      wordWrap: "break-word",
    },
    ...modalBasic,
    ...deleteDialogStyles,
  }),
);

const AddNamespaceModal = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

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
        <React.Fragment>
          {addNamespaceLoading && <LinearProgress />}
          <DialogContentText>
            Are you sure you want to add a namespace called
            <br />
            <b className={classes.wrapText}>{namespace}</b>?
          </DialogContentText>
        </React.Fragment>
      }
    />
  );
};

export default AddNamespaceModal;
