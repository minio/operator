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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { Box, Button, FormLayout, InputBox, Switch } from "mds";
import { modalStyleUtils } from "../../Common/FormComponents/common/styleLibrary";
import { ErrorResponseHandler } from "../../../../common/types";
import {
  setModalErrorSnackMessage,
  setSnackBarMessage,
} from "../../../../systemSlice";
import { useAppDispatch } from "../../../../store";
import ModalWrapper from "../../Common/ModalWrapper/ModalWrapper";
import api from "../../../../common/api";

interface IUpdateTenantModal {
  open: boolean;
  closeModalAndRefresh: (update: boolean) => any;
  namespace: string;
  idTenant: string;
}

const UpdateTenantModal = ({
  open,
  closeModalAndRefresh,
  namespace,
  idTenant,
}: IUpdateTenantModal) => {
  const dispatch = useAppDispatch();
  const [isSending, setIsSending] = useState<boolean>(false);
  const [minioImage, setMinioImage] = useState<string>("");
  const [imageRegistry, setImageRegistry] = useState<boolean>(false);
  const [imageRegistryEndpoint, setImageRegistryEndpoint] =
    useState<string>("");
  const [imageRegistryUsername, setImageRegistryUsername] =
    useState<string>("");
  const [imageRegistryPassword, setImageRegistryPassword] =
    useState<string>("");
  const [validMinioImage, setValidMinioImage] = useState<boolean>(true);

  const validateImage = useCallback(
    (fieldToCheck: string) => {
      const pattern = new RegExp("^$|^((.*?)/(.*?):(.+))$");

      switch (fieldToCheck) {
        case "minioImage":
          setValidMinioImage(pattern.test(minioImage));
          break;
      }
    },
    [minioImage],
  );

  useEffect(() => {
    validateImage("minioImage");
  }, [minioImage, validateImage]);

  const closeAction = () => {
    closeModalAndRefresh(false);
  };

  const resetForm = () => {
    setMinioImage("");
    setImageRegistry(false);
    setImageRegistryEndpoint("");
    setImageRegistryUsername("");
    setImageRegistryPassword("");
  };

  const updateMinIOImage = () => {
    setIsSending(true);

    let payload = {
      image: minioImage,
    };

    if (imageRegistry) {
      const registry: any = {
        image_registry: {
          registry: imageRegistryEndpoint,
          username: imageRegistryUsername,
          password: imageRegistryPassword,
        },
      };
      payload = {
        ...payload,
        ...registry,
      };
    }

    api
      .invoke(
        "PUT",
        `/api/v1/namespaces/${namespace}/tenants/${idTenant}`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        dispatch(setSnackBarMessage(`Image updated successfully`));
        closeModalAndRefresh(true);
      })
      .catch((error: ErrorResponseHandler) => {
        dispatch(setModalErrorSnackMessage(error));
        setIsSending(false);
      });
  };

  return (
    <ModalWrapper
      title={"Update MinIO Version"}
      modalOpen={open}
      onClose={closeAction}
    >
      <FormLayout withBorders={false} containerPadding={false}>
        <Box>
          <Box
            sx={{
              fontSize: 14,
            }}
          >
            Please enter the MinIO image from dockerhub to use. If blank, then
            latest build will be used.
          </Box>
          <br />
          <InputBox
            value={minioImage}
            label={"MinIO's Image"}
            id={"minioImage"}
            name={"minioImage"}
            placeholder={"E.g. minio/minio:RELEASE.2022-02-26T02-54-46Z"}
            onChange={(e) => {
              setMinioImage(e.target.value);
            }}
          />
          <Switch
            value="imageRegistry"
            id="setImageRegistry"
            name="setImageRegistry"
            checked={imageRegistry}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              setImageRegistry(!imageRegistry);
            }}
            label={"Set Custom Image Registry"}
            indicatorLabels={["Yes", "No"]}
          />
          {imageRegistry && (
            <Fragment>
              <InputBox
                value={imageRegistryEndpoint}
                label={"Endpoint"}
                id={"imageRegistry"}
                name={"imageRegistry"}
                placeholder={"E.g. https://index.docker.io/v1/"}
                onChange={(e) => {
                  setImageRegistryEndpoint(e.target.value);
                }}
              />
              <InputBox
                value={imageRegistryUsername}
                label={"Username"}
                id={"imageRegistryUsername"}
                name={"imageRegistryUsername"}
                placeholder={"Enter image registry username"}
                onChange={(e) => {
                  setImageRegistryUsername(e.target.value);
                }}
              />
              <InputBox
                value={imageRegistryPassword}
                label={"Password"}
                id={"imageRegistryPassword"}
                name={"imageRegistryPassword"}
                placeholder={"Enter image registry password"}
                onChange={(e) => {
                  setImageRegistryPassword(e.target.value);
                }}
              />
            </Fragment>
          )}
          <Box sx={modalStyleUtils.modalButtonBar}>
            <Button
              id={"clear"}
              variant="regular"
              onClick={resetForm}
              label="Clear"
            />
            <Button
              id={"save-tenant"}
              type="submit"
              variant="callAction"
              disabled={
                !validMinioImage ||
                (imageRegistry &&
                  (imageRegistryEndpoint.trim() === "" ||
                    imageRegistryUsername.trim() === "" ||
                    imageRegistryPassword.trim() === "")) ||
                isSending
              }
              onClick={updateMinIOImage}
              label={"Save"}
            />
          </Box>
        </Box>
      </FormLayout>
    </ModalWrapper>
  );
};

export default UpdateTenantModal;
