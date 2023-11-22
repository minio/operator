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
import {
  Box,
  FormLayout,
  Grid,
  IconButton,
  InputBox,
  RemoveIcon,
  Select,
  Switch,
  AddIcon,
} from "mds";
import { useSelector } from "react-redux";
import styled from "styled-components";
import { AppState, useAppDispatch } from "../../../../../store";
import { clearValidationError } from "../../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import {
  addNewMinIODomain,
  isPageValid,
  removeMinIODomain,
  setEnvVars,
  updateAddField,
} from "../createTenantSlice";
import H3Section from "../../../Common/H3Section";

const ConfigureMain = styled.div(() => ({
  "& .configSectionItem": {
    marginRight: 15,
    marginBottom: 15,
  },
  "& .containerItem": {
    marginRight: 15,
  },
  "& .responsiveSectionItem": {
    "&.doubleElement": {
      display: "flex",
      "& div": {
        flexGrow: 1,
      },
    },
    "@media (max-width: 900px)": {
      flexFlow: "column",
      alignItems: "flex-start",

      "& div > div": {
        marginBottom: 5,
        marginRight: 0,
      },
    },
  },
  "& .wrapperContainer": {
    display: "flex",
    alignItems: "center",
  },
  "& .envVarRow": {
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
  "& .fileItem": {
    marginRight: 10,
    display: "flex",
    "& div label": {
      minWidth: 50,
    },

    "@media (max-width: 900px)": {
      flexFlow: "column",
    },
  },
  "& .rowActions": {
    display: "flex",
    justifyContent: "flex-end",
    "@media (max-width: 900px)": {
      flex: 1,
    },
  },
  "& .overlayAction": {
    marginLeft: 10,
    marginBottom: 15,
  },
}));

const Configure = () => {
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
    <ConfigureMain>
      <FormLayout withBorders={false} containerPadding={false}>
        <Box className={"inputItem"}>
          <H3Section>Configure</H3Section>
          <span className={"muted"}>
            Basic configurations for tenant management
          </span>
        </Box>
        <Box className={"inputItem"}>
          <h4 style={{ margin: "10px 0px 0px" }}>Services</h4>
          <span className={"muted"}>
            Whether the tenant's services should request an external IP via
            LoadBalancer service type.
          </span>
        </Box>
        <Switch
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
        <Switch
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
        <Switch
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
        <Switch
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
        {setDomains && (
          <Grid item xs={12} className={"inputItem"}>
            <fieldset>
              <legend>Custom Domains for MinIO</legend>
              <Grid item xs={12} className={"configSectionItem"}>
                <Box className={"inputItem"}>
                  <InputBox
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
                </Box>
                <Box>
                  <h4>MinIO Domains</h4>
                  <Box className={"responsiveSectionItem"}>
                    {minioDomains.map((domain, index) => {
                      return (
                        <Box
                          className={`containerItem wrapperContainer`}
                          key={`minio-domain-key-${index.toString()}`}
                        >
                          <InputBox
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
                          <Box className={"overlayAction"}>
                            <IconButton
                              size={"small"}
                              onClick={() => dispatch(addNewMinIODomain())}
                              disabled={index !== minioDomains.length - 1}
                            >
                              <AddIcon />
                            </IconButton>
                          </Box>

                          <Box className={"overlayAction"}>
                            <IconButton
                              size={"small"}
                              onClick={() => dispatch(removeMinIODomain(index))}
                              disabled={minioDomains.length <= 1}
                            >
                              <RemoveIcon />
                            </IconButton>
                          </Box>
                        </Box>
                      );
                    })}
                  </Box>
                </Box>
              </Grid>
            </fieldset>
          </Grid>
        )}

        <Switch
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
        {tenantCustom && (
          <Grid item xs={12} className={"inputItem"}>
            <fieldset>
              <legend>Security Context for MinIO</legend>
              <Grid item xs={12} className={`configSectionItem`}>
                <Box className={`responsiveSectionItem doubleElement`}>
                  <Box className={"containerItem"}>
                    <InputBox
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
                        validationErrors["tenant_securityContext_runAsUser"] ||
                        ""
                      }
                      min="0"
                    />
                  </Box>
                  <Box className={"containerItem"}>
                    <InputBox
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
                  </Box>
                </Box>
              </Grid>
              <br />
              <Grid item xs={12} className={`configSectionItem`}>
                <Box className={`responsiveSectionItem doubleElement`}>
                  <Box className={"containerItem"}>
                    <InputBox
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
                  </Box>
                  <Box className={"containerItem"}>
                    <Select
                      label="FsGroupChangePolicy"
                      id="securityContext_fsGroupChangePolicy"
                      name="securityContext_fsGroupChangePolicy"
                      value={tenantSecurityContext.fsGroupChangePolicy!}
                      onChange={(value) => {
                        updateField("tenantSecurityContext", {
                          ...tenantSecurityContext,
                          fsGroupChangePolicy: value,
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
                  </Box>
                </Box>
              </Grid>
              <br />
              <Grid item xs={12} className={"configSectionItem"}>
                <Switch
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
              </Grid>
            </fieldset>
          </Grid>
        )}
        <Switch
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
        {customRuntime && (
          <Grid item xs={12} className={"inputItem"}>
            <fieldset>
              <legend>Custom Runtime Configurations</legend>
              <InputBox
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
            </fieldset>
          </Grid>
        )}
        <hr />

        <Box className={"inputItem"}>
          <H3Section>Additional Environment Variables</H3Section>
          <span className={"muted"}>
            Define additional environment variables to be used by your MinIO
            pods
          </span>
        </Box>
        <Grid container>
          {tenantEnvVars.map((envVar, index) => (
            <Grid
              item
              xs={12}
              className={`formFieldRow envVarRow`}
              key={`tenant-envVar-${index.toString()}`}
            >
              <Grid item xs={5} className={"fileItem"}>
                <InputBox
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
              <Grid item xs={5} className={"fileItem"}>
                <InputBox
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
              <Grid item xs={2} className={"rowActions"}>
                <Box className={"overlayAction"}>
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
                </Box>
                <Box className={"overlayAction"}>
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
                </Box>
              </Grid>
            </Grid>
          ))}
        </Grid>
      </FormLayout>
    </ConfigureMain>
  );
};

export default Configure;
