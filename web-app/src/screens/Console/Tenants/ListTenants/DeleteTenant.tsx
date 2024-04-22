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
import {
  ConfirmDeleteIcon,
  FormLayout,
  InformativeMessage,
  InputBox,
  Switch,
  Grid,
} from "mds";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";
import { Tenant } from "../../../../api/operatorApi";
import useApi from "../../Common/Hooks/useApi";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";

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
        <FormLayout withBorders={false} containerPadding={false}>
          {deleteVolumes && (
            <Grid item xs={12} className={"inputItem"}>
              <InformativeMessage
                variant={"error"}
                title={"WARNING"}
                message={
                  "Delete Volumes: Data will be permanently deleted. Please proceed with caution."
                }
              />
            </Grid>
          )}
          To continue please type <b>{selectedTenant.name}</b> in the box.
          <Grid item xs={12}>
            <InputBox
              id="retype-tenant"
              name="retype-tenant"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                setRetypeTenant(event.target.value);
              }}
              label=""
              value={retypeTenant}
            />
            <Switch
              checked={deleteVolumes}
              id={`delete-volumes`}
              label={"Delete Volumes"}
              name={`delete-volumes`}
              onChange={() => {
                setDeleteVolumes(!deleteVolumes);
              }}
            />
          </Grid>
        </FormLayout>
      }
    />
  );
};

export default DeleteTenant;
