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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/minio/madmin-go/v2"
	mcCmd "github.com/minio/mc/cmd"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/operator/pkg"
	iampolicy "github.com/minio/pkg/iam/policy"
)

// NewAdminClientWithInsecure gives a new madmin client interface either secure or insecure based on parameter
func NewAdminClientWithInsecure(url, accessKey, secretKey, sessionToken string, insecure bool) (*madmin.AdminClient, *probe.Error) {
	admClient, err := s3AdminNew(&mcCmd.Config{
		HostURL:      url,
		AccessKey:    accessKey,
		SecretKey:    secretKey,
		SessionToken: sessionToken,
		AppName:      globalAppName,
		AppVersion:   pkg.Version,
		Insecure:     insecure,
	})
	if err != nil {
		return nil, err.Trace(url)
	}
	stsClient := PrepareConsoleHTTPClient(insecure)
	admClient.SetCustomTransport(stsClient.Transport)
	// set user-agent to differentiate Console UI requests for auditing.
	admClient.SetAppInfo("MinIO Console", pkg.Version)
	return admClient, nil
}

// s3AdminNew returns an initialized minioAdmin structure. If debug is enabled,
// it also enables an internal trace transport.
var s3AdminNew = mcCmd.NewAdminFactory()

// MinioAdmin interface with all functions to be implemented
// by mock when testing, it should include all MinioAdmin respective api calls
// that are used within this project.
type MinioAdmin interface {
	listUsers(ctx context.Context) (map[string]madmin.UserInfo, error)
	addUser(ctx context.Context, acessKey, SecretKey string) error
	removeUser(ctx context.Context, accessKey string) error
	getUserInfo(ctx context.Context, accessKey string) (madmin.UserInfo, error)
	setUserStatus(ctx context.Context, accessKey string, status madmin.AccountStatus) error
	listGroups(ctx context.Context) ([]string, error)
	updateGroupMembers(ctx context.Context, greq madmin.GroupAddRemove) error
	getGroupDescription(ctx context.Context, group string) (*madmin.GroupDesc, error)
	setGroupStatus(ctx context.Context, group string, status madmin.GroupStatus) error
	listPolicies(ctx context.Context) (map[string]*iampolicy.Policy, error)
	getPolicy(ctx context.Context, name string) (*iampolicy.Policy, error)
	removePolicy(ctx context.Context, name string) error
	addPolicy(ctx context.Context, name string, policy *iampolicy.Policy) error
	setPolicy(ctx context.Context, policyName, entityName string, isGroup bool) error
	getConfigKV(ctx context.Context, key string) ([]byte, error)
	helpConfigKV(ctx context.Context, subSys, key string, envOnly bool) (madmin.Help, error)
	helpConfigKVGlobal(ctx context.Context, envOnly bool) (madmin.Help, error)
	setConfigKV(ctx context.Context, kv string) (restart bool, err error)
	delConfigKV(ctx context.Context, kv string) (err error)
	serviceRestart(ctx context.Context) error
	serverInfo(ctx context.Context) (madmin.InfoMessage, error)
	startProfiling(ctx context.Context, profiler madmin.ProfilerType) ([]madmin.StartProfilingResult, error)
	stopProfiling(ctx context.Context) (io.ReadCloser, error)
	serviceTrace(ctx context.Context, threshold int64, s3, internal, storage, os, errTrace bool) <-chan madmin.ServiceTraceInfo
	getLogs(ctx context.Context, node string, lineCnt int, logKind string) <-chan madmin.LogInfo
	AccountInfo(ctx context.Context) (madmin.AccountInfo, error)
	heal(ctx context.Context, bucket, prefix string, healOpts madmin.HealOpts, clientToken string,
		forceStart, forceStop bool) (healStart madmin.HealStartSuccess, healTaskStatus madmin.HealTaskStatus, err error)
	// Service Accounts
	addServiceAccount(ctx context.Context, policy *iampolicy.Policy, user string, accessKey string, secretKey string) (madmin.Credentials, error)
	listServiceAccounts(ctx context.Context, user string) (madmin.ListServiceAccountsResp, error)
	deleteServiceAccount(ctx context.Context, serviceAccount string) error
	infoServiceAccount(ctx context.Context, serviceAccount string) (madmin.InfoServiceAccountResp, error)
	updateServiceAccount(ctx context.Context, serviceAccount string, opts madmin.UpdateServiceAccountReq) error
	// Remote Buckets
	listRemoteBuckets(ctx context.Context, bucket, arnType string) (targets []madmin.BucketTarget, err error)
	getRemoteBucket(ctx context.Context, bucket, arnType string) (targets *madmin.BucketTarget, err error)
	removeRemoteBucket(ctx context.Context, bucket, arn string) error
	addRemoteBucket(ctx context.Context, bucket string, target *madmin.BucketTarget) (string, error)
	// Account password management
	changePassword(ctx context.Context, accessKey, secretKey string) error
	serverHealthInfo(ctx context.Context, healthDataTypes []madmin.HealthDataType, deadline time.Duration) (interface{}, string, error)
	// List Tiers
	listTiers(ctx context.Context) ([]*madmin.TierConfig, error)
	// Tier Info
	tierStats(ctx context.Context) ([]madmin.TierInfo, error)
	// Add Tier
	addTier(ctx context.Context, tier *madmin.TierConfig) error
	// Edit Tier Credentials
	editTierCreds(ctx context.Context, tierName string, creds madmin.TierCreds) error
	// verify Tier status
	verifyTierStatus(ctx context.Context, tierName string) error
	// Speedtest
	speedtest(ctx context.Context, opts madmin.SpeedtestOpts) (chan madmin.SpeedTestResult, error)
	// Site Relication
	getSiteReplicationInfo(ctx context.Context) (*madmin.SiteReplicationInfo, error)
	addSiteReplicationInfo(ctx context.Context, sites []madmin.PeerSite) (*madmin.ReplicateAddStatus, error)
	editSiteReplicationInfo(ctx context.Context, site madmin.PeerInfo) (*madmin.ReplicateEditStatus, error)
	deleteSiteReplicationInfo(ctx context.Context, removeReq madmin.SRRemoveReq) (*madmin.ReplicateRemoveStatus, error)

	// Replication status
	getSiteReplicationStatus(ctx context.Context, params madmin.SRStatusOptions) (*madmin.SRStatusInfo, error)

	// KMS
	kmsStatus(ctx context.Context) (madmin.KMSStatus, error)
	kmsMetrics(ctx context.Context) (*madmin.KMSMetrics, error)
	kmsAPIs(ctx context.Context) ([]madmin.KMSAPI, error)
	kmsVersion(ctx context.Context) (*madmin.KMSVersion, error)
	createKey(ctx context.Context, key string) error
	importKey(ctx context.Context, key string, content []byte) error
	listKeys(ctx context.Context, pattern string) ([]madmin.KMSKeyInfo, error)
	keyStatus(ctx context.Context, key string) (*madmin.KMSKeyStatus, error)
	deleteKey(ctx context.Context, key string) error
	setKMSPolicy(ctx context.Context, policy string, content []byte) error
	assignPolicy(ctx context.Context, policy string, content []byte) error
	describePolicy(ctx context.Context, policy string) (*madmin.KMSDescribePolicy, error)
	getKMSPolicy(ctx context.Context, policy string) (*madmin.KMSPolicy, error)
	listKMSPolicies(ctx context.Context, pattern string) ([]madmin.KMSPolicyInfo, error)
	deletePolicy(ctx context.Context, policy string) error
	describeIdentity(ctx context.Context, identity string) (*madmin.KMSDescribeIdentity, error)
	describeSelfIdentity(ctx context.Context) (*madmin.KMSDescribeSelfIdentity, error)
	deleteIdentity(ctx context.Context, identity string) error
	listIdentities(ctx context.Context, pattern string) ([]madmin.KMSIdentityInfo, error)

	// IDP
	addOrUpdateIDPConfig(ctx context.Context, idpType, cfgName, cfgData string, update bool) (restart bool, err error)
	listIDPConfig(ctx context.Context, idpType string) ([]madmin.IDPListItem, error)
	deleteIDPConfig(ctx context.Context, idpType, cfgName string) (restart bool, err error)
	getIDPConfig(ctx context.Context, cfgType, cfgName string) (c madmin.IDPConfig, err error)
}

// AdminClient Interface implementation
//
// Define the structure of a minIO Client and define the functions that are actually used
// from minIO api.
type AdminClient struct {
	Client *madmin.AdminClient
}

func (ac AdminClient) changePassword(ctx context.Context, accessKey, secretKey string) error {
	return ac.Client.SetUser(ctx, accessKey, secretKey, madmin.AccountEnabled)
}

// implements madmin.ListUsers()
func (ac AdminClient) listUsers(ctx context.Context) (map[string]madmin.UserInfo, error) {
	return ac.Client.ListUsers(ctx)
}

// implements madmin.AddUser()
func (ac AdminClient) addUser(ctx context.Context, accessKey, secretKey string) error {
	return ac.Client.AddUser(ctx, accessKey, secretKey)
}

// implements madmin.RemoveUser()
func (ac AdminClient) removeUser(ctx context.Context, accessKey string) error {
	return ac.Client.RemoveUser(ctx, accessKey)
}

// implements madmin.GetUserInfo()
func (ac AdminClient) getUserInfo(ctx context.Context, accessKey string) (madmin.UserInfo, error) {
	return ac.Client.GetUserInfo(ctx, accessKey)
}

// implements madmin.SetUserStatus()
func (ac AdminClient) setUserStatus(ctx context.Context, accessKey string, status madmin.AccountStatus) error {
	return ac.Client.SetUserStatus(ctx, accessKey, status)
}

// implements madmin.ListGroups()
func (ac AdminClient) listGroups(ctx context.Context) ([]string, error) {
	return ac.Client.ListGroups(ctx)
}

// implements madmin.UpdateGroupMembers()
func (ac AdminClient) updateGroupMembers(ctx context.Context, greq madmin.GroupAddRemove) error {
	return ac.Client.UpdateGroupMembers(ctx, greq)
}

// implements madmin.GetGroupDescription(group)
func (ac AdminClient) getGroupDescription(ctx context.Context, group string) (*madmin.GroupDesc, error) {
	return ac.Client.GetGroupDescription(ctx, group)
}

// implements madmin.SetGroupStatus(group, status)
func (ac AdminClient) setGroupStatus(ctx context.Context, group string, status madmin.GroupStatus) error {
	return ac.Client.SetGroupStatus(ctx, group, status)
}

// implements madmin.ListCannedPolicies()
func (ac AdminClient) listPolicies(ctx context.Context) (map[string]*iampolicy.Policy, error) {
	policyMap, err := ac.Client.ListCannedPolicies(ctx)
	if err != nil {
		return nil, err
	}
	policies := make(map[string]*iampolicy.Policy, len(policyMap))
	for k, v := range policyMap {
		p, err := iampolicy.ParseConfig(bytes.NewReader(v))
		if err != nil {
			return nil, err
		}
		policies[k] = p
	}
	return policies, nil
}

// implements madmin.ListCannedPolicies()
func (ac AdminClient) getPolicy(ctx context.Context, name string) (*iampolicy.Policy, error) {
	praw, err := ac.Client.InfoCannedPolicy(ctx, name)
	if err != nil {
		return nil, err
	}
	return iampolicy.ParseConfig(bytes.NewReader(praw))
}

// implements madmin.RemoveCannedPolicy()
func (ac AdminClient) removePolicy(ctx context.Context, name string) error {
	return ac.Client.RemoveCannedPolicy(ctx, name)
}

// implements madmin.AddCannedPolicy()
func (ac AdminClient) addPolicy(ctx context.Context, name string, policy *iampolicy.Policy) error {
	buf, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return ac.Client.AddCannedPolicy(ctx, name, buf)
}

// implements madmin.SetPolicy()
func (ac AdminClient) setPolicy(ctx context.Context, policyName, entityName string, isGroup bool) error {
	return ac.Client.SetPolicy(ctx, policyName, entityName, isGroup)
}

// implements madmin.GetConfigKV()
func (ac AdminClient) getConfigKV(ctx context.Context, key string) ([]byte, error) {
	return ac.Client.GetConfigKV(ctx, key)
}

// implements madmin.HelpConfigKV()
func (ac AdminClient) helpConfigKV(ctx context.Context, subSys, key string, envOnly bool) (madmin.Help, error) {
	return ac.Client.HelpConfigKV(ctx, subSys, key, envOnly)
}

// implements madmin.helpConfigKVGlobal()
func (ac AdminClient) helpConfigKVGlobal(ctx context.Context, envOnly bool) (madmin.Help, error) {
	return ac.Client.HelpConfigKV(ctx, "", "", envOnly)
}

// implements madmin.SetConfigKV()
func (ac AdminClient) setConfigKV(ctx context.Context, kv string) (restart bool, err error) {
	return ac.Client.SetConfigKV(ctx, kv)
}

// implements madmin.DelConfigKV()
func (ac AdminClient) delConfigKV(ctx context.Context, kv string) (err error) {
	_, err = ac.Client.DelConfigKV(ctx, kv)
	return err
}

// implements madmin.ServiceRestart()
func (ac AdminClient) serviceRestart(ctx context.Context) (err error) {
	return ac.Client.ServiceRestart(ctx)
}

// implements madmin.ServerInfo()
func (ac AdminClient) serverInfo(ctx context.Context) (madmin.InfoMessage, error) {
	return ac.Client.ServerInfo(ctx)
}

// implements madmin.StartProfiling()
func (ac AdminClient) startProfiling(ctx context.Context, profiler madmin.ProfilerType) ([]madmin.StartProfilingResult, error) {
	return ac.Client.StartProfiling(ctx, profiler)
}

// implements madmin.DownloadProfilingData()
func (ac AdminClient) stopProfiling(ctx context.Context) (io.ReadCloser, error) {
	return ac.Client.DownloadProfilingData(ctx)
}

// implements madmin.ServiceTrace()
func (ac AdminClient) serviceTrace(ctx context.Context, threshold int64, s3, internal, storage, os, errTrace bool) <-chan madmin.ServiceTraceInfo {
	thresholdT := time.Duration(threshold)

	tracingOptions := madmin.ServiceTraceOpts{
		S3:         true,
		OnlyErrors: errTrace,
		Internal:   internal,
		Storage:    storage,
		OS:         os,
		Threshold:  thresholdT,
	}

	return ac.Client.ServiceTrace(ctx, tracingOptions)
}

// implements madmin.GetLogs()
func (ac AdminClient) getLogs(ctx context.Context, node string, lineCnt int, logKind string) <-chan madmin.LogInfo {
	return ac.Client.GetLogs(ctx, node, lineCnt, logKind)
}

// implements madmin.AddServiceAccount()
func (ac AdminClient) addServiceAccount(ctx context.Context, policy *iampolicy.Policy, user string, accessKey string, secretKey string) (madmin.Credentials, error) {
	buf, err := json.Marshal(policy)
	if err != nil {
		return madmin.Credentials{}, err
	}
	return ac.Client.AddServiceAccount(ctx, madmin.AddServiceAccountReq{
		Policy:     buf,
		TargetUser: user,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
	})
}

// implements madmin.ListServiceAccounts()
func (ac AdminClient) listServiceAccounts(ctx context.Context, user string) (madmin.ListServiceAccountsResp, error) {
	// TODO: Fix this
	return ac.Client.ListServiceAccounts(ctx, user)
}

// implements madmin.DeleteServiceAccount()
func (ac AdminClient) deleteServiceAccount(ctx context.Context, serviceAccount string) error {
	return ac.Client.DeleteServiceAccount(ctx, serviceAccount)
}

// implements madmin.InfoServiceAccount()
func (ac AdminClient) infoServiceAccount(ctx context.Context, serviceAccount string) (madmin.InfoServiceAccountResp, error) {
	return ac.Client.InfoServiceAccount(ctx, serviceAccount)
}

// implements madmin.UpdateServiceAccount()
func (ac AdminClient) updateServiceAccount(ctx context.Context, serviceAccount string, opts madmin.UpdateServiceAccountReq) error {
	return ac.Client.UpdateServiceAccount(ctx, serviceAccount, opts)
}

// AccountInfo implements madmin.AccountInfo()
func (ac AdminClient) AccountInfo(ctx context.Context) (madmin.AccountInfo, error) {
	return ac.Client.AccountInfo(ctx, madmin.AccountOpts{})
}

func (ac AdminClient) heal(ctx context.Context, bucket, prefix string, healOpts madmin.HealOpts, clientToken string,
	forceStart, forceStop bool,
) (healStart madmin.HealStartSuccess, healTaskStatus madmin.HealTaskStatus, err error) {
	return ac.Client.Heal(ctx, bucket, prefix, healOpts, clientToken, forceStart, forceStop)
}

// listRemoteBuckets - return a list of remote buckets
func (ac AdminClient) listRemoteBuckets(ctx context.Context, bucket, arnType string) (targets []madmin.BucketTarget, err error) {
	return ac.Client.ListRemoteTargets(ctx, bucket, arnType)
}

// getRemoteBucket - gets remote bucked based on a given bucket name
func (ac AdminClient) getRemoteBucket(ctx context.Context, bucket, arnType string) (*madmin.BucketTarget, error) {
	targets, err := ac.Client.ListRemoteTargets(ctx, bucket, arnType)
	if err != nil {
		return nil, err
	}
	if len(targets) > 0 {
		return &targets[0], nil
	}
	return nil, err
}

// removeRemoteBucket removes a remote target associated with particular ARN for this bucket
func (ac AdminClient) removeRemoteBucket(ctx context.Context, bucket, arn string) error {
	return ac.Client.RemoveRemoteTarget(ctx, bucket, arn)
}

// addRemoteBucket sets up a remote target for this bucket
func (ac AdminClient) addRemoteBucket(ctx context.Context, bucket string, target *madmin.BucketTarget) (string, error) {
	return ac.Client.SetRemoteTarget(ctx, bucket, target)
}

// serverHealthInfo implements mc.ServerHealthInfo - Connect to a minio server and call Health Info Management API
func (ac AdminClient) serverHealthInfo(ctx context.Context, healthDataTypes []madmin.HealthDataType, deadline time.Duration) (interface{}, string, error) {
	resp, version, err := ac.Client.ServerHealthInfo(ctx, healthDataTypes, deadline)
	if err != nil {
		return nil, version, err
	}

	var healthInfo interface{}

	decoder := json.NewDecoder(resp.Body)
	switch version {
	case madmin.HealthInfoVersion0:
		info := madmin.HealthInfoV0{}
		for {
			if err = decoder.Decode(&info); err != nil {
				break
			}
		}

		// Old minio versions don't return the MinIO info in
		// response of the healthinfo api. So fetch it separately
		minioInfo, err := ac.Client.ServerInfo(ctx)
		if err != nil {
			info.Minio.Error = err.Error()
		} else {
			info.Minio.Info = minioInfo
		}

		healthInfo = mcCmd.MapHealthInfoToV1(info, nil)
		version = madmin.HealthInfoVersion1
	case madmin.HealthInfoVersion:
		info := madmin.HealthInfo{}
		for {
			if err = decoder.Decode(&info); err != nil {
				break
			}
		}
		healthInfo = info
	}

	return healthInfo, version, nil
}

// implements madmin.listTiers()
func (ac AdminClient) listTiers(ctx context.Context) ([]*madmin.TierConfig, error) {
	return ac.Client.ListTiers(ctx)
}

// implements madmin.tierStats()
func (ac AdminClient) tierStats(ctx context.Context) ([]madmin.TierInfo, error) {
	return ac.Client.TierStats(ctx)
}

// implements madmin.AddTier()
func (ac AdminClient) addTier(ctx context.Context, cfg *madmin.TierConfig) error {
	return ac.Client.AddTier(ctx, cfg)
}

// implements madmin.EditTier()
func (ac AdminClient) editTierCreds(ctx context.Context, tierName string, creds madmin.TierCreds) error {
	return ac.Client.EditTier(ctx, tierName, creds)
}

// implements madmin.VerifyTier()
func (ac AdminClient) verifyTierStatus(ctx context.Context, tierName string) error {
	return ac.Client.VerifyTier(ctx, tierName)
}

// httpClient is a custom http client, this client should not be called directly and instead be
// called using GetConsoleHTTPClient() to ensure is initialized and the certificates are loaded correctly
var httpClients = struct {
	sync.Mutex
	m map[string]*http.Client
}{
	m: make(map[string]*http.Client),
}

// isLocalAddress returns true if the url contains an IPv4/IPv6 hostname
// that points to the local machine - FQDN are not supported
func isLocalIPAddress(ipAddr string) bool {
	if ipAddr == "" {
		return false
	}
	ip := net.ParseIP(ipAddr)
	return ip != nil && ip.IsLoopback()
}

// GetConsoleHTTPClient caches different http clients depending on the target endpoint while taking
// in consideration CA certs stored in ${HOME}/.console/certs/CAs and ${HOME}/.minio/certs/CAs
// If the target endpoint points to a loopback device, skip the TLS verification.
func GetConsoleHTTPClient(address string) *http.Client {
	u, err := url.Parse(address)
	if err == nil {
		address = u.Hostname()
	}

	httpClients.Lock()
	client, ok := httpClients.m[address]
	httpClients.Unlock()
	if ok {
		return client
	}

	client = PrepareConsoleHTTPClient(isLocalIPAddress(address))
	httpClients.Lock()
	httpClients.m[address] = client
	httpClients.Unlock()
	return client
}

func (ac AdminClient) speedtest(ctx context.Context, opts madmin.SpeedtestOpts) (chan madmin.SpeedTestResult, error) {
	return ac.Client.Speedtest(ctx, opts)
}

// Site Replication
func (ac AdminClient) getSiteReplicationInfo(ctx context.Context) (*madmin.SiteReplicationInfo, error) {
	res, err := ac.Client.SiteReplicationInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &madmin.SiteReplicationInfo{
		Enabled:                 res.Enabled,
		Name:                    res.Name,
		Sites:                   res.Sites,
		ServiceAccountAccessKey: res.ServiceAccountAccessKey,
	}, nil
}

func (ac AdminClient) addSiteReplicationInfo(ctx context.Context, sites []madmin.PeerSite) (*madmin.ReplicateAddStatus, error) {
	res, err := ac.Client.SiteReplicationAdd(ctx, sites)
	if err != nil {
		return nil, err
	}

	return &madmin.ReplicateAddStatus{
		Success:                 res.Success,
		Status:                  res.Status,
		ErrDetail:               res.ErrDetail,
		InitialSyncErrorMessage: res.InitialSyncErrorMessage,
	}, nil
}

func (ac AdminClient) editSiteReplicationInfo(ctx context.Context, site madmin.PeerInfo) (*madmin.ReplicateEditStatus, error) {
	res, err := ac.Client.SiteReplicationEdit(ctx, site)
	if err != nil {
		return nil, err
	}
	return &madmin.ReplicateEditStatus{
		Success:   res.Success,
		Status:    res.Status,
		ErrDetail: res.ErrDetail,
	}, nil
}

func (ac AdminClient) deleteSiteReplicationInfo(ctx context.Context, removeReq madmin.SRRemoveReq) (*madmin.ReplicateRemoveStatus, error) {
	res, err := ac.Client.SiteReplicationRemove(ctx, removeReq)
	if err != nil {
		return nil, err
	}
	return &madmin.ReplicateRemoveStatus{
		Status:    res.Status,
		ErrDetail: res.ErrDetail,
	}, nil
}

func (ac AdminClient) getSiteReplicationStatus(ctx context.Context, params madmin.SRStatusOptions) (*madmin.SRStatusInfo, error) {
	res, err := ac.Client.SRStatusInfo(ctx, params)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ac AdminClient) kmsStatus(ctx context.Context) (madmin.KMSStatus, error) {
	return ac.Client.KMSStatus(ctx)
}

func (ac AdminClient) kmsMetrics(ctx context.Context) (*madmin.KMSMetrics, error) {
	return ac.Client.KMSMetrics(ctx)
}

func (ac AdminClient) kmsAPIs(ctx context.Context) ([]madmin.KMSAPI, error) {
	return ac.Client.KMSAPIs(ctx)
}

func (ac AdminClient) kmsVersion(ctx context.Context) (*madmin.KMSVersion, error) {
	return ac.Client.KMSVersion(ctx)
}

func (ac AdminClient) createKey(ctx context.Context, key string) error {
	return ac.Client.CreateKey(ctx, key)
}

func (ac AdminClient) importKey(ctx context.Context, key string, content []byte) error {
	return ac.Client.ImportKey(ctx, key, content)
}

func (ac AdminClient) listKeys(ctx context.Context, pattern string) ([]madmin.KMSKeyInfo, error) {
	return ac.Client.ListKeys(ctx, pattern)
}

func (ac AdminClient) keyStatus(ctx context.Context, key string) (*madmin.KMSKeyStatus, error) {
	return ac.Client.GetKeyStatus(ctx, key)
}

func (ac AdminClient) deleteKey(ctx context.Context, key string) error {
	return ac.Client.DeleteKey(ctx, key)
}

func (ac AdminClient) setKMSPolicy(ctx context.Context, policy string, content []byte) error {
	return ac.Client.SetKMSPolicy(ctx, policy, content)
}

func (ac AdminClient) assignPolicy(ctx context.Context, policy string, content []byte) error {
	return ac.Client.AssignPolicy(ctx, policy, content)
}

func (ac AdminClient) describePolicy(ctx context.Context, policy string) (*madmin.KMSDescribePolicy, error) {
	return ac.Client.DescribePolicy(ctx, policy)
}

func (ac AdminClient) getKMSPolicy(ctx context.Context, policy string) (*madmin.KMSPolicy, error) {
	return ac.Client.GetPolicy(ctx, policy)
}

func (ac AdminClient) listKMSPolicies(ctx context.Context, pattern string) ([]madmin.KMSPolicyInfo, error) {
	return ac.Client.ListPolicies(ctx, pattern)
}

func (ac AdminClient) deletePolicy(ctx context.Context, policy string) error {
	return ac.Client.DeletePolicy(ctx, policy)
}

func (ac AdminClient) describeIdentity(ctx context.Context, identity string) (*madmin.KMSDescribeIdentity, error) {
	return ac.Client.DescribeIdentity(ctx, identity)
}

func (ac AdminClient) describeSelfIdentity(ctx context.Context) (*madmin.KMSDescribeSelfIdentity, error) {
	return ac.Client.DescribeSelfIdentity(ctx)
}

func (ac AdminClient) deleteIdentity(ctx context.Context, identity string) error {
	return ac.Client.DeleteIdentity(ctx, identity)
}

func (ac AdminClient) listIdentities(ctx context.Context, pattern string) ([]madmin.KMSIdentityInfo, error) {
	return ac.Client.ListIdentities(ctx, pattern)
}

func (ac AdminClient) addOrUpdateIDPConfig(ctx context.Context, idpType, cfgName, cfgData string, update bool) (restart bool, err error) {
	return ac.Client.AddOrUpdateIDPConfig(ctx, idpType, cfgName, cfgData, update)
}

func (ac AdminClient) listIDPConfig(ctx context.Context, idpType string) ([]madmin.IDPListItem, error) {
	return ac.Client.ListIDPConfig(ctx, idpType)
}

func (ac AdminClient) deleteIDPConfig(ctx context.Context, idpType, cfgName string) (restart bool, err error) {
	return ac.Client.DeleteIDPConfig(ctx, idpType, cfgName)
}

func (ac AdminClient) getIDPConfig(ctx context.Context, idpType, cfgName string) (c madmin.IDPConfig, err error) {
	return ac.Client.GetIDPConfig(ctx, idpType, cfgName)
}
