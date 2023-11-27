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

import React, { useState, Fragment } from "react";
import { Grid, InputBox } from "mds";

import { ErrorResponseHandler } from "../../../../common/types";
import useApi from "../../Common/Hooks/useApi";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import { ConfirmDeleteIcon } from "mds";
import { IStoragePVCs } from "../../Storage/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";

interface IDeletePVC {
  deleteOpen: boolean;
  selectedPVC: IStoragePVCs;
  closeDeleteModalAndRefresh: (refreshList: boolean) => any;
}

const DeletePVC = ({
  deleteOpen,
  selectedPVC,
  closeDeleteModalAndRefresh,
}: IDeletePVC) => {
  const dispatch = useAppDispatch();
  const [retypePVC, setRetypePVC] = useState("");

  const onDelSuccess = () => closeDeleteModalAndRefresh(true);
  const onDelError = (err: ErrorResponseHandler) =>
    dispatch(setErrorSnackMessage(err));
  const onClose = () => closeDeleteModalAndRefresh(false);

  const [deleteLoading, invokeDeleteApi] = useApi(onDelSuccess, onDelError);

  const onConfirmDelete = () => {
    if (retypePVC !== selectedPVC.name) {
      dispatch(
        setErrorSnackMessage({
          errorMessage: "PVC name is incorrect",
          detailedError: "",
        }),
      );
      return;
    }
    invokeDeleteApi(
      "DELETE",
      `/api/v1/namespaces/${selectedPVC.namespace}/tenants/${selectedPVC.tenant}/pvc/${selectedPVC.name}`,
    );
  };

  return (
    <ConfirmDialog
      title={`Delete PVC`}
      confirmText={"Delete"}
      isOpen={deleteOpen}
      titleIcon={<ConfirmDeleteIcon />}
      isLoading={deleteLoading}
      onConfirm={onConfirmDelete}
      onClose={onClose}
      confirmButtonProps={{
        disabled: retypePVC !== selectedPVC.name || deleteLoading,
      }}
      confirmationContent={
        <Fragment>
          To continue please type <b>{selectedPVC.name}</b> in the box.
          <Grid item xs={12} sx={{ marginTop: 15 }}>
            <InputBox
              id="retype-PVC"
              name="retype-PVC"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                setRetypePVC(event.target.value);
              }}
              label=""
              value={retypePVC}
            />
          </Grid>
        </Fragment>
      }
    />
  );
};

export default DeletePVC;
