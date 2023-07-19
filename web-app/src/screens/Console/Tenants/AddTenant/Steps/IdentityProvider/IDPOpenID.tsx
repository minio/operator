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

import { Grid } from "@mui/material";
import InputBoxWrapper from "../../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import React, { Fragment, useCallback, useEffect, useState } from "react";
import makeStyles from "@mui/styles/makeStyles";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import {
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../../Common/FormComponents/common/styleLibrary";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../../store";
import { isPageValid, updateAddField } from "../../createTenantSlice";
import { clearValidationError } from "../../../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../../utils/validationFunctions";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    buttonTray: {
      marginLeft: 10,
      display: "flex",
      height: 38,
      "& button": {
        background: "#EAEAEA",
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
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  }),
);

const IDPOpenID = () => {
  const dispatch = useAppDispatch();
  const classes = useStyles();

  const idpSelection = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.idpSelection,
  );
  const openIDConfigurationURL = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDConfigurationURL,
  );
  const openIDClientID = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDClientID,
  );
  const openIDSecretID = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDSecretID,
  );
  const openIDCallbackURL = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDCallbackURL,
  );
  const openIDClaimName = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDClaimName,
  );
  const openIDScopes = useSelector(
    (state: AppState) =>
      state.createTenant.fields.identityProvider.openIDScopes,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({
          pageName: "identityProvider",
          field: field,
          value: value,
        }),
      );
    },
    [dispatch],
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  // Validation
  useEffect(() => {
    let customIDPValidation: IValidation[] = [];

    if (idpSelection === "OpenID") {
      customIDPValidation = [
        ...customIDPValidation,
        {
          fieldKey: "openID_CONFIGURATION_URL",
          required: true,
          value: openIDConfigurationURL,
        },
        {
          fieldKey: "openID_clientID",
          required: true,
          value: openIDClientID,
        },
        {
          fieldKey: "openID_secretID",
          required: true,
          value: openIDSecretID,
        },
        {
          fieldKey: "openID_claimName",
          required: true,
          value: openIDClaimName,
        },
      ];
    }

    const commonVal = commonFormValidation(customIDPValidation);

    dispatch(
      isPageValid({
        pageName: "identityProvider",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [
    idpSelection,
    openIDClientID,
    openIDSecretID,
    openIDConfigurationURL,
    openIDClaimName,
    dispatch,
  ]);

  return (
    <Fragment>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_CONFIGURATION_URL"
          name="openID_CONFIGURATION_URL"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDConfigurationURL", e.target.value);
            cleanValidation("openID_CONFIGURATION_URL");
          }}
          label="Configuration URL"
          value={openIDConfigurationURL}
          placeholder="https://your-identity-provider.com/.well-known/openid-configuration"
          error={validationErrors["openID_CONFIGURATION_URL"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_clientID"
          name="openID_clientID"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDClientID", e.target.value);
            cleanValidation("openID_clientID");
          }}
          label="Client ID"
          value={openIDClientID}
          error={validationErrors["openID_clientID"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_secretID"
          name="openID_secretID"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDSecretID", e.target.value);
            cleanValidation("openID_secretID");
          }}
          label="Secret ID"
          value={openIDSecretID}
          error={validationErrors["openID_secretID"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_callbackURL"
          name="openID_callbackURL"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDCallbackURL", e.target.value);
            cleanValidation("openID_callbackURL");
          }}
          label="Callback URL"
          value={openIDCallbackURL}
          placeholder="https://your-console-endpoint:9443/oauth_callback"
          error={validationErrors["openID_callbackURL"] || ""}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_claimName"
          name="openID_claimName"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDClaimName", e.target.value);
            cleanValidation("openID_claimName");
          }}
          label="Claim Name"
          value={openIDClaimName}
          error={validationErrors["openID_claimName"] || ""}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="openID_scopes"
          name="openID_scopes"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("openIDScopes", e.target.value);
            cleanValidation("openID_scopes");
          }}
          label="Scopes"
          value={openIDScopes}
        />
      </Grid>
    </Fragment>
  );
};

export default IDPOpenID;
