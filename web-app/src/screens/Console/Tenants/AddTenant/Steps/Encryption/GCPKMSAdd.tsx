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
import { InputBox } from "mds";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../../store";
import { updateAddField } from "../../createTenantSlice";

const GCPKMSAdd = () => {
  const dispatch = useAppDispatch();

  const gcpProjectID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpProjectID,
  );
  const gcpEndpoint = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpEndpoint,
  );
  const gcpClientEmail = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpClientEmail,
  );
  const gcpClientID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpClientID,
  );
  const gcpPrivateKeyID = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpPrivateKeyID,
  );
  const gcpPrivateKey = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.gcpPrivateKey,
  );

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "encryption", field: field, value: value }),
      );
    },
    [dispatch],
  );

  return (
    <Fragment>
      <InputBox
        id="gcp_project_id"
        name="gcp_project_id"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          updateField("gcpProjectID", e.target.value);
        }}
        label="Project ID"
        tooltip="ProjectID is the GCP project ID."
        value={gcpProjectID}
      />
      <InputBox
        id="gcp_endpoint"
        name="gcp_endpoint"
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          updateField("gcpEndpoint", e.target.value);
        }}
        label="Endpoint"
        tooltip="Endpoint is the GCP project ID. If empty defaults to: secretmanager.googleapis.com:443"
        value={gcpEndpoint}
      />
      <fieldset className={"inputItem"}>
        <legend>Credentials</legend>
        <InputBox
          id="gcp_client_email"
          name="gcp_client_email"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpClientEmail", e.target.value);
          }}
          label="Client Email"
          tooltip="Is the Client email of the GCP service account used to access the SecretManager"
          value={gcpClientEmail}
        />
        <InputBox
          id="gcp_client_id"
          name="gcp_client_id"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpClientID", e.target.value);
          }}
          label="Client ID"
          tooltip="Is the Client ID of the GCP service account used to access the SecretManager"
          value={gcpClientID}
        />
        <InputBox
          id="gcp_private_key_id"
          name="gcp_private_key_id"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpPrivateKeyID", e.target.value);
          }}
          label="Private Key ID"
          tooltip="Is the private key ID of the GCP service account used to access the SecretManager"
          value={gcpPrivateKeyID}
        />
        <InputBox
          id="gcp_private_key"
          name="gcp_private_key"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            updateField("gcpPrivateKey", e.target.value);
          }}
          label="Private Key"
          tooltip="Is the private key of the GCP service account used to access the SecretManager"
          value={gcpPrivateKey}
        />
      </fieldset>
    </Fragment>
  );
};

export default GCPKMSAdd;
