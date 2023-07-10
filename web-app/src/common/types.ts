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

export interface IDomainsRequest {
  console?: string;
  minio?: string[];
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

export interface IAWSConfig {
  secretsmanager: ISecretsManagerConfig;
}

export interface IGemaltoCredentials {
  token: string;
  domain: string;
  retry?: string;
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

export interface ErrorResponseHandler {
  errorMessage: string;
  detailedError: string;
  statusCode?: number;
  message?: string; // if any type error thrown
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
