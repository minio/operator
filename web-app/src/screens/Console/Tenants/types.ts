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

import {
  IAWSConfig,
  IAzureConfig,
  IErasureCodeCalc,
  IGCPConfig,
  IGemaltoCredentials,
} from "../../../common/types";
import { IResourcesSize } from "./ListTenants/types";
import { KeyPair, Opts } from "./ListTenants/utils";
import { IntegrationConfiguration } from "./AddTenant/Steps/TenantResources/utils";
import { SecurityContext } from "../../../api/operatorApi";

export interface ICertificateInfo {
  name: string;
  serialNumber: string;
  domains: string[];
  expiry: string;
}

export interface ICustomCertificates {
  minio: ICertificateInfo[];
  client: ICertificateInfo[];
  minioCAs: ICertificateInfo[];
  console: ICertificateInfo[];
  consoleCAs: ICertificateInfo[];
}

export interface ITenantConfigurationResponse {
  environmentVariables: LabelKeyPair[];
}

export interface ITenantConfigurationRequest {
  environmentVariables: LabelKeyPair[];
  keysToBeDeleted: string[];
}

export interface ITenantSecurityResponse {
  autoCert: boolean;
  customCertificates: ICustomCertificates;
  securityContext: SecurityContext;
}

export interface IVaultAppRole {
  engine: string;
  id: string;
  secret: string;
  retry: string;
}

export interface IVaultStatus {
  ping: string;
}

export interface IVaultConfiguration {
  endpoint: string;
  engine: string;
  namespace: string;
  prefix: string;
  approle: IVaultAppRole;
  status: IVaultStatus;
}

export interface IKeysecureConfiguration {
  endpoint: string;
  credentials: IGemaltoCredentials;
}

export interface IGemaltoConfiguration {
  keysecure: IKeysecureConfiguration;
}

export interface ITenantEncryptionResponse {
  raw: string;
  image: string;
  replicas: string;
  policies?: Record<string, any>;
  securityContext: SecurityContext;
  server_tls: ICertificateInfo;
  minio_mtls: ICertificateInfo;
  kms_mtls: {
    crt: ICertificateInfo;
    ca: ICertificateInfo;
  };
  vault?: IVaultConfiguration;
  aws?: IAWSConfig;
  gemalto?: IGemaltoConfiguration;
  gcp?: IGCPConfig;
  azure?: IAzureConfig;
}

export interface ICertificatesItems {
  minioServerCertificates: KeyPair[];
  minioClientCertificates: KeyPair[];
  minioCAsCertificates: KeyPair[];
  kesServerCertificate: KeyPair;
  minioMTLSCertificate: KeyPair;
  kmsMTLSCertificate: KeyPair;
  kmsCA: KeyPair;
}

export interface IFieldStore {
  nameTenant: INameTenantFields;
  configure: IConfigureFields;
  identityProvider: IIdentityProviderFields;
  security: ISecurityFields;
  encryption: IEncryptionFields;
  tenantSize: ITenantSizeFields;
  affinity: ITenantAffinity;
}

export interface INameTenantFields {
  tenantName: string;
  namespace: string;
  selectedStorageClass: string;
  selectedStorageType: string;
}

export interface LabelKeyPair {
  key: string;
  value: string;
}

export type fsGroupChangePolicyType = "Always" | "OnRootMismatch";

export interface IConfigureFields {
  customImage: boolean;
  imageName: string;
  customDockerhub: boolean;
  imageRegistry: string;
  imageRegistryUsername: string;
  imageRegistryPassword: string;
  exposeMinIO: boolean;
  exposeConsole: boolean;
  tenantCustom: boolean;
  customRuntime: boolean;
  runtimeClassName: string;
  envVars: LabelKeyPair[];
  kesImage: string;
  setDomains: boolean;
  consoleDomain: string;
  minioDomains: string[];
  tenantSecurityContext: SecurityContext;
}

export interface IIdentityProviderFields {
  idpSelection: string;
  accessKeys: string[];
  secretKeys: string[];
  openIDConfigurationURL: string;
  openIDClientID: string;
  openIDSecretID: string;
  openIDCallbackURL: string;
  openIDClaimName: string;
  openIDScopes: string;
  ADURL: string;
  ADSkipTLS: boolean;
  ADServerInsecure: boolean;
  ADGroupSearchBaseDN: string;
  ADGroupSearchFilter: string;
  ADUserDNs: string[];
  ADGroupDNs: string[];
  ADLookupBindDN: string;
  ADLookupBindPassword: string;
  ADUserDNSearchBaseDN: string;
  ADUserDNSearchFilter: string;
  ADServerStartTLS: boolean;
}

export interface ISecurityFields {
  enableTLS: boolean;
  enableAutoCert: boolean;
  enableCustomCerts: boolean;
}

export interface IEncryptionFields {
  rawConfiguration: string;
  encryptionTab: number;
  enableEncryption: boolean;
  encryptionType: string;
  gemaltoEndpoint: string;
  gemaltoToken: string;
  gemaltoDomain: string;
  gemaltoRetry: string;
  awsEndpoint: string;
  awsRegion: string;
  awsKMSKey: string;
  awsAccessKey: string;
  awsSecretKey: string;
  awsToken: string;
  vaultEndpoint: string;
  vaultEngine: string;
  vaultNamespace: string;
  vaultPrefix: string;
  vaultAppRoleEngine: string;
  vaultId: string;
  vaultSecret: string;
  vaultRetry: string;
  vaultPing: string;
  azureEndpoint: string;
  azureTenantID: string;
  azureClientID: string;
  azureClientSecret: string;
  gcpProjectID: string;
  gcpEndpoint: string;
  gcpClientEmail: string;
  gcpClientID: string;
  gcpPrivateKeyID: string;
  gcpPrivateKey: string;
  enableCustomCertsForKES: boolean;
  replicas: string;
  kesSecurityContext: SecurityContext;
}

export interface ITenantSizeFields {
  volumeSize: string;
  sizeFactor: string;
  drivesPerServer: string;
  nodes: string;
  memoryNode: string;
  ecParity: string;
  ecParityChoices: Opts[];
  cleanECChoices: string[];
  untouchedECField: boolean;
  resourcesSize: IResourcesSize;
  distribution: any;
  ecParityCalc: IErasureCodeCalc;
  cpuToUse: string;
  limitSize: any;
  maxAllocatableResources: AllocableResourcesResponse;
  maxCPUsUse: string;
  maxMemorySize: string;
  integrationSelection: IntegrationConfiguration;

  resourcesSpecifyLimit: boolean;

  resourcesCPURequestError: string;
  resourcesCPURequest: string;
  resourcesCPULimitError: string;
  resourcesCPULimit: string;

  resourcesMemoryRequestError: string;
  resourcesMemoryRequest: string;
  resourcesMemoryLimitError: string;
  resourcesMemoryLimit: string;
}

export interface ITenantAffinity {
  podAffinity: "default" | "nodeSelector" | "none";
  nodeSelectorLabels: string;
  withPodAntiAffinity: boolean;
}

export interface ILabelKeyPair {
  labelKey: string;
  labelValue: string;
}

export interface AllocableResourcesResponse {
  min_allocatable_mem?: number;
  min_allocatable_cpu?: number;
  cpu_priority: NodeMaxAllocatableResources;
  mem_priority: NodeMaxAllocatableResources;
}

export interface NodeMaxAllocatableResources {
  max_allocatable_cpu: number;
  max_allocatable_mem: number;
}

export interface IAddPoolSetup {
  numberOfNodes: number;
  volumeSize: number;
  volumesPerServer: number;
  storageClass: string;
}

export interface IPoolConfiguration {
  securityContextEnabled: boolean;
  securityContext: SecurityContext;
  customRuntime: boolean;
  runtimeClassName: string;
}

export interface ITenantIdentityProviderResponse {
  oidc?: {
    callback_url: string;
    claim_name: string;
    client_id: string;
    configuration_url: string;
    scopes: string;
    secret_id: string;
  };
  active_directory?: {
    lookup_bind_dn: string;
    lookup_bind_password: string;
    server_start_tls: boolean;
    skip_tls_verification: boolean;
    url: string;
    group_search_base_dn: string;
    group_search_filter: string;
    server_insecure: boolean;
    user_dn_search_base_dn: string;
    user_dn_search_filter: string;
  };
}

export interface ITenantSetAdministratorsRequest {
  user_dns?: string[];
  group_dns?: string[];
}

export interface IEditMonitoringSecurityContext {
  securityContextEnabled: boolean;
  runAsUser: string;
  runAsGroup: string;
  fsGroup: string;
  runAsNonRoot: boolean;
}

export interface IEditTenantSecurityContext {
  securityContextEnabled: boolean;
  runAsUser: string;
  runAsGroup: string;
  fsGroup: string;
  fsGroupChangePolicy: fsGroupChangePolicyType;
  runAsNonRoot: boolean;
}
