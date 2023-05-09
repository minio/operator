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
import { isPageValid, updateAddField } from "../../createTenantSlice";
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

const GemaltoKMSAdd = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const encryptionTab = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionTab
  );
  const gemaltoEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gemaltoEndpoint
  );
  const gemaltoToken = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gemaltoToken
  );
  const gemaltoDomain = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gemaltoDomain
  );
  const gemaltoRetry = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gemaltoRetry
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  // Validation
  useEffect(() => {
    let encryptionValidation: IValidation[] = [];

    if (!encryptionTab) {
      encryptionValidation = [
        ...encryptionValidation,
        {
          fieldKey: "gemalto_endpoint",
          required: true,
          value: gemaltoEndpoint,
        },
        {
          fieldKey: "gemalto_token",
          required: true,
          value: gemaltoToken,
        },
        {
          fieldKey: "gemalto_domain",
          required: true,
          value: gemaltoDomain,
        },
        {
          fieldKey: "gemalto_retry",
          required: false,
          value: gemaltoRetry,
          customValidation: parseInt(gemaltoRetry) < 0,
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
    gemaltoEndpoint,
    gemaltoToken,
    gemaltoDomain,
    gemaltoRetry,
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
          id="gemalto_endpoint"
          name="gemalto_endpoint"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gemaltoEndpoint", e.target.value);
            cleanValidation("gemalto_endpoint");
          }}
          label="Endpoint"
          tooltip="Endpoint is the endpoint to the KeySecure server"
          value={gemaltoEndpoint}
          error={validationErrors["gemalto_endpoint"] || ""}
          required
        />
      </Grid>
      <Grid
        item
        xs={12}
        style={{
          marginBottom: 15,
        }}
      >
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>Credentials</legend>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gemalto_token"
              name="gemalto_token"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gemaltoToken", e.target.value);
                cleanValidation("gemalto_token");
              }}
              label="Token"
              tooltip="Token is the refresh authentication token to access the KeySecure server"
              value={gemaltoToken}
              error={validationErrors["gemalto_token"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gemalto_domain"
              name="gemalto_domain"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gemaltoDomain", e.target.value);
                cleanValidation("gemalto_domain");
              }}
              label="Domain"
              tooltip="Domain is the isolated namespace within the KeySecure server. If empty, defaults to the top-level / root domain"
              value={gemaltoDomain}
              error={validationErrors["gemalto_domain"] || ""}
              required
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              type="number"
              min="0"
              id="gemalto_retry"
              name="gemalto_retry"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gemaltoRetry", e.target.value);
                cleanValidation("gemalto_retry");
              }}
              label="Retry (seconds)"
              value={gemaltoRetry}
              error={validationErrors["gemalto_retry"] || ""}
            />
          </Grid>
        </fieldset>
      </Grid>
    </Fragment>
  );
};

export default GemaltoKMSAdd;
