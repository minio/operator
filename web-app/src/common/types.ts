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
  ILabelKeyPair,
  ISecurityContext,
} from "../screens/Console/Tenants/types";

export interface ITenantsObject {
  tenants: ITenant[];
}

export interface ITenant {
  creation_date: string;
  deletion_date: string;
  currentState: string;
  image: string;
  instance_count: string;
  name: string;
  namespace?: string;
  total_size: string;
  used_size: string;
  volume_count: string;
  volume_size: string;
  volumes_per_server?: string;
  pool_count: string;
  pools?: IPoolModel[];
  used_capacity?: string;
  endpoint?: string;
  storage_class?: string;
  enable_prometheus: boolean;
}

export interface IVolumeConfiguration {
  size: string;
  storage_class_name: string;
  labels?: any;
}

export interface IDomainsRequest {
  console?: string;
  minio?: string[];
}

export interface ITenantCreator {
  name: string;
  service_name: string;
  enable_console: boolean;
  enable_prometheus: boolean;
  enable_tls: boolean;
  access_key: string;
  secret_key: string;
  access_keys: string[];
  secret_keys: string[];
  image: string;
  expose_minio: boolean;
  expose_console: boolean;
  pools: IPoolModel[];
  namespace: string;
  erasureCodingParity: number;
  tls?: ITLSTenantConfiguration;
  encryption?: IEncryptionConfiguration;
  idp?: IIDPConfiguration;
  annotations?: Object;
  image_registry?: ImageRegistry;
  logSearchConfiguration?: LogSearchConfiguration;
  prometheusConfiguration?: PrometheusConfiguration;
  affinity?: AffinityConfiguration;
  domains?: IDomainsRequest;
}

export interface ImageRegistry {
  registry: string;
  username: string;
  password: string;
}

export interface ITenantUpdateObject {
  image: string;
  image_registry?: IRegistryObject;
}

export interface IRegistryObject {
  registry: string;
  username: string;
  password: string;
}

export interface ITenantUsage {
  used: string;
  disk_used: string;
}

export interface IAffinityModel {
  podAntiAffinity?: IPodAntiAffinityModel;
  nodeAffinity?: INodeAffinityModel;
}

export interface IPodAntiAffinityModel {
  requiredDuringSchedulingIgnoredDuringExecution: IPodAffinityTerm[];
}

export interface IPodAffinityTerm {
  labelSelector: IPodAffinityTermLabelSelector;
  topologyKey: string;
}

export interface IPodAffinityTermLabelSelector {
  matchExpressions: IMatchExpressionItem[];
}

export interface INodeAffinityModel {
  requiredDuringSchedulingIgnoredDuringExecution: INodeAffinityTerms;
}

export interface INodeAffinityTerms {
  nodeSelectorTerms: INodeAffinityLabelsSelector[];
}

export interface INodeAffinityLabelsSelector {
  matchExpressions: IMatchExpressionItem[];
}

export interface IMatchExpressionItem {
  key: string;
  operator: string;
  values: string[];
}

export enum ITolerationEffect {
  "NoSchedule" = "NoSchedule",
  "PreferNoSchedule" = "PreferNoSchedule",
  "NoExecute" = "NoExecute",
}

export enum ITolerationOperator {
  "Equal" = "Equal",
  "Exists" = "Exists",
}

export interface ITolerationModel {
  effect: ITolerationEffect;
  key: string;
  operator: ITolerationOperator;
  value?: string;
  tolerationSeconds?: ITolerationSeconds;
}

export interface ITolerationSeconds {
  seconds: number;
}

export interface IResourceModel {
  requests?: IResourceRequests;
  limits?: IResourceLimits;
}

export interface IResourceRequests {
  memory?: number;
  cpu?: number;
}

export interface IResourceLimits {
  memory?: number;
  cpu?: number;
}

export interface ITLSTenantConfiguration {
  minio: ITLSConfiguration;
  console: ITLSConfiguration;
}

export interface ITLSConfiguration {
  crt: string;
  key: string;
}

export interface IEncryptionConfiguration {
  server: ITLSConfiguration;
  client: ITLSConfiguration;
  master_key?: string;
  gemalto?: IGemaltoConfig;
  aws?: IAWSConfig;
  vault?: IVaultConfig;
  azure?: IAzureConfig;
  gcp?: IGCPConfig;
}

export interface IGCPCredentials {
  client_email: string;
  client_id: string;
  private_key_id: string;
  private_key: string;
}

export interface IGCPSecretManager {
  project_id: string;
  endpoint?: string;
  credentials?: IGCPCredentials;
}

export interface IGCPConfig {
  secretmanager: IGCPSecretManager;
}

export interface IAzureCredentials {
  tenant_id: string;
  client_id: string;
  client_secret: string;
}

export interface IAzureKeyVault {
  endpoint: string;
  credentials?: IAzureCredentials;
}

export interface IAzureConfig {
  keyvault: IAzureKeyVault;
}

export interface IVaultConfig {
  endpoint: string;
  engine?: string;
  namespace?: string;
  prefix?: string;
  approle: IApproleConfig;
  tls: IVaultTLSConfig;
  status: IVaultStatusConfig;
}

export interface IGemaltoConfig {
  keysecure: IKeysecureConfig;
}

export interface IAWSConfig {
  secretsmanager: ISecretsManagerConfig;
}

export interface IApproleConfig {
  engine: string;
  id: string;
  secret: string;
  retry: number;
}

export interface IVaultTLSConfig {
  key: string;
  crt: string;
  ca: string;
}

export interface IVaultStatusConfig {
  ping: number;
}

export interface IKeysecureConfig {
  endpoint: string;
  credentials: IGemaltoCredentials;
  tls: IGemaltoTLSConfig;
}

export interface IGemaltoCredentials {
  token: string;
  domain: string;
  retry?: string;
}

export interface IGemaltoTLSConfig {
  ca: string;
}

export interface ISecretsManagerConfig {
  endpoint: string;
  region: string;
  kmskey?: string;
  credentials: IAWSCredentials;
}

export interface IAWSCredentials {
  accesskey: string;
  secretkey: string;
  token?: string;
}

export interface IIDPConfiguration {
  oidc?: IOpenIDConfiguration;
  active_directory: IActiveDirectoryConfiguration;
}

export interface IOpenIDConfiguration {
  url: string;
  client_id: string;
  secret_id: string;
}

export interface IActiveDirectoryConfiguration {
  url: string;
  skip_tls_verification: boolean;
  server_insecure: boolean;
  server_start_tls: boolean;
  username_search_filter: string;
  group_Search_base_dn: string;
  group_search_filter: string;
  group_name_attribute: string;
  user_dns: string[];
  lookup_bind_dn: string;
  lookup_bind_password: string;
  user_dn_search_base_dn: string;
  user_dn_search_filter: string;
}

export interface IStorageDistribution {
  error: number | string;
  nodes: number;
  persistentVolumes: number;
  disks: number;
  pvSize: number;
}

export interface IStorageFactors {
  erasureCode: string;
  storageFactor: number;
  maxCapacity: string;
  maxFailureTolerations: number;
}

export interface ITenantHealthInList {
  name: string;
  namespace: string;
  status?: string;
  message?: string;
}

export interface ITenantsListHealthRequest {
  tenants: ITenantHealthInList[];
}

export interface IMaxAllocatableMemoryRequest {
  num_nodes: number;
}

export interface IMaxAllocatableMemoryResponse {
  max_memory: number;
}

export interface IEncryptionUpdateRequest {
  encryption: IEncryptionConfiguration;
}

export interface IArchivedTenantsList {
  tenants: IArchivedTenant[];
}

export interface IArchivedTenant {
  namespace: string;
  tenant: string;
  number_volumes: number;
  capacity: number;
}

export interface IPoolModel {
  name?: string;
  servers: number;
  volumes_per_server: number;
  volume_configuration: IVolumeConfiguration;
  affinity?: IAffinityModel;
  tolerations?: ITolerationModel[];
  resources?: IResourceModel;
  securityContext?: ISecurityContext | null;
}

export interface IUpdatePool {
  pools: IPoolModel[];
}

export interface INode {
  name: string;
  freeSpace: string;
  totalSpace: string;
  disks: IDisk[];
}

export interface IStorageType {
  freeSpace: string;
  totalSpace: string;
  storageClasses: string[];
  nodes: INode[];
  schedulableNodes: INode[];
}

export interface IDisk {
  name: string;
  freeSpace: string;
  totalSpace: string;
}

export interface ICapacity {
  value: string;
  unit: string;
}

export interface IErasureCodeCalc {
  error: number;
  maxEC: string;
  erasureCodeSet: number;
  rawCapacity: string;
  defaultEC: string;
  storageFactors: IStorageFactors[];
}

export interface LogSearchConfiguration {
  storageClass?: string;
  storageSize?: number;
  image: string;
  postgres_image: string;
  postgres_init_image: string;
  securityContext?: ISecurityContext;
  postgres_securityContext?: ISecurityContext;
}

export interface PrometheusConfiguration {
  storageClass?: string;
  storageSize?: number;
  image: string;
  sidecar_image: string;
  init_image: string;
  securityContext?: ISecurityContext;
}

export interface AffinityConfiguration {
  affinityType: "default" | "nodeSelector" | "none";
  nodeSelectorLabels?: ILabelKeyPair[];
  withPodAntiAffinity?: boolean;
}

export interface ErrorResponseHandler {
  errorMessage: string;
  detailedError: string;
  statusCode?: number;
}

export interface IRetentionConfig {
  mode: string;
  unit: string;
  validity: number;
}

export interface IBytesCalc {
  total: number;
  unit: string;
}

export interface IEmbeddedCustomButton {
  backgroundColor?: string;
  textColor?: string;
  hoverColor?: string;
  hoverText?: string;
  activeColor?: string;
  activeText?: string;
}

export interface IEmbeddedCustomStyles {
  backgroundColor: string;
  fontColor: string;
  buttonStyles: IEmbeddedCustomButton;
}
