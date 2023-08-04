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

import { createAsyncThunk } from "@reduxjs/toolkit";
import { AppState } from "../../../../../store";
import { generatePoolName, getBytes } from "../../../../../common/utils";
import { getDefaultAffinity, getNodeSelector } from "../../TenantDetails/utils";
import { KeyPair } from "../../ListTenants/utils";
import { createTenantCall } from "../createTenantAPI";
import { setErrorSnackMessage } from "../../../../../systemSlice";
import { CreateTenantRequest } from "../../../../../api/operatorApi";

export const createTenantAsync = createAsyncThunk(
  "createTenant/createTenantAsync",
  async (_, { getState, rejectWithValue, dispatch }) => {
    const state = getState() as AppState;

    let fields = state.createTenant.fields;
    let certificates = state.createTenant.certificates;

    const tenantName = fields.nameTenant.tenantName;
    const selectedStorageClass = fields.nameTenant.selectedStorageClass;
    const imageName = fields.configure.imageName;
    const customDockerhub = fields.configure.customDockerhub;
    const imageRegistry = fields.configure.imageRegistry;
    const imageRegistryUsername = fields.configure.imageRegistryUsername;
    const imageRegistryPassword = fields.configure.imageRegistryPassword;
    const exposeMinIO = fields.configure.exposeMinIO;
    const exposeConsole = fields.configure.exposeConsole;
    const exposeSFTP = fields.configure.exposeSFTP;
    const idpSelection = fields.identityProvider.idpSelection;
    const openIDConfigurationURL =
      fields.identityProvider.openIDConfigurationURL;
    const openIDClientID = fields.identityProvider.openIDClientID;
    const openIDClaimName = fields.identityProvider.openIDClaimName;
    const openIDCallbackURL = fields.identityProvider.openIDCallbackURL;
    const openIDScopes = fields.identityProvider.openIDScopes;
    const openIDSecretID = fields.identityProvider.openIDSecretID;
    const ADURL = fields.identityProvider.ADURL;
    const ADSkipTLS = fields.identityProvider.ADSkipTLS;
    const ADServerInsecure = fields.identityProvider.ADServerInsecure;
    const ADGroupSearchBaseDN = fields.identityProvider.ADGroupSearchBaseDN;
    const ADGroupSearchFilter = fields.identityProvider.ADGroupSearchFilter;
    const ADUserDNs = fields.identityProvider.ADUserDNs;
    const ADGroupDNs = fields.identityProvider.ADGroupDNs;
    const ADLookupBindDN = fields.identityProvider.ADLookupBindDN;
    const ADLookupBindPassword = fields.identityProvider.ADLookupBindPassword;
    const ADUserDNSearchBaseDN = fields.identityProvider.ADUserDNSearchBaseDN;
    const ADUserDNSearchFilter = fields.identityProvider.ADUserDNSearchFilter;
    const ADServerStartTLS = fields.identityProvider.ADServerStartTLS;
    const accessKeys = fields.identityProvider.accessKeys;
    const secretKeys = fields.identityProvider.secretKeys;
    const minioServerCertificates = certificates.minioServerCertificates;
    const minioClientCertificates = certificates.minioClientCertificates;
    const minioCAsCertificates = certificates.minioCAsCertificates;
    const kesServerCertificate = certificates.kesServerCertificate;
    const minioMTLSCertificate = certificates.minioMTLSCertificate;
    const kmsMTLSCertificate = certificates.kmsMTLSCertificate;
    const kmsCA = certificates.kmsCA;
    const rawConfiguration = fields.encryption.rawConfiguration;
    const encryptionTab = fields.encryption.encryptionTab;
    const enableEncryption = fields.encryption.enableEncryption;
    const encryptionType = fields.encryption.encryptionType;
    const gemaltoEndpoint = fields.encryption.gemaltoEndpoint;
    const gemaltoToken = fields.encryption.gemaltoToken;
    const gemaltoDomain = fields.encryption.gemaltoDomain;
    const gemaltoRetry = fields.encryption.gemaltoRetry;
    const awsEndpoint = fields.encryption.awsEndpoint;
    const awsRegion = fields.encryption.awsRegion;
    const awsKMSKey = fields.encryption.awsKMSKey;
    const awsAccessKey = fields.encryption.awsAccessKey;
    const awsSecretKey = fields.encryption.awsSecretKey;
    const awsToken = fields.encryption.awsToken;
    const vaultEndpoint = fields.encryption.vaultEndpoint;
    const vaultEngine = fields.encryption.vaultEngine;
    const vaultNamespace = fields.encryption.vaultNamespace;
    const vaultPrefix = fields.encryption.vaultPrefix;
    const vaultAppRoleEngine = fields.encryption.vaultAppRoleEngine;
    const vaultId = fields.encryption.vaultId;
    const vaultSecret = fields.encryption.vaultSecret;
    const vaultRetry = fields.encryption.vaultRetry;
    const vaultPing = fields.encryption.vaultPing;
    const azureEndpoint = fields.encryption.azureEndpoint;
    const azureTenantID = fields.encryption.azureTenantID;
    const azureClientID = fields.encryption.azureClientID;
    const azureClientSecret = fields.encryption.azureClientSecret;
    const gcpProjectID = fields.encryption.gcpProjectID;
    const gcpEndpoint = fields.encryption.gcpEndpoint;
    const gcpClientEmail = fields.encryption.gcpClientEmail;
    const gcpClientID = fields.encryption.gcpClientID;
    const gcpPrivateKeyID = fields.encryption.gcpPrivateKeyID;
    const gcpPrivateKey = fields.encryption.gcpPrivateKey;
    const enableAutoCert = fields.security.enableAutoCert;
    const enableTLS = fields.security.enableTLS;
    const ecParity = fields.tenantSize.ecParity;
    const distribution = fields.tenantSize.distribution;
    const tenantCustom = fields.configure.tenantCustom;
    const kesImage = fields.configure.kesImage;
    const affinityType = fields.affinity.podAffinity;
    const nodeSelectorLabels = fields.affinity.nodeSelectorLabels;
    const withPodAntiAffinity = fields.affinity.withPodAntiAffinity;

    const tenantSecurityContext = fields.configure.tenantSecurityContext;
    const kesSecurityContext = fields.encryption.kesSecurityContext;
    const kesReplicas = fields.encryption.replicas;
    const setDomains = fields.configure.setDomains;
    const minioDomains = fields.configure.minioDomains;
    const consoleDomain = fields.configure.consoleDomain;
    const environmentVariables = fields.configure.envVars;
    const customRuntime = fields.configure.customRuntime;
    const runtimeClassName = fields.configure.runtimeClassName;

    let tolerations = state.createTenant.tolerations;
    let namespace = state.createTenant.fields.nameTenant.namespace;

    const tolerationValues = tolerations.filter(
      (toleration) => toleration.key.trim() !== "",
    );

    const poolName = generatePoolName([]);

    let affinityObject = {};

    switch (affinityType) {
      case "default":
        affinityObject = {
          affinity: getDefaultAffinity(tenantName, poolName),
        };
        break;
      case "nodeSelector":
        affinityObject = {
          affinity: getNodeSelector(
            nodeSelectorLabels,
            withPodAntiAffinity,
            tenantName,
            poolName,
          ),
        };
        break;
    }

    const erasureCode = ecParity.split(":")[1];

    let runtimeClass = {};

    if (customRuntime) {
      runtimeClass = {
        runtimeClassName,
      };
    }

    let dataSend: CreateTenantRequest = {
      name: tenantName,
      namespace: namespace,
      access_key: "",
      secret_key: "",
      enable_tls: enableTLS && enableAutoCert,
      enable_console: true,
      image: imageName,
      expose_minio: exposeMinIO,
      expose_console: exposeConsole,
      expose_sftp: exposeSFTP,
      pools: [
        {
          name: poolName,
          servers: distribution.nodes,
          volumes_per_server: distribution.disks,
          volume_configuration: {
            size: distribution.pvSize,
            storage_class_name: selectedStorageClass,
          },
          securityContext: tenantCustom ? tenantSecurityContext : undefined,
          tolerations: tolerationValues,
          ...affinityObject,
          ...runtimeClass,
        },
      ],
      erasureCodingParity: parseInt(erasureCode, 10),
    };

    // Set Resources
    if (
      fields.tenantSize.resourcesCPURequest !== "" ||
      fields.tenantSize.resourcesCPULimit !== "" ||
      fields.tenantSize.resourcesMemoryRequest !== "" ||
      fields.tenantSize.resourcesMemoryLimit !== ""
    ) {
      dataSend.pools[0].resources = {};
      // requests
      if (
        fields.tenantSize.resourcesCPURequest !== "" ||
        fields.tenantSize.resourcesMemoryRequest !== ""
      ) {
        dataSend.pools[0].resources.requests = {};
        if (fields.tenantSize.resourcesCPURequest !== "") {
          dataSend.pools[0].resources.requests.cpu = parseInt(
            fields.tenantSize.resourcesCPURequest,
          );
        }
        if (fields.tenantSize.resourcesMemoryRequest !== "") {
          dataSend.pools[0].resources.requests.memory = parseInt(
            getBytes(fields.tenantSize.resourcesMemoryRequest, "Gi", true),
          );
        }
      }
      // limits
      if (
        fields.tenantSize.resourcesCPULimit !== "" ||
        fields.tenantSize.resourcesMemoryLimit !== ""
      ) {
        dataSend.pools[0].resources.limits = {};
        if (fields.tenantSize.resourcesCPULimit !== "") {
          dataSend.pools[0].resources.limits.cpu = parseInt(
            fields.tenantSize.resourcesCPULimit,
          );
        }
        if (fields.tenantSize.resourcesMemoryLimit !== "") {
          dataSend.pools[0].resources.limits.memory = parseInt(
            getBytes(fields.tenantSize.resourcesMemoryLimit, "Gi", true),
          );
        }
      }
    }
    if (customDockerhub) {
      dataSend = {
        ...dataSend,
        image_registry: {
          registry: imageRegistry,
          username: imageRegistryUsername,
          password: imageRegistryPassword,
        },
      };
    }

    let tenantServerCertificates: any = null;
    let tenantClientCertificates: any = null;
    let tenantCAsCertificates: any = null;

    if (enableTLS && minioServerCertificates.length > 0) {
      tenantServerCertificates = {
        minioServerCertificates: minioServerCertificates
          .map((keyPair: KeyPair) => ({
            crt: keyPair.encoded_cert,
            key: keyPair.encoded_key,
          }))
          .filter((keyPair) => keyPair.crt && keyPair.key),
      };
    }

    if (enableTLS && minioClientCertificates.length > 0) {
      tenantClientCertificates = {
        minioClientCertificates: minioClientCertificates
          .map((keyPair: KeyPair) => ({
            crt: keyPair.encoded_cert,
            key: keyPair.encoded_key,
          }))
          .filter((keyPair) => keyPair.crt && keyPair.key),
      };
    }

    if (enableTLS && minioCAsCertificates.length > 0) {
      tenantCAsCertificates = {
        minioCAsCertificates: minioCAsCertificates
          .map((keyPair: KeyPair) => keyPair.encoded_cert)
          .filter((keyPair) => keyPair),
      };
    }

    if (
      minioServerCertificates ||
      minioClientCertificates ||
      minioCAsCertificates
    ) {
      dataSend = {
        ...dataSend,
        tls: {
          ...tenantServerCertificates,
          ...tenantClientCertificates,
          ...tenantCAsCertificates,
        },
      };
    }

    if (enableEncryption) {
      let insertEncrypt = {};

      switch (encryptionType) {
        case "gemalto":
          insertEncrypt = {
            gemalto: {
              keysecure: {
                endpoint: gemaltoEndpoint,
                credentials: {
                  token: gemaltoToken,
                  domain: gemaltoDomain,
                  retry: parseInt(gemaltoRetry),
                },
              },
            },
          };
          break;
        case "aws":
          insertEncrypt = {
            aws: {
              secretsmanager: {
                endpoint: awsEndpoint,
                region: awsRegion,
                kmskey: awsKMSKey,
                credentials: {
                  accesskey: awsAccessKey,
                  secretkey: awsSecretKey,
                  token: awsToken,
                },
              },
            },
          };
          break;
        case "azure":
          insertEncrypt = {
            azure: {
              keyvault: {
                endpoint: azureEndpoint,
                credentials: {
                  tenant_id: azureTenantID,
                  client_id: azureClientID,
                  client_secret: azureClientSecret,
                },
              },
            },
          };
          break;
        case "gcp":
          insertEncrypt = {
            gcp: {
              secretmanager: {
                project_id: gcpProjectID,
                endpoint: gcpEndpoint,
                credentials: {
                  client_email: gcpClientEmail,
                  client_id: gcpClientID,
                  private_key_id: gcpPrivateKeyID,
                  private_key: gcpPrivateKey,
                },
              },
            },
          };
          break;
        case "vault":
          insertEncrypt = {
            vault: {
              endpoint: vaultEndpoint,
              engine: vaultEngine,
              namespace: vaultNamespace,
              prefix: vaultPrefix,
              approle: {
                engine: vaultAppRoleEngine,
                id: vaultId,
                secret: vaultSecret,
                retry: parseInt(vaultRetry),
              },
              status: {
                ping: parseInt(vaultPing),
              },
            },
          };
          break;
      }

      let encryptionServerKeyPair: any = {};
      let encryptionClientKeyPair: any = {};
      let encryptionKMSCertificates: any = {};

      // MinIO -> KES (mTLS certificates)
      if (
        minioMTLSCertificate.encoded_key !== "" &&
        minioMTLSCertificate.encoded_cert !== ""
      ) {
        encryptionClientKeyPair = {
          minio_mtls: {
            key: minioMTLSCertificate.encoded_key,
            crt: minioMTLSCertificate.encoded_cert,
          },
        };
      }

      // KES server certificates
      if (
        kesServerCertificate.encoded_key !== "" &&
        kesServerCertificate.encoded_cert !== ""
      ) {
        encryptionServerKeyPair = {
          server_tls: {
            key: kesServerCertificate.encoded_key,
            crt: kesServerCertificate.encoded_cert,
          },
        };
      }

      // KES -> KMS (mTLS certificates)
      let kmsMTLSKeyPair = null;
      let kmsCAInsert = null;
      if (
        kmsMTLSCertificate.encoded_key !== "" &&
        kmsMTLSCertificate.encoded_cert !== ""
      ) {
        kmsMTLSKeyPair = {
          key: kmsMTLSCertificate.encoded_key,
          crt: kmsMTLSCertificate.encoded_cert,
        };
      }
      if (kmsCA.encoded_cert !== "") {
        kmsCAInsert = {
          ca: kmsCA.encoded_cert,
        };
      }
      if (kmsMTLSKeyPair || kmsCAInsert) {
        encryptionKMSCertificates = {
          kms_mtls: {
            ...kmsMTLSKeyPair,
            ...kmsCAInsert,
          },
        };
      }

      dataSend = {
        ...dataSend,
        encryption: {
          raw: encryptionTab ? rawConfiguration : "",
          replicas: kesReplicas,
          securityContext: kesSecurityContext,
          image: kesImage,
          ...encryptionClientKeyPair,
          ...encryptionServerKeyPair,
          ...encryptionKMSCertificates,
          ...insertEncrypt,
        },
      };
    }

    let dataIDP: any = {};
    switch (idpSelection) {
      case "Built-in":
        let keyarray = [];
        for (let i = 0; i < accessKeys.length; i++) {
          keyarray.push({
            access_key: accessKeys[i],
            secret_key: secretKeys[i],
          });
        }
        dataIDP = {
          keys: keyarray,
        };
        break;
      case "OpenID":
        dataIDP = {
          oidc: {
            configuration_url: openIDConfigurationURL,
            client_id: openIDClientID,
            secret_id: openIDSecretID,
            claim_name: openIDClaimName,
            callback_url: openIDCallbackURL,
            scopes: openIDScopes,
          },
        };
        break;
      case "AD":
        dataIDP = {
          active_directory: {
            url: ADURL,
            skip_tls_verification: ADSkipTLS,
            server_insecure: ADServerInsecure,
            group_search_base_dn: ADGroupSearchBaseDN,
            group_search_filter: ADGroupSearchFilter,
            user_dns: ADUserDNs.filter((user) => user.trim() !== ""),
            group_dns: ADGroupDNs.filter((group) => group.trim() !== ""),
            lookup_bind_dn: ADLookupBindDN,
            lookup_bind_password: ADLookupBindPassword,
            user_dn_search_base_dn: ADUserDNSearchBaseDN,
            user_dn_search_filter: ADUserDNSearchFilter,
            server_start_tls: ADServerStartTLS,
          },
        };
        break;
    }

    let domains: any = {};
    let sendDomain: any = {};
    let sendEnvironmentVariables: any = {};

    if (setDomains) {
      if (consoleDomain !== "") {
        domains.console = consoleDomain;
      }

      const filteredDomains = minioDomains.filter((dom) => dom.trim() !== "");

      if (filteredDomains.length > 0) {
        domains.minio = filteredDomains;
      }

      if (Object.keys(domains).length > 0) {
        sendDomain.domains = domains;
      }
    }

    sendEnvironmentVariables.environmentVariables = environmentVariables
      .map((envVar) => ({
        key: envVar.key.trim(),
        value: envVar.value.trim(),
      }))
      .filter((envVar) => envVar.key !== "");

    dataSend = {
      ...dataSend,
      ...sendDomain,
      ...sendEnvironmentVariables,
      idp: { ...dataIDP },
    };

    return createTenantCall(dataSend)
      .then((resp) => {
        return resp;
      })
      .catch((err) => {
        dispatch(setErrorSnackMessage(err));
        return rejectWithValue(err);
      });
  },
);
