// Copyright (C) 2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package v1

// Condition Clauses keys
const (
	// names
	StringEquals              ConditionClauseKey = "StringEquals"
	StringNotEquals                              = "StringNotEquals"
	StringEqualsIgnoreCase                       = "StringEqualsIgnoreCase"
	StringNotEqualsIgnoreCase                    = "StringNotEqualsIgnoreCase"
	StringLike                                   = "StringLike"
	StringNotLike                                = "StringNotLike"
	BinaryEquals                                 = "BinaryEquals"
	IpAddress                                    = "IpAddress"
	NotIPAddress                                 = "NotIpAddress"
	Null                                         = "Null"
	Boolean                                      = "Bool"
	NumericEquals                                = "NumericEquals"
	NumericNotEquals                             = "NumericNotEquals"
	NumericLessThan                              = "NumericLessThan"
	NumericLessThanEquals                        = "NumericLessThanEquals"
	NumericGreaterThan                           = "NumericGreaterThan"
	NumericGreaterThanEquals                     = "NumericGreaterThanEquals"
	DateEquals                                   = "DateEquals"
	DateNotEquals                                = "DateNotEquals"
	DateLessThan                                 = "DateLessThan"
	DateLessThanEquals                           = "DateLessThanEquals"
	DateGreaterThan                              = "DateGreaterThan"
	DateGreaterThanEquals                        = "DateGreaterThanEquals"

	// qualifiers
	// refer https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_multi-value-conditions.html#reference_policies_multi-key-or-value-conditions
	ForAllValues ConditionClauseKey = "ForAllValues"
	ForAnyValue                     = "ForAnyValue"
)

// Condition key names.
const (
	// S3XAmzCopySource - key representing x-amz-copy-source HTTP header applicable to PutObject API only.
	S3XAmzCopySource ConditionKeyName = "s3:x-amz-copy-source"

	// S3XAmzServerSideEncryption - key representing x-amz-server-side-encryption HTTP header applicable
	// to PutObject API only.
	S3XAmzServerSideEncryption ConditionKeyName = "s3:x-amz-server-side-encryption"

	// S3XAmzServerSideEncryptionCustomerAlgorithm - key representing
	// x-amz-server-side-encryption-customer-algorithm HTTP header applicable to PutObject API only.
	S3XAmzServerSideEncryptionCustomerAlgorithm ConditionKeyName = "s3:x-amz-server-side-encryption-customer-algorithm"

	// S3XAmzMetadataDirective - key representing x-amz-metadata-directive HTTP header applicable to
	// PutObject API only.
	S3XAmzMetadataDirective ConditionKeyName = "s3:x-amz-metadata-directive"

	// S3XAmzContentSha256 - set a static content-sha256 for all calls for a given action.
	S3XAmzContentSha256 ConditionKeyName = "s3:x-amz-content-sha256"

	// S3XAmzStorageClass - key representing x-amz-storage-class HTTP header applicable to PutObject API
	// only.
	S3XAmzStorageClass ConditionKeyName = "s3:x-amz-storage-class"

	// S3LocationConstraint - key representing LocationConstraint XML tag of CreateBucket API only.
	S3LocationConstraint ConditionKeyName = "s3:LocationConstraint"

	// S3Prefix - key representing prefix query parameter of ListBucket API only.
	S3Prefix ConditionKeyName = "s3:prefix"

	// S3Delimiter - key representing delimiter query parameter of ListBucket API only.
	S3Delimiter ConditionKeyName = "s3:delimiter"

	// S3VersionID - Enables you to limit the permission for the
	// s3:PutObjectVersionTagging action to a specific object version.
	S3VersionID ConditionKeyName = "s3:versionid"

	// S3MaxKeys - key representing max-keys query parameter of ListBucket API only.
	S3MaxKeys ConditionKeyName = "s3:max-keys"

	// S3ObjectLockRemainingRetentionDays - key representing object-lock-remaining-retention-days
	// Enables enforcement of an object relative to the remaining retention days, you can set
	// minimum and maximum allowable retention periods for a bucket using a bucket policy.
	// This key are specific for s3:PutObjectRetention API.
	S3ObjectLockRemainingRetentionDays ConditionKeyName = "s3:object-lock-remaining-retention-days"

	// S3ObjectLockMode - key representing object-lock-mode
	// Enables enforcement of the specified object retention mode
	S3ObjectLockMode ConditionKeyName = "s3:object-lock-mode"

	// S3ObjectLockRetainUntilDate - key representing object-lock-retain-util-date
	// Enables enforcement of a specific retain-until-date
	S3ObjectLockRetainUntilDate ConditionKeyName = "s3:object-lock-retain-until-date"

	// S3ObjectLockLegalHold - key representing object-local-legal-hold
	// Enables enforcement of the specified object legal hold status
	S3ObjectLockLegalHold ConditionKeyName = "s3:object-lock-legal-hold"

	// AWSReferer - key representing Referer header of any API.
	AWSReferer ConditionKeyName = "aws:Referer"

	// AWSSourceIP - key representing client's IP address (not intermittent proxies) of any API.
	AWSSourceIP ConditionKeyName = "aws:SourceIp"

	// AWSUserAgent - key representing UserAgent header for any API.
	AWSUserAgent ConditionKeyName = "aws:UserAgent"

	// AWSSecureTransport - key representing if the clients request is authenticated or not.
	AWSSecureTransport ConditionKeyName = "aws:SecureTransport"

	// AWSCurrentTime - key representing the current time.
	AWSCurrentTime ConditionKeyName = "aws:CurrentTime"

	// AWSEpochTime - key representing the current epoch time.
	AWSEpochTime ConditionKeyName = "aws:EpochTime"

	// AWSPrincipalType - user principal type currently supported values are "User" and "Anonymous".
	AWSPrincipalType ConditionKeyName = "aws:principaltype"

	// AWSUserID - user unique ID, in MinIO this value is same as your user Access Key.
	AWSUserID ConditionKeyName = "aws:userid"

	// AWSUsername - user friendly name, in MinIO this value is same as your user Access Key.
	AWSUsername ConditionKeyName = "aws:username"

	// S3SignatureVersion - identifies the version of AWS Signature that you want to support for authenticated requests.
	S3SignatureVersion ConditionKeyName = "s3:signatureversion"

	// S3AuthType - optionally use this condition key to restrict incoming requests to use a specific authentication method.
	S3AuthType ConditionKeyName = "s3:authType"

	// Refer https://docs.aws.amazon.com/AmazonS3/latest/userguide/tagging-and-policies.html
	ExistingObjectTag    ConditionKeyName = "s3:ExistingObjectTag"
	RequestObjectTagKeys ConditionKeyName = "s3:RequestObjectTagKeys"
	RequestObjectTag     ConditionKeyName = "s3:RequestObjectTag"
)

// JWT claims supported substitutions.
// https://www.iana.org/assignments/jwt/jwt.xhtml#claims
const (
	// JWTSub - JWT subject claim substitution.
	JWTSub ConditionKeyName = "jwt:sub"

	// JWTIss issuer claim substitution.
	JWTIss ConditionKeyName = "jwt:iss"

	// JWTAud audience claim substitution.
	JWTAud ConditionKeyName = "jwt:aud"

	// JWTJti JWT unique identifier claim substitution.
	JWTJti ConditionKeyName = "jwt:jti"

	JWTUpn          ConditionKeyName = "jwt:upn"
	JWTName         ConditionKeyName = "jwt:name"
	JWTGroups       ConditionKeyName = "jwt:groups"
	JWTGivenName    ConditionKeyName = "jwt:given_name"
	JWTFamilyName   ConditionKeyName = "jwt:family_name"
	JWTMiddleName   ConditionKeyName = "jwt:middle_name"
	JWTNickName     ConditionKeyName = "jwt:nickname"
	JWTPrefUsername ConditionKeyName = "jwt:preferred_username"
	JWTProfile      ConditionKeyName = "jwt:profile"
	JWTPicture      ConditionKeyName = "jwt:picture"
	JWTWebsite      ConditionKeyName = "jwt:website"
	JWTEmail        ConditionKeyName = "jwt:email"
	JWTGender       ConditionKeyName = "jwt:gender"
	JWTBirthdate    ConditionKeyName = "jwt:birthdate"
	JWTPhoneNumber  ConditionKeyName = "jwt:phone_number"
	JWTAddress      ConditionKeyName = "jwt:address"
	JWTScope        ConditionKeyName = "jwt:scope"
	JWTClientID     ConditionKeyName = "jwt:client_id"
)

const (
	// LDAPUser - LDAP username, in MinIO this value is equal to your authenticating LDAP user.
	LDAPUser ConditionKeyName = "ldap:user"

	// LDAPUsername - LDAP username, in MinIO is the authenticated simply user.
	LDAPUsername ConditionKeyName = "ldap:username"
)
