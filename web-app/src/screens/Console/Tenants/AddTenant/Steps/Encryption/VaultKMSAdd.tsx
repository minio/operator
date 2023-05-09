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
import Grid from "@mui/material/Grid";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";

import { isPageValid, updateAddField } from "../../createTenantSlice";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../../store";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import makeStyles from "@mui/styles/makeStyles";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";
import { clearValidationError } from "../../../utils";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  })
);

const VaultKMSAdd = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const encryptionTab = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionTab
  );
  const vaultEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultEndpoint
  );
  const vaultEngine = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultEngine
  );
  const vaultNamespace = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultNamespace
  );
  const vaultPrefix = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultPrefix
  );
  const vaultAppRoleEngine = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultAppRoleEngine
  );
  const vaultId = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultId
  );
  const vaultSecret = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultSecret
  );
  const vaultRetry = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultRetry
  );
  const vaultPing = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.vaultPing
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Validation
  useEffect(() => {
    let encryptionValidation: IValidation[] = [];

    if (!encryptionTab) {
      encryptionValidation = [
        ...encryptionValidation,
        {
          fieldKey: "vault_endpoint",
          required: true,
          value: vaultEndpoint,
        },
        {
          fieldKey: "vault_id",
          required: true,
          value: vaultId,
        },
        {
          fieldKey: "vault_secret",
          required: true,
          value: vaultSecret,
        },
        {
          fieldKey: "vault_ping",
          required: false,
          value: vaultPing,
          customValidation: parseInt(vaultPing) < 0,
          customValidationMessage: "Value needs to be 0 or greater",
        },
        {
          fieldKey: "vault_retry",
          required: false,
          value: vaultRetry,
          customValidation: parseInt(vaultRetry) < 0,
          customValidationMessage: "Value needs to be 0 or greater",
        },
      ];
    }

    const commonVal = commonFormValidation(encryptionValidation);

    dispatch(
      isPageValid({
        pageName: "encryption",
        valid: Object.keys(commonVal).length === 0,
      })
    );

    setValidationErrors(commonVal);
  }, [
    encryptionTab,
    vaultEndpoint,
    vaultEngine,
    vaultId,
    vaultSecret,
    vaultPing,
    vaultRetry,
    dispatch,
  ]);

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "encryption", field: field, value: value })
      );
    },
    [dispatch]
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  return (
    <Fragment>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="vault_endpoint"
          name="vault_endpoint"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("vaultEndpoint", e.target.value);
            cleanValidation("vault_endpoint");
          }}
          label="Endpoint"
          tooltip="Endpoint is the Hashicorp Vault endpoint"
          value={vaultEndpoint}
          error={validationErrors["vault_endpoint"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="vault_engine"
          name="vault_engine"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("vaultEngine", e.target.value);
            cleanValidation("vault_engine");
          }}
          label="Engine"
          tooltip="Engine is the Hashicorp Vault K/V engine path. If empty, defaults to 'kv'"
          value={vaultEngine}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="vault_namespace"
          name="vault_namespace"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("vaultNamespace", e.target.value);
          }}
          label="Namespace"
          tooltip="Namespace is an optional Hashicorp Vault namespace. An empty namespace means no particular namespace is used."
          value={vaultNamespace}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="vault_prefix"
          name="vault_prefix"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("vaultPrefix", e.target.value);
          }}
          label="Prefix"
          tooltip="Prefix is an optional prefix / directory within the K/V engine. If empty, keys will be stored at the K/V engine top level"
          value={vaultPrefix}
        />
      </Grid>

      <Grid item xs={12}>
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>App Role</legend>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="vault_approle_engine"
              name="vault_approle_engine"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("vaultAppRoleEngine", e.target.value);
              }}
              label="Engine"
              tooltip="AppRoleEngine is the AppRole authentication engine path. If empty, defaults to 'approle'"
              value={vaultAppRoleEngine}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="vault_id"
              name="vault_id"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("vaultId", e.target.value);
                cleanValidation("vault_id");
              }}
              label="AppRole ID"
              tooltip="AppRoleSecret is the AppRole access secret for authenticating to Hashicorp Vault via the AppRole method"
              value={vaultId}
              error={validationErrors["vault_id"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="vault_secret"
              name="vault_secret"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("vaultSecret", e.target.value);
                cleanValidation("vault_secret");
              }}
              label="AppRole Secret"
              tooltip="AppRoleSecret is the AppRole access secret for authenticating to Hashicorp Vault via the AppRole method"
              value={vaultSecret}
              error={validationErrors["vault_secret"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              type="number"
              min="0"
              id="vault_retry"
              name="vault_retry"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("vaultRetry", e.target.value);
                cleanValidation("vault_retry");
              }}
              label="Retry (Seconds)"
              value={vaultRetry}
              error={validationErrors["vault_retry"] || ""}
            />
          </Grid>
        </fieldset>
      </Grid>
      <Grid
        item
        xs={12}
        className={classes.formFieldRow}
        style={{ marginTop: 15 }}
      >
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>Status</legend>
          <InputBoxWrapper
            type="number"
            min="0"
            id="vault_ping"
            name="vault_ping"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("vaultPing", e.target.value);
              cleanValidation("vault_ping");
            }}
            label="Ping (Seconds)"
            value={vaultPing}
            error={validationErrors["vault_ping"] || ""}
          />
        </fieldset>
      </Grid>
    </Fragment>
  );
};

export default VaultKMSAdd;
