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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Grid, Paper, SelectChangeEvent } from "@mui/material";
import {
  createTenantCommon,
  modalBasic,
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
import SelectWrapper from "../../../Common/FormComponents/SelectWrapper/SelectWrapper";
import InputUnitMenu from "../../../Common/FormComponents/InputUnitMenu/InputUnitMenu";
import SectionH1 from "../../../Common/SectionH1";
import { isPageValid, updateAddField } from "../createTenantSlice";
import FormHr from "../../../Common/FormHr";

interface IConfigureProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    configSectionItem: {
      marginRight: 15,

      "& .multiContainer": {
        border: "1px solid red",
      },
    },
    fieldGroup: {
      ...createTenantCommon.fieldGroup,
      paddingTop: 15,
      marginBottom: 25,
    },
    responsiveSectionItem: {
      "@media (max-width: 900px)": {
        flexFlow: "column",
        alignItems: "flex-start",

        "& div > div": {
          marginBottom: 5,
          marginRight: 0,
        },
      },
    },

    fieldSpaceTop: {
      marginTop: 15,
    },
    ...modalBasic,
    ...wizardCommon,
  });

const ConfigLogSearch = ({ classes }: IConfigureProps) => {
  const dispatch = useAppDispatch();

  const storageClasses = useSelector(
    (state: AppState) => state.createTenant.storageClasses
  );
  const logSearchEnabled = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchEnabled
  );
  const logSearchVolumeSize = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchVolumeSize
  );
  const logSearchSelectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchSelectedStorageClass
  );
  const logSearchImage = useSelector(
    (state: AppState) => state.createTenant.fields.configure.logSearchImage
  );
  const logSearchPostgresImage = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchPostgresImage
  );
  const logSearchPostgresInitImage = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchPostgresInitImage
  );
  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass
  );
  const tenantSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.tenantSecurityContext
  );
  const logSearchSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchSecurityContext
  );
  const logSearchPostgresSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.logSearchPostgresSecurityContext
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  const configureSTClasses = [
    { label: "Default", value: "default" },
    ...storageClasses,
  ];

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

    if (logSearchEnabled) {
      customAccountValidation = [
        {
          fieldKey: "log_search_storage_class",
          required: true,
          value: logSearchSelectedStorageClass!,
          customValidation: logSearchSelectedStorageClass === "",
          customValidationMessage: "Field cannot be empty",
        },
        {
          fieldKey: "log_search_volume_size",
          required: true,
          value: `${logSearchVolumeSize}`,
          customValidation:
            logSearchVolumeSize === "" || parseInt(logSearchVolumeSize) <= 0,
          customValidationMessage: `Volume size must be present and be greatter than 0`,
        },
        {
          fieldKey: "logSearch_securityContext_runAsUser",
          required: true,
          value: logSearchSecurityContext.runAsUser,
          customValidation:
            logSearchSecurityContext.runAsUser === "" ||
            parseInt(logSearchSecurityContext.runAsUser) < 0,
          customValidationMessage: `runAsUser must be present and be 0 or more`,
        },
        {
          fieldKey: "logSearch_securityContext_runAsGroup",
          required: true,
          value: logSearchSecurityContext.runAsGroup,
          customValidation:
            logSearchSecurityContext.runAsGroup === "" ||
            parseInt(logSearchSecurityContext.runAsGroup) < 0,
          customValidationMessage: `runAsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "logSearch_securityContext_fsGroup",
          required: true,
          value: logSearchSecurityContext.fsGroup!,
          customValidation:
            logSearchSecurityContext.fsGroup === "" ||
            parseInt(logSearchSecurityContext.fsGroup!) < 0,
          customValidationMessage: `fsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "postgres_securityContext_runAsUser",
          required: true,
          value: logSearchPostgresSecurityContext.runAsUser,
          customValidation:
            logSearchPostgresSecurityContext.runAsUser === "" ||
            parseInt(logSearchPostgresSecurityContext.runAsUser) < 0,
          customValidationMessage: `runAsUser must be present and be 0 or more`,
        },
        {
          fieldKey: "postgres_securityContext_runAsGroup",
          required: true,
          value: logSearchSecurityContext.runAsGroup,
          customValidation:
            logSearchPostgresSecurityContext.runAsGroup === "" ||
            parseInt(logSearchPostgresSecurityContext.runAsGroup) < 0,
          customValidationMessage: `runAsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "postgres_securityContext_fsGroup",
          required: true,
          value: logSearchPostgresSecurityContext.fsGroup!,
          customValidation:
            logSearchPostgresSecurityContext.fsGroup === "" ||
            parseInt(logSearchPostgresSecurityContext.fsGroup!) < 0,
          customValidationMessage: `fsGroup must be present and be 0 or more`,
        },
      ];
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
    logSearchImage,
    logSearchPostgresImage,
    logSearchPostgresInitImage,
    dispatch,
    logSearchEnabled,
    logSearchSelectedStorageClass,
    logSearchVolumeSize,
    tenantSecurityContext,
    logSearchSecurityContext,
    logSearchPostgresSecurityContext,
  ]);

  useEffect(() => {
    // New default values in current selection is invalid
    if (storageClasses.length > 0) {
      const filterLogSearch = storageClasses.filter(
        (item: any) => item.value === logSearchSelectedStorageClass
      );
      if (filterLogSearch.length === 0) {
        updateField("logSearchSelectedStorageClass", "default");
      }
    }
  }, [
    logSearchSelectedStorageClass,
    selectedStorageClass,
    storageClasses,
    updateField,
  ]);

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  return (
    <Paper className={classes.paperWrapper}>
      <Grid container alignItems={"center"}>
        <Grid item xs>
          <SectionH1>Audit Log</SectionH1>
        </Grid>
        <Grid item xs={4}>
          <FormSwitchWrapper
            value="enableLogging"
            id="enableLogging"
            name="enableLogging"
            checked={logSearchEnabled}
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;

              updateField("logSearchEnabled", checked);
            }}
            indicatorLabels={["Enabled", "Disabled"]}
          />
        </Grid>
      </Grid>
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <span className={classes.descriptionText}>
            Deploys a small PostgreSQL database and stores access logs of all
            calls into the tenant.
          </span>
        </Grid>
        <Grid xs={12}>
          <FormHr />
        </Grid>
        {logSearchEnabled && (
          <Fragment>
            <Grid item xs={12}>
              <SelectWrapper
                id="log_search_storage_class"
                name="log_search_storage_class"
                onChange={(e: SelectChangeEvent<string>) => {
                  updateField(
                    "logSearchSelectedStorageClass",
                    e.target.value as string
                  );
                }}
                label="Log Search Storage Class"
                value={logSearchSelectedStorageClass}
                options={configureSTClasses}
                disabled={configureSTClasses.length < 1}
              />
            </Grid>
            <Grid item xs={12}>
              <div className={classes.multiContainer}>
                <InputBoxWrapper
                  type="number"
                  id="log_search_volume_size"
                  name="log_search_volume_size"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    updateField("logSearchVolumeSize", e.target.value);
                    cleanValidation("log_search_volume_size");
                  }}
                  label="Storage Size"
                  overlayObject={
                    <InputUnitMenu
                      id={"size-unit"}
                      onUnitChange={() => {}}
                      unitSelected={"Gi"}
                      unitsList={[{ label: "Gi", value: "Gi" }]}
                      disabled={true}
                    />
                  }
                  value={logSearchVolumeSize}
                  required
                  error={validationErrors["log_search_volume_size"] || ""}
                  min="0"
                />
              </div>
            </Grid>

            <fieldset
              className={`${classes.fieldGroup} ${classes.fieldSpaceTop}`}
            >
              <legend className={classes.descriptionText}>
                SecurityContext for LogSearch
              </legend>

              <Grid item xs={12}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="logSearch_securityContext_runAsUser"
                      name="logSearch_securityContext_runAsUser"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchSecurityContext", {
                          ...logSearchSecurityContext,
                          runAsUser: e.target.value,
                        });
                        cleanValidation("logSearch_securityContext_runAsUser");
                      }}
                      label="Run As User"
                      value={logSearchSecurityContext.runAsUser}
                      required
                      error={
                        validationErrors[
                          "logSearch_securityContext_runAsUser"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="logSearch_securityContext_runAsGroup"
                      name="logSearch_securityContext_runAsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchSecurityContext", {
                          ...logSearchSecurityContext,
                          runAsGroup: e.target.value,
                        });
                        cleanValidation("logSearch_securityContext_runAsGroup");
                      }}
                      label="Run As Group"
                      value={logSearchSecurityContext.runAsGroup}
                      required
                      error={
                        validationErrors[
                          "logSearch_securityContext_runAsGroup"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                </div>
              </Grid>
              <br />
              <Grid item xs={12}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="logSearch_securityContext_fsGroup"
                      name="logSearch_securityContext_fsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchSecurityContext", {
                          ...logSearchSecurityContext,
                          fsGroup: e.target.value,
                        });
                        cleanValidation("logSearch_securityContext_fsGroup");
                      }}
                      label="FsGroup"
                      value={logSearchSecurityContext.fsGroup!}
                      required
                      error={
                        validationErrors["logSearch_securityContext_fsGroup"] ||
                        ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <SelectWrapper
                      label="FsGroupChangePolicy"
                      id="securityContext_fsGroupChangePolicy"
                      name="securityContext_fsGroupChangePolicy"
                      value={logSearchSecurityContext.fsGroupChangePolicy!}
                      onChange={(e: SelectChangeEvent<string>) => {
                        updateField("logSearchSecurityContext", {
                          ...logSearchSecurityContext,
                          fsGroupChangePolicy: e.target.value,
                        });
                      }}
                      options={[
                        {
                          label: "Always",
                          value: "Always",
                        },
                        {
                          label: "OnRootMismatch",
                          value: "OnRootMismatch",
                        },
                      ]}
                    />
                  </div>
                </div>
              </Grid>
              <br />
              <Grid item xs={12}>
                <div className={classes.multiContainer}>
                  <FormSwitchWrapper
                    value="logSearchSecurityContextRunAsNonRoot"
                    id="logSearch_securityContext_runAsNonRoot"
                    name="logSearch_securityContext_runAsNonRoot"
                    checked={logSearchSecurityContext.runAsNonRoot}
                    onChange={(e) => {
                      const targetD = e.target;
                      const checked = targetD.checked;
                      updateField("logSearchSecurityContext", {
                        ...logSearchSecurityContext,
                        runAsNonRoot: checked,
                      });
                    }}
                    label={"Do not run as Root"}
                  />
                </div>
              </Grid>
            </fieldset>
            <fieldset className={classes.fieldGroup}>
              <legend className={classes.descriptionText}>
                SecurityContext for PostgreSQL
              </legend>

              <Grid item xs={12}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="postgres_securityContext_runAsUser"
                      name="postgres_securityContext_runAsUser"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchPostgresSecurityContext", {
                          ...logSearchPostgresSecurityContext,
                          runAsUser: e.target.value,
                        });
                        cleanValidation("postgres_securityContext_runAsUser");
                      }}
                      label="Run As User"
                      value={logSearchPostgresSecurityContext.runAsUser}
                      required
                      error={
                        validationErrors[
                          "postgres_securityContext_runAsUser"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="postgres_securityContext_runAsGroup"
                      name="postgres_securityContext_runAsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchPostgresSecurityContext", {
                          ...logSearchPostgresSecurityContext,
                          runAsGroup: e.target.value,
                        });
                        cleanValidation("postgres_securityContext_runAsGroup");
                      }}
                      label="Run As Group"
                      value={logSearchPostgresSecurityContext.runAsGroup}
                      required
                      error={
                        validationErrors[
                          "postgres_securityContext_runAsGroup"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                </div>
              </Grid>
              <br />
              <Grid item xs={12}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="postgres_securityContext_fsGroup"
                      name="postgres_securityContext_fsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("logSearchPostgresSecurityContext", {
                          ...logSearchPostgresSecurityContext,
                          fsGroup: e.target.value,
                        });
                        cleanValidation("postgres_securityContext_fsGroup");
                      }}
                      label="FsGroup"
                      value={logSearchPostgresSecurityContext.fsGroup!}
                      required
                      error={
                        validationErrors["postgres_securityContext_fsGroup"] ||
                        ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <SelectWrapper
                      label="FsGroupChangePolicy"
                      id="securityContext_fsGroupChangePolicy"
                      name="securityContext_fsGroupChangePolicy"
                      value={
                        logSearchPostgresSecurityContext.fsGroupChangePolicy!
                      }
                      onChange={(e: SelectChangeEvent<string>) => {
                        updateField("logSearchPostgresSecurityContext", {
                          ...logSearchPostgresSecurityContext,
                          fsGroupChangePolicy: e.target.value,
                        });
                      }}
                      options={[
                        {
                          label: "Always",
                          value: "Always",
                        },
                        {
                          label: "OnRootMismatch",
                          value: "OnRootMismatch",
                        },
                      ]}
                    />
                  </div>
                </div>
              </Grid>
              <br />
              <Grid item xs={12}>
                <div className={classes.multiContainer}>
                  <FormSwitchWrapper
                    value="postgresSecurityContextRunAsNonRoot"
                    id="postgres_securityContext_runAsNonRoot"
                    name="postgres_securityContext_runAsNonRoot"
                    checked={logSearchPostgresSecurityContext.runAsNonRoot}
                    onChange={(e) => {
                      const targetD = e.target;
                      const checked = targetD.checked;
                      updateField("logSearchPostgresSecurityContext", {
                        ...logSearchPostgresSecurityContext,
                        runAsNonRoot: checked,
                      });
                    }}
                    label={"Do not run as Root"}
                  />
                </div>
              </Grid>
            </fieldset>
          </Fragment>
        )}
      </Grid>
    </Paper>
  );
};

export default withStyles(styles)(ConfigLogSearch);
