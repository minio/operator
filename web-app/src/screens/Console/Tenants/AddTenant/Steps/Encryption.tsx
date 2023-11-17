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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import {
  Box,
  CodeEditor,
  FileSelector,
  FormLayout,
  Grid,
  InputBox,
  RadioGroup,
  Select,
  SimpleHeader,
  Switch,
  Tabs,
} from "mds";
import { useSelector } from "react-redux";
import { AppState, useAppDispatch } from "../../../../../store";
import { clearValidationError } from "../../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import {
  addFileKESServerCert,
  addFileKMSCa,
  addFileKMSMTLSCert,
  addFileMinIOMTLSCert,
  isPageValid,
  updateAddField,
} from "../createTenantSlice";
import VaultKMSAdd from "./Encryption/VaultKMSAdd";
import AzureKMSAdd from "./Encryption/AzureKMSAdd";
import GCPKMSAdd from "./Encryption/GCPKMSAdd";
import GemaltoKMSAdd from "./Encryption/GemaltoKMSAdd";
import AWSKMSAdd from "./Encryption/AWSKMSAdd";
import H3Section from "../../../Common/H3Section";

const Encryption = () => {
  const dispatch = useAppDispatch();

  const replicas = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.replicas,
  );
  const rawConfiguration = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.rawConfiguration,
  );
  const encryptionTab = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionTab,
  );
  const enableEncryption = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.enableEncryption,
  );
  const encryptionType = useSelector(
    (state: AppState) => state.createTenant.fields.encryption.encryptionType,
  );

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
  const enableCustomCertsForKES = useSelector(
    (state: AppState) =>
      state.createTenant.fields.encryption.enableCustomCertsForKES,
  );
  const enableAutoCert = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableAutoCert,
  );
  const enableTLS = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableTLS,
  );
  const minioServerCertificates = useSelector(
    (state: AppState) =>
      state.createTenant.certificates.minioServerCertificates,
  );
  const kesServerCertificate = useSelector(
    (state: AppState) => state.createTenant.certificates.kesServerCertificate,
  );
  const minioMTLSCertificate = useSelector(
    (state: AppState) => state.createTenant.certificates.minioMTLSCertificate,
  );
  const kmsMTLSCertificate = useSelector(
    (state: AppState) => state.createTenant.certificates.kmsMTLSCertificate,
  );
  const kmsCA = useSelector(
    (state: AppState) => state.createTenant.certificates.kmsCA,
  );
  const enableCustomCerts = useSelector(
    (state: AppState) => state.createTenant.fields.security.enableCustomCerts,
  );
  const kesSecurityContext = useSelector(
    (state: AppState) =>
      state.createTenant.fields.encryption.kesSecurityContext,
  );

  const [validationErrors, setValidationErrors] = useState<any>({});

  let encryptionAvailable = false;
  if (
    enableTLS &&
    (enableAutoCert ||
      (minioServerCertificates &&
        minioServerCertificates.filter(
          (item) => item.encoded_key && item.encoded_cert,
        ).length > 0))
  ) {
    encryptionAvailable = true;
  }

  // Common
  const updateField = useCallback(
    (field: string, value: any) => {
      dispatch(
        updateAddField({ pageName: "encryption", field: field, value: value }),
      );
    },
    [dispatch],
  );

  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };

  // Validation
  useEffect(() => {
    let encryptionValidation: IValidation[] = [];

    if (enableEncryption) {
      encryptionValidation = [
        {
          fieldKey: "rawConfiguration",
          required: encryptionTab === "kms-raw-configuration",
          value: rawConfiguration,
        },
        {
          fieldKey: "replicas",
          required: true,
          value: replicas,
          customValidation: parseInt(replicas) < 1,
          customValidationMessage: "Replicas needs to be 1 or greater",
        },
        {
          fieldKey: "kes_securityContext_runAsUser",
          required: true,
          value: kesSecurityContext.runAsUser,
          customValidation:
            kesSecurityContext.runAsUser === "" ||
            parseInt(kesSecurityContext.runAsUser) < 0,
          customValidationMessage: `runAsUser must be present and be 0 or more`,
        },
        {
          fieldKey: "kes_securityContext_runAsGroup",
          required: true,
          value: kesSecurityContext.runAsGroup,
          customValidation:
            kesSecurityContext.runAsGroup === "" ||
            parseInt(kesSecurityContext.runAsGroup) < 0,
          customValidationMessage: `runAsGroup must be present and be 0 or more`,
        },
        {
          fieldKey: "kes_securityContext_fsGroup",
          required: true,
          value: kesSecurityContext.fsGroup!,
          customValidation:
            kesSecurityContext.fsGroup === "" ||
            parseInt(kesSecurityContext.fsGroup!) < 0,
          customValidationMessage: `fsGroup must be present and be 0 or more`,
        },
      ];

      if (enableCustomCerts) {
        encryptionValidation = [
          ...encryptionValidation,
          {
            fieldKey: "serverKey",
            required: !enableAutoCert,
            value: kesServerCertificate.encoded_key,
          },
          {
            fieldKey: "serverCert",
            required: !enableAutoCert,
            value: kesServerCertificate.encoded_cert,
          },
          {
            fieldKey: "clientKey",
            required: !enableAutoCert,
            value: minioMTLSCertificate.encoded_key,
          },
          {
            fieldKey: "clientCert",
            required: !enableAutoCert,
            value: minioMTLSCertificate.encoded_cert,
          },
        ];
      }
    }

    const commonVal = commonFormValidation(encryptionValidation);
    dispatch(
      isPageValid({
        pageName: "encryption",
        valid: Object.keys(commonVal).length === 0,
      }),
    );

    setValidationErrors(commonVal);
  }, [
    rawConfiguration,
    encryptionTab,
    enableEncryption,
    encryptionType,
    gcpProjectID,
    gcpEndpoint,
    gcpClientEmail,
    gcpClientID,
    gcpPrivateKeyID,
    gcpPrivateKey,
    dispatch,
    enableAutoCert,
    enableCustomCerts,
    kesServerCertificate.encoded_key,
    kesServerCertificate.encoded_cert,
    minioMTLSCertificate.encoded_key,
    minioMTLSCertificate.encoded_cert,
    kesSecurityContext,
    replicas,
  ]);

  return (
    <FormLayout
      withBorders={false}
      containerPadding={false}
      sx={{
        "& .tabs-container": { height: "inherit" },
        "& .rightSpacer": {
          marginRight: 15,
        },
        "& .responsiveContainer": {
          "@media (max-width: 900px)": {
            display: "flex",
            flexFlow: "column",
          },
        },
        "& .multiContainer": {
          display: "flex",
          alignItems: "center",
          justifyContent: "flex-start",
        },
      }}
    >
      <Box
        className={"inputItem"}
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <H3Section>Encryption</H3Section>
        <Switch
          label={""}
          indicatorLabels={["Enabled", "Disabled"]}
          checked={enableEncryption}
          value={"tenant_encryption"}
          id="tenant-encryption"
          name="tenant-encryption"
          onChange={(e) => {
            const targetD = e.target;
            const checked = targetD.checked;

            updateField("enableEncryption", checked);
          }}
          description=""
          disabled={!encryptionAvailable}
        />
      </Box>
      <Box className={"muted inputItem"}>
        MinIO Server-Side Encryption (SSE) protects objects as part of write
        operations, allowing clients to take advantage of server processing
        power to secure objects at the storage layer (encryption-at-rest). SSE
        also provides key functionality to regulatory and compliance
        requirements around secure locking and erasure.
      </Box>
      <hr />

      {enableEncryption && (
        <Fragment>
          <Tabs
            horizontal
            currentTabOrPath={encryptionTab}
            onTabClick={(value: string) => {
              updateField("encryptionTab", value);
            }}
            sx={{
              height: "initial",
            }}
            options={[
              {
                tabConfig: {
                  label: "Options",
                  id: "kms-options",
                },
                content: (
                  <Fragment>
                    <RadioGroup
                      currentValue={encryptionType}
                      id="encryptionType"
                      name="encryptionType"
                      label="KMS"
                      onChange={(e) => {
                        updateField("encryptionType", e.target.value);
                      }}
                      selectorOptions={[
                        { label: "Vault", value: "vault" },
                        { label: "AWS", value: "aws" },
                        { label: "Gemalto", value: "gemalto" },
                        { label: "GCP", value: "gcp" },
                        { label: "Azure", value: "azure" },
                      ]}
                    />
                    {encryptionType === "vault" && <VaultKMSAdd />}
                    {encryptionType === "azure" && <AzureKMSAdd />}
                    {encryptionType === "gcp" && <GCPKMSAdd />}
                    {encryptionType === "aws" && <AWSKMSAdd />}
                    {encryptionType === "gemalto" && <GemaltoKMSAdd />}
                  </Fragment>
                ),
              },
              {
                tabConfig: {
                  label: "Raw Edit",
                  id: "kms-raw-configuration",
                },
                content: (
                  <Fragment>
                    <Grid item xs={12}>
                      <CodeEditor
                        value={rawConfiguration}
                        mode={"yaml"}
                        onChange={(value) => {
                          updateField("rawConfiguration", value);
                        }}
                        editorHeight={"550px"}
                      />
                    </Grid>
                  </Fragment>
                ),
              },
            ]}
          />
          <SimpleHeader
            label={"Additional Configurations"}
            sx={{ margin: "0px 0px 10px" }}
          />
          <Switch
            value="enableCustomCertsForKES"
            id="enableCustomCertsForKES"
            name="enableCustomCertsForKES"
            checked={enableCustomCertsForKES || !enableAutoCert}
            onChange={(e) => {
              const targetD = e.target;
              const checked = targetD.checked;

              updateField("enableCustomCertsForKES", checked);
            }}
            label={"Custom Certificates"}
            disabled={!enableAutoCert}
          />
          {(enableCustomCertsForKES || !enableAutoCert) && (
            <Fragment>
              <fieldset className={"inputItem"}>
                <legend>Encryption server certificates</legend>
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileKESServerCert({
                          key: "key",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("serverKey");
                    }
                  }}
                  accept=".key,.pem"
                  id="serverKey"
                  name="serverKey"
                  label="Key"
                  error={validationErrors["serverKey"] || ""}
                  value={kesServerCertificate.key}
                  required={!enableAutoCert}
                  returnEncodedData
                />
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileKESServerCert({
                          key: "cert",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("serverCert");
                    }
                  }}
                  accept=".cer,.crt,.cert,.pem"
                  id="serverCert"
                  name="serverCert"
                  label="Cert"
                  error={validationErrors["serverCert"] || ""}
                  value={kesServerCertificate.cert}
                  required={!enableAutoCert}
                  returnEncodedData
                />
              </fieldset>
              <fieldset className={"inputItem"}>
                <legend>
                  MinIO mTLS certificates (connection between MinIO and the
                  Encryption server)
                </legend>
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileMinIOMTLSCert({
                          key: "key",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("clientKey");
                    }
                  }}
                  accept=".key,.pem"
                  id="clientKey"
                  name="clientKey"
                  label="Key"
                  error={validationErrors["clientKey"] || ""}
                  value={minioMTLSCertificate.key}
                  required={!enableAutoCert}
                  returnEncodedData
                />
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileMinIOMTLSCert({
                          key: "cert",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("clientCert");
                    }
                  }}
                  accept=".cer,.crt,.cert,.pem"
                  id="clientCert"
                  name="clientCert"
                  label="Cert"
                  error={validationErrors["clientCert"] || ""}
                  value={minioMTLSCertificate.cert}
                  required={!enableAutoCert}
                  returnEncodedData
                />
              </fieldset>
              <fieldset className={"inputItem"}>
                <legend>
                  KMS mTLS certificates (connection between the Encryption
                  server and the KMS)
                </legend>
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileKMSMTLSCert({
                          key: "key",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("vault_key");
                    }
                  }}
                  accept=".key,.pem"
                  id="vault_key"
                  name="vault_key"
                  label="Key"
                  value={kmsMTLSCertificate.key}
                  returnEncodedData
                />
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileKMSMTLSCert({
                          key: "cert",
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("vault_cert");
                    }
                  }}
                  accept=".cer,.crt,.cert,.pem"
                  id="vault_cert"
                  name="vault_cert"
                  label="Cert"
                  value={kmsMTLSCertificate.cert}
                  returnEncodedData
                />
                <FileSelector
                  onChange={(event, fileName, encodedValue) => {
                    if (encodedValue) {
                      dispatch(
                        addFileKMSCa({
                          fileName: fileName,
                          value: encodedValue,
                        }),
                      );
                      cleanValidation("vault_ca");
                    }
                  }}
                  accept=".cer,.crt,.cert,.pem"
                  id="vault_ca"
                  name="vault_ca"
                  label="CA"
                  value={kmsCA.cert}
                  returnEncodedData
                />
              </fieldset>
            </Fragment>
          )}
          <InputBox
            type="number"
            min="1"
            id="replicas"
            name="replicas"
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
              updateField("replicas", e.target.value);
              cleanValidation("replicas");
            }}
            label="Replicas"
            value={replicas}
            required
            error={validationErrors["replicas"] || ""}
            sx={{ marginBottom: 10 }}
          />

          <fieldset className={"inputItem"}>
            <legend>SecurityContext for KES pods</legend>
            <Grid item xs={12}>
              <div className={`multiContainer responsiveContainer`}>
                <div className={`rightSpacer`}>
                  <InputBox
                    type="number"
                    id="kes_securityContext_runAsUser"
                    name="kes_securityContext_runAsUser"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("kesSecurityContext", {
                        ...kesSecurityContext,
                        runAsUser: e.target.value,
                      });
                      cleanValidation("kes_securityContext_runAsUser");
                    }}
                    label="Run As User"
                    value={kesSecurityContext.runAsUser}
                    required
                    error={
                      validationErrors["kes_securityContext_runAsUser"] || ""
                    }
                    min="0"
                  />
                </div>
                <div className={`rightSpacer`}>
                  <InputBox
                    type="number"
                    id="kes_securityContext_runAsGroup"
                    name="kes_securityContext_runAsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("kesSecurityContext", {
                        ...kesSecurityContext,
                        runAsGroup: e.target.value,
                      });
                      cleanValidation("kes_securityContext_runAsGroup");
                    }}
                    label="Run As Group"
                    value={kesSecurityContext.runAsGroup}
                    required
                    error={
                      validationErrors["kes_securityContext_runAsGroup"] || ""
                    }
                    min="0"
                  />
                </div>
              </div>
            </Grid>
            <br />
            <Grid item xs={12}>
              <div className={`multiContainer responsiveContainer`}>
                <div className={`rightSpacer`}>
                  <InputBox
                    type="number"
                    id="kes_securityContext_fsGroup"
                    name="kes_securityContext_fsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      updateField("kesSecurityContext", {
                        ...kesSecurityContext,
                        fsGroup: e.target.value,
                      });
                      cleanValidation("kes_securityContext_fsGroup");
                    }}
                    label="FsGroup"
                    value={kesSecurityContext.fsGroup!}
                    required
                    error={
                      validationErrors["kes_securityContext_fsGroup"] || ""
                    }
                    min="0"
                  />
                </div>
                <div className={`rightSpacer`}>
                  <Select
                    label="FsGroupChangePolicy"
                    id="securityContext_fsGroupChangePolicy"
                    name="securityContext_fsGroupChangePolicy"
                    value={kesSecurityContext.fsGroupChangePolicy!}
                    onChange={(value) => {
                      updateField("kesSecurityContext", {
                        ...kesSecurityContext,
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
                </div>
              </div>
            </Grid>
            <br />
            <Switch
              value="kesSecurityContextRunAsNonRoot"
              id="kes_securityContext_runAsNonRoot"
              name="kes_securityContext_runAsNonRoot"
              checked={kesSecurityContext.runAsNonRoot}
              onChange={(e) => {
                const targetD = e.target;
                const checked = targetD.checked;
                updateField("kesSecurityContext", {
                  ...kesSecurityContext,
                  runAsNonRoot: checked,
                });
              }}
              label={"Do not run as Root"}
            />
          </fieldset>
        </Fragment>
      )}
    </FormLayout>
  );
};

export default Encryption;
