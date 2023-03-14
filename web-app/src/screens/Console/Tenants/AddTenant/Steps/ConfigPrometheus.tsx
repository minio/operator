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

const ConfigPrometheus = ({ classes }: IConfigureProps) => {
  const dispatch = useAppDispatch();

  const storageClasses = useSelector(
    (state: AppState) => state.createTenant.storageClasses
  );
  const prometheusEnabled = useSelector(
    (state: AppState) => state.createTenant.fields.configure.prometheusEnabled
  );
  const prometheusVolumeSize = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.prometheusVolumeSize
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
  const selectedStorageClass = useSelector(
    (state: AppState) =>
      state.createTenant.fields.nameTenant.selectedStorageClass
  );
  const tenantSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.tenantSecurityContext
  );
  const prometheusSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.prometheusSecurityContext
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

    if (prometheusEnabled) {
      customAccountValidation = [
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
          customValidationMessage: `Volume size must be present and be greater than 0`,
        },
        {
          fieldKey: "prometheus_securityContext_runAsUser",
          required: true,
          value: prometheusSecurityContext.runAsUser,
          customValidation:
            prometheusSecurityContext.runAsUser === "" ||
            parseInt(prometheusSecurityContext.runAsUser) < 0,
          customValidationMessage: `runAsUser must be present and be 0 or more`,
        },
        {
          fieldKey: "prometheus_securityContext_runAsGroup",
          required: true,
          value: prometheusSecurityContext.runAsGroup,
          customValidation:
            prometheusSecurityContext.runAsGroup === "" ||
            parseInt(prometheusSecurityContext.runAsGroup) < 0,
          customValidationMessage: `runAsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "prometheus_securityContext_fsGroup",
          required: true,
          value: prometheusSecurityContext.fsGroup!,
          customValidation:
            prometheusSecurityContext.fsGroup === "" ||
            parseInt(prometheusSecurityContext.fsGroup!) < 0,
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
    prometheusImage,
    prometheusSidecarImage,
    prometheusInitImage,
    dispatch,
    prometheusEnabled,
    prometheusSelectedStorageClass,
    prometheusVolumeSize,
    tenantSecurityContext,
    prometheusSecurityContext,
  ]);

  useEffect(() => {
    // New default values in current selection is invalid
    if (storageClasses.length > 0) {
      const filterPrometheus = storageClasses.filter(
        (item: any) => item.value === prometheusSelectedStorageClass
      );
      if (filterPrometheus.length === 0) {
        updateField("prometheusSelectedStorageClass", "default");
      }
    }
  }, [
    prometheusSelectedStorageClass,
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
          <SectionH1>Monitoring</SectionH1>
        </Grid>
        <Grid item xs={4}>
          <FormSwitchWrapper
            indicatorLabels={["Enabled", "Disabled"]}
            checked={prometheusEnabled}
            value={"monitoring_status"}
            id="monitoring-status"
            name="monitoring-status"
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;

              updateField("prometheusEnabled", checked);
            }}
            description=""
          />
        </Grid>
      </Grid>
      <Grid item xs={12}>
        <span className={classes.descriptionText}>
          A small Prometheus will be deployed to keep metrics about the tenant.
        </span>
      </Grid>
      <Grid xs={12}>
        <FormHr />
      </Grid>
      <Grid container spacing={1}>
        {prometheusEnabled && (
          <Fragment>
            <Grid item xs={12}>
              <SelectWrapper
                id="prometheus_storage_class"
                name="prometheus_storage_class"
                onChange={(e: SelectChangeEvent<string>) => {
                  updateField(
                    "prometheusSelectedStorageClass",
                    e.target.value as string
                  );
                }}
                label="Storage Class"
                value={prometheusSelectedStorageClass}
                options={configureSTClasses}
                disabled={configureSTClasses.length < 1}
              />
            </Grid>
            <Grid item xs={12}>
              <div className={classes.multiContainer}>
                <InputBoxWrapper
                  type="number"
                  id="prometheus_volume_size"
                  name="prometheus_volume_size"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    updateField("prometheusVolumeSize", e.target.value);
                    cleanValidation("prometheus_volume_size");
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
                  value={prometheusVolumeSize}
                  required
                  error={validationErrors["prometheus_volume_size"] || ""}
                  min="0"
                />
              </div>
            </Grid>
            <fieldset
              className={`${classes.fieldGroup} ${classes.fieldSpaceTop}`}
            >
              <legend className={classes.descriptionText}>
                SecurityContext
              </legend>
              <Grid item xs={12} className={classes.configSectionItem}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="prometheus_securityContext_runAsUser"
                      name="prometheus_securityContext_runAsUser"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("prometheusSecurityContext", {
                          ...prometheusSecurityContext,
                          runAsUser: e.target.value,
                        });
                        cleanValidation("prometheus_securityContext_runAsUser");
                      }}
                      label="Run As User"
                      value={prometheusSecurityContext.runAsUser}
                      required
                      error={
                        validationErrors[
                          "prometheus_securityContext_runAsUser"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="prometheus_securityContext_runAsGroup"
                      name="prometheus_securityContext_runAsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("prometheusSecurityContext", {
                          ...prometheusSecurityContext,
                          runAsGroup: e.target.value,
                        });
                        cleanValidation(
                          "prometheus_securityContext_runAsGroup"
                        );
                      }}
                      label="Run As Group"
                      value={prometheusSecurityContext.runAsGroup}
                      required
                      error={
                        validationErrors[
                          "prometheus_securityContext_runAsGroup"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                </div>
              </Grid>
              <br />
              <Grid item xs={12} className={classes.configSectionItem}>
                <div
                  className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
                >
                  <div className={classes.configSectionItem}>
                    <InputBoxWrapper
                      type="number"
                      id="prometheus_securityContext_fsGroup"
                      name="prometheus_securityContext_fsGroup"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        updateField("prometheusSecurityContext", {
                          ...prometheusSecurityContext,
                          fsGroup: e.target.value,
                        });
                        cleanValidation("prometheus_securityContext_fsGroup");
                      }}
                      label="FsGroup"
                      value={prometheusSecurityContext.fsGroup!}
                      required
                      error={
                        validationErrors[
                          "prometheus_securityContext_fsGroup"
                        ] || ""
                      }
                      min="0"
                    />
                  </div>
                  <div className={classes.configSectionItem}>
                    <SelectWrapper
                      label="FsGroupChangePolicy"
                      id="securityContext_fsGroupChangePolicy"
                      name="securityContext_fsGroupChangePolicy"
                      value={prometheusSecurityContext.fsGroupChangePolicy!}
                      onChange={(e: SelectChangeEvent<string>) => {
                        updateField("prometheusSecurityContext", {
                          ...prometheusSecurityContext,
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
              <Grid item xs={12} className={classes.configSectionItem}>
                <div
                  className={`${classes.multiContainer} ${classes.fieldSpaceTop}`}
                >
                  <FormSwitchWrapper
                    value="prometheusSecurityContextRunAsNonRoot"
                    id="prometheus_securityContext_runAsNonRoot"
                    name="prometheus_securityContext_runAsNonRoot"
                    checked={prometheusSecurityContext.runAsNonRoot}
                    onChange={(e) => {
                      const targetD = e.target;
                      const checked = targetD.checked;
                      updateField("prometheusSecurityContext", {
                        ...prometheusSecurityContext,
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

export default withStyles(styles)(ConfigPrometheus);
