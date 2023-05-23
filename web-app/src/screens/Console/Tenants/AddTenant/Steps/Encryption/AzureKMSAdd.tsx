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
import { isPageValid, updateAddField } from "../../createTenantSlice";
import { clearValidationError } from "../../../utils";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  })
);

const AzureKMSAdd = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const encryptionTab = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionTab
  );
  const azureEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.azureEndpoint
  );
  const azureTenantID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.azureTenantID
  );
  const azureClientID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.azureClientID
  );
  const azureClientSecret = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.azureClientSecret
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Validation
  useEffect(() => {
    let encryptionValidation: IValidation[] = [];

    if (!encryptionTab) {
      encryptionValidation = [
        ...encryptionValidation,
        {
          fieldKey: "azure_endpoint",
          required: true,
          value: azureEndpoint,
        },
        {
          fieldKey: "azure_tenant_id",
          required: true,
          value: azureTenantID,
        },
        {
          fieldKey: "azure_client_id",
          required: true,
          value: azureClientID,
        },
        {
          fieldKey: "azure_client_secret",
          required: true,
          value: azureClientSecret,
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
    azureEndpoint,
    azureTenantID,
    azureClientID,
    azureClientSecret,
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
          id="azure_endpoint"
          name="azure_endpoint"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("azureEndpoint", e.target.value);
            cleanValidation("azure_endpoint");
          }}
          label="Endpoint"
          tooltip="Endpoint is the Azure KeyVault endpoint"
          value={azureEndpoint}
          error={validationErrors["azure_endpoint"] || ""}
        />
      </Grid>
      <Grid item xs={12}>
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>Credentials</legend>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="azure_tenant_id"
              name="azure_tenant_id"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("azureTenantID", e.target.value);
                cleanValidation("azure_tenant_id");
              }}
              label="Tenant ID"
              tooltip="TenantID is the ID of the Azure KeyVault tenant"
              value={azureTenantID}
              error={validationErrors["azure_tenant_id"] || ""}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="azure_client_id"
              name="azure_client_id"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("azureClientID", e.target.value);
                cleanValidation("azure_client_id");
              }}
              label="Client ID"
              tooltip="ClientID is the ID of the client accessing Azure KeyVault"
              value={azureClientID}
              error={validationErrors["azure_client_id"] || ""}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="azure_client_secret"
              name="azure_client_secret"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("azureClientSecret", e.target.value);
                cleanValidation("azure_client_secret");
              }}
              label="Client Secret"
              tooltip="ClientSecret is the client secret accessing the Azure KeyVault"
              value={azureClientSecret}
              error={validationErrors["azure_client_secret"] || ""}
            />
          </Grid>
        </fieldset>
      </Grid>
    </Fragment>
  );
};

export default AzureKMSAdd;
