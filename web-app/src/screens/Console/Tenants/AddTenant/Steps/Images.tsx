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
import { useSelector } from "react-redux";
import { Box, FormLayout, InputBox, Switch } from "mds";
import { AppState, useAppDispatch } from "../../../../../store";
import { clearValidationError } from "../../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import { isPageValid, updateAddField } from "../createTenantSlice";
import H3Section from "../../../Common/H3Section";

const Images = () => {
  const dispatch = useAppDispatch();

  const customImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.customImage,
  );
  const imageName = useSelector(
    (state: AppState) => state.createTenant.fields.configure.imageName,
  );
  const customDockerhub = useSelector(
    (state: AppState) => state.createTenant.fields.configure.customDockerhub,
  );
  const imageRegistry = useSelector(
    (state: AppState) => state.createTenant.fields.configure.imageRegistry,
  );
  const imageRegistryUsername = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.imageRegistryUsername,
  );
  const imageRegistryPassword = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.imageRegistryPassword,
  );

  const tenantCustom = useSelector(
    (state: AppState) => state.createTenant.fields.configure.tenantCustom,
  );

  const kesImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.kesImage,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "configure", field: field, value: value }),
      );
    },
    [dispatch],
  );

  // Validation
  useEffect(() => {
    let customAccountValidation: IValidation[] = [];

    if (customImage) {
      customAccountValidation = [
        ...customAccountValidation,
        {
          fieldKey: "image",
          required: false,
          value: imageName,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage: "Format must be of form: 'minio/minio:VERSION'",
        },
        {
          fieldKey: "kesImage",
          required: false,
          value: kesImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage: "Format must be of form: 'minio/kes:VERSION'",
        },
      ];
      if (customDockerhub) {
        customAccountValidation = [
          ...customAccountValidation,
          {
            fieldKey: "registry",
            required: true,
            value: imageRegistry,
          },
          {
            fieldKey: "registryUsername",
            required: true,
            value: imageRegistryUsername,
          },
          {
            fieldKey: "registryPassword",
            required: true,
            value: imageRegistryPassword,
          },
        ];
      }
    }

    const commonVal = commonFormValidation(customAccountValidation);

    dispatch(
      isPageValid({
        pageName: "configure",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [
    customImage,
    imageName,
    kesImage,
    customDockerhub,
    imageRegistry,
    imageRegistryUsername,
    imageRegistryPassword,
    dispatch,
    tenantCustom,
  ]);

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  return (
    <FormLayout withBorders={false} containerPadding={false}>
      <Box className={"inputItem"}>
        <H3Section>Container Images</H3Section>
        <span className={"muted"}>
          Specify the container images used by the Tenant and its features.
        </span>
      </Box>

      <InputBox
        id="image"
        name="image"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          updateField("imageName", e.target.value);
          cleanValidation("image");
        }}
        label="MinIO"
        value={imageName}
        error={validationErrors["image"] || ""}
        placeholder="minio/minio:RELEASE.2023-11-15T20-43-25Z"
      />
      <InputBox
        id="kesImage"
        name="kesImage"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          updateField("kesImage", e.target.value);
          cleanValidation("kesImage");
        }}
        label="KES"
        value={kesImage}
        error={validationErrors["kesImage"] || ""}
        placeholder="minio/kes:2023-11-10T10-44-28Z"
      />

      {customImage && (
        <Fragment>
          <Box className={"inputItem"}>
            <h4>Custom Container Registry</h4>
          </Box>
          <Switch
            value="custom_docker_hub"
            id="custom_docker_hub"
            name="custom_docker_hub"
            checked={customDockerhub}
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;

              updateField("customDockerhub", checked);
            }}
            label={"Use a private container registry"}
          />
        </Fragment>
      )}
      {customDockerhub && (
        <Fragment>
          <InputBox
            id="registry"
            name="registry"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("imageRegistry", e.target.value);
            }}
            label="Endpoint"
            value={imageRegistry}
            error={validationErrors["registry"] || ""}
            placeholder="https://index.docker.io/v1/"
            required
          />
          <InputBox
            id="registryUsername"
            name="registryUsername"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("imageRegistryUsername", e.target.value);
            }}
            label="Username"
            value={imageRegistryUsername}
            error={validationErrors["registryUsername"] || ""}
            required
          />
          <InputBox
            id="registryPassword"
            name="registryPassword"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("imageRegistryPassword", e.target.value);
            }}
            label="Password"
            value={imageRegistryPassword}
            error={validationErrors["registryPassword"] || ""}
            required
          />
        </Fragment>
      )}
    </FormLayout>
  );
};

export default Images;
