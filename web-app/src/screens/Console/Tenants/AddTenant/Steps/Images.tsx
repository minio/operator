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
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid, Paper } from "@mui/material";
import {
  formFieldStyles,
  wizardCommon,
} from "../../../Common/FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../../store";
import { clearValidationError } from "../../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import FormSwitchWrapper from "../../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import InputBoxWrapper from "../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { isPageValid, updateAddField } from "../createTenantSlice";
import H3Section from "../../../Common/H3Section";

interface IImagesProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...formFieldStyles,
    ...wizardCommon,
  });

const Images = ({ classes }: IImagesProps) => {
  const dispatch = useAppDispatch();

  const customImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.customImage
  );
  const imageName = useSelector(
    (state: AppState) => state.createTenant.fields.configure.imageName
  );
  const customDockerhub = useSelector(
    (state: AppState) => state.createTenant.fields.configure.customDockerhub
  );
  const imageRegistry = useSelector(
    (state: AppState) => state.createTenant.fields.configure.imageRegistry
  );
  const imageRegistryUsername = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.imageRegistryUsername
  );
  const imageRegistryPassword = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.imageRegistryPassword
  );

  const prometheusCustom = useSelector(
    (state: AppState) => state.createTenant.fields.configure.prometheusEnabled
  );
  const tenantCustom = useSelector(
    (state: AppState) => state.createTenant.fields.configure.tenantCustom
  );
  const logSearchCustom = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchEnabled
  );
  const logSearchVolumeSize = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchVolumeSize
  );

  const prometheusVolumeSize = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.prometheusVolumeSize
  );

  const logSearchSelectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchSelectedStorageClass
  );
  const logSearchImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchImage
  );
  const kesImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.kesImage
  );
  const logSearchPostgresImage = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchPostgresImage
  );
  const logSearchPostgresInitImage = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchPostgresInitImage
  );
  const prometheusSelectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.prometheusSelectedStorageClass
  );
  const prometheusImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.prometheusImage
  );
  const prometheusSidecarImage = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.prometheusSidecarImage
  );
  const prometheusInitImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.prometheusInitImage
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "configure", field: field, value: value })
      );
    },
    [dispatch]
  );

  // Validation
  useEffect(() => {
    let customAccountValidation: IValidation[] = [];

    if (prometheusCustom) {
      customAccountValidation = [
        ...customAccountValidation,
        {
          fieldKey: "prometheus_storage_class",
          required: true,
          value: prometheusSelectedStorageClass,
          customValidation: prometheusSelectedStorageClass === "",
          customValidationMessage: "Field cannot be empty",
        },
        {
          fieldKey: "prometheus_volume_size",
          required: true,
          value: prometheusVolumeSize,
          customValidation:
            prometheusVolumeSize === "" || parseInt(prometheusVolumeSize) <= 0,
          customValidationMessage: `Volume size must be present and be greatter than 0`,
        },
      ];
    }
    if (logSearchCustom) {
      customAccountValidation = [
        ...customAccountValidation,
        {
          fieldKey: "log_search_storage_class",
          required: true,
          value: logSearchSelectedStorageClass,
          customValidation: logSearchSelectedStorageClass === "",
          customValidationMessage: "Field cannot be empty",
        },
        {
          fieldKey: "log_search_volume_size",
          required: true,
          value: logSearchVolumeSize,
          customValidation:
            logSearchVolumeSize === "" || parseInt(logSearchVolumeSize) <= 0,
          customValidationMessage: `Volume size must be present and be greatter than 0`,
        },
      ];
    }

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
          fieldKey: "logSearchImage",
          required: false,
          value: logSearchImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'minio/operator:VERSION'",
        },
        {
          fieldKey: "kesImage",
          required: false,
          value: kesImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage: "Format must be of form: 'minio/kes:VERSION'",
        },
        {
          fieldKey: "logSearchPostgresImage",
          required: false,
          value: logSearchPostgresImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'library/postgres:VERSION'",
        },
        {
          fieldKey: "logSearchPostgresInitImage",
          required: false,
          value: logSearchPostgresInitImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'library/busybox:VERSION'",
        },
        {
          fieldKey: "prometheusImage",
          required: false,
          value: prometheusImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'minio/prometheus:VERSION'",
        },
        {
          fieldKey: "prometheusSidecarImage",
          required: false,
          value: prometheusSidecarImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'project/container:VERSION'",
        },
        {
          fieldKey: "prometheusInitImage",
          required: false,
          value: prometheusInitImage,
          pattern: /^((.*?)\/(.*?):(.+))$/,
          customPatternMessage:
            "Format must be of form: 'library/busybox:VERSION'",
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
      })
    );

    setValidationErrors(commonVal);
  }, [
    customImage,
    imageName,
    logSearchImage,
    kesImage,
    logSearchPostgresImage,
    logSearchPostgresInitImage,
    prometheusImage,
    prometheusSidecarImage,
    prometheusInitImage,
    customDockerhub,
    imageRegistry,
    imageRegistryUsername,
    imageRegistryPassword,
    dispatch,
    prometheusCustom,
    tenantCustom,
    logSearchCustom,
    prometheusSelectedStorageClass,
    prometheusVolumeSize,
    logSearchSelectedStorageClass,
    logSearchVolumeSize,
  ]);

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  return (
    <Paper className={classes.paperWrapper}>
      <div className={classes.headerElement}>
        <H3Section>Container Images</H3Section>
        <span className={classes.descriptionText}>
          Specify the container images used by the Tenant and it's features.
        </span>
      </div>

      <Fragment>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="image"
            name="image"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("imageName", e.target.value);
              cleanValidation("image");
            }}
            label="MinIO"
            value={imageName}
            error={validationErrors["image"] || ""}
            placeholder="minio/minio:RELEASE.2022-02-26T02-54-46Z"
          />
        </Grid>

        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="kesImage"
            name="kesImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("kesImage", e.target.value);
              cleanValidation("kesImage");
            }}
            label="KES"
            value={kesImage}
            error={validationErrors["kesImage"] || ""}
            placeholder="minio/kes:v0.17.6"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <h4>Log Search</h4>
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="logSearchImage"
            name="logSearchImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("logSearchImage", e.target.value);
              cleanValidation("logSearchImage");
            }}
            label="API"
            value={logSearchImage}
            error={validationErrors["logSearchImage"] || ""}
            placeholder="minio/operator:v4.4.22"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="logSearchPostgresImage"
            name="logSearchPostgresImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("logSearchPostgresImage", e.target.value);
              cleanValidation("logSearchPostgresImage");
            }}
            label="PostgreSQL"
            value={logSearchPostgresImage}
            error={validationErrors["logSearchPostgresImage"] || ""}
            placeholder="library/postgres:13"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="logSearchPostgresInitImage"
            name="logSearchPostgresInitImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("logSearchPostgresInitImage", e.target.value);
              cleanValidation("logSearchPostgresInitImage");
            }}
            label="PostgreSQL Init"
            value={logSearchPostgresInitImage}
            error={validationErrors["logSearchPostgresInitImage"] || ""}
            placeholder="library/busybox:1.33.1"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <h4>Monitoring</h4>
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="prometheusImage"
            name="prometheusImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("prometheusImage", e.target.value);
              cleanValidation("prometheusImage");
            }}
            label="Prometheus"
            value={prometheusImage}
            error={validationErrors["prometheusImage"] || ""}
            placeholder="quay.io/prometheus/prometheus:latest"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="prometheusSidecarImage"
            name="prometheusSidecarImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("prometheusSidecarImage", e.target.value);
              cleanValidation("prometheusSidecarImage");
            }}
            label="Prometheus Sidecar"
            value={prometheusSidecarImage}
            error={validationErrors["prometheusSidecarImage"] || ""}
            placeholder="library/alpine:latest"
          />
        </Grid>
        <Grid item xs={12} className={classes.formFieldRow}>
          <InputBoxWrapper
            id="prometheusInitImage"
            name="prometheusInitImage"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("prometheusInitImage", e.target.value);
              cleanValidation("prometheusInitImage");
            }}
            label="Prometheus Init"
            value={prometheusInitImage}
            error={validationErrors["prometheusInitImage"] || ""}
            placeholder="library/busybox:1.33.1"
          />
        </Grid>
      </Fragment>

      {customImage && (
        <Fragment>
          <Grid item xs={12} className={classes.formFieldRow}>
            <h4>Custom Container Registry</h4>
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <FormSwitchWrapper
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
          </Grid>
        </Fragment>
      )}
      {customDockerhub && (
        <Fragment>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
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
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
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
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
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
          </Grid>
        </Fragment>
      )}
    </Paper>
  );
};

export default withStyles(styles)(Images);
