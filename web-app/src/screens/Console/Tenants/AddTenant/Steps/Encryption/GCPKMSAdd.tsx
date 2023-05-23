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

import React, { Fragment, useCallback } from "react";
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
import { updateAddField } from "../../createTenantSlice";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  })
);

const GCPKMSAdd = () => {
  const classes = useStyles();
  const dispatch = useAppDispatch();

  const gcpProjectID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpProjectID
  );
  const gcpEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpEndpoint
  );
  const gcpClientEmail = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpClientEmail
  );
  const gcpClientID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpClientID
  );
  const gcpPrivateKeyID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpPrivateKeyID
  );
  const gcpPrivateKey = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpPrivateKey
  );

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "encryption", field: field, value: value })
      );
    },
    [dispatch]
  );

  return (
    <Fragment>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="gcp_project_id"
          name="gcp_project_id"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpProjectID", e.target.value);
          }}
          label="Project ID"
          tooltip="ProjectID is the GCP project ID."
          value={gcpProjectID}
        />
      </Grid>
      <Grid item xs={12} className={classes.formFieldRow}>
        <InputBoxWrapper
          id="gcp_endpoint"
          name="gcp_endpoint"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpEndpoint", e.target.value);
          }}
          label="Endpoint"
          tooltip="Endpoint is the GCP project ID. If empty defaults to: secretmanager.googleapis.com:443"
          value={gcpEndpoint}
        />
      </Grid>
      <Grid item xs={12}>
        <fieldset className={classes.fieldGroup}>
          <legend className={classes.descriptionText}>Credentials</legend>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gcp_client_email"
              name="gcp_client_email"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gcpClientEmail", e.target.value);
              }}
              label="Client Email"
              tooltip="Is the Client email of the GCP service account used to access the SecretManager"
              value={gcpClientEmail}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gcp_client_id"
              name="gcp_client_id"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gcpClientID", e.target.value);
              }}
              label="Client ID"
              tooltip="Is the Client ID of the GCP service account used to access the SecretManager"
              value={gcpClientID}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gcp_private_key_id"
              name="gcp_private_key_id"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gcpPrivateKeyID", e.target.value);
              }}
              label="Private Key ID"
              tooltip="Is the private key ID of the GCP service account used to access the SecretManager"
              value={gcpPrivateKeyID}
            />
          </Grid>
          <Grid item xs={12} className={classes.formFieldRow}>
            <InputBoxWrapper
              id="gcp_private_key"
              name="gcp_private_key"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                updateField("gcpPrivateKey", e.target.value);
              }}
              label="Private Key"
              tooltip="Is the private key of the GCP service account used to access the SecretManager"
              value={gcpPrivateKey}
            />
          </Grid>
        </fieldset>
      </Grid>
    </Fragment>
  );
};

export default GCPKMSAdd;
