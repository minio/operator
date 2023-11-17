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

import React, { Fragment, useCallback, useEffect } from "react";
import {
  AddIcon,
  Box,
  breakPoints,
  FileSelector,
  FormLayout,
  Grid,
  IconButton,
  RemoveIcon,
  Switch,
} from "mds";
import { useSelector } from "react-redux";
import get from "lodash/get";
import styled from "styled-components";
import { AppState, useAppDispatch } from "../../../../../store";
import { KeyPair } from "../../ListTenants/utils";
import {
  addCaCertificate,
  addClientKeyPair,
  addFileToCaCertificates,
  addFileToClientKeyPair,
  addFileToKeyPair,
  addKeyPair,
  deleteCaCertificate,
  deleteClientKeyPair,
  deleteKeyPair,
  isPageValid,
  updateAddField,
} from "../createTenantSlice";
import TLSHelpBox from "../../HelpBox/TLSHelpBox";
import H3Section from "../../../Common/H3Section";

const CertificateRow = styled.div(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-start",
  padding: 8,
  borderBottom: `1px solid ${get(theme, "borderColor", "#E2E2E2")}`,
  "& .fileItem": {
    display: "flex",
    "& .inputItem:not(:last-of-type)": {
      marginBottom: 0,
    },
    [`@media (max-width: ${breakPoints.md}px)`]: {
      flexFlow: "column",
      "& .inputItem:not(:last-of-type)": {
        marginBottom: 10,
      },
    },
  },
  "& .rowActions": {
    display: "flex",
    justifyContent: "flex-end",
    alignItems: "center",
    gap: 10,
    "@media (max-width: 900px)": {
      flex: 1,
    },
  },
}));

const Security = () => {
  const dispatch = useAppDispatch();

  const enableTLS = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableTLS,
  );
  const enableAutoCert = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableAutoCert,
  );
  const enableCustomCerts = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableCustomCerts,
  );
  const minioCertificates = useSelector(
    (state: AppState) =>
      state.createTenant.certificates.minioServerCertificates,
  );
  const minioClientCertificates = useSelector(
    (state: AppState) =>
      state.createTenant.certificates.minioClientCertificates,
  );
  const caCertificates = useSelector(
    (state: AppState) => state.createTenant.certificates.minioCAsCertificates,
  );

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "security", field: field, value: value }),
      );
    },
    [dispatch],
  );

  // Validation

  useEffect(() => {
    if (!enableTLS) {
      dispatch(isPageValid({ pageName: "security", valid: true }));
      return;
    }
    if (enableAutoCert) {
      dispatch(isPageValid({ pageName: "security", valid: true }));
      return;
    }
    if (enableCustomCerts) {
      dispatch(isPageValid({ pageName: "security", valid: true }));
      return;
    }
    dispatch(isPageValid({ pageName: "security", valid: false }));
  }, [enableTLS, enableAutoCert, enableCustomCerts, dispatch]);

  return (
    <FormLayout withBorders={false} containerPadding={false}>
      <Box className={"inputItem"}>
        <H3Section>Security</H3Section>
      </Box>
      <Switch
        value="enableTLS"
        id="enableTLS"
        name="enableTLS"
        checked={enableTLS}
        onChange={(e) => {
          const targetD = e.target;
          const checked = targetD.checked;

          updateField("enableTLS", checked);
        }}
        label={"TLS"}
        description={
          "Securing all the traffic using TLS. This is required for Encryption Configuration"
        }
      />
      {enableTLS && (
        <Fragment>
          <Switch
            value="enableAutoCert"
            id="enableAutoCert"
            name="enableAutoCert"
            checked={enableAutoCert}
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;
              updateField("enableAutoCert", checked);
            }}
            label={"AutoCert"}
            description={
              "The internode certificates will be generated and managed by MinIO Operator"
            }
          />
          <Switch
            value="enableCustomCerts"
            id="enableCustomCerts"
            name="enableCustomCerts"
            checked={enableCustomCerts}
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;
              updateField("enableCustomCerts", checked);
            }}
            label={"Custom Certificates"}
            description={"Certificates used to terminated TLS at MinIO"}
          />
          {enableCustomCerts && (
            <Fragment>
              {!enableAutoCert && <TLSHelpBox />}
              <fieldset className="inputItem">
                <legend>MinIO Server Certificates</legend>

                {minioCertificates.map((keyPair: KeyPair, index) => (
                  <CertificateRow key={`minio-certs-${keyPair.id}`}>
                    <Grid item xs={10} className={"fileItem"}>
                      <FileSelector
                        onChange={(e, fileName, encodedValue) => {
                          if (encodedValue) {
                            dispatch(
                              addFileToKeyPair({
                                id: keyPair.id,
                                key: "cert",
                                fileName: fileName,
                                value: encodedValue,
                              }),
                            );
                          }
                        }}
                        accept=".cer,.crt,.cert,.pem"
                        id="tlsCert"
                        name="tlsCert"
                        label="Cert"
                        value={keyPair.cert}
                        returnEncodedData
                      />
                      <FileSelector
                        onChange={(event, fileName, encodedValue) => {
                          if (encodedValue) {
                            dispatch(
                              addFileToKeyPair({
                                id: keyPair.id,
                                key: "key",
                                fileName: fileName,
                                value: encodedValue,
                              }),
                            );
                          }
                        }}
                        accept=".key,.pem"
                        id="tlsKey"
                        name="tlsKey"
                        label="Key"
                        value={keyPair.key}
                        returnEncodedData
                      />
                    </Grid>

                    <Grid item xs={2} className={"rowActions"}>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(addKeyPair());
                        }}
                        disabled={index !== minioCertificates.length - 1}
                      >
                        <AddIcon />
                      </IconButton>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(deleteKeyPair(keyPair.id));
                        }}
                        disabled={minioCertificates.length <= 1}
                      >
                        <RemoveIcon />
                      </IconButton>
                    </Grid>
                  </CertificateRow>
                ))}
              </fieldset>
              <fieldset className="inputItem">
                <legend>MinIO Client Certificates</legend>
                {minioClientCertificates.map((keyPair: KeyPair, index) => (
                  <CertificateRow key={`minio-certs-${keyPair.id}`}>
                    <Grid item xs={10} className={"fileItem"}>
                      <FileSelector
                        onChange={(event, fileName, encodedValue) => {
                          if (encodedValue) {
                            dispatch(
                              addFileToClientKeyPair({
                                id: keyPair.id,
                                key: "cert",
                                fileName: fileName,
                                value: encodedValue,
                              }),
                            );
                          }
                        }}
                        accept=".cer,.crt,.cert,.pem"
                        id="tlsCert"
                        name="tlsCert"
                        label="Cert"
                        value={keyPair.cert}
                        returnEncodedData
                      />
                      <FileSelector
                        onChange={(event, fileName, encodedValue) => {
                          if (encodedValue) {
                            dispatch(
                              addFileToClientKeyPair({
                                id: keyPair.id,
                                key: "key",
                                fileName: fileName,
                                value: encodedValue,
                              }),
                            );
                          }
                        }}
                        accept=".key,.pem"
                        id="tlsKey"
                        name="tlsKey"
                        label="Key"
                        value={keyPair.key}
                        returnEncodedData
                      />
                    </Grid>

                    <Grid item xs={2} className={"rowActions"}>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(addClientKeyPair());
                        }}
                        disabled={index !== minioClientCertificates.length - 1}
                      >
                        <AddIcon />
                      </IconButton>
                      <IconButton
                        size={"small"}
                        onClick={() => {
                          dispatch(deleteClientKeyPair(keyPair.id));
                        }}
                        disabled={minioClientCertificates.length <= 1}
                      >
                        <RemoveIcon />
                      </IconButton>
                    </Grid>
                  </CertificateRow>
                ))}
              </fieldset>
              <fieldset className="inputItem">
                <legend>MinIO CA Certificates</legend>
                {caCertificates.map((keyPair: KeyPair, index) => (
                  <CertificateRow key={`minio-CA-certs-${keyPair.id}`}>
                    <Grid item xs={6} className={"fileItem"}>
                      <FileSelector
                        onChange={(event, fileName, encodedValue) => {
                          if (encodedValue) {
                            dispatch(
                              addFileToCaCertificates({
                                id: keyPair.id,
                                key: "cert",
                                fileName: fileName,
                                value: encodedValue,
                              }),
                            );
                          }
                        }}
                        accept=".cer,.crt,.cert,.pem"
                        id="tlsCert"
                        name="tlsCert"
                        label="Cert"
                        value={keyPair.cert}
                        returnEncodedData
                      />
                    </Grid>
                    <Grid item xs={6}>
                      <div className={"rowActions"}>
                        <IconButton
                          size={"small"}
                          onClick={() => {
                            dispatch(addCaCertificate());
                          }}
                          disabled={index !== caCertificates.length - 1}
                        >
                          <AddIcon />
                        </IconButton>
                        <IconButton
                          size={"small"}
                          onClick={() => {
                            dispatch(deleteCaCertificate(keyPair.id));
                          }}
                          disabled={caCertificates.length <= 1}
                        >
                          <RemoveIcon />
                        </IconButton>
                      </div>
                    </Grid>
                  </CertificateRow>
                ))}
              </fieldset>
            </Fragment>
          )}
        </Fragment>
      )}
    </FormLayout>
  );
};

export default Security;
