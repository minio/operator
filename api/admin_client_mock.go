// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
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

package api

import (
	"context"
	"io"
	"time"

	"github.com/minio/madmin-go/v2"
	iampolicy "github.com/minio/pkg/iam/policy"
)

// AdminClientMock mock client
type AdminClientMock struct{}

var (
	// MinioServerInfoMock is the mc admin info server client mock
	MinioServerInfoMock     func(ctx context.Context) (madmin.InfoMessage, error)
	minioChangePasswordMock func(ctx context.Context, accessKey, secretKey string) error

	minioHelpConfigKVMock       func(subSys, key string, envOnly bool) (madmin.Help, error)
	minioGetConfigKVMock        func(key string) ([]byte, error)
	minioSetConfigKVMock        func(kv string) (restart bool, err error)
	minioDelConfigKVMock        func(name string) (err error)
	minioHelpConfigKVGlobalMock func(envOnly bool) (madmin.Help, error)

	minioGetLogsMock func(ctx context.Context, node string, lineCnt int, logKind string) <-chan madmin.LogInfo

	minioListGroupsMock          func() ([]string, error)
	minioUpdateGroupMembersMock  func(madmin.GroupAddRemove) error
	minioGetGroupDescriptionMock func(group string) (*madmin.GroupDesc, error)
	minioSetGroupStatusMock      func(group string, status madmin.GroupStatus) error

	minioHealMock func(ctx context.Context, bucket, prefix string, healOpts madmin.HealOpts, clientToken string,
		forceStart, forceStop bool) (healStart madmin.HealStartSuccess, healTaskStatus madmin.HealTaskStatus, err error)

	minioServerHealthInfoMock func(ctx context.Context, healthDataTypes []madmin.HealthDataType, deadline time.Duration) (interface{}, string, error)

	minioListPoliciesMock func() (map[string]*iampolicy.Policy, error)
	minioGetPolicyMock    func(name string) (*iampolicy.Policy, error)
	minioRemovePolicyMock func(name string) error
	minioAddPolicyMock    func(name string, policy *iampolicy.Policy) error
	minioSetPolicyMock    func(policyName, entityName string, isGroup bool) error

	minioStartProfiling func(profiler madmin.ProfilerType) ([]madmin.StartProfilingResult, error)
	minioStopProfiling  func() (io.ReadCloser, error)

	minioServiceRestartMock func(ctx context.Context) error

	getSiteReplicationInfo        func(ctx context.Context) (*madmin.SiteReplicationInfo, error)
	addSiteReplicationInfo        func(ctx context.Context, sites []madmin.PeerSite) (*madmin.ReplicateAddStatus, error)
	editSiteReplicationInfo       func(ctx context.Context, site madmin.PeerInfo) (*madmin.ReplicateEditStatus, error)
	deleteSiteReplicationInfoMock func(ctx context.Context, removeReq madmin.SRRemoveReq) (*madmin.ReplicateRemoveStatus, error)
	getSiteReplicationStatus      func(ctx context.Context, params madmin.SRStatusOptions) (*madmin.SRStatusInfo, error)

	minioListTiersMock func(ctx context.Context) ([]*madmin.TierConfig, error)
	minioTierStatsMock func(ctx context.Context) ([]madmin.TierInfo, error)
	minioAddTiersMock  func(ctx context.Context, tier *madmin.TierConfig) error
	minioEditTiersMock func(ctx context.Context, tierName string, creds madmin.TierCreds) error

	minioServiceTraceMock func(ctx context.Context, threshold int64, s3, internal, storage, os, errTrace bool) <-chan madmin.ServiceTraceInfo

	minioListUsersMock     func() (map[string]madmin.UserInfo, error)
	minioAddUserMock       func(accessKey, secreyKey string) error
	minioRemoveUserMock    func(accessKey string) error
	minioGetUserInfoMock   func(accessKey string) (madmin.UserInfo, error)
	minioSetUserStatusMock func(accessKey string, status madmin.AccountStatus) error

	minioAccountInfoMock          func(ctx context.Context) (madmin.AccountInfo, error)
	minioAddServiceAccountMock    func(ctx context.Context, policy *iampolicy.Policy, user string, accessKey string, secretKey string) (madmin.Credentials, error)
	minioListServiceAccountsMock  func(ctx context.Context, user string) (madmin.ListServiceAccountsResp, error)
	minioDeleteServiceAccountMock func(ctx context.Context, serviceAccount string) error
	minioInfoServiceAccountMock   func(ctx context.Context, serviceAccount string) (madmin.InfoServiceAccountResp, error)
	minioUpdateServiceAccountMock func(ctx context.Context, serviceAccount string, opts madmin.UpdateServiceAccountReq) error
)

func (ac AdminClientMock) serverInfo(ctx context.Context) (madmin.InfoMessage, error) {
	return MinioServerInfoMock(ctx)
}

func (ac AdminClientMock) listRemoteBuckets(ctx context.Context, bucket, arnType string) (targets []madmin.BucketTarget, err error) {
	return nil, nil
}

func (ac AdminClientMock) getRemoteBucket(ctx context.Context, bucket, arnType string) (targets *madmin.BucketTarget, err error) {
	return nil, nil
}

func (ac AdminClientMock) removeRemoteBucket(ctx context.Context, bucket, arn string) error {
	return nil
}

func (ac AdminClientMock) addRemoteBucket(ctx context.Context, bucket string, target *madmin.BucketTarget) (string, error) {
	return "", nil
}

func (ac AdminClientMock) changePassword(ctx context.Context, accessKey, secretKey string) error {
	return minioChangePasswordMock(ctx, accessKey, secretKey)
}

func (ac AdminClientMock) speedtest(ctx context.Context, opts madmin.SpeedtestOpts) (chan madmin.SpeedTestResult, error) {
	return nil, nil
}

func (ac AdminClientMock) verifyTierStatus(ctx context.Context, tierName string) error {
	return nil
}

// mock function helpConfigKV()
func (ac AdminClientMock) helpConfigKV(ctx context.Context, subSys, key string, envOnly bool) (madmin.Help, error) {
	return minioHelpConfigKVMock(subSys, key, envOnly)
}

// mock function getConfigKV()
func (ac AdminClientMock) getConfigKV(ctx context.Context, name string) ([]byte, error) {
	return minioGetConfigKVMock(name)
}

// mock function setConfigKV()
func (ac AdminClientMock) setConfigKV(ctx context.Context, kv string) (restart bool, err error) {
	return minioSetConfigKVMock(kv)
}

// mock function helpConfigKV()
func (ac AdminClientMock) helpConfigKVGlobal(ctx context.Context, envOnly bool) (madmin.Help, error) {
	return minioHelpConfigKVGlobalMock(envOnly)
}

func (ac AdminClientMock) delConfigKV(ctx context.Context, name string) (err error) {
	return minioDelConfigKVMock(name)
}

func (ac AdminClientMock) getLogs(ctx context.Context, node string, lineCnt int, logKind string) <-chan madmin.LogInfo {
	return minioGetLogsMock(ctx, node, lineCnt, logKind)
}

func (ac AdminClientMock) listGroups(ctx context.Context) ([]string, error) {
	return minioListGroupsMock()
}

func (ac AdminClientMock) updateGroupMembers(ctx context.Context, req madmin.GroupAddRemove) error {
	return minioUpdateGroupMembersMock(req)
}

func (ac AdminClientMock) getGroupDescription(ctx context.Context, group string) (*madmin.GroupDesc, error) {
	return minioGetGroupDescriptionMock(group)
}

func (ac AdminClientMock) setGroupStatus(ctx context.Context, group string, status madmin.GroupStatus) error {
	return minioSetGroupStatusMock(group, status)
}

func (ac AdminClientMock) heal(ctx context.Context, bucket, prefix string, healOpts madmin.HealOpts, clientToken string,
	forceStart, forceStop bool,
) (healStart madmin.HealStartSuccess, healTaskStatus madmin.HealTaskStatus, err error) {
	return minioHealMock(ctx, bucket, prefix, healOpts, clientToken, forceStart, forceStop)
}

func (ac AdminClientMock) serverHealthInfo(ctx context.Context, healthDataTypes []madmin.HealthDataType, deadline time.Duration) (interface{}, string, error) {
	return minioServerHealthInfoMock(ctx, healthDataTypes, deadline)
}

func (ac AdminClientMock) addOrUpdateIDPConfig(ctx context.Context, idpType, cfgName, cfgData string, update bool) (restart bool, err error) {
	return true, nil
}

func (ac AdminClientMock) listIDPConfig(ctx context.Context, idpType string) ([]madmin.IDPListItem, error) {
	return []madmin.IDPListItem{{Name: "mock"}}, nil
}

func (ac AdminClientMock) deleteIDPConfig(ctx context.Context, idpType, cfgName string) (restart bool, err error) {
	return true, nil
}

func (ac AdminClientMock) getIDPConfig(ctx context.Context, cfgType, cfgName string) (c madmin.IDPConfig, err error) {
	return madmin.IDPConfig{Info: []madmin.IDPCfgInfo{{Key: "mock", Value: "mock"}}}, nil
}

func (ac AdminClientMock) kmsStatus(ctx context.Context) (madmin.KMSStatus, error) {
	return madmin.KMSStatus{Name: "name", DefaultKeyID: "key", Endpoints: map[string]madmin.ItemState{"localhost": madmin.ItemState("online")}}, nil
}

func (ac AdminClientMock) kmsAPIs(ctx context.Context) ([]madmin.KMSAPI, error) {
	return []madmin.KMSAPI{{Method: "GET", Path: "/mock"}}, nil
}

func (ac AdminClientMock) kmsMetrics(ctx context.Context) (*madmin.KMSMetrics, error) {
	return &madmin.KMSMetrics{}, nil
}

func (ac AdminClientMock) kmsVersion(ctx context.Context) (*madmin.KMSVersion, error) {
	return &madmin.KMSVersion{Version: "test-version"}, nil
}

func (ac AdminClientMock) createKey(ctx context.Context, key string) error {
	return nil
}

func (ac AdminClientMock) importKey(ctx context.Context, key string, content []byte) error {
	return nil
}

func (ac AdminClientMock) listKeys(ctx context.Context, pattern string) ([]madmin.KMSKeyInfo, error) {
	return []madmin.KMSKeyInfo{{
		Name:      "name",
		CreatedBy: "by",
	}}, nil
}

func (ac AdminClientMock) keyStatus(ctx context.Context, key string) (*madmin.KMSKeyStatus, error) {
	return &madmin.KMSKeyStatus{KeyID: "key"}, nil
}

func (ac AdminClientMock) deleteKey(ctx context.Context, key string) error {
	return nil
}

func (ac AdminClientMock) setKMSPolicy(ctx context.Context, policy string, content []byte) error {
	return nil
}

func (ac AdminClientMock) assignPolicy(ctx context.Context, policy string, content []byte) error {
	return nil
}

func (ac AdminClientMock) describePolicy(ctx context.Context, policy string) (*madmin.KMSDescribePolicy, error) {
	return &madmin.KMSDescribePolicy{Name: "name"}, nil
}

func (ac AdminClientMock) getKMSPolicy(ctx context.Context, policy string) (*madmin.KMSPolicy, error) {
	return &madmin.KMSPolicy{Allow: []string{""}, Deny: []string{""}}, nil
}

func (ac AdminClientMock) listKMSPolicies(ctx context.Context, pattern string) ([]madmin.KMSPolicyInfo, error) {
	return []madmin.KMSPolicyInfo{{
		Name:      "name",
		CreatedBy: "by",
	}}, nil
}

func (ac AdminClientMock) deletePolicy(ctx context.Context, policy string) error {
	return nil
}

func (ac AdminClientMock) describeIdentity(ctx context.Context, identity string) (*madmin.KMSDescribeIdentity, error) {
	return &madmin.KMSDescribeIdentity{}, nil
}

func (ac AdminClientMock) describeSelfIdentity(ctx context.Context) (*madmin.KMSDescribeSelfIdentity, error) {
	return &madmin.KMSDescribeSelfIdentity{
		Policy: &madmin.KMSPolicy{Allow: []string{}, Deny: []string{}},
	}, nil
}

func (ac AdminClientMock) deleteIdentity(ctx context.Context, identity string) error {
	return nil
}

func (ac AdminClientMock) listIdentities(ctx context.Context, pattern string) ([]madmin.KMSIdentityInfo, error) {
	return []madmin.KMSIdentityInfo{{Identity: "identity"}}, nil
}

func (ac AdminClientMock) listPolicies(ctx context.Context) (map[string]*iampolicy.Policy, error) {
	return minioListPoliciesMock()
}

func (ac AdminClientMock) getPolicy(ctx context.Context, name string) (*iampolicy.Policy, error) {
	return minioGetPolicyMock(name)
}

func (ac AdminClientMock) removePolicy(ctx context.Context, name string) error {
	return minioRemovePolicyMock(name)
}

func (ac AdminClientMock) addPolicy(ctx context.Context, name string, policy *iampolicy.Policy) error {
	return minioAddPolicyMock(name, policy)
}

func (ac AdminClientMock) setPolicy(ctx context.Context, policyName, entityName string, isGroup bool) error {
	return minioSetPolicyMock(policyName, entityName, isGroup)
}

// mock function for startProfiling()
func (ac AdminClientMock) startProfiling(ctx context.Context, profiler madmin.ProfilerType) ([]madmin.StartProfilingResult, error) {
	return minioStartProfiling(profiler)
}

// mock function for stopProfiling()
func (ac AdminClientMock) stopProfiling(ctx context.Context) (io.ReadCloser, error) {
	return minioStopProfiling()
}

// mock function of serviceRestart()
func (ac AdminClientMock) serviceRestart(ctx context.Context) error {
	return minioServiceRestartMock(ctx)
}

func (ac AdminClientMock) getSiteReplicationInfo(ctx context.Context) (*madmin.SiteReplicationInfo, error) {
	return getSiteReplicationInfo(ctx)
}

func (ac AdminClientMock) addSiteReplicationInfo(ctx context.Context, sites []madmin.PeerSite) (*madmin.ReplicateAddStatus, error) {
	return addSiteReplicationInfo(ctx, sites)
}

func (ac AdminClientMock) editSiteReplicationInfo(ctx context.Context, site madmin.PeerInfo) (*madmin.ReplicateEditStatus, error) {
	return editSiteReplicationInfo(ctx, site)
}

func (ac AdminClientMock) deleteSiteReplicationInfo(ctx context.Context, removeReq madmin.SRRemoveReq) (*madmin.ReplicateRemoveStatus, error) {
	return deleteSiteReplicationInfoMock(ctx, removeReq)
}

func (ac AdminClientMock) getSiteReplicationStatus(ctx context.Context, params madmin.SRStatusOptions) (*madmin.SRStatusInfo, error) {
	return getSiteReplicationStatus(ctx, params)
}

func (ac AdminClientMock) listTiers(ctx context.Context) ([]*madmin.TierConfig, error) {
	return minioListTiersMock(ctx)
}

func (ac AdminClientMock) tierStats(ctx context.Context) ([]madmin.TierInfo, error) {
	return minioTierStatsMock(ctx)
}

func (ac AdminClientMock) addTier(ctx context.Context, tier *madmin.TierConfig) error {
	return minioAddTiersMock(ctx, tier)
}

func (ac AdminClientMock) editTierCreds(ctx context.Context, tierName string, creds madmin.TierCreds) error {
	return minioEditTiersMock(ctx, tierName, creds)
}

func (ac AdminClientMock) serviceTrace(ctx context.Context, threshold int64, s3, internal, storage, os, errTrace bool) <-chan madmin.ServiceTraceInfo {
	return minioServiceTraceMock(ctx, threshold, s3, internal, storage, os, errTrace)
}

func (ac AdminClientMock) listUsers(ctx context.Context) (map[string]madmin.UserInfo, error) {
	return minioListUsersMock()
}

func (ac AdminClientMock) addUser(ctx context.Context, accessKey, secretKey string) error {
	return minioAddUserMock(accessKey, secretKey)
}

func (ac AdminClientMock) removeUser(ctx context.Context, accessKey string) error {
	return minioRemoveUserMock(accessKey)
}

func (ac AdminClientMock) getUserInfo(ctx context.Context, accessKey string) (madmin.UserInfo, error) {
	return minioGetUserInfoMock(accessKey)
}

func (ac AdminClientMock) setUserStatus(ctx context.Context, accessKey string, status madmin.AccountStatus) error {
	return minioSetUserStatusMock(accessKey, status)
}

// AccountInfo mock
func (ac AdminClientMock) AccountInfo(ctx context.Context) (madmin.AccountInfo, error) {
	return minioAccountInfoMock(ctx)
}

func (ac AdminClientMock) addServiceAccount(ctx context.Context, policy *iampolicy.Policy, user string, accessKey string, secretKey string) (madmin.Credentials, error) {
	return minioAddServiceAccountMock(ctx, policy, user, accessKey, secretKey)
}

func (ac AdminClientMock) listServiceAccounts(ctx context.Context, user string) (madmin.ListServiceAccountsResp, error) {
	return minioListServiceAccountsMock(ctx, user)
}

func (ac AdminClientMock) deleteServiceAccount(ctx context.Context, serviceAccount string) error {
	return minioDeleteServiceAccountMock(ctx, serviceAccount)
}

func (ac AdminClientMock) infoServiceAccount(ctx context.Context, serviceAccount string) (madmin.InfoServiceAccountResp, error) {
	return minioInfoServiceAccountMock(ctx, serviceAccount)
}

func (ac AdminClientMock) updateServiceAccount(ctx context.Context, serviceAccount string, opts madmin.UpdateServiceAccountReq) error {
	return minioUpdateServiceAccountMock(ctx, serviceAccount, opts)
}
