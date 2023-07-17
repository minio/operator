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
import { useSelector } from "react-redux";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import { Paper, SelectChangeEvent } from "@mui/material";
import Grid from "@mui/material/Grid";

import {
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  wizardCommon,
} from "../../../Common/FormComponents/common/styleLibrary";
import { AppState, useAppDispatch } from "../../../../../store";
import { clearValidationError } from "../../utils";
import InputBoxWrapper from "../../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import FormSwitchWrapper from "../../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import FileSelector from "../../../Common/FormComponents/FileSelector/FileSelector";
import RadioGroupSelector from "../../../Common/FormComponents/RadioGroupSelector/RadioGroupSelector";
import {
  commonFormValidation,
  IValidation,
} from "../../../../../utils/validationFunctions";
import SectionH1 from "../../../Common/SectionH1";
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
import SelectWrapper from "../../../Common/FormComponents/SelectWrapper/SelectWrapper";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import CodeMirrorWrapper from "../../../Common/FormComponents/CodeMirrorWrapper/CodeMirrorWrapper";
import FormHr from "../../../Common/FormHr";

interface IEncryptionProps {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    encryptionTypeOptions: {
      marginBottom: 15,
    },
    mutualTlsConfig: {
      marginTop: 15,
      "& fieldset": {
        flex: 1,
      },
    },
    rightSpacer: {
      marginRight: 15,
    },
    responsiveContainer: {
      "@media (max-width: 900px)": {
        display: "flex",
        flexFlow: "column",
      },
    },
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  });

const Encryption = ({ classes }: IEncryptionProps) => {
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
          required: encryptionTab > 0,
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
    <Paper className={classes.paperWrapper}>
      <Grid container alignItems={"center"}>
        <Grid item xs>
          <SectionH1>Encryption</SectionH1>
        </Grid>
        <Grid item xs={4} justifyContent={"end"} textAlign={"right"}>
          <FormSwitchWrapper
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
        </Grid>
      </Grid>
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <span className={classes.descriptionText}>
            MinIO Server-Side Encryption (SSE) protects objects as part of write
            operations, allowing clients to take advantage of server processing
            power to secure objects at the storage layer (encryption-at-rest).
            SSE also provides key functionality to regulatory and compliance
            requirements around secure locking and erasure.
          </span>
        </Grid>
        <Grid xs={12}>
          <FormHr />
        </Grid>

        {enableEncryption && (
          <Fragment>
            <Grid item xs={12}>
              <Tabs
                value={encryptionTab}
                onChange={(e: React.ChangeEvent<{}>, value: number) => {
                  updateField("encryptionTab", value);
                }}
                indicatorColor="primary"
                textColor="primary"
                aria-label="cluster-tabs"
                variant="scrollable"
                scrollButtons="auto"
              >
                <Tab id="kms-options" label="Options" />
                <Tab id="kms-raw-configuration" label="Raw Edit" />
              </Tabs>
            </Grid>

            {encryptionTab ? (
              <Fragment>
                <Grid item xs={12}>
                  <CodeMirrorWrapper
                    value={rawConfiguration}
                    mode={"yaml"}
                    onBeforeChange={(editor, data, value) => {
                      updateField("rawConfiguration", value);
                    }}
                    editorHeight={"550px"}
                  />
                </Grid>
              </Fragment>
            ) : (
              <Fragment>
                <Grid item xs={12} className={classes.encryptionTypeOptions}>
                  <RadioGroupSelector
                    currentSelection={encryptionType}
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
                </Grid>
                {encryptionType === "vault" && <VaultKMSAdd />}
                {encryptionType === "azure" && <AzureKMSAdd />}
                {encryptionType === "gcp" && <GCPKMSAdd />}
                {encryptionType === "aws" && <AWSKMSAdd />}
                {encryptionType === "gemalto" && <GemaltoKMSAdd />}
              </Fragment>
            )}

            <div className={classes.headerElement}>
              <h4 className={classes.h3Section}>Additional Configurations</h4>
            </div>
            <Grid item xs={12}>
              <FormSwitchWrapper
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
            </Grid>
            {(enableCustomCertsForKES || !enableAutoCert) && (
              <Fragment>
                <Grid container>
                  <Grid item xs={12} style={{ marginBottom: 15 }}>
                    <fieldset className={classes.fieldGroup}>
                      <legend className={classes.descriptionText}>
                        Encryption server certificates
                      </legend>
                      <FileSelector
                        onChange={(encodedValue, fileName) => {
                          dispatch(
                            addFileKESServerCert({
                              key: "key",
                              fileName: fileName,
                              value: encodedValue,
                            }),
                          );
                          cleanValidation("serverKey");
                        }}
                        accept=".key,.pem"
                        id="serverKey"
                        name="serverKey"
                        label="Key"
                        error={validationErrors["serverKey"] || ""}
                        value={kesServerCertificate.key}
                        required={!enableAutoCert}
                      />
                      <FileSelector
                        onChange={(encodedValue, fileName) => {
                          dispatch(
                            addFileKESServerCert({
                              key: "cert",
                              fileName: fileName,
                              value: encodedValue,
                            }),
                          );
                          cleanValidation("serverCert");
                        }}
                        accept=".cer,.crt,.cert,.pem"
                        id="serverCert"
                        name="serverCert"
                        label="Cert"
                        error={validationErrors["serverCert"] || ""}
                        value={kesServerCertificate.cert}
                        required={!enableAutoCert}
                      />
                    </fieldset>
                  </Grid>
                </Grid>
                <Grid container style={{ marginBottom: 15 }}>
                  <Grid item xs={12}>
                    <fieldset className={classes.fieldGroup}>
                      <legend className={classes.descriptionText}>
                        MinIO mTLS certificates (connection between MinIO and
                        the Encryption server)
                      </legend>
                      <FileSelector
                        onChange={(encodedValue, fileName) => {
                          dispatch(
                            addFileMinIOMTLSCert({
                              key: "key",
                              fileName: fileName,
                              value: encodedValue,
                            }),
                          );
                          cleanValidation("clientKey");
                        }}
                        accept=".key,.pem"
                        id="clientKey"
                        name="clientKey"
                        label="Key"
                        error={validationErrors["clientKey"] || ""}
                        value={minioMTLSCertificate.key}
                        required={!enableAutoCert}
                      />
                      <FileSelector
                        onChange={(encodedValue, fileName) => {
                          dispatch(
                            addFileMinIOMTLSCert({
                              key: "cert",
                              fileName: fileName,
                              value: encodedValue,
                            }),
                          );
                          cleanValidation("clientCert");
                        }}
                        accept=".cer,.crt,.cert,.pem"
                        id="clientCert"
                        name="clientCert"
                        label="Cert"
                        error={validationErrors["clientCert"] || ""}
                        value={minioMTLSCertificate.cert}
                        required={!enableAutoCert}
                      />
                    </fieldset>
                  </Grid>
                </Grid>
                <Grid container className={classes.mutualTlsConfig}>
                  <fieldset className={classes.fieldGroup}>
                    <legend className={classes.descriptionText}>
                      KMS mTLS certificates (connection between the Encryption
                      server and the KMS)
                    </legend>
                    <FileSelector
                      onChange={(encodedValue, fileName) => {
                        dispatch(
                          addFileKMSMTLSCert({
                            key: "key",
                            fileName: fileName,
                            value: encodedValue,
                          }),
                        );
                        cleanValidation("vault_key");
                      }}
                      accept=".key,.pem"
                      id="vault_key"
                      name="vault_key"
                      label="Key"
                      value={kmsMTLSCertificate.key}
                    />
                    <FileSelector
                      onChange={(encodedValue, fileName) => {
                        dispatch(
                          addFileKMSMTLSCert({
                            key: "cert",
                            fileName: fileName,
                            value: encodedValue,
                          }),
                        );
                        cleanValidation("vault_cert");
                      }}
                      accept=".cer,.crt,.cert,.pem"
                      id="vault_cert"
                      name="vault_cert"
                      label="Cert"
                      value={kmsMTLSCertificate.cert}
                    />
                    <FileSelector
                      onChange={(encodedValue, fileName) => {
                        dispatch(
                          addFileKMSCa({
                            fileName: fileName,
                            value: encodedValue,
                          }),
                        );
                        cleanValidation("vault_ca");
                      }}
                      accept=".cer,.crt,.cert,.pem"
                      id="vault_ca"
                      name="vault_ca"
                      label="CA"
                      value={kmsCA.cert}
                    />
                  </fieldset>
                </Grid>
              </Fragment>
            )}
            <Grid item xs={12}>
              <Grid item xs={12} classes={classes.formFieldRow}>
                <InputBoxWrapper
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
                />
              </Grid>

              <fieldset
                className={classes.fieldGroup}
                style={{ marginTop: 15 }}
              >
                <legend className={classes.descriptionText}>
                  SecurityContext for KES pods
                </legend>
                <Grid item xs={12} className={classes.kesSecurityContext}>
                  <div
                    className={`${classes.multiContainer} ${classes.responsiveContainer}`}
                  >
                    <div
                      className={`${classes.formFieldRow} ${classes.rightSpacer}`}
                    >
                      <InputBoxWrapper
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
                          validationErrors["kes_securityContext_runAsUser"] ||
                          ""
                        }
                        min="0"
                      />
                    </div>
                    <div
                      className={`${classes.formFieldRow} ${classes.rightSpacer}`}
                    >
                      <InputBoxWrapper
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
                          validationErrors["kes_securityContext_runAsGroup"] ||
                          ""
                        }
                        min="0"
                      />
                    </div>
                  </div>
                </Grid>
                <br />
                <Grid item xs={12} className={classes.kesSecurityContext}>
                  <div
                    className={`${classes.multiContainer} ${classes.responsiveContainer}`}
                  >
                    <div
                      className={`${classes.formFieldRow} ${classes.rightSpacer}`}
                    >
                      <InputBoxWrapper
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
                    <div
                      className={`${classes.formFieldRow} ${classes.rightSpacer}`}
                    >
                      <SelectWrapper
                        label="FsGroupChangePolicy"
                        id="securityContext_fsGroupChangePolicy"
                        name="securityContext_fsGroupChangePolicy"
                        value={kesSecurityContext.fsGroupChangePolicy!}
                        onChange={(e: SelectChangeEvent<string>) => {
                          updateField("kesSecurityContext", {
                            ...kesSecurityContext,
                            fsGroupChangePolicy: e.target.value,
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
                <Grid item xs={12}>
                  <div className={classes.multiContainer}>
                    <FormSwitchWrapper
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
                  </div>
                </Grid>
              </fieldset>
            </Grid>
          </Fragment>
        )}
      </Grid>
    </Paper>
  );
};

export default withStyles(styles)(Encryption);
