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

import React, { useState, Fragment } from "react";
import { ConfirmDeleteIcon, Box, InputBox } from "mds";
import { IPodListElement } from "../ListTenants/types";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";
import useApi from "../../Common/Hooks/useApi";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";

interface IDeletePod {
  deleteOpen: boolean;
  selectedPod: IPodListElement;
  closeDeleteModalAndRefresh: (refreshList: boolean) => any;
}

const DeletePod = ({
  deleteOpen,
  selectedPod,
  closeDeleteModalAndRefresh,
}: IDeletePod) => {
  const dispatch = useAppDispatch();
  const [retypePod, setRetypePod] = useState("");

  const onDelSuccess = () => closeDeleteModalAndRefresh(true);
  const onDelError = (err: ErrorResponseHandler) =>
    dispatch(setErrorSnackMessage(err));
  const onClose = () => closeDeleteModalAndRefresh(false);

  const [deleteLoading, invokeDeleteApi] = useApi(onDelSuccess, onDelError);

  const onConfirmDelete = () => {
    if (retypePod !== selectedPod.name) {
      setErrorSnackMessage({
        errorMessage: "Tenant name is incorrect",
        detailedError: "",
      });
      return;
    }
    invokeDeleteApi(
      "DELETE",
      `/api/v1/namespaces/${selectedPod.namespace}/tenants/${selectedPod.tenant}/pods/${selectedPod.name}`,
    );
  };

  return (
    <ConfirmDialog
      title={`Delete Pod`}
      confirmText={"Delete"}
      isOpen={deleteOpen}
      titleIcon={<ConfirmDeleteIcon />}
      isLoading={deleteLoading}
      onConfirm={onConfirmDelete}
      onClose={onClose}
      confirmButtonProps={{
        disabled: retypePod !== selectedPod.name || deleteLoading,
      }}
      confirmationContent={
        <Fragment>
          To continue please type <b>{selectedPod.name}</b> in the box.
          <Box sx={{ marginTop: 15 }}>
            <InputBox
              id="retype-pod"
              name="retype-pod"
              onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                setRetypePod(event.target.value);
              }}
              label=""
              value={retypePod}
            />
          </Box>
        </Fragment>
      }
    />
  );
};

export default DeletePod;
