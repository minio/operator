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

export const IAM_ROLES = {
  BUCKET_OWNER: "BUCKET_OWNER", // upload/delete objects from the bucket
  BUCKET_VIEWER: "BUCKET_VIEWER", // only view objects on the bucket
  BUCKET_ADMIN: "BUCKET_ADMIN", // administrate the bucket
  BUCKET_LIFECYCLE: "BUCKET_LIFECYCLE", // can manage bucket lifecycle
};

export const IAM_SCOPES = {
  S3_STAR_BUCKET: "s3:*Bucket",
  S3_LIST_BUCKET: "s3:ListBucket",
  S3_ALL_LIST_BUCKET: "s3:List*",
  S3_GET_BUCKET_POLICY: "s3:GetBucketPolicy",
  S3_PUT_BUCKET_POLICY: "s3:PutBucketPolicy",
  S3_GET_OBJECT: "s3:GetObject",
  S3_PUT_OBJECT: "s3:PutObject",
  S3_GET_ACTIONS: "s3:Get*",
  S3_PUT_ACTIONS: "s3:Put*",
  S3_GET_OBJECT_LEGAL_HOLD: "s3:GetObjectLegalHold",
  S3_PUT_OBJECT_LEGAL_HOLD: "s3:PutObjectLegalHold",
  S3_DELETE_OBJECT: "s3:DeleteObject",
  S3_GET_BUCKET_VERSIONING: "s3:GetBucketVersioning",
  S3_PUT_BUCKET_VERSIONING: "s3:PutBucketVersioning",
  S3_GET_OBJECT_RETENTION: "s3:GetObjectRetention",
  S3_PUT_OBJECT_RETENTION: "s3:PutObjectRetention",
  S3_GET_OBJECT_TAGGING: "s3:GetObjectTagging",
  S3_PUT_OBJECT_TAGGING: "s3:PutObjectTagging",
  S3_DELETE_OBJECT_TAGGING: "s3:DeleteObjectTagging",
  S3_GET_BUCKET_ENCRYPTION_CONFIGURATION: "s3:GetEncryptionConfiguration",
  S3_PUT_BUCKET_ENCRYPTION_CONFIGURATION: "s3:PutEncryptionConfiguration",
  S3_CREATE_BUCKET: "s3:CreateBucket",
  S3_DELETE_BUCKET: "s3:DeleteBucket",
  S3_FORCE_DELETE_BUCKET: "s3:ForceDeleteBucket",
  S3_GET_BUCKET_NOTIFICATIONS: "s3:GetBucketNotification",
  S3_LISTEN_BUCKET_NOTIFICATIONS: "s3:ListenBucketNotification",
  S3_PUT_BUCKET_NOTIFICATIONS: "s3:PutBucketNotification",
  S3_GET_REPLICATION_CONFIGURATION: "s3:GetReplicationConfiguration",
  S3_PUT_REPLICATION_CONFIGURATION: "s3:PutReplicationConfiguration",
  S3_GET_LIFECYCLE_CONFIGURATION: "s3:GetLifecycleConfiguration",
  S3_PUT_LIFECYCLE_CONFIGURATION: "s3:PutLifecycleConfiguration",
  S3_GET_BUCKET_OBJECT_LOCK_CONFIGURATION:
    "s3:GetBucketObjectLockConfiguration",
  S3_PUT_BUCKET_OBJECT_LOCK_CONFIGURATION:
    "s3:PutBucketObjectLockConfiguration",
  ADMIN_GET_POLICY: "admin:GetPolicy",
  ADMIN_LIST_USERS: "admin:ListUsers",
  ADMIN_CREATE_USER: "admin:CreateUser",
  ADMIN_DELETE_USER: "admin:DeleteUser",
  ADMIN_ENABLE_USER: "admin:EnableUser",
  ADMIN_DISABLE_USER: "admin:DisableUser",
  ADMIN_GET_USER: "admin:GetUser",
  ADMIN_LIST_USER_POLICIES: "admin:ListUserPolicies",
  ADMIN_SERVER_INFO: "admin:ServerInfo",
  ADMIN_GET_BUCKET_QUOTA: "admin:GetBucketQuota",
  ADMIN_SET_BUCKET_QUOTA: "admin:SetBucketQuota",
  ADMIN_LIST_TIERS: "admin:ListTier",
  ADMIN_SET_TIER: "admin:SetTier",
  ADMIN_LIST_GROUPS: "admin:ListGroups",
  S3_GET_OBJECT_VERSION_FOR_REPLICATION: "s3:GetObjectVersionForReplication",
  S3_REPLICATE_TAGS: "s3:ReplicateTags",
  S3_REPLICATE_DELETE: "s3:ReplicateDelete",
  S3_REPLICATE_OBJECT: "s3:ReplicateObject",
  S3_PUT_OBJECT_VERSION_TAGGING: "s3:PutObjectVersionTagging",
  S3_DELETE_OBJECT_VERSION_TAGGING: "s3:DeleteObjectVersionTagging",
  S3_DELETE_OBJECT_VERSION: "s3:DeleteObjectVersion",
  S3_GET_OBJECT_VERSION_TAGGING: "s3:GetObjectVersionTagging",
  S3_GET_OBJECT_VERSION: "s3:GetObjectVersion",
  S3_PUT_BUCKET_TAGGING: "s3:PutBucketTagging",
  S3_GET_BUCKET_TAGGING: "s3:GetBucketTagging",
  S3_BYPASS_GOVERNANCE_RETENTION: "s3:BypassGovernanceRetention",
  S3_LIST_MULTIPART_UPLOAD_PARTS: "s3:ListMultipartUploadParts",
  S3_LISTEN_NOTIFICATIONS: "s3:ListenNotification",
  S3_LIST_BUCKET_MULTIPART_UPLOADS: "s3:ListBucketMultipartUploads",
  S3_LIST_BUCKET_VERSIONS: "s3:ListBucketVersions",
  S3_GET_BUCKET_POLICY_STATUS: "s3:GetBucketPolicyStatus",
  S3_LIST_ALL_MY_BUCKETS: "s3:ListAllMyBuckets",
  S3_HEAD_BUCKET: "s3:HeadBucket",
  S3_GET_BUCKET_LOCATION: "s3:GetBucketLocation",
  S3_DELETE_BUCKET_POLICY: "s3:DeleteBucketPolicy",
  S3_ABORT_MULTIPART_UPLOAD: "s3:AbortMultipartUpload",
  ADMIN_ADD_USER_TO_GROUP: "admin:AddUserToGroup",
  ADMIN_REMOVE_USER_FROM_GROUP: "admin:RemoveUserFromGroup",
  ADMIN_GET_GROUP: "admin:GetGroup",
  ADMIN_ENABLE_GROUP: "admin:EnableGroup",
  ADMIN_DISABLE_GROUP: "admin:DisableGroup",
  ADMIN_CREATE_POLICY: "admin:CreatePolicy",
  ADMIN_DELETE_POLICY: "admin:DeletePolicy",
  ADMIN_ATTACH_USER_OR_GROUP_POLICY: "admin:AttachUserOrGroupPolicy",
  ADMIN_CREATE_SERVICEACCOUNT: "admin:CreateServiceAccount",
  ADMIN_UPDATE_SERVICEACCOUNT: "admin:UpdateServiceAccount",
  ADMIN_REMOVE_SERVICEACCOUNT: "admin:RemoveServiceAccount",
  ADMIN_LIST_SERVICEACCOUNTS: "admin:ListServiceAccounts",
  ADMIN_CONFIG_UPDATE: "admin:ConfigUpdate",
  ADMIN_GET_CONSOLE_LOG: "admin:ConsoleLog",
  ADMIN_SERVER_TRACE: "admin:ServerTrace",
  ADMIN_HEALTH_INFO: "admin:OBDInfo",
  ADMIN_HEAL: "admin:Heal",
  ADMIN_INSPECT_DATA: "admin:InspectData",
  S3_ALL_ACTIONS: "s3:*",
  ADMIN_ALL_ACTIONS: "admin:*",
  KMS_ALL_ACTIONS: "kms:*",
  KMS_STATUS: "kms:Status",
  KMS_METRICS: "kms:Metrics",
  KMS_APIS: "kms:API",
  KMS_Version: "kms:Version",
  KMS_CREATE_KEY: "kms:CreateKey",
  KMS_DELETE_KEY: "kms:DeleteKey",
  KMS_LIST_KEYS: "kms:ListKeys",
  KMS_IMPORT_KEY: "kms:ImportKey",
  KMS_KEY_STATUS: "kms:KeyStatus",
  KMS_DESCRIBE_POLICY: "kms:DescribePolicy",
  KMS_ASSIGN_POLICY: "kms:AssignPolicy",
  KMS_DELETE_POLICY: "kms:DeletePolicy",
  KMS_SET_POLICY: "kms:SetPolicy",
  KMS_GET_POLICY: "kms:GetPolicy",
  KMS_LIST_POLICIES: "kms:ListPolicies",
  KMS_DESCRIBE_IDENTITY: "kms:DescribeIdentity",
  KMS_DESCRIBE_SELF_IDENTITY: "kms:DescribeSelfIdentity",
  KMS_DELETE_IDENTITY: "kms:DeleteIdentity",
  KMS_LIST_IDENTITIES: "kms:ListIdentities",
};

export const IAM_PAGES = {
  /* Buckets */
  BUCKETS: "/buckets",
  ADD_BUCKETS: "add-bucket",
  BUCKETS_ADMIN_VIEW: ":bucketName/admin/*",
  /* Object Browser */
  OBJECT_BROWSER_VIEW: "/browser",
  OBJECT_BROWSER_BUCKET_VIEW: "/browser/:bucketName",
  OBJECT_BROWSER_BUCKET_DETAILS_VIEW: "/browser/:bucketName/*",
  /* Identity */
  IDENTITY: "/identity",
  USERS: "/identity/users",
  USERS_VIEW: "/identity/users/:userName",
  USER_ADD: "/identity/users/add-user",
  GROUPS: "/identity/groups",
  GROUPS_ADD: "/identity/groups/create-group",
  GROUPS_VIEW: "/identity/groups/:groupName",
  ACCOUNT: "/access-keys",
  ACCOUNT_ADD: "/access-keys/new-account",
  USER_SA_ACCOUNT_ADD: "/identity/users/new-user-sa/:userName",

  /* IDP */
  IDP_LDAP_CONFIGURATIONS: "/identity/idp/ldap/configurations",
  IDP_LDAP_CONFIGURATIONS_VIEW: "/identity/idp/ldap/configurations/:idpName",
  IDP_LDAP_CONFIGURATIONS_ADD: "/identity/idp/ldap/configurations/add-idp",

  IDP_OPENID_CONFIGURATIONS: "/identity/idp/openid/configurations",
  IDP_OPENID_CONFIGURATIONS_VIEW:
    "/identity/idp/openid/configurations/:idpName",
  IDP_OPENID_CONFIGURATIONS_ADD: "/identity/idp/openid/configurations/add-idp",

  POLICIES: "/policies",
  POLICY_ADD: "/add-policy",
  POLICIES_VIEW: "/policies/*",
  /* Monitoring */
  TOOLS_LOGS: "/tools/logs",
  TOOLS_AUDITLOGS: "/tools/audit-logs",
  TOOLS_TRACE: "/tools/trace",
  DASHBOARD: "/tools/metrics",
  TOOLS_HEAL: "/tools/heal",
  TOOLS_WATCH: "/tools/watch",

  /* KMS */
  KMS: "/kms",
  KMS_STATUS: "/kms/status",
  KMS_KEYS: "/kms/keys",
  KMS_KEYS_ADD: "/kms/add-key/",
  KMS_KEYS_IMPORT: "/kms/import-key/",

  /* Support */
  TOOLS: "/support",
  REGISTER_SUPPORT: "/support/register",
  TOOLS_DIAGNOSTICS: "/support/diagnostics",
  TOOLS_SPEEDTEST: "/support/speedtest",
  CALL_HOME: "/support/call-home",
  PROFILE: "/support/profile",
  SUPPORT_INSPECT: "/support/inspect",

  /** License **/
  LICENSE: "/license",
  /* Settings **/
  SETTINGS: "/settings/configurations",
  SETTINGS_VIEW: "/settings/configurations/:option",
  /* Documentation **/
  DOCUMENTATION: "/documentation",
  /* TBD ? */
  EVENT_DESTINATIONS: "/settings/event-destinations",
  EVENT_DESTINATIONS_ADD: "/settings/event-destinations/add",
  EVENT_DESTINATIONS_ADD_SERVICE: "/settings/event-destinations/add/:service",
  TIERS: "/settings/tiers",
  TIERS_ADD: "/settings/tiers/add",
  TIERS_ADD_SERVICE: "/settings/tiers/add/:service",
  SITE_REPLICATION: "/settings/site-replication",
  SITE_REPLICATION_STATUS: "/settings/site-replication/status",
  SITE_REPLICATION_ADD: "/settings/site-replication/add",

  /* Operator */
  TENANTS: "/tenants",
  TENANTS_ADD: "/tenants/add",
  NAMESPACE_TENANT: "/namespaces/:tenantNamespace/tenants/:tenantName",
  NAMESPACE_TENANT_HOP: "/namespaces/:tenantNamespace/tenants/:tenantName/hop",
  NAMESPACE_TENANT_PODS:
    "/namespaces/:tenantNamespace/tenants/:tenantName/pods/:podName",
  NAMESPACE_TENANT_PVCS:
    "/namespaces/:tenantNamespace/tenants/:tenantName/pvcs/:PVCName",
  NAMESPACE_TENANT_PODS_LIST:
    "/namespaces/:tenantNamespace/tenants/:tenantName/pods",
  NAMESPACE_TENANT_SUMMARY:
    "/namespaces/:tenantNamespace/tenants/:tenantName/summary",
  NAMESPACE_TENANT_METRICS:
    "/namespaces/:tenantNamespace/tenants/:tenantName/metrics",
  NAMESPACE_TENANT_TRACE:
    "/namespaces/:tenantNamespace/tenants/:tenantName/trace",
  NAMESPACE_TENANT_POOLS:
    "/namespaces/:tenantNamespace/tenants/:tenantName/pools",
  NAMESPACE_TENANT_POOLS_ADD:
    "/namespaces/:tenantNamespace/tenants/:tenantName/add-pool",
  NAMESPACE_TENANT_POOLS_EDIT:
    "/namespaces/:tenantNamespace/tenants/:tenantName/edit-pool",
  NAMESPACE_TENANT_VOLUMES:
    "/namespaces/:tenantNamespace/tenants/:tenantName/volumes",
  NAMESPACE_TENANT_LICENSE:
    "/namespaces/:tenantNamespace/tenants/:tenantName/license",
  NAMESPACE_TENANT_IDENTITY_PROVIDER:
    "/namespaces/:tenantNamespace/tenants/:tenantName/identity-provider",
  NAMESPACE_TENANT_SECURITY:
    "/namespaces/:tenantNamespace/tenants/:tenantName/security",
  NAMESPACE_TENANT_ENCRYPTION:
    "/namespaces/:tenantNamespace/tenants/:tenantName/encryption",
  NAMESPACE_TENANT_MONITORING:
    "/namespaces/:tenantNamespace/tenants/:tenantName/monitoring",
  NAMESPACE_TENANT_LOGGING:
    "/namespaces/:tenantNamespace/tenants/:tenantName/logging",
  NAMESPACE_TENANT_EVENTS:
    "/namespaces/:tenantNamespace/tenants/:tenantName/events",
  NAMESPACE_TENANT_CSR: "/namespaces/:tenantNamespace/tenants/:tenantName/csr",
  OPERATOR_MARKETPLACE: "/marketplace",

  /* DirectPV */
  DIRECTPV_STORAGE: "/storage",
  DIRECTPV_DRIVES: "/drives",
  DIRECTPV_VOLUMES: "/volumes",
};

// roles
export const IAM_PERMISSIONS = {
  [IAM_ROLES.BUCKET_OWNER]: [
    IAM_SCOPES.S3_PUT_OBJECT,
    IAM_SCOPES.S3_PUT_ACTIONS,
    IAM_SCOPES.S3_DELETE_OBJECT,
  ],
  [IAM_ROLES.BUCKET_VIEWER]: [
    IAM_SCOPES.S3_LIST_BUCKET,
    IAM_SCOPES.S3_ALL_LIST_BUCKET,
  ],
  [IAM_ROLES.BUCKET_ADMIN]: [
    IAM_SCOPES.S3_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.S3_REPLICATE_OBJECT,
    IAM_SCOPES.S3_REPLICATE_DELETE,
    IAM_SCOPES.S3_REPLICATE_TAGS,
    IAM_SCOPES.S3_GET_OBJECT_VERSION_FOR_REPLICATION,
    IAM_SCOPES.S3_PUT_REPLICATION_CONFIGURATION,
    IAM_SCOPES.S3_GET_REPLICATION_CONFIGURATION,
    IAM_SCOPES.S3_GET_BUCKET_VERSIONING,
    IAM_SCOPES.S3_PUT_BUCKET_VERSIONING,
    IAM_SCOPES.S3_GET_BUCKET_ENCRYPTION_CONFIGURATION,
    IAM_SCOPES.S3_PUT_BUCKET_ENCRYPTION_CONFIGURATION,
    IAM_SCOPES.S3_DELETE_OBJECT_TAGGING,
    IAM_SCOPES.S3_PUT_OBJECT_TAGGING,
    IAM_SCOPES.S3_GET_OBJECT_TAGGING,
    IAM_SCOPES.S3_PUT_OBJECT_VERSION_TAGGING,
    IAM_SCOPES.S3_DELETE_OBJECT_VERSION_TAGGING,
    IAM_SCOPES.S3_DELETE_OBJECT_VERSION,
    IAM_SCOPES.S3_GET_OBJECT_VERSION_TAGGING,
    IAM_SCOPES.S3_GET_OBJECT_VERSION,
    IAM_SCOPES.S3_PUT_BUCKET_TAGGING,
    IAM_SCOPES.S3_GET_BUCKET_TAGGING,
    IAM_SCOPES.S3_PUT_BUCKET_OBJECT_LOCK_CONFIGURATION,
    IAM_SCOPES.S3_GET_BUCKET_OBJECT_LOCK_CONFIGURATION,
    IAM_SCOPES.S3_PUT_OBJECT_LEGAL_HOLD,
    IAM_SCOPES.S3_GET_OBJECT_LEGAL_HOLD,
    IAM_SCOPES.S3_GET_OBJECT_RETENTION,
    IAM_SCOPES.S3_PUT_OBJECT_RETENTION,
    IAM_SCOPES.S3_BYPASS_GOVERNANCE_RETENTION,
    IAM_SCOPES.S3_PUT_BUCKET_POLICY,
    IAM_SCOPES.S3_PUT_BUCKET_NOTIFICATIONS,
    IAM_SCOPES.S3_GET_LIFECYCLE_CONFIGURATION,
    IAM_SCOPES.S3_PUT_LIFECYCLE_CONFIGURATION,
    IAM_SCOPES.S3_LIST_MULTIPART_UPLOAD_PARTS,
    IAM_SCOPES.S3_LISTEN_BUCKET_NOTIFICATIONS,
    IAM_SCOPES.S3_LISTEN_NOTIFICATIONS,
    IAM_SCOPES.S3_LIST_BUCKET_MULTIPART_UPLOADS,
    IAM_SCOPES.S3_LIST_BUCKET_VERSIONS,
    IAM_SCOPES.S3_GET_BUCKET_POLICY_STATUS,
    IAM_SCOPES.S3_LIST_ALL_MY_BUCKETS,
    IAM_SCOPES.S3_HEAD_BUCKET,
    IAM_SCOPES.S3_GET_BUCKET_POLICY,
    IAM_SCOPES.S3_GET_BUCKET_NOTIFICATIONS,
    IAM_SCOPES.S3_GET_BUCKET_LOCATION,
    IAM_SCOPES.S3_DELETE_BUCKET_POLICY,
    IAM_SCOPES.S3_FORCE_DELETE_BUCKET,
    IAM_SCOPES.S3_DELETE_BUCKET,
    IAM_SCOPES.S3_CREATE_BUCKET,
    IAM_SCOPES.S3_ABORT_MULTIPART_UPLOAD,
    IAM_SCOPES.ADMIN_GET_POLICY,
    IAM_SCOPES.ADMIN_LIST_USER_POLICIES,
    IAM_SCOPES.ADMIN_LIST_USERS,
    IAM_SCOPES.ADMIN_HEAL,
    IAM_SCOPES.S3_GET_ACTIONS,
    IAM_SCOPES.S3_PUT_ACTIONS,
  ],
  [IAM_ROLES.BUCKET_LIFECYCLE]: [
    IAM_SCOPES.S3_GET_LIFECYCLE_CONFIGURATION,
    IAM_SCOPES.S3_PUT_LIFECYCLE_CONFIGURATION,
    IAM_SCOPES.S3_GET_ACTIONS,
    IAM_SCOPES.S3_PUT_ACTIONS,
    IAM_SCOPES.ADMIN_LIST_TIERS,
    IAM_SCOPES.ADMIN_SET_TIER,
  ],
};

// application pages/routes and required scopes/roles
export const IAM_PAGES_PERMISSIONS = {
  [IAM_PAGES.ADD_BUCKETS]: [
    IAM_SCOPES.S3_CREATE_BUCKET, // create bucket page
  ],
  [IAM_PAGES.BUCKETS_ADMIN_VIEW]: [
    ...IAM_PERMISSIONS[IAM_ROLES.BUCKET_ADMIN], // bucket admin page
  ],
  [IAM_PAGES.OBJECT_BROWSER_VIEW]: [
    ...IAM_PERMISSIONS[IAM_ROLES.BUCKET_OWNER],
    ...IAM_PERMISSIONS[IAM_ROLES.BUCKET_VIEWER],
  ],
  [IAM_PAGES.GROUPS]: [
    IAM_SCOPES.ADMIN_LIST_GROUPS, // displays groups
    IAM_SCOPES.ADMIN_ADD_USER_TO_GROUP, // displays create group button
  ],
  [IAM_PAGES.GROUPS_VIEW]: [
    IAM_SCOPES.ADMIN_GET_GROUP,
    IAM_SCOPES.ADMIN_DISABLE_GROUP,
    IAM_SCOPES.ADMIN_ENABLE_GROUP,
    IAM_SCOPES.ADMIN_REMOVE_USER_FROM_GROUP,
    IAM_SCOPES.ADMIN_LIST_USER_POLICIES,
    IAM_SCOPES.ADMIN_ADD_USER_TO_GROUP, // display "edit members" button in groups detail page
    IAM_SCOPES.ADMIN_ATTACH_USER_OR_GROUP_POLICY, // display "set policy" button in groups details page
  ],
  [IAM_PAGES.GROUPS_ADD]: [
    IAM_SCOPES.ADMIN_LIST_USERS, // displays users
    IAM_SCOPES.ADMIN_CREATE_USER, // displays create user button
  ],
  [IAM_PAGES.USERS]: [
    IAM_SCOPES.ADMIN_LIST_USERS, // displays users
    IAM_SCOPES.ADMIN_CREATE_USER, // displays create user button
  ],
  [IAM_PAGES.USERS_VIEW]: [
    IAM_SCOPES.ADMIN_GET_USER, // displays list of users
    IAM_SCOPES.ADMIN_ADD_USER_TO_GROUP, // displays "add to gorups" button in user details page
    IAM_SCOPES.ADMIN_ENABLE_USER,
    IAM_SCOPES.ADMIN_DISABLE_USER,
    IAM_SCOPES.ADMIN_DELETE_USER,
  ],
  [IAM_PAGES.USER_SA_ACCOUNT_ADD]: [
    IAM_SCOPES.ADMIN_CREATE_SERVICEACCOUNT,
    IAM_SCOPES.ADMIN_UPDATE_SERVICEACCOUNT,
    IAM_SCOPES.ADMIN_REMOVE_SERVICEACCOUNT,
    IAM_SCOPES.ADMIN_LIST_SERVICEACCOUNTS,
  ],
  [IAM_PAGES.USER_ADD]: [IAM_SCOPES.ADMIN_CREATE_USER], // displays create user button
  [IAM_PAGES.ACCOUNT_ADD]: [IAM_SCOPES.ADMIN_CREATE_SERVICEACCOUNT],
  [IAM_PAGES.DASHBOARD]: [
    IAM_SCOPES.ADMIN_SERVER_INFO, // displays dashboard information
  ],
  [IAM_PAGES.POLICIES_VIEW]: [
    IAM_SCOPES.ADMIN_DELETE_POLICY,
    IAM_SCOPES.ADMIN_LIST_GROUPS,
    IAM_SCOPES.ADMIN_GET_GROUP,
    IAM_SCOPES.ADMIN_GET_POLICY,
    IAM_SCOPES.ADMIN_CREATE_POLICY,
  ],
  [IAM_PAGES.POLICIES]: [
    IAM_SCOPES.ADMIN_LIST_USER_POLICIES, // displays policies
    IAM_SCOPES.ADMIN_CREATE_POLICY, // displays create policy button
  ],
  [IAM_PAGES.POLICY_ADD]: [
    IAM_SCOPES.ADMIN_CREATE_POLICY, // displays create policy button
  ],
  [IAM_PAGES.SETTINGS]: [
    IAM_SCOPES.ADMIN_CONFIG_UPDATE, // displays configuration list
  ],
  [IAM_PAGES.SETTINGS_VIEW]: [
    IAM_SCOPES.ADMIN_CONFIG_UPDATE, // displays configuration list
  ],
  [IAM_PAGES.EVENT_DESTINATIONS_ADD_SERVICE]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.EVENT_DESTINATIONS_ADD]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.EVENT_DESTINATIONS]: [
    IAM_SCOPES.ADMIN_SERVER_INFO, // displays notifications endpoints
    IAM_SCOPES.ADMIN_CONFIG_UPDATE, // displays create notification button
  ],
  [IAM_PAGES.TIERS]: [
    IAM_SCOPES.ADMIN_LIST_TIERS, // display tiers list
  ],
  [IAM_PAGES.TIERS_ADD]: [
    IAM_SCOPES.ADMIN_SET_TIER, // display "add tier" button / shows add service tier page
    IAM_SCOPES.ADMIN_LIST_TIERS, // display tiers list
  ],
  [IAM_PAGES.TIERS_ADD_SERVICE]: [
    IAM_SCOPES.ADMIN_SET_TIER, // display "add tier" button / shows add service tier page
    IAM_SCOPES.ADMIN_LIST_TIERS, // display tiers list
  ],
  [IAM_PAGES.TOOLS]: [
    IAM_SCOPES.S3_LISTEN_NOTIFICATIONS, // displays watch notifications
    IAM_SCOPES.S3_LISTEN_BUCKET_NOTIFICATIONS, // display watch notifications
    IAM_SCOPES.ADMIN_GET_CONSOLE_LOG, // display minio console logs
    IAM_SCOPES.ADMIN_SERVER_TRACE, // display minio trace
    IAM_SCOPES.ADMIN_HEAL, // display heal
    IAM_SCOPES.ADMIN_HEALTH_INFO, // display diagnostics / display speedtest / display audit log
    IAM_SCOPES.ADMIN_SERVER_INFO, // display diagnostics
  ],
  [IAM_PAGES.TOOLS_LOGS]: [IAM_SCOPES.ADMIN_GET_CONSOLE_LOG],
  [IAM_PAGES.TOOLS_AUDITLOGS]: [IAM_SCOPES.ADMIN_HEALTH_INFO],
  [IAM_PAGES.TOOLS_WATCH]: [
    IAM_SCOPES.S3_LISTEN_NOTIFICATIONS, // displays watch notifications
    IAM_SCOPES.S3_LISTEN_BUCKET_NOTIFICATIONS, // display watch notifications
  ],
  [IAM_PAGES.TOOLS_TRACE]: [IAM_SCOPES.ADMIN_SERVER_TRACE],
  [IAM_PAGES.TOOLS_HEAL]: [IAM_SCOPES.ADMIN_HEAL],
  [IAM_PAGES.TOOLS_DIAGNOSTICS]: [
    IAM_SCOPES.ADMIN_HEALTH_INFO,
    IAM_SCOPES.ADMIN_SERVER_INFO,
  ],
  [IAM_PAGES.TOOLS_SPEEDTEST]: [IAM_SCOPES.ADMIN_HEALTH_INFO],
  [IAM_PAGES.REGISTER_SUPPORT]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.CALL_HOME]: [IAM_SCOPES.ADMIN_HEALTH_INFO],
  [IAM_PAGES.PROFILE]: [IAM_SCOPES.ADMIN_HEALTH_INFO],
  [IAM_PAGES.SUPPORT_INSPECT]: [IAM_SCOPES.ADMIN_HEALTH_INFO],
  [IAM_PAGES.LICENSE]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.SITE_REPLICATION]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.SITE_REPLICATION_STATUS]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.SITE_REPLICATION_ADD]: [
    IAM_SCOPES.ADMIN_SERVER_INFO,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.KMS]: [IAM_SCOPES.KMS_ALL_ACTIONS],
  [IAM_PAGES.KMS_STATUS]: [IAM_SCOPES.KMS_ALL_ACTIONS, IAM_SCOPES.KMS_STATUS],
  [IAM_PAGES.KMS_KEYS]: [
    IAM_SCOPES.KMS_ALL_ACTIONS,
    IAM_SCOPES.KMS_CREATE_KEY,
    IAM_SCOPES.KMS_DELETE_KEY,
    IAM_SCOPES.KMS_LIST_KEYS,
    IAM_SCOPES.KMS_IMPORT_KEY,
    IAM_SCOPES.KMS_KEY_STATUS,
  ],
  [IAM_PAGES.KMS_KEYS_ADD]: [
    IAM_SCOPES.KMS_ALL_ACTIONS,
    IAM_SCOPES.KMS_CREATE_KEY,
  ],
  [IAM_PAGES.KMS_KEYS_IMPORT]: [
    IAM_SCOPES.KMS_ALL_ACTIONS,
    IAM_SCOPES.KMS_IMPORT_KEY,
  ],
  [IAM_PAGES.IDP_LDAP_CONFIGURATIONS]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.IDP_LDAP_CONFIGURATIONS_ADD]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.IDP_LDAP_CONFIGURATIONS_VIEW]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.IDP_OPENID_CONFIGURATIONS]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.IDP_OPENID_CONFIGURATIONS_ADD]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
  [IAM_PAGES.IDP_OPENID_CONFIGURATIONS_VIEW]: [
    IAM_SCOPES.ADMIN_ALL_ACTIONS,
    IAM_SCOPES.ADMIN_CONFIG_UPDATE,
  ],
};

export const CONSOLE_UI_RESOURCE = "console-ui";
