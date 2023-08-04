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

import React, { useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  Divider,
  Grid,
  IconButton,
  Paper,
  SelectChangeEvent,
} from "@mui/material";
import {
  createTenantCommon,
  formFieldStyles,
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
import AddIcon from "@mui/icons-material/Add";
import { RemoveIcon } from "mds";
import {
  addNewMinIODomain,
  isPageValid,
  removeMinIODomain,
  setEnvVars,
  updateAddField,
} from "../createTenantSlice";
import SelectWrapper from "../../../Common/FormComponents/SelectWrapper/SelectWrapper";
import H3Section from "../../../Common/H3Section";

interface IConfigureProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    configSectionItem: {
      marginRight: 15,
      marginBottom: 15,

      "& .multiContainer": {
        border: "1px solid red",
      },
    },
    tenantCustomizationFields: {
      marginLeft: 30, // 2nd Level(15+15)
      width: "88%",
      margin: "auto",
    },
    containerItem: {
      marginRight: 15,
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
    wrapperContainer: {
      display: "flex",
      marginBottom: 15,
    },
    envVarRow: {
      display: "flex",
      alignItems: "center",
      justifyContent: "flex-start",
      "&:last-child": {
        borderBottom: 0,
      },
      "@media (max-width: 900px)": {
        flex: 1,

        "& div label": {
          minWidth: 50,
        },
      },
    },
    fileItem: {
      marginRight: 10,
      display: "flex",
      "& div label": {
        minWidth: 50,
      },

      "@media (max-width: 900px)": {
        flexFlow: "column",
      },
    },
    rowActions: {
      display: "flex",
      justifyContent: "flex-end",
      "@media (max-width: 900px)": {
        flex: 1,
      },
    },
    overlayAction: {
      marginLeft: 10,
      "& svg": {
        maxWidth: 15,
        maxHeight: 15,
      },
      "& button": {
        background: "#EAEAEA",
      },
    },
    ...modalBasic,
    ...wizardCommon,
    ...formFieldStyles,
  });

const Configure = ({ classes }: IConfigureProps) => {
  const dispatch = useAppDispatch();

  const exposeMinIO = useSelector(
    (state: AppState) => state.createTenant.fields.configure.exposeMinIO,
  );
  const exposeConsole = useSelector(
    (state: AppState) => state.createTenant.fields.configure.exposeConsole,
  );
  const exposeSFTP = useSelector(
    (state: AppState) => state.createTenant.fields.configure.exposeSFTP,
  );
  const setDomains = useSelector(
    (state: AppState) => state.createTenant.fields.configure.setDomains,
  );
  const consoleDomain = useSelector(
    (state: AppState) => state.createTenant.fields.configure.consoleDomain,
  );
  const minioDomains = useSelector(
    (state: AppState) => state.createTenant.fields.configure.minioDomains,
  );
  const tenantCustom = useSelector(
    (state: AppState) => state.createTenant.fields.configure.tenantCustom,
  );
  const tenantEnvVars = useSelector(
    (state: AppState) => state.createTenant.fields.configure.envVars,
  );
  const tenantSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.configure.tenantSecurityContext,
  );
  const customRuntime = useSelector(
    (state: AppState) => state.createTenant.fields.configure.customRuntime,
  );
  const runtimeClassName = useSelector(
    (state: AppState) => state.createTenant.fields.configure.runtimeClassName,
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
    if (tenantCustom) {
      customAccountValidation = [
        {
          fieldKey: "tenant_securityContext_runAsUser",
          required: true,
          value: tenantSecurityContext.runAsUser,
          customValidation:
            tenantSecurityContext.runAsUser === "" ||
            parseInt(tenantSecurityContext.runAsUser) < 0,
          customValidationMessage: `runAsUser must be present and be 0 or more`,
        },
        {
          fieldKey: "tenant_securityContext_runAsGroup",
          required: true,
          value: tenantSecurityContext.runAsGroup,
          customValidation:
            tenantSecurityContext.runAsGroup === "" ||
            parseInt(tenantSecurityContext.runAsGroup) < 0,
          customValidationMessage: `runAsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "tenant_securityContext_fsGroup",
          required: true,
          value: tenantSecurityContext.fsGroup!,
          customValidation:
            tenantSecurityContext.fsGroup === "" ||
            parseInt(tenantSecurityContext.fsGroup!) < 0,
          customValidationMessage: `fsGroup must be present and be 0 or more`,
        },
      ];
    }

    if (setDomains) {
      const minioExtraValidations = minioDomains.map((validation, index) => {
        return {
          fieldKey: `minio-domain-${index.toString()}`,
          required: false,
          value: validation,
          pattern: /^(https?):\/\/([a-zA-Z0-9\-.]+)(:[0-9]+)?$/,
          customPatternMessage:
            "MinIO domain is not in the form of http|https://subdomain.domain",
        };
      });

      customAccountValidation = [
        ...customAccountValidation,
        ...minioExtraValidations,
        {
          fieldKey: "console_domain",
          required: false,
          value: consoleDomain,
          pattern:
            /^(https?):\/\/([a-zA-Z0-9\-.]+)(:[0-9]+)?(\/[a-zA-Z0-9\-./]*)?$/,
          customPatternMessage:
            "Console domain is not in the form of http|https://subdomain.domain:port/subpath1/subpath2",
        },
      ];
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
    dispatch,
    tenantCustom,
    tenantSecurityContext,
    setDomains,
    consoleDomain,
    minioDomains,
  ]);

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  const updateMinIODomain = (value: string, index: number) => {
    const copyDomains = [...minioDomains];
    copyDomains[index] = value;

    updateField("minioDomains", copyDomains);
  };

  return (
    <Paper className={classes.paperWrapper}>
      <div className={classes.headerElement}>
        <H3Section>Configure</H3Section>
        <span className={classes.descriptionText}>
          Basic configurations for tenant management
        </span>
      </div>
      <div className={classes.headerElement}>
        <h4 className={classes.h3Section}>Services</h4>
        <span className={classes.descriptionText}>
          Whether the tenant's services should request an external IP via
          LoadBalancer service type.
        </span>
      </div>
      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="expose_minio"
          id="expose_minio"
          name="expose_minio"
          checked={exposeMinIO}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("exposeMinIO", checked);
          }}
          label={"Expose MinIO Service"}
        />
      </Grid>
      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="expose_console"
          id="expose_console"
          name="expose_console"
          checked={exposeConsole}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("exposeConsole", checked);
          }}
          label={"Expose Console Service"}
        />
      </Grid>
      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="expose_sftp"
          id="expose_sftp"
          name="expose_sftp"
          checked={exposeSFTP}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("exposeSFTP", checked);
          }}
          label={"Expose SFTP Service"}
        />
      </Grid>
      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="custom_domains"
          id="custom_domains"
          name="custom_domains"
          checked={setDomains}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("setDomains", checked);
          }}
          label={"Set Custom Domains"}
        />
      </Grid>
      {setDomains && (
        <Grid item xs={12} className={classes.tenantCustomizationFields}>
          <fieldset className={classes.fieldGroup}>
            <legend className={classes.descriptionText}>
              Custom Domains for MinIO
            </legend>
            <Grid item xs={12} className={`${classes.configSectionItem}`}>
              <div className={classes.containerItem}>
                <InputBoxWrapper
                  id="console_domain"
                  name="console_domain"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    updateField("consoleDomain", e.target.value);
                    cleanValidation("tenant_securityContext_runAsUser");
                  }}
                  label="Console Domain"
                  value={consoleDomain}
                  placeholder={
                    "Eg. http://subdomain.domain:port/subpath1/subpath2"
                  }
                  error={validationErrors["console_domain"] || ""}
                />
              </div>
              <div>
                <h4>MinIO Domains</h4>
                <div className={`${classes.responsiveSectionItem}`}>
                  {minioDomains.map((domain, index) => {
                    return (
                      <div
                        className={`${classes.containerItem} ${classes.wrapperContainer}`}
                        key={`minio-domain-key-${index.toString()}`}
                      >
                        <InputBoxWrapper
                          id={`minio-domain-${index.toString()}`}
                          name={`minio-domain-${index.toString()}`}
                          onChange={(
                            e: React.ChangeEvent<HTMLInputElement>,
                          ) => {
                            updateMinIODomain(e.target.value, index);
                          }}
                          label={`MinIO Domain ${index + 1}`}
                          value={domain}
                          placeholder={"Eg. http://subdomain.domain"}
                          error={
                            validationErrors[
                              `minio-domain-${index.toString()}`
                            ] || ""
                          }
                        />
                        <div className={classes.overlayAction}>
                          <IconButton
                            size={"small"}
                            onClick={() => dispatch(addNewMinIODomain())}
                            disabled={index !== minioDomains.length - 1}
                          >
                            <AddIcon />
                          </IconButton>
                        </div>

                        <div className={classes.overlayAction}>
                          <IconButton
                            size={"small"}
                            onClick={() => dispatch(removeMinIODomain(index))}
                            disabled={minioDomains.length <= 1}
                          >
                            <RemoveIcon />
                          </IconButton>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </Grid>
          </fieldset>
        </Grid>
      )}

      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="tenantConfig"
          id="tenant_configuration"
          name="tenant_configuration"
          checked={tenantCustom}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("tenantCustom", checked);
          }}
          label={"Security Context"}
        />
      </Grid>
      {tenantCustom && (
        <Grid item xs={12} className={classes.tenantCustomizationFields}>
          <fieldset className={classes.fieldGroup}>
            <legend className={classes.descriptionText}>
              SecurityContext for MinIO
            </legend>
            <Grid item xs={12} className={`${classes.configSectionItem}`}>
              <div
                className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
              >
                <div className={classes.containerItem}>
                  <InputBoxWrapper
                    type="number"
                    id="tenant_securityContext_runAsUser"
                    name="tenant_securityContext_runAsUser"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("tenantSecurityContext", {
                        ...tenantSecurityContext,
                        runAsUser: e.target.value,
                      });
                      cleanValidation("tenant_securityContext_runAsUser");
                    }}
                    label="Run As User"
                    value={tenantSecurityContext.runAsUser}
                    required
                    error={
                      validationErrors["tenant_securityContext_runAsUser"] || ""
                    }
                    min="0"
                  />
                </div>
                <div className={classes.containerItem}>
                  <InputBoxWrapper
                    type="number"
                    id="tenant_securityContext_runAsGroup"
                    name="tenant_securityContext_runAsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("tenantSecurityContext", {
                        ...tenantSecurityContext,
                        runAsGroup: e.target.value,
                      });
                      cleanValidation("tenant_securityContext_runAsGroup");
                    }}
                    label="Run As Group"
                    value={tenantSecurityContext.runAsGroup}
                    required
                    error={
                      validationErrors["tenant_securityContext_runAsGroup"] ||
                      ""
                    }
                    min="0"
                  />
                </div>
              </div>
            </Grid>
            <br />
            <Grid item xs={12} className={`${classes.configSectionItem}`}>
              <div
                className={`${classes.multiContainer} ${classes.responsiveSectionItem}`}
              >
                <div className={classes.containerItem}>
                  <InputBoxWrapper
                    type="number"
                    id="tenant_securityContext_fsGroup"
                    name="tenant_securityContext_fsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("tenantSecurityContext", {
                        ...tenantSecurityContext,
                        fsGroup: e.target.value,
                      });
                      cleanValidation("tenant_securityContext_fsGroup");
                    }}
                    label="FsGroup"
                    value={tenantSecurityContext.fsGroup!}
                    required
                    error={
                      validationErrors["tenant_securityContext_fsGroup"] || ""
                    }
                    min="0"
                  />
                </div>
                <div className={classes.containerItem}>
                  <div className={classes.configSectionItem}>
                    <SelectWrapper
                      label="FsGroupChangePolicy"
                      id="securityContext_fsGroupChangePolicy"
                      name="securityContext_fsGroupChangePolicy"
                      value={tenantSecurityContext.fsGroupChangePolicy!}
                      onChange={(e: SelectChangeEvent<string>) => {
                        updateField("tenantSecurityContext", {
                          ...tenantSecurityContext,
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
              </div>
            </Grid>
            <br />
            <Grid item xs={12} className={classes.configSectionItem}>
              <div className={classes.multiContainer}>
                <FormSwitchWrapper
                  value="tenantSecurityContextRunAsNonRoot"
                  id="tenant_securityContext_runAsNonRoot"
                  name="tenant_securityContext_runAsNonRoot"
                  checked={tenantSecurityContext.runAsNonRoot}
                  onChange={(e) => {
                    const targetD = e.target;
                    const checked = targetD.checked;
                    updateField("tenantSecurityContext", {
                      ...tenantSecurityContext,
                      runAsNonRoot: checked,
                    });
                  }}
                  label={"Do not run as Root"}
                />
              </div>
            </Grid>
          </fieldset>
        </Grid>
      )}
      <Grid item xs={12} className={classes.configSectionItem}>
        <FormSwitchWrapper
          value="customRuntime"
          id="tenant_custom_runtime"
          name="tenant_custom_runtime"
          checked={customRuntime}
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("customRuntime", checked);
          }}
          label={"Custom Runtime Configurations"}
        />
      </Grid>
      {customRuntime && (
        <Grid item xs={12} className={classes.tenantCustomizationFields}>
          <fieldset className={classes.fieldGroup}>
            <legend className={classes.descriptionText}>
              Custom Runtime Configurations
            </legend>
            <Grid item xs={12} className={`${classes.configSectionItem}`}>
              <div className={classes.containerItem}>
                <InputBoxWrapper
                  id="tenant_runtime_runtimeClassName"
                  name="tenant_runtime_runtimeClassName"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    updateField("runtimeClassName", e.target.value);
                    cleanValidation("tenant_runtime_runtimeClassName");
                  }}
                  label="Runtime Class Name"
                  value={runtimeClassName}
                  error={
                    validationErrors["tenant_runtime_runtimeClassName"] || ""
                  }
                />
              </div>
            </Grid>
          </fieldset>
        </Grid>
      )}
      <Divider />

      <div className={classes.headerElement}>
        <H3Section>Additional Environment Variables</H3Section>
        <span className={classes.descriptionText}>
          Define additional environment variables to be used by your MinIO pods
        </span>
      </div>
      <Grid container>
        {tenantEnvVars.map((envVar, index) => (
          <Grid
            item
            xs={12}
            className={`${classes.formFieldRow} ${classes.envVarRow}`}
            key={`tenant-envVar-${index.toString()}`}
          >
            <Grid item xs={5} className={classes.fileItem}>
              <InputBoxWrapper
                id="env_var_key"
                name="env_var_key"
                label="Key"
                value={envVar.key}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  const existingEnvVars = [...tenantEnvVars];
                  dispatch(
                    setEnvVars(
                      existingEnvVars.map((keyPair, i) =>
                        i === index
                          ? { key: e.target.value, value: keyPair.value }
                          : keyPair,
                      ),
                    ),
                  );
                }}
                index={index}
                key={`env_var_key_${index.toString()}`}
              />
            </Grid>
            <Grid item xs={5} className={classes.fileItem}>
              <InputBoxWrapper
                id="env_var_value"
                name="env_var_value"
                label="Value"
                value={envVar.value}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  const existingEnvVars = [...tenantEnvVars];
                  dispatch(
                    setEnvVars(
                      existingEnvVars.map((keyPair, i) =>
                        i === index
                          ? { key: keyPair.key, value: e.target.value }
                          : keyPair,
                      ),
                    ),
                  );
                }}
                index={index}
                key={`env_var_value_${index.toString()}`}
              />
            </Grid>
            <Grid item xs={2} className={classes.rowActions}>
              <div className={classes.overlayAction}>
                <IconButton
                  size={"small"}
                  onClick={() => {
                    const existingEnvVars = [...tenantEnvVars];
                    existingEnvVars.push({ key: "", value: "" });

                    dispatch(setEnvVars(existingEnvVars));
                  }}
                  disabled={index !== tenantEnvVars.length - 1}
                >
                  <AddIcon />
                </IconButton>
              </div>
              <div className={classes.overlayAction}>
                <IconButton
                  size={"small"}
                  onClick={() => {
                    const existingEnvVars = tenantEnvVars.filter(
                      (item, fIndex) => fIndex !== index,
                    );
                    dispatch(setEnvVars(existingEnvVars));
                  }}
                  disabled={tenantEnvVars.length <= 1}
                >
                  <RemoveIcon />
                </IconButton>
              </div>
            </Grid>
          </Grid>
        ))}
      </Grid>
    </Paper>
  );
};

export default withStyles(styles)(Configure);
