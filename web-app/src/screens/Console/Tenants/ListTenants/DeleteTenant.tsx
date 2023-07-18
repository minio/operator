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

import React, { useState } from "react";
import { DialogContentText } from "@mui/material";

import { ErrorResponseHandler } from "../../../../common/types";
import InputBoxWrapper from "../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import Grid from "@mui/material/Grid";
import useApi from "../../Common/Hooks/useApi";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import { ConfirmDeleteIcon } from "mds";
import WarningMessage from "../../Common/WarningMessage/WarningMessage";
import FormSwitchWrapper from "../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";
import { Tenant } from "../../../../api/operatorApi";

interface IDeleteTenant {
  deleteOpen: boolean;
  selectedTenant: Tenant;
  closeDeleteModalAndRefresh: (refreshList: boolean) => any;
}

const DeleteTenant = ({
  deleteOpen,
  selectedTenant,
  closeDeleteModalAndRefresh,
}: IDeleteTenant) => {
  const dispatch = useAppDispatch();
  const [retypeTenant, setRetypeTenant] = useState("");

  const onDelSuccess = () => closeDeleteModalAndRefresh(true);
  const onDelError = (err: ErrorResponseHandler) =>
    dispatch(setErrorSnackMessage(err));
  const onClose = () => closeDeleteModalAndRefresh(false);

  const [deleteVolumes, setDeleteVolumes] = useState<boolean>(false);

  const [deleteLoading, invokeDeleteApi] = useApi(onDelSuccess, onDelError);

  const onConfirmDelete = () => {
    if (retypeTenant !== selectedTenant.name) {
      setErrorSnackMessage({
        errorMessage: "Tenant name is incorrect",
        detailedError: "",
      });
      return;
    }
    invokeDeleteApi(
      "DELETE",
      `/api/v1/namespaces/${selectedTenant.namespace}/tenants/${selectedTenant.name}`,
      { delete_pvcs: deleteVolumes },
    );
  };

  return (
    <ConfirmDialog
      title={`Delete Tenant`}
      confirmText={"Delete"}
      isOpen={deleteOpen}
      titleIcon={<ConfirmDeleteIcon />}
      isLoading={deleteLoading}
      onConfirm={onConfirmDelete}
      onClose={onClose}
      confirmButtonProps={{
        disabled: retypeTenant !== selectedTenant.name || deleteLoading,
      }}
      confirmationContent={
        <DialogContentText>
          {deleteVolumes && (
            <Grid item xs={12}>
              <WarningMessage
                title={"WARNING"}
                label={
                  "Delete Volumes: Data will be permanently deleted. Please proceed with caution."
                }
              />
            </Grid>
          )}
          To continue please type <b>{selectedTenant.name}</b> in the box.
          <Grid item xs={12}>
            <InputBoxWrapper
              id="retype-tenant"
              name="retype-tenant"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                setRetypeTenant(event.target.value);
              }}
              label=""
              value={retypeTenant}
            />
            <br />
            <FormSwitchWrapper
              checked={deleteVolumes}
              id={`delete-volumes`}
              label={"Delete Volumes"}
              name={`delete-volumes`}
              onChange={() => {
                setDeleteVolumes(!deleteVolumes);
              }}
              value={deleteVolumes}
            />
          </Grid>
        </DialogContentText>
      }
    />
  );
};

export default DeleteTenant;
