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
import {
  AddIcon,
  Box,
  Button,
  ConfirmModalIcon,
  FileSelector,
  FormLayout,
  Grid,
  IconButton,
  Loader,
  RemoveIcon,
  SectionTitle,
  Switch,
} from "mds";
import {
  fsGroupChangePolicyType,
  ICertificateInfo,
  ITenantSecurityResponse,
} from "../types";
import { modalStyleUtils } from "../../Common/FormComponents/common/styleLibrary";

import { KeyPair } from "../ListTenants/utils";
import { AppState, useAppDispatch } from "../../../../store";
import { ErrorResponseHandler } from "../../../../common/types";
import { setErrorSnackMessage } from "../../../../systemSlice";
import api from "../../../../common/api";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import TLSCertificate from "../../Common/TLSCertificate/TLSCertificate";
import SecurityContextSelector from "../securityContextSelector";
import {
  setFSGroup,
  setFSGroupChangePolicy,
  setRunAsGroup,
  setRunAsNonRoot,
  setRunAsUser,
} from "../tenantSecurityContextSlice";
import TLSHelpBox from "../HelpBox/TLSHelpBox";

const TenantSecurity = () => {
  const dispatch = useAppDispatch();

  const tenant = useSelector((state: AppState) => state.tenants.tenantInfo);
  const loadingTenant = useSelector(
    (state: AppState) => state.tenants.loadingTenant,
  );

  const [isSending, setIsSending] = useState<boolean>(false);
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [enableTLS, setEnableTLS] = useState<boolean>(false);
  const [enableAutoCert, setEnableAutoCert] = useState<boolean>(false);
  const [enableCustomCerts, setEnableCustomCerts] = useState<boolean>(false);
  const [certificatesToBeRemoved, setCertificatesToBeRemoved] = useState<
    string[]
  >([]);
  // MinIO certificates
  const [minioServerCertificates, setMinioServerCertificates] = useState<
    KeyPair[]
  >([
    {
      id: Date.now().toString(),
      key: "",
      cert: "",
      encoded_key: "",
      encoded_cert: "",
    },
  ]);
  const [minioClientCertificates, setMinioClientCertificates] = useState<
    KeyPair[]
  >([
    {
      id: Date.now().toString(),
      key: "",
      cert: "",
      encoded_key: "",
      encoded_cert: "",
    },
  ]);
  const [minioCaCertificates, setMinioCaCertificates] = useState<KeyPair[]>([
    {
      id: Date.now().toString(),
      key: "",
      cert: "",
      encoded_key: "",
      encoded_cert: "",
    },
  ]);
  const [minioServerCertificateSecrets, setMinioServerCertificateSecrets] =
    useState<ICertificateInfo[]>([]);
  const [minioClientCertificateSecrets, setMinioClientCertificateSecrets] =
    useState<ICertificateInfo[]>([]);
  const [minioTLSCaCertificateSecrets, setMinioTLSCaCertificateSecrets] =
    useState<ICertificateInfo[]>([]);

  const runAsGroup = useSelector(
    (state: AppState) => state.editTenantSecurityContext.runAsGroup,
  );
  const runAsUser = useSelector(
    (state: AppState) => state.editTenantSecurityContext.runAsUser,
  );
  const fsGroup = useSelector(
    (state: AppState) => state.editTenantSecurityContext.fsGroup,
  );
  const runAsNonRoot = useSelector(
    (state: AppState) => state.editTenantSecurityContext.runAsNonRoot,
  );
  const fsGroupChangePolicy = useSelector(
    (state: AppState) => state.editTenantSecurityContext.fsGroupChangePolicy,
  );

  const getTenantSecurityInfo = useCallback(() => {
    api
      .invoke(
        "GET",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/security`,
      )
      .then((res: ITenantSecurityResponse) => {
        setEnableAutoCert(res.autoCert);
        setEnableTLS(res.autoCert);
        if (
          res.customCertificates.minio ||
          res.customCertificates.client ||
          res.customCertificates.minioCAs
        ) {
          setEnableCustomCerts(true);
          setEnableTLS(true);
        }
        setMinioServerCertificateSecrets(res.customCertificates.minio || []);
        setMinioClientCertificateSecrets(res.customCertificates.client || []);
        setMinioTLSCaCertificateSecrets(res.customCertificates.minioCAs || []);
        dispatch(setRunAsGroup(res.securityContext.runAsGroup));
        dispatch(setRunAsUser(res.securityContext.runAsUser));
        dispatch(setFSGroup(res.securityContext.fsGroup!));
        dispatch(setRunAsNonRoot(res.securityContext.runAsNonRoot));
        dispatch(
          setFSGroupChangePolicy(
            res.securityContext.fsGroupChangePolicy as fsGroupChangePolicyType,
          ),
        );
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
      });
  }, [tenant, dispatch]);

  useEffect(() => {
    if (tenant) {
      getTenantSecurityInfo();
    }
  }, [tenant, getTenantSecurityInfo]);

  const updateTenantSecurity = () => {
    setIsSending(true);
    let payload = {
      autoCert: enableAutoCert,
      customCertificates: {},
      securityContext: {
        runAsGroup: runAsGroup,
        runAsUser: runAsUser,
        runAsNonRoot: runAsNonRoot,
        fsGroup: fsGroup,
        fsGroupChangePolicy: fsGroupChangePolicy,
      },
    };
    if (enableCustomCerts) {
      payload["customCertificates"] = {
        secretsToBeDeleted: certificatesToBeRemoved,
        minioServerCertificates: minioServerCertificates
          .map((keyPair: KeyPair) => ({
            crt: keyPair.encoded_cert,
            key: keyPair.encoded_key,
          }))
          .filter((cert: any) => cert.crt && cert.key),
        minioClientCertificates: minioClientCertificates
          .map((keyPair: KeyPair) => ({
            crt: keyPair.encoded_cert,
            key: keyPair.encoded_key,
          }))
          .filter((cert: any) => cert.crt && cert.key),
        minioCAsCertificates: minioCaCertificates
          .map((keyPair: KeyPair) => keyPair.encoded_cert)
          .filter((cert: any) => cert),
      };
    } else {
      payload["customCertificates"] = {
        secretsToBeDeleted: [
          ...minioServerCertificateSecrets.map((cert) => cert.name),
          ...minioClientCertificateSecrets.map((cert) => cert.name),
          ...minioTLSCaCertificateSecrets.map((cert) => cert.name),
        ],
        minioServerCertificates: [],
        minioClientCertificates: [],
        minioCAsCertificates: [],
      };
    }
    api
      .invoke(
        "POST",
        `/api/v1/namespaces/${tenant?.namespace}/tenants/${tenant?.name}/security`,
        payload,
      )
      .then(() => {
        setIsSending(false);
        // Close confirmation modal
        setDialogOpen(false);
        // Refresh Information and reset forms
        setMinioServerCertificates([
          {
            cert: "",
            encoded_cert: "",
            encoded_key: "",
            id: Date.now().toString(),
            key: "",
          },
        ]);
        setMinioClientCertificates([
          {
            cert: "",
            encoded_cert: "",
            encoded_key: "",
            id: Date.now().toString(),
            key: "",
          },
        ]);
        setMinioCaCertificates([
          {
            cert: "",
            encoded_cert: "",
            encoded_key: "",
            id: Date.now().toString(),
            key: "",
          },
        ]);
        getTenantSecurityInfo();
      })
      .catch((err: ErrorResponseHandler) => {
        dispatch(setErrorSnackMessage(err));
        setIsSending(false);
      });
  };

  const removeCertificate = (certificateInfo: ICertificateInfo) => {
    // TLS certificate secrets can be referenced MinIO, Console or KES, we need to remove the secret from all list and update
    // the arrays
    // Add certificate to the global list of secrets to be removed
    setCertificatesToBeRemoved([
      ...certificatesToBeRemoved,
      certificateInfo.name,
    ]);

    // Update MinIO server TLS certificate secrets
    const updatedMinioServerCertificateSecrets =
      minioServerCertificateSecrets.filter(
        (certificateSecret) => certificateSecret.name !== certificateInfo.name,
      );
    // Update MinIO client TLS certificate secrets
    const updatedMinioClientCertificateSecrets =
      minioClientCertificateSecrets.filter(
        (certificateSecret) => certificateSecret.name !== certificateInfo.name,
      );
    const updatedMinIOTLSCaCertificateSecrets =
      minioTLSCaCertificateSecrets.filter(
        (certificateSecret) => certificateSecret.name !== certificateInfo.name,
      );
    setMinioServerCertificateSecrets(updatedMinioServerCertificateSecrets);
    setMinioClientCertificateSecrets(updatedMinioClientCertificateSecrets);
    setMinioTLSCaCertificateSecrets(updatedMinIOTLSCaCertificateSecrets);
  };

  const addFileToKeyPair = (
    type: string,
    id: string,
    key: string,
    fileName: string,
    value: string,
  ) => {
    let certificates = minioServerCertificates;
    let updateCertificates: any = () => {};

    switch (type) {
      case "minio": {
        certificates = minioServerCertificates;
        updateCertificates = setMinioServerCertificates;
        break;
      }
      case "client": {
        certificates = minioClientCertificates;
        updateCertificates = setMinioClientCertificates;
        break;
      }
      case "minioCAs": {
        certificates = minioCaCertificates;
        updateCertificates = setMinioCaCertificates;
        break;
      }
      default:
    }

    const NCertList = certificates.map((item: KeyPair) => {
      if (item.id === id) {
        return {
          ...item,
          [key]: fileName,
          [`encoded_${key}`]: value,
        };
      }
      return item;
    });
    updateCertificates(NCertList);
  };

  const deleteKeyPair = (type: string, id: string) => {
    let certificates = minioServerCertificates;
    let updateCertificates: any = () => {};

    switch (type) {
      case "minio": {
        certificates = minioServerCertificates;
        updateCertificates = setMinioServerCertificates;
        break;
      }
      case "client": {
        certificates = minioClientCertificates;
        updateCertificates = setMinioClientCertificates;
        break;
      }
      case "minioCAs": {
        certificates = minioCaCertificates;
        updateCertificates = setMinioCaCertificates;
        break;
      }
      default:
    }

    if (certificates.length > 1) {
      const cleanCertsList = certificates.filter(
        (item: KeyPair) => item.id !== id,
      );
      updateCertificates(cleanCertsList);
    }
  };

  const addKeyPair = (type: string) => {
    let certificates = minioServerCertificates;
    let updateCertificates: any = () => {};

    switch (type) {
      case "minio": {
        certificates = minioServerCertificates;
        updateCertificates = setMinioServerCertificates;
        break;
      }
      case "client": {
        certificates = minioClientCertificates;
        updateCertificates = setMinioClientCertificates;
        break;
      }
      case "minioCAs": {
        certificates = minioCaCertificates;
        updateCertificates = setMinioCaCertificates;
        break;
      }
      default:
    }
    const updatedCertificates = [
      ...certificates,
      {
        id: Date.now().toString(),
        key: "",
        cert: "",
        encoded_key: "",
        encoded_cert: "",
      },
    ];
    updateCertificates(updatedCertificates);
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
        onConfirm={updateTenantSecurity}
        confirmationContent={
          <Fragment>
            Are you sure you want to save the changes and restart the service?
          </Fragment>
        }
      />
      {loadingTenant ? (
        <Box
          sx={{
            textAlign: "center",
          }}
        >
          <Loader />
        </Box>
      ) : (
        <Fragment>
          <Box>
            <SectionTitle separator sx={{ marginBottom: 15 }}>
              Security
            </SectionTitle>
          </Box>
          <FormLayout
            withBorders={false}
            containerPadding={false}
            sx={{
              "& .minioCertificateRows": {
                display: "flex",
                alignItems: "center",
                justifyContent: "flex-start",
                padding: 5,
                "& .inputItem": {
                  marginBottom: 0,
                },
                "&:last-child": {
                  borderBottom: 0,
                },
                "@media (max-width: 900px)": {
                  flex: 1,
                },
              },
              "& .overlayAction": {
                marginLeft: 10,
              },
              "& .rowActions": {
                display: "flex",
                justifyContent: "flex-end",
                "@media (max-width: 900px)": {
                  flex: 1,
                },
              },
            }}
          >
            <Switch
              value="enableTLS"
              id="enableTLS"
              name="enableTLS"
              checked={enableTLS}
              onChange={(e) => {
                const targetD = e.target;
                const checked = targetD.checked;
                setEnableTLS(checked);
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
                    setEnableAutoCert(checked);
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
                    setEnableCustomCerts(checked);
                  }}
                  label={"Custom Certificates"}
                  description={"Certificates used to terminated TLS at MinIO"}
                />

                {enableCustomCerts && (
                  <Fragment>
                    {!enableAutoCert && (
                      <Grid item xs={12}>
                        <TLSHelpBox />
                      </Grid>
                    )}
                    <Grid item xs={12} className={"inputItem"}>
                      <h5>MinIO Server Certificates</h5>
                    </Grid>
                    <Grid item xs={12}>
                      {minioServerCertificateSecrets.map(
                        (certificateInfo: ICertificateInfo) => (
                          <TLSCertificate
                            certificateInfo={certificateInfo}
                            onDelete={() => removeCertificate(certificateInfo)}
                          />
                        ),
                      )}
                    </Grid>
                    <Grid item xs={12}>
                      {minioServerCertificates.map((keyPair, index) => (
                        <Grid
                          item
                          xs={12}
                          key={keyPair.id}
                          className={"minioCertificateRows"}
                        >
                          <FileSelector
                            onChange={(event, fileName, encodedValue) => {
                              if (encodedValue) {
                                addFileToKeyPair(
                                  "minio",
                                  keyPair.id,
                                  "cert",
                                  fileName,
                                  encodedValue,
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
                                addFileToKeyPair(
                                  "minio",
                                  keyPair.id,
                                  "key",
                                  fileName,
                                  encodedValue,
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
                          <Grid item xs={2} className={"rowActions"}>
                            <div className={"overlayAction"}>
                              <IconButton
                                size={"small"}
                                onClick={() => addKeyPair("minio")}
                                disabled={
                                  index !== minioServerCertificates.length - 1
                                }
                              >
                                <AddIcon />
                              </IconButton>
                            </div>
                            <div className={"overlayAction"}>
                              <IconButton
                                size={"small"}
                                onClick={() =>
                                  deleteKeyPair("minio", keyPair.id)
                                }
                                disabled={minioServerCertificates.length <= 1}
                              >
                                <RemoveIcon />
                              </IconButton>
                            </div>
                          </Grid>
                        </Grid>
                      ))}
                    </Grid>

                    <Grid item xs={12} className={"inputItem"}>
                      <h5>MinIO Client Certificates</h5>
                    </Grid>
                    <Grid item xs={12}>
                      {minioClientCertificateSecrets.map(
                        (certificateInfo: ICertificateInfo) => (
                          <TLSCertificate
                            certificateInfo={certificateInfo}
                            onDelete={() => removeCertificate(certificateInfo)}
                          />
                        ),
                      )}
                    </Grid>
                    <Grid item xs={12} className={"inputItem"}>
                      {minioClientCertificates.map((keyPair, index) => (
                        <Grid
                          item
                          xs={12}
                          key={keyPair.id}
                          className={"minioCertificateRows"}
                        >
                          <FileSelector
                            onChange={(event, fileName, encodedValue) => {
                              if (encodedValue) {
                                addFileToKeyPair(
                                  "client",
                                  keyPair.id,
                                  "cert",
                                  fileName,
                                  encodedValue,
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
                                addFileToKeyPair(
                                  "client",
                                  keyPair.id,
                                  "key",
                                  fileName,
                                  encodedValue,
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
                          <Grid item xs={2} className={"rowActions"}>
                            <div className={"overlayAction"}>
                              <IconButton
                                size={"small"}
                                onClick={() => addKeyPair("client")}
                                disabled={
                                  index !== minioClientCertificates.length - 1
                                }
                              >
                                <AddIcon />
                              </IconButton>
                            </div>
                            <div className={"overlayAction"}>
                              <IconButton
                                size={"small"}
                                onClick={() =>
                                  deleteKeyPair("client", keyPair.id)
                                }
                                disabled={minioClientCertificates.length <= 1}
                              >
                                <RemoveIcon />
                              </IconButton>
                            </div>
                          </Grid>
                        </Grid>
                      ))}
                    </Grid>

                    <Grid item xs={12}>
                      <h5>MinIO CA Certificates</h5>
                    </Grid>
                    <Grid item xs={12}>
                      {minioTLSCaCertificateSecrets.map(
                        (certificateInfo: ICertificateInfo) => (
                          <TLSCertificate
                            certificateInfo={certificateInfo}
                            onDelete={() => removeCertificate(certificateInfo)}
                          />
                        ),
                      )}
                    </Grid>
                    <Grid item xs={12}>
                      {minioCaCertificates.map((keyPair: KeyPair, index) => (
                        <Grid
                          item
                          xs={12}
                          key={keyPair.id}
                          className={"minioCertificateRows"}
                        >
                          <Grid item xs={10}>
                            <FileSelector
                              onChange={(event, fileName, encodedValue) => {
                                if (encodedValue) {
                                  addFileToKeyPair(
                                    "minioCAs",
                                    keyPair.id,
                                    "cert",
                                    fileName,
                                    encodedValue,
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
                          <Grid item xs={2}>
                            <div className={"rowActions"}>
                              <div className={"overlayAction"}>
                                <IconButton
                                  size={"small"}
                                  onClick={() => addKeyPair("minioCAs")}
                                  disabled={
                                    index !== minioCaCertificates.length - 1
                                  }
                                >
                                  <AddIcon />
                                </IconButton>
                              </div>
                              <div className={"overlayAction"}>
                                <IconButton
                                  size={"small"}
                                  onClick={() =>
                                    deleteKeyPair("minioCAs", keyPair.id)
                                  }
                                  disabled={minioCaCertificates.length <= 1}
                                >
                                  <RemoveIcon />
                                </IconButton>
                              </div>
                            </div>
                          </Grid>
                        </Grid>
                      ))}
                    </Grid>
                  </Fragment>
                )}
              </Fragment>
            )}
            <Grid item xs={12} className={"inputItem"}>
              <SectionTitle separator>Security Context</SectionTitle>
            </Grid>
            <Grid item xs={12} className={"inputItem"}>
              <SecurityContextSelector
                runAsGroup={runAsGroup}
                runAsUser={runAsUser}
                fsGroup={fsGroup}
                runAsNonRoot={runAsNonRoot}
                fsGroupChangePolicy={fsGroupChangePolicy}
                setFSGroup={(value: string) => dispatch(setFSGroup(value))}
                setRunAsUser={(value: string) => dispatch(setRunAsUser(value))}
                setRunAsGroup={(value: string) =>
                  dispatch(setRunAsGroup(value))
                }
                setRunAsNonRoot={(value: boolean) =>
                  dispatch(setRunAsNonRoot(value))
                }
                setFSGroupChangePolicy={(value: fsGroupChangePolicyType) =>
                  dispatch(setFSGroupChangePolicy(value))
                }
              />
            </Grid>
            <Grid item xs={12} sx={modalStyleUtils.modalButtonBar}>
              <Button
                id={"save-security"}
                type="submit"
                variant="callAction"
                disabled={dialogOpen || isSending}
                onClick={() => setDialogOpen(true)}
                label={"Save"}
              />
            </Grid>
          </FormLayout>
        </Fragment>
      )}
    </React.Fragment>
  );
};

export default TenantSecurity;
