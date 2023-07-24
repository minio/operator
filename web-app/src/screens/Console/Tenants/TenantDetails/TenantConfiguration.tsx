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

import React, { useCallback, useEffect, useState } from "react";
import { connect, useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import FormSwitchWrapper from "../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import { DialogContentText, IconButton } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { Button, ConfirmModalIcon, Loader, RemoveIcon } from "mds";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import Grid from "@mui/material/Grid";
import {
  ITenantConfigurationRequest,
  ITenantConfigurationResponse,
  LabelKeyPair,
} from "../types";
import {
  containerForHeader,
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  spacingUtils,
  tenantDetailsStyles,
  wizardCommon,
} from "../../Common/FormComponents/common/styleLibrary";
import InputBoxWrapper from "../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import api from "../../../../common/api";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import { MinIOEnvVarsSettings } from "../../../../common/utils";
import FormHr from "../../Common/FormHr";

interface ITenantConfiguration {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...spacingUtils,
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
        width: 15,
        height: 15,
        maxWidth: 15,
        maxHeight: 15,
      },
      "& button": {
        background: "#EAEAEA",
      },
    },
    loaderAlign: {
      textAlign: "center",
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
    ...containerForHeader,
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  });

const TenantConfiguration = ({ classes }: ITenantConfiguration) => {
  const dispatch = useAppDispatch();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [isSending, setIsSending] = useState<boolean>(false);
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [envVars, setEnvVars] = useState<LabelKeyPair[]>([]);
  const [envVarsToBeDeleted, setEnvVarsToBeDeleted] = useState<string[]>([]);
  const [sftpExposed, setSftpEnabled] = useState<boolean>(false);

  const getTenantConfigurationInfo = useCallback(() => {
    api
      .invoke(
        "GET",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/configuration`,
      )
      .then((res: ITenantConfigurationResponse) => {
        if (res.environmentVariables) {
          setEnvVars(res.environmentVariables);
          setSftpEnabled(res.sftpExposed);
        }
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  }, [tenant, dispatch]);

  useEffect(() => {
    if (tenant) {
      getTenantConfigurationInfo();
    }
  }, [tenant, getTenantConfigurationInfo]);

  const updateTenantConfiguration = () => {
    setIsSending(true);
    let payload: ITenantConfigurationRequest = {
      environmentVariables: envVars.filter((env) => env.key !== ""),
      keysToBeDeleted: envVarsToBeDeleted,
      sftpExposed: sftpExposed,
    };
    api
      .invoke(
        "PATCH",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/configuration`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        setDialogOpen(false);
        getTenantConfigurationInfo();
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        setIsSending(false);
      });
  };

  return (
    <React.Fragment>
      <ConfirmDialog
        title={"Save and Restart"}
        confirmText={"Restart"}
        cancelText="Cancel"
        titleIcon={<ConfirmModalIcon />}
        isLoading={isSending}
        onClose={() => setDialogOpen(false)}
        isOpen={dialogOpen}
        onConfirm={updateTenantConfiguration}
        confirmationContent={
          <DialogContentText>
            Are you sure you want to save the changes and restart the service?
          </DialogContentText>
        }
      />
      {loadingTenant ? (
        <div className={classes.loaderAlign}>
          <Loader />
        </div>
      ) : (
        <Grid container spacing={1}>
          <Grid item xs={12}>
            <h1 className={classes.sectionTitle}>Configuration</h1>
            <FormHr />
          </Grid>
          <Grid container spacing={1}>
            {envVars.map((envVar, index) => (
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
                      const existingEnvVars = [...envVars];

                      setEnvVars(
                        existingEnvVars.map((keyPair, i) =>
                          i === index
                            ? { key: e.target.value, value: keyPair.value }
                            : keyPair,
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
                      const existingEnvVars = [...envVars];
                      setEnvVars(
                        existingEnvVars.map((keyPair, i) =>
                          i === index
                            ? { key: keyPair.key, value: e.target.value }
                            : keyPair,
                        ),
                      );
                    }}
                    index={index}
                    key={`env_var_value_${index.toString()}`}
                    type={
                      MinIOEnvVarsSettings[envVar.key] &&
                      MinIOEnvVarsSettings[envVar.key].secret
                        ? "password"
                        : "text"
                    }
                  />
                </Grid>
                <Grid item xs={2} className={classes.rowActions}>
                  <div className={classes.overlayAction}>
                    <IconButton
                      size={"small"}
                      onClick={() => {
                        const existingEnvVars = [...envVars];
                        existingEnvVars.push({ key: "", value: "" });

                        setEnvVars(existingEnvVars);
                      }}
                      disabled={index !== envVars.length - 1}
                    >
                      <AddIcon />
                    </IconButton>
                  </div>
                  <div className={classes.overlayAction}>
                    <IconButton
                      size={"small"}
                      onClick={() => {
                        const existingEnvVars = envVars.filter(
                          (item, fIndex) => fIndex !== index,
                        );
                        setEnvVars(existingEnvVars);
                        setEnvVarsToBeDeleted([
                          ...envVarsToBeDeleted,
                          envVar.key,
                        ]);
                      }}
                      disabled={envVars.length <= 1}
                    >
                      <RemoveIcon />
                    </IconButton>
                  </div>
                </Grid>
              </Grid>
            ))}
          </Grid>
          <Grid container spacing={1}>
            <Grid
              item
              xs={12}
              justifyContent={"end"}
              textAlign={"right"}
              className={classes.configSectionItem}
            >
              <FormSwitchWrapper
                label={"SFTP"}
                indicatorLabels={["Enabled", "Disabled"]}
                checked={sftpExposed}
                value={"expose_sftp"}
                id="expose-sftp"
                name="expose-sftp"
                onChange={() => {
                  setSftpEnabled(!sftpExposed);
                }}
                description=""
              />
            </Grid>
          </Grid>
          <Grid
            item
            xs={12}
            sx={{ display: "flex", justifyContent: "flex-end" }}
          >
            <Button
              id={"save-environment-variables"}
              type="submit"
              variant="callAction"
              disabled={dialogOpen || isSending}
              onClick={() => setDialogOpen(true)}
              label={"Save"}
            />
          </Grid>
        </Grid>
      )}
    </React.Fragment>
  );
};

const mapState = (state: AppState) => ({
  loadingTenant: state.tenants.loadingTenant,
  selectedTenant: state.tenants.currentTenant,
  tenant: state.tenants.tenantInfo,
});

const connector = connect(mapState, null);

export default withStyles(styles)(connector(TenantConfiguration));
