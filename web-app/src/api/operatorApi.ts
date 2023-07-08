/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface Error {
  /** @format int32 */
  code?: number;
  message: string;
  detailedMessage: string;
}

export interface LoginDetails {
  loginStrategy?:
    | "form"
    | "redirect"
    | "service-account"
    | "redirect-service-account";
  redirectRules?: RedirectRule[];
  isK8S?: boolean;
}

export interface LoginRequest {
  accessKey?: string;
  secretKey?: string;
  sts?: string;
  features?: {
    hide_menu?: boolean;
  };
}

export interface LoginOauth2AuthRequest {
  state: string;
  code: string;
}

export interface LoginOperatorRequest {
  jwt: string;
}

export interface Principal {
  STSAccessKeyID?: string;
  STSSecretAccessKey?: string;
  STSSessionToken?: string;
  accountAccessKey?: string;
  hm?: boolean;
  ob?: boolean;
  customStyleOb?: string;
}

export interface LoginResponse {
  sessionId?: string;
  IDPRefreshToken?: string;
}

export interface OperatorSessionResponse {
  features?: string[];
  status?: "ok";
  operator?: boolean;
  permissions?: Record<string, string[]>;
}

export interface TenantStatus {
  /** @format int32 */
  write_quorum?: number;
  /** @format int32 */
  drives_online?: number;
  /** @format int32 */
  drives_offline?: number;
  /** @format int32 */
  drives_healing?: number;
  health_status?: string;
  usage?: {
    /** @format int64 */
    raw?: number;
    /** @format int64 */
    raw_usage?: number;
    /** @format int64 */
    capacity?: number;
    /** @format int64 */
    capacity_usage?: number;
  };
}

export interface TenantConfigurationResponse {
  environmentVariables?: EnvironmentVariable[];
}

export interface UpdateTenantConfigurationRequest {
  keysToBeDeleted?: string[];
  environmentVariables?: EnvironmentVariable[];
}

export interface TenantSecurityResponse {
  autoCert?: boolean;
  customCertificates?: {
    minio?: CertificateInfo[];
    client?: CertificateInfo[];
    minioCAs?: CertificateInfo[];
  };
  securityContext?: SecurityContext;
}

export interface UpdateTenantSecurityRequest {
  autoCert?: boolean;
  customCertificates?: {
    secretsToBeDeleted?: string[];
    minioServerCertificates?: KeyPairConfiguration[];
    minioClientCertificates?: KeyPairConfiguration[];
    minioCAsCertificates?: string[];
  };
  securityContext?: SecurityContext;
}

export interface CertificateInfo {
  serialNumber?: string;
  name?: string;
  domains?: string[];
  expiry?: string;
}

export interface Tenant {
  name?: string;
  creation_date?: string;
  deletion_date?: string;
  currentState?: string;
  pools?: Pool[];
  image?: string;
  namespace?: string;
  /** @format int64 */
  total_size?: number;
  subnet_license?: License;
  endpoints?: {
    minio?: string;
    console?: string;
  };
  idpAdEnabled?: boolean;
  idpOidcEnabled?: boolean;
  encryptionEnabled?: boolean;
  status?: TenantStatus;
  minioTLS?: boolean;
  domains?: DomainsConfiguration;
  tiers?: TenantTierElement[];
}

export interface TenantUsage {
  /** @format int64 */
  used?: number;
  /** @format int64 */
  disk_used?: number;
}

export interface TenantList {
  name?: string;
  pool_count?: number;
  instance_count?: number;
  total_size?: number;
  volume_count?: number;
  creation_date?: string;
  deletion_date?: string;
  currentState?: string;
  namespace?: string;
  health_status?: string;
  /** @format int64 */
  capacity_raw?: number;
  /** @format int64 */
  capacity_raw_usage?: number;
  /** @format int64 */
  capacity?: number;
  /** @format int64 */
  capacity_usage?: number;
  tiers?: TenantTierElement[];
  domains?: DomainsConfiguration;
}

export interface ListTenantsResponse {
  /** list of resulting tenants */
  tenants?: TenantList[];
  /**
   * number of tenants accessible to tenant user
   * @format int64
   */
  total?: number;
}

export interface UpdateTenantRequest {
  /** @pattern ^((.*?)/(.*?):(.+))$ */
  image?: string;
  image_registry?: ImageRegistry;
  image_pull_secret?: string;
}

export interface ImageRegistry {
  registry: string;
  username: string;
  password: string;
}

export interface CsrElements {
  csrElement?: CsrElement[];
}

export interface CsrElement {
  status?: string;
  name?: string;
  generate_name?: string;
  namespace?: string;
  resource_version?: string;
  /** @format int64 */
  generation?: number;
  /** @format int64 */
  deletion_grace_period_seconds?: number;
  annotations?: Annotation[];
}

export interface CreateTenantRequest {
  /** @pattern ^[a-z0-9-]{3,63}$ */
  name: string;
  image?: string;
  pools: Pool[];
  mount_path?: string;
  access_key?: string;
  secret_key?: string;
  /** @default true */
  enable_console?: boolean;
  /** @default true */
  enable_tls?: boolean;
  namespace: string;
  erasureCodingParity?: number;
  annotations?: Record<string, string>;
  labels?: Record<string, string>;
  image_registry?: ImageRegistry;
  image_pull_secret?: string;
  idp?: IdpConfiguration;
  tls?: TlsConfiguration;
  encryption?: EncryptionConfiguration;
  expose_minio?: boolean;
  expose_console?: boolean;
  domains?: DomainsConfiguration;
  environmentVariables?: EnvironmentVariable[];
}

export interface MetadataFields {
  annotations?: Record<string, string>;
  labels?: Record<string, string>;
  node_selector?: Record<string, string>;
}

export interface KeyPairConfiguration {
  crt: string;
  key: string;
}

export interface TlsConfiguration {
  minioServerCertificates?: KeyPairConfiguration[];
  minioClientCertificates?: KeyPairConfiguration[];
  minioCAsCertificates?: string[];
}

export interface SetAdministratorsRequest {
  user_dns?: string[];
  group_dns?: string[];
}

export interface IdpConfiguration {
  oidc?: {
    configuration_url: string;
    client_id: string;
    secret_id: string;
    callback_url?: string;
    claim_name: string;
    scopes?: string;
  };
  keys?: {
    access_key: string;
    secret_key: string;
  }[];
  active_directory?: {
    url: string;
    group_search_base_dn?: string;
    group_search_filter?: string;
    skip_tls_verification?: boolean;
    server_insecure?: boolean;
    server_start_tls?: boolean;
    lookup_bind_dn: string;
    lookup_bind_password?: string;
    user_dn_search_base_dn?: string;
    user_dn_search_filter?: string;
    user_dns?: string[];
  };
}

export type EncryptionConfiguration = MetadataFields & {
  raw?: string;
  image?: string;
  replicas?: string;
  secretsToBeDeleted?: string[];
  server_tls?: KeyPairConfiguration;
  minio_mtls?: KeyPairConfiguration;
  kms_mtls?: {
    key?: string;
    crt?: string;
    ca?: string;
  };
  policies?: object;
  gemalto?: GemaltoConfiguration;
  aws?: AwsConfiguration;
  vault?: VaultConfiguration;
  gcp?: GcpConfiguration;
  azure?: AzureConfiguration;
  securityContext?: SecurityContext;
};

export type EncryptionConfigurationResponse = MetadataFields & {
  raw?: string;
  policies?: object;
  image?: string;
  replicas?: string;
  server_tls?: CertificateInfo;
  minio_mtls?: CertificateInfo;
  kms_mtls?: {
    crt?: CertificateInfo;
    ca?: CertificateInfo;
  };
  gemalto?: GemaltoConfigurationResponse;
  aws?: AwsConfiguration;
  vault?: VaultConfigurationResponse;
  gcp?: GcpConfiguration;
  azure?: AzureConfiguration;
  securityContext?: SecurityContext;
};

export interface VaultConfiguration {
  endpoint: string;
  engine?: string;
  namespace?: string;
  prefix?: string;
  approle: {
    engine?: string;
    id: string;
    secret: string;
    /** @format int64 */
    retry?: number;
  };
  status?: {
    /** @format int64 */
    ping?: number;
  };
}

export interface VaultConfigurationResponse {
  endpoint: string;
  engine?: string;
  namespace?: string;
  prefix?: string;
  approle: {
    engine?: string;
    id: string;
    secret: string;
    /** @format int64 */
    retry?: number;
  };
  status?: {
    /** @format int64 */
    ping?: number;
  };
}

export interface AwsConfiguration {
  secretsmanager: {
    endpoint: string;
    region: string;
    kmskey?: string;
    credentials: {
      accesskey: string;
      secretkey: string;
      token?: string;
    };
  };
}

export interface GemaltoConfiguration {
  keysecure: {
    endpoint: string;
    credentials: {
      token: string;
      domain: string;
      /** @format int64 */
      retry?: number;
    };
  };
}

export interface GemaltoConfigurationResponse {
  keysecure: {
    endpoint: string;
    credentials: {
      token: string;
      domain: string;
      /** @format int64 */
      retry?: number;
    };
  };
}

export interface GcpConfiguration {
  secretmanager: {
    project_id: string;
    endpoint?: string;
    credentials?: {
      client_email?: string;
      client_id?: string;
      private_key_id?: string;
      private_key?: string;
    };
  };
}

export interface AzureConfiguration {
  keyvault: {
    endpoint: string;
    credentials?: {
      tenant_id: string;
      client_id: string;
      client_secret: string;
    };
  };
}

export interface CreateTenantResponse {
  externalIDP?: boolean;
  console?: TenantResponseItem[];
}

export interface TenantResponseItem {
  access_key?: string;
  secret_key?: string;
  url?: string;
}

export interface TenantPod {
  name: string;
  status?: string;
  timeCreated?: number;
  podIP?: string;
  restarts?: number;
  node?: string;
}

export interface Pool {
  name?: string;
  servers: number;
  /** @format int32 */
  volumes_per_server: number;
  volume_configuration: {
    size: number;
    storage_class_name?: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
  };
  /** If provided, use these requests and limit for cpu/memory resource allocation */
  resources?: PoolResources;
  /** NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node's labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ */
  node_selector?: Record<string, string>;
  /** If specified, affinity will define the pod's scheduling constraints */
  affinity?: PoolAffinity;
  runtimeClassName?: string;
  /** Tolerations allows users to set entries like effect, key, operator, value. */
  tolerations?: PoolTolerations;
  securityContext?: SecurityContext;
}

/** Tolerations allows users to set entries like effect, key, operator, value. */
export type PoolTolerations = {
  /** Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute. */
  effect?: string;
  /** Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys. */
  key?: string;
  /** Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category. */
  operator?: string;
  /** TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system. */
  tolerationSeconds?: PoolTolerationSeconds;
  /** Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string. */
  value?: string;
}[];

/** TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system. */
export interface PoolTolerationSeconds {
  /** @format int64 */
  seconds: number;
}

/** If provided, use these requests and limit for cpu/memory resource allocation */
export interface PoolResources {
  /** Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/ */
  limits?: Record<string, number>;
  /** Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/ */
  requests?: Record<string, number>;
}

/** If specified, affinity will define the pod's scheduling constraints */
export interface PoolAffinity {
  /** Describes node affinity scheduling rules for the pod. */
  nodeAffinity?: {
    /** The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred. */
    preferredDuringSchedulingIgnoredDuringExecution?: {
      /** A node selector term, associated with the corresponding weight. */
      preference: NodeSelectorTerm;
      /**
       * Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
       * @format int32
       */
      weight: number;
    }[];
    /** If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node. */
    requiredDuringSchedulingIgnoredDuringExecution?: {
      /** Required. A list of node selector terms. The terms are ORed. */
      nodeSelectorTerms: NodeSelectorTerm[];
    };
  };
  /** Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, pool, etc. as some other pod(s)). */
  podAffinity?: {
    /** The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred. */
    preferredDuringSchedulingIgnoredDuringExecution?: {
      /** Required. A pod affinity term, associated with the corresponding weight. */
      podAffinityTerm: PodAffinityTerm;
      /**
       * weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
       * @format int32
       */
      weight: number;
    }[];
    /** If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied. */
    requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  };
  /** Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, pool, etc. as some other pod(s)). */
  podAntiAffinity?: {
    /** The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred. */
    preferredDuringSchedulingIgnoredDuringExecution?: {
      /** Required. A pod affinity term, associated with the corresponding weight. */
      podAffinityTerm: PodAffinityTerm;
      /**
       * weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
       * @format int32
       */
      weight: number;
    }[];
    /** If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied. */
    requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  };
}

/** A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm. */
export interface NodeSelectorTerm {
  /** A list of node selector requirements by node's labels. */
  matchExpressions?: {
    /** The label key that the selector applies to. */
    key: string;
    /** Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt. */
    operator: string;
    /** An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch. */
    values?: string[];
  }[];
  /** A list of node selector requirements by node's fields. */
  matchFields?: {
    /** The label key that the selector applies to. */
    key: string;
    /** Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt. */
    operator: string;
    /** An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch. */
    values?: string[];
  }[];
}

/** Required. A pod affinity term, associated with the corresponding weight. */
export interface PodAffinityTerm {
  /** A label query over a set of resources, in this case pods. */
  labelSelector?: {
    /** matchExpressions is a list of label selector requirements. The requirements are ANDed. */
    matchExpressions?: {
      /** key is the label key that the selector applies to. */
      key: string;
      /** operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist. */
      operator: string;
      /** values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch. */
      values?: string[];
    }[];
    /** matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed. */
    matchLabels?: Record<string, string>;
  };
  /** namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means "this pod's namespace" */
  namespaces?: string[];
  /** This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed. */
  topologyKey: string;
}

export interface ResourceQuota {
  name?: string;
  elements?: ResourceQuotaElement[];
}

export interface ResourceQuotaElement {
  name?: string;
  /** @format int64 */
  hard?: number;
  /** @format int64 */
  used?: number;
}

export interface DeleteTenantRequest {
  delete_pvcs?: boolean;
}

export interface PoolUpdateRequest {
  pools: Pool[];
}

export interface MaxAllocatableMemResponse {
  /** @format int64 */
  max_memory?: number;
}

export type ParityResponse = string[];

export interface SubscriptionValidateRequest {
  license?: string;
  email?: string;
  password?: string;
}

export interface License {
  email?: string;
  organization?: string;
  account_id?: number;
  storage_capacity?: number;
  plan?: string;
  expires_at?: string;
}

export interface TenantYAML {
  yaml?: string;
}

export interface ListPVCsResponse {
  pvcs?: PvcsListResponse[];
}

export interface PvcsListResponse {
  namespace?: string;
  name?: string;
  status?: string;
  volume?: string;
  tenant?: string;
  capacity?: string;
  storageClass?: string;
  age?: string;
}

export type NodeLabels = Record<string, string[]>;

export interface Namespace {
  name: string;
}

export type EventListWrapper = EventListElement[];

export interface EventListElement {
  namespace?: string;
  /** @format int64 */
  last_seen?: number;
  event_type?: string;
  reason?: string;
  object?: string;
  message?: string;
}

export interface DescribePodWrapper {
  name?: string;
  namespace?: string;
  priority?: number;
  priorityClassName?: string;
  nodeName?: string;
  startTime?: string;
  labels?: Label[];
  annotations?: Annotation[];
  deletionTimestamp?: string;
  deletionGracePeriodSeconds?: number;
  phase?: string;
  reason?: string;
  message?: string;
  podIP?: string;
  controllerRef?: string;
  containers?: Container[];
  conditions?: Condition[];
  volumes?: Volume[];
  qosClass?: string;
  nodeSelector?: NodeSelector[];
  tolerations?: Toleration[];
}

export interface DescribePVCWrapper {
  name?: string;
  namespace?: string;
  storageClass?: string;
  status?: string;
  volume?: string;
  labels?: Label[];
  annotations?: Annotation[];
  finalizers?: string[];
  capacity?: string;
  accessModes?: string[];
  volumeMode?: string;
}

export interface Container {
  name?: string;
  containerID?: string;
  image?: string;
  imageID?: string;
  ports?: string[];
  hostPorts?: string[];
  args?: string[];
  state?: State;
  lastState?: State;
  ready?: boolean;
  restartCount?: number;
  environmentVariables?: EnvironmentVariable[];
  mounts?: Mount[];
}

export interface State {
  state?: string;
  reason?: string;
  message?: string;
  exitCode?: number;
  started?: string;
  finished?: string;
  signal?: number;
}

export interface EnvironmentVariable {
  key?: string;
  value?: string;
}

export interface Mount {
  mountPath?: string;
  name?: string;
  readOnly?: boolean;
  subPath?: string;
}

export interface Condition {
  type?: string;
  status?: string;
}

export interface Volume {
  name?: string;
  pvc?: Pvc;
  projected?: ProjectedVolume;
}

export interface Pvc {
  claimName?: string;
  readOnly?: boolean;
}

export interface ProjectedVolume {
  sources?: ProjectedVolumeSource[];
}

export interface ProjectedVolumeSource {
  secret?: Secret;
  downwardApi?: boolean;
  configMap?: ConfigMap;
  serviceAccountToken?: ServiceAccountToken;
}

export interface Secret {
  name?: string;
  optional?: boolean;
}

export interface ConfigMap {
  name?: string;
  optional?: boolean;
}

export interface ServiceAccountToken {
  expirationSeconds?: number;
}

export interface Toleration {
  tolerationSeconds?: number;
  key?: string;
  value?: string;
  effect?: string;
  operator?: string;
}

export interface Label {
  key?: string;
  value?: string;
}

export interface Annotation {
  key?: string;
  value?: string;
}

export interface NodeSelector {
  key?: string;
  value?: string;
}

export interface SecurityContext {
  runAsUser: string;
  runAsGroup: string;
  runAsNonRoot: boolean;
  fsGroup?: string;
  fsGroupChangePolicy?: string;
}

export interface AllocatableResourcesResponse {
  /** @format int64 */
  min_allocatable_mem?: number;
  /** @format int64 */
  min_allocatable_cpu?: number;
  cpu_priority?: NodeMaxAllocatableResources;
  mem_priority?: NodeMaxAllocatableResources;
}

export interface NodeMaxAllocatableResources {
  /** @format int64 */
  max_allocatable_cpu?: number;
  /** @format int64 */
  max_allocatable_mem?: number;
}

export interface CheckOperatorVersionResponse {
  current_version?: string;
  latest_version?: string;
}

export interface TenantTierElement {
  name?: string;
  type?: string;
  /** @format int64 */
  size?: number;
}

export interface DomainsConfiguration {
  minio?: string[];
  console?: string;
}

export interface UpdateDomainsRequest {
  domains?: DomainsConfiguration;
}

export interface MpIntegration {
  email?: string;
  isInEU?: boolean;
}

export interface OperatorSubnetLoginRequest {
  username?: string;
  password?: string;
}

export interface OperatorSubnetLoginResponse {
  access_token?: string;
  mfa_token?: string;
}

export interface OperatorSubnetLoginMFARequest {
  username: string;
  otp: string;
  mfa_token: string;
}

export interface OperatorSubnetAPIKey {
  apiKey?: string;
}

export interface OperatorSubnetRegisterAPIKeyResponse {
  registered?: boolean;
}

export interface RedirectRule {
  redirect?: string;
  displayName?: string;
}

/** @default "user" */
export enum PolicyEntity {
  User = "user",
  Group = "group",
}

export interface ServerDrives {
  uuid?: string;
  state?: string;
  endpoint?: string;
  drivePath?: string;
  rootDisk?: boolean;
  healing?: boolean;
  model?: string;
  totalSpace?: number;
  usedSpace?: number;
  availableSpace?: number;
}

export interface ServerProperties {
  state?: string;
  endpoint?: string;
  uptime?: number;
  version?: string;
  commitID?: string;
  poolNumber?: number;
  network?: Record<string, string>;
  drives?: ServerDrives[];
}

export interface BackendProperties {
  backendType?: string;
  rrSCParity?: number;
  standardSCParity?: number;
}

export interface TenantLogReport {
  filename?: string;
  blob?: string;
}

export type QueryParamsType = Record<string | number, any>;
export type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;

export interface FullRequestParams extends Omit<RequestInit, "body"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseFormat;
  /** request body */
  body?: unknown;
  /** base url */
  baseUrl?: string;
  /** request cancellation token */
  cancelToken?: CancelToken;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown> {
  baseUrl?: string;
  baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
  securityWorker?: (
    securityData: SecurityDataType | null
  ) => Promise<RequestParams | void> | RequestParams | void;
  customFetch?: typeof fetch;
}

export interface HttpResponse<D extends unknown, E extends unknown = unknown>
  extends Response {
  data: D;
  error: E;
}

type CancelToken = Symbol | string | number;

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public baseUrl: string = "/api/v1";
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private abortControllers = new Map<CancelToken, AbortController>();
  private customFetch = (...fetchParams: Parameters<typeof fetch>) =>
    fetch(...fetchParams);

  private baseApiParams: RequestParams = {
    credentials: "same-origin",
    headers: {},
    redirect: "follow",
    referrerPolicy: "no-referrer",
  };

  constructor(apiConfig: ApiConfig<SecurityDataType> = {}) {
    Object.assign(this, apiConfig);
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected encodeQueryParam(key: string, value: any) {
    const encodedKey = encodeURIComponent(key);
    return `${encodedKey}=${encodeURIComponent(
      typeof value === "number" ? value : `${value}`
    )}`;
  }

  protected addQueryParam(query: QueryParamsType, key: string) {
    return this.encodeQueryParam(key, query[key]);
  }

  protected addArrayQueryParam(query: QueryParamsType, key: string) {
    const value = query[key];
    return value.map((v: any) => this.encodeQueryParam(key, v)).join("&");
  }

  protected toQueryString(rawQuery?: QueryParamsType): string {
    const query = rawQuery || {};
    const keys = Object.keys(query).filter(
      (key) => "undefined" !== typeof query[key]
    );
    return keys
      .map((key) =>
        Array.isArray(query[key])
          ? this.addArrayQueryParam(query, key)
          : this.addQueryParam(query, key)
      )
      .join("&");
  }

  protected addQueryParams(rawQuery?: QueryParamsType): string {
    const queryString = this.toQueryString(rawQuery);
    return queryString ? `?${queryString}` : "";
  }

  private contentFormatters: Record<ContentType, (input: any) => any> = {
    [ContentType.Json]: (input: any) =>
      input !== null && (typeof input === "object" || typeof input === "string")
        ? JSON.stringify(input)
        : input,
    [ContentType.Text]: (input: any) =>
      input !== null && typeof input !== "string"
        ? JSON.stringify(input)
        : input,
    [ContentType.FormData]: (input: any) =>
      Object.keys(input || {}).reduce((formData, key) => {
        const property = input[key];
        formData.append(
          key,
          property instanceof Blob
            ? property
            : typeof property === "object" && property !== null
            ? JSON.stringify(property)
            : `${property}`
        );
        return formData;
      }, new FormData()),
    [ContentType.UrlEncoded]: (input: any) => this.toQueryString(input),
  };

  protected mergeRequestParams(
    params1: RequestParams,
    params2?: RequestParams
  ): RequestParams {
    return {
      ...this.baseApiParams,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...(this.baseApiParams.headers || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected createAbortSignal = (
    cancelToken: CancelToken
  ): AbortSignal | undefined => {
    if (this.abortControllers.has(cancelToken)) {
      const abortController = this.abortControllers.get(cancelToken);
      if (abortController) {
        return abortController.signal;
      }
      return void 0;
    }

    const abortController = new AbortController();
    this.abortControllers.set(cancelToken, abortController);
    return abortController.signal;
  };

  public abortRequest = (cancelToken: CancelToken) => {
    const abortController = this.abortControllers.get(cancelToken);

    if (abortController) {
      abortController.abort();
      this.abortControllers.delete(cancelToken);
    }
  };

  public request = async <T = any, E = any>({
    body,
    secure,
    path,
    type,
    query,
    format,
    baseUrl,
    cancelToken,
    ...params
  }: FullRequestParams): Promise<HttpResponse<T, E>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.baseApiParams.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const queryString = query && this.toQueryString(query);
    const payloadFormatter = this.contentFormatters[type || ContentType.Json];
    const responseFormat = format || requestParams.format;

    return this.customFetch(
      `${baseUrl || this.baseUrl || ""}${path}${
        queryString ? `?${queryString}` : ""
      }`,
      {
        ...requestParams,
        headers: {
          ...(requestParams.headers || {}),
          ...(type && type !== ContentType.FormData
            ? { "Content-Type": type }
            : {}),
        },
        signal: cancelToken
          ? this.createAbortSignal(cancelToken)
          : requestParams.signal,
        body:
          typeof body === "undefined" || body === null
            ? null
            : payloadFormatter(body),
      }
    ).then(async (response) => {
      const r = response as HttpResponse<T, E>;
      r.data = null as unknown as T;
      r.error = null as unknown as E;

      const data = !responseFormat
        ? r
        : await response[responseFormat]()
            .then((data) => {
              if (r.ok) {
                r.data = data;
              } else {
                r.error = data;
              }
              return r;
            })
            .catch((e) => {
              r.error = e;
              return r;
            });

      if (cancelToken) {
        this.abortControllers.delete(cancelToken);
      }

      if (!response.ok) throw data;
      return data;
    });
  };
}

/**
 * @title MinIO Console Server
 * @version 0.1.0
 * @baseUrl /api/v1
 */
export class Api<
  SecurityDataType extends unknown
> extends HttpClient<SecurityDataType> {
  login = {
    /**
     * No description
     *
     * @tags Auth
     * @name LoginDetail
     * @summary Returns login strategy, form or sso.
     * @request GET:/login
     */
    loginDetail: (params: RequestParams = {}) =>
      this.request<LoginDetails, Error>({
        path: `/login`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Auth
     * @name LoginOperator
     * @summary Login to Operator Console.
     * @request POST:/login/operator
     */
    loginOperator: (body: LoginOperatorRequest, params: RequestParams = {}) =>
      this.request<void, Error>({
        path: `/login/operator`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags Auth
     * @name LoginOauth2Auth
     * @summary Identity Provider oauth2 callback endpoint.
     * @request POST:/login/oauth2/auth
     */
    loginOauth2Auth: (
      body: LoginOauth2AuthRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/login/oauth2/auth`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),
  };
  logout = {
    /**
     * No description
     *
     * @tags Auth
     * @name Logout
     * @summary Logout from Operator.
     * @request POST:/logout
     * @secure
     */
    logout: (params: RequestParams = {}) =>
      this.request<void, Error>({
        path: `/logout`,
        method: "POST",
        secure: true,
        ...params,
      }),
  };
  session = {
    /**
     * No description
     *
     * @tags Auth
     * @name SessionCheck
     * @summary Endpoint to check if your session is still valid
     * @request GET:/session
     * @secure
     */
    sessionCheck: (params: RequestParams = {}) =>
      this.request<OperatorSessionResponse, Error>({
        path: `/session`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  checkVersion = {
    /**
     * No description
     *
     * @tags UserAPI
     * @name CheckMinIoVersion
     * @summary Checks the current Operator version against the latest
     * @request GET:/check-version
     */
    checkMinIoVersion: (params: RequestParams = {}) =>
      this.request<CheckOperatorVersionResponse, Error>({
        path: `/check-version`,
        method: "GET",
        format: "json",
        ...params,
      }),
  };
  subscription = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name SubscriptionInfo
     * @summary Subscription info
     * @request GET:/subscription/info
     * @secure
     */
    subscriptionInfo: (params: RequestParams = {}) =>
      this.request<License, Error>({
        path: `/subscription/info`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name SubscriptionValidate
     * @summary Validates subscription license
     * @request POST:/subscription/validate
     * @secure
     */
    subscriptionValidate: (
      body: SubscriptionValidateRequest,
      params: RequestParams = {}
    ) =>
      this.request<License, Error>({
        path: `/subscription/validate`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name SubscriptionRefresh
     * @summary Refresh existing subscription license
     * @request POST:/subscription/refresh
     * @secure
     */
    subscriptionRefresh: (params: RequestParams = {}) =>
      this.request<License, Error>({
        path: `/subscription/refresh`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name SubscriptionActivate
     * @summary Activate a particular tenant using the existing subscription license
     * @request POST:/subscription/namespaces/{namespace}/tenants/{tenant}/activate
     * @secure
     */
    subscriptionActivate: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/subscription/namespaces/${namespace}/tenants/${tenant}/activate`,
        method: "POST",
        secure: true,
        ...params,
      }),
  };
  tenants = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListAllTenants
     * @summary List Tenant of All Namespaces
     * @request GET:/tenants
     * @secure
     */
    listAllTenants: (
      query?: {
        sort_by?: string;
        /** @format int32 */
        offset?: number;
        /** @format int32 */
        limit?: number;
      },
      params: RequestParams = {}
    ) =>
      this.request<ListTenantsResponse, Error>({
        path: `/tenants`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name CreateTenant
     * @summary Create Tenant
     * @request POST:/tenants
     * @secure
     */
    createTenant: (body: CreateTenantRequest, params: RequestParams = {}) =>
      this.request<CreateTenantResponse, Error>({
        path: `/tenants`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  namespace = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name CreateNamespace
     * @summary Creates a new Namespace with given information
     * @request POST:/namespace
     * @secure
     */
    createNamespace: (body: Namespace, params: RequestParams = {}) =>
      this.request<void, Error>({
        path: `/namespace`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),
  };
  namespaces = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListTenants
     * @summary List Tenants by Namespace
     * @request GET:/namespaces/{namespace}/tenants
     * @secure
     */
    listTenants: (
      namespace: string,
      query?: {
        sort_by?: string;
        /** @format int32 */
        offset?: number;
        /** @format int32 */
        limit?: number;
      },
      params: RequestParams = {}
    ) =>
      this.request<ListTenantsResponse, Error>({
        path: `/namespaces/${namespace}/tenants`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListTenantCertificateSigningRequest
     * @summary List Tenant Certificate Signing Request
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/csr
     * @secure
     */
    listTenantCertificateSigningRequest: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<CsrElements, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/csr`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantIdentityProvider
     * @summary Tenant Identity Provider
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/identity-provider
     * @secure
     */
    tenantIdentityProvider: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<IdpConfiguration, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/identity-provider`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name UpdateTenantIdentityProvider
     * @summary Update Tenant Identity Provider
     * @request POST:/namespaces/{namespace}/tenants/{tenant}/identity-provider
     * @secure
     */
    updateTenantIdentityProvider: (
      namespace: string,
      tenant: string,
      body: IdpConfiguration,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/identity-provider`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name SetTenantAdministrators
     * @summary Set the consoleAdmin policy to the specified users and groups
     * @request POST:/namespaces/{namespace}/tenants/{tenant}/set-administrators
     * @secure
     */
    setTenantAdministrators: (
      namespace: string,
      tenant: string,
      body: SetAdministratorsRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/set-administrators`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantConfiguration
     * @summary Tenant Configuration
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/configuration
     * @secure
     */
    tenantConfiguration: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantConfigurationResponse, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/configuration`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name UpdateTenantConfiguration
     * @summary Update Tenant Configuration
     * @request PATCH:/namespaces/{namespace}/tenants/{tenant}/configuration
     * @secure
     */
    updateTenantConfiguration: (
      namespace: string,
      tenant: string,
      body: UpdateTenantConfigurationRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/configuration`,
        method: "PATCH",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantSecurity
     * @summary Tenant Security
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/security
     * @secure
     */
    tenantSecurity: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantSecurityResponse, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/security`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name UpdateTenantSecurity
     * @summary Update Tenant Security
     * @request POST:/namespaces/{namespace}/tenants/{tenant}/security
     * @secure
     */
    updateTenantSecurity: (
      namespace: string,
      tenant: string,
      body: UpdateTenantSecurityRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/security`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantDetails
     * @summary Tenant Details
     * @request GET:/namespaces/{namespace}/tenants/{tenant}
     * @secure
     */
    tenantDetails: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<Tenant, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name DeleteTenant
     * @summary Delete tenant and underlying pvcs
     * @request DELETE:/namespaces/{namespace}/tenants/{tenant}
     * @secure
     */
    deleteTenant: (
      namespace: string,
      tenant: string,
      body: DeleteTenantRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}`,
        method: "DELETE",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name UpdateTenant
     * @summary Update Tenant
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}
     * @secure
     */
    updateTenant: (
      namespace: string,
      tenant: string,
      body: UpdateTenantRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantAddPool
     * @summary Tenant Add Pool
     * @request POST:/namespaces/{namespace}/tenants/{tenant}/pools
     * @secure
     */
    tenantAddPool: (
      namespace: string,
      tenant: string,
      body: Pool,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pools`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantUpdatePools
     * @summary Tenant Update Pools
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}/pools
     * @secure
     */
    tenantUpdatePools: (
      namespace: string,
      tenant: string,
      body: PoolUpdateRequest,
      params: RequestParams = {}
    ) =>
      this.request<Tenant, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pools`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListPvCsForTenant
     * @summary List all PVCs from given Tenant
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pvcs
     * @secure
     */
    listPvCsForTenant: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<ListPVCsResponse, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pvcs`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetTenantUsage
     * @summary Get Usage For The Tenant
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/usage
     * @secure
     */
    getTenantUsage: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantUsage, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/usage`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetTenantPods
     * @summary Get Pods For The Tenant
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pods
     * @secure
     */
    getTenantPods: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantPod[], Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pods`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetTenantEvents
     * @summary Get Events for given Tenant
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/events
     * @secure
     */
    getTenantEvents: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<EventListWrapper, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/events`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetTenantLogReport
     * @summary Get Tenant Log Report
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/log-report
     * @secure
     */
    getTenantLogReport: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantLogReport, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/log-report`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetPodLogs
     * @summary Get Logs for Pod
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pods/{podName}
     * @secure
     */
    getPodLogs: (
      namespace: string,
      tenant: string,
      podName: string,
      params: RequestParams = {}
    ) =>
      this.request<string, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pods/${podName}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name DeletePod
     * @summary Delete pod
     * @request DELETE:/namespaces/{namespace}/tenants/{tenant}/pods/{podName}
     * @secure
     */
    deletePod: (
      namespace: string,
      tenant: string,
      podName: string,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pods/${podName}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetPodEvents
     * @summary Get Events for Pod
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pods/{podName}/events
     * @secure
     */
    getPodEvents: (
      namespace: string,
      tenant: string,
      podName: string,
      params: RequestParams = {}
    ) =>
      this.request<EventListWrapper, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pods/${podName}/events`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name DescribePod
     * @summary Describe Pod
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pods/{podName}/describe
     * @secure
     */
    describePod: (
      namespace: string,
      tenant: string,
      podName: string,
      params: RequestParams = {}
    ) =>
      this.request<DescribePodWrapper, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pods/${podName}/describe`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantUpdateCertificate
     * @summary Tenant Update Certificates
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}/certificates
     * @secure
     */
    tenantUpdateCertificate: (
      namespace: string,
      tenant: string,
      body: TlsConfiguration,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/certificates`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantDeleteEncryption
     * @summary Tenant Delete Encryption
     * @request DELETE:/namespaces/{namespace}/tenants/{tenant}/encryption
     * @secure
     */
    tenantDeleteEncryption: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/encryption`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantUpdateEncryption
     * @summary Tenant Update Encryption
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}/encryption
     * @secure
     */
    tenantUpdateEncryption: (
      namespace: string,
      tenant: string,
      body: EncryptionConfiguration,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/encryption`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name TenantEncryptionInfo
     * @summary Tenant Encryption Info
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/encryption
     * @secure
     */
    tenantEncryptionInfo: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<EncryptionConfigurationResponse, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/encryption`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetTenantYaml
     * @summary Get the Tenant YAML
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/yaml
     * @secure
     */
    getTenantYaml: (
      namespace: string,
      tenant: string,
      params: RequestParams = {}
    ) =>
      this.request<TenantYAML, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/yaml`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name PutTenantYaml
     * @summary Put the Tenant YAML
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}/yaml
     * @secure
     */
    putTenantYaml: (
      namespace: string,
      tenant: string,
      body: TenantYAML,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/yaml`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name UpdateTenantDomains
     * @summary Update Domains for a Tenant
     * @request PUT:/namespaces/{namespace}/tenants/{tenant}/domains
     * @secure
     */
    updateTenantDomains: (
      namespace: string,
      tenant: string,
      body: UpdateDomainsRequest,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/domains`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetResourceQuota
     * @summary Get Resource Quota
     * @request GET:/namespaces/{namespace}/resourcequotas/{resource-quota-name}
     * @secure
     */
    getResourceQuota: (
      namespace: string,
      resourceQuotaName: string,
      params: RequestParams = {}
    ) =>
      this.request<ResourceQuota, Error>({
        path: `/namespaces/${namespace}/resourcequotas/${resourceQuotaName}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name DeletePvc
     * @summary Delete PVC
     * @request DELETE:/namespaces/{namespace}/tenants/{tenant}/pvc/{PVCName}
     * @secure
     */
    deletePvc: (
      namespace: string,
      tenant: string,
      pvcName: string,
      params: RequestParams = {}
    ) =>
      this.request<void, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pvc/${pvcName}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetPvcEvents
     * @summary Get Events for PVC
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pvcs/{PVCName}/events
     * @secure
     */
    getPvcEvents: (
      namespace: string,
      tenant: string,
      pvcName: string,
      params: RequestParams = {}
    ) =>
      this.request<EventListWrapper, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pvcs/${pvcName}/events`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetPvcDescribe
     * @summary Get Describe output for PVC
     * @request GET:/namespaces/{namespace}/tenants/{tenant}/pvcs/{PVCName}/describe
     * @secure
     */
    getPvcDescribe: (
      namespace: string,
      tenant: string,
      pvcName: string,
      params: RequestParams = {}
    ) =>
      this.request<DescribePVCWrapper, Error>({
        path: `/namespaces/${namespace}/tenants/${tenant}/pvcs/${pvcName}/describe`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  cluster = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetMaxAllocatableMem
     * @summary Get maximum allocatable memory for given number of nodes
     * @request GET:/cluster/max-allocatable-memory
     * @secure
     */
    getMaxAllocatableMem: (
      query: {
        /**
         * @format int32
         * @min 1
         */
        num_nodes: number;
      },
      params: RequestParams = {}
    ) =>
      this.request<MaxAllocatableMemResponse, Error>({
        path: `/cluster/max-allocatable-memory`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetAllocatableResources
     * @summary Get allocatable cpu and memory for given number of nodes
     * @request GET:/cluster/allocatable-resources
     * @secure
     */
    getAllocatableResources: (
      query: {
        /**
         * @format int32
         * @min 1
         */
        num_nodes: number;
      },
      params: RequestParams = {}
    ) =>
      this.request<AllocatableResourcesResponse, Error>({
        path: `/cluster/allocatable-resources`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),
  };
  getParity = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetParity
     * @summary Gets parity by sending number of nodes & number of disks
     * @request GET:/get-parity/{nodes}/{disksPerNode}
     * @secure
     */
    getParity: (
      nodes: number,
      disksPerNode: number,
      params: RequestParams = {}
    ) =>
      this.request<ParityResponse, Error>({
        path: `/get-parity/${nodes}/${disksPerNode}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  listPvcs = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListPvCs
     * @summary List all PVCs from namespaces that the user has access to
     * @request GET:/list-pvcs
     * @secure
     */
    listPvCs: (params: RequestParams = {}) =>
      this.request<ListPVCsResponse, Error>({
        path: `/list-pvcs`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  mpIntegration = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name GetMpIntegration
     * @summary Returns email registered for marketplace integration
     * @request GET:/mp-integration
     * @secure
     */
    getMpIntegration: (params: RequestParams = {}) =>
      this.request<
        {
          isEmailSet?: boolean;
        },
        Error
      >({
        path: `/mp-integration`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name PostMpIntegration
     * @summary Set email to register for marketplace integration
     * @request POST:/mp-integration
     * @secure
     */
    postMpIntegration: (body: MpIntegration, params: RequestParams = {}) =>
      this.request<void, Error>({
        path: `/mp-integration`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        ...params,
      }),
  };
  nodes = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name ListNodeLabels
     * @summary List node labels
     * @request GET:/nodes/labels
     * @secure
     */
    listNodeLabels: (params: RequestParams = {}) =>
      this.request<NodeLabels, Error>({
        path: `/nodes/labels`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  subnet = {
    /**
     * No description
     *
     * @tags OperatorAPI
     * @name OperatorSubnetLogin
     * @summary Login to subnet
     * @request POST:/subnet/login
     * @secure
     */
    operatorSubnetLogin: (
      body: OperatorSubnetLoginRequest,
      params: RequestParams = {}
    ) =>
      this.request<OperatorSubnetLoginResponse, Error>({
        path: `/subnet/login`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name OperatorSubnetLoginMfa
     * @summary Login to subnet using mfa
     * @request POST:/subnet/login/mfa
     * @secure
     */
    operatorSubnetLoginMfa: (
      body: OperatorSubnetLoginMFARequest,
      params: RequestParams = {}
    ) =>
      this.request<OperatorSubnetLoginResponse, Error>({
        path: `/subnet/login/mfa`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name OperatorSubnetApiKey
     * @summary Subnet api key
     * @request GET:/subnet/apikey
     * @secure
     */
    operatorSubnetApiKey: (
      query: {
        token: string;
      },
      params: RequestParams = {}
    ) =>
      this.request<OperatorSubnetAPIKey, Error>({
        path: `/subnet/apikey`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name OperatorSubnetRegisterApiKey
     * @summary Register Operator with Subnet
     * @request POST:/subnet/apikey/register
     * @secure
     */
    operatorSubnetRegisterApiKey: (
      body: OperatorSubnetAPIKey,
      params: RequestParams = {}
    ) =>
      this.request<OperatorSubnetRegisterAPIKeyResponse, Error>({
        path: `/subnet/apikey/register`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OperatorAPI
     * @name OperatorSubnetApiKeyInfo
     * @summary Subnet API key info
     * @request GET:/subnet/apikey/info
     * @secure
     */
    operatorSubnetApiKeyInfo: (params: RequestParams = {}) =>
      this.request<OperatorSubnetRegisterAPIKeyResponse, Error>({
        path: `/subnet/apikey/info`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
