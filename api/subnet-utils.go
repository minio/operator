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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/madmin-go/v2"
	mc "github.com/minio/mc/cmd"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	subnetRespBodyLimit = 1 << 20 // 1 MiB
)

var (
	subnetBaseURLEnvVar = "SUBNET_BASE_URL"
	httpClient          *http.Client
)

// LicenseTokenConfig stores subnet license information
type LicenseTokenConfig struct {
	APIKey  string
	License string
	Proxy   string
}

type subnetMFAReq struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
	Token    string `json:"token"`
}

// BaseURL returns subnet base url
func BaseURL() string {
	url := os.Getenv(subnetBaseURLEnvVar)
	if url != "" {
		return url
	}
	return "https://subnet.min.io"
}

// RegisterURL returns subnet register url
func RegisterURL() string {
	return BaseURL() + "/api/cluster/register"
}

// LoginURL returns subnet login url
func LoginURL() string {
	return BaseURL() + "/api/auth/login"
}

// MFAURL returns subnet mfa url
func MFAURL() string {
	return BaseURL() + "/api/auth/mfa-login"
}

// KeyURL returns subnet api key url
func KeyURL() string {
	return BaseURL() + "/api/auth/api-key"
}

// GetHTTPClient returns a http client to communicate with subnet
func GetHTTPClient() *http.Client {
	if httpClient == nil {
		httpClient = prepareHTTPClient(false)
	}
	return httpClient
}

func prepareHTTPClient(insecure bool) *http.Client {
	return &http.Client{Transport: PrepareClientTransport(insecure)}
}

// PrepareClientTransport prepares http transport
func PrepareClientTransport(insecure bool) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 15 * time.Second,
	}
	DefaultTransport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		DisableCompression:    true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
	return DefaultTransport
}

// RegisterWithAPIKey registers minio instance in subnet
func RegisterWithAPIKey(admInfo madmin.InfoMessage, apiKey string) (*LicenseTokenConfig, error) {
	regInfo := getClusterRegInfo(admInfo)
	regURL := fmt.Sprintf("%s?api_key=%s", RegisterURL(), apiKey)
	regToken, err := generateRegToken(regInfo)
	if err != nil {
		return nil, err
	}
	reqPayload := mc.ClusterRegistrationReq{Token: regToken}
	resp, err := subnetPostReq(regURL, reqPayload, map[string]string{})
	if err != nil {
		return nil, err
	}
	respJSON := gjson.Parse(resp)
	subnetAPIKey := respJSON.Get("api_key").String()
	licenseJwt := respJSON.Get("license").String()

	if subnetAPIKey != "" || licenseJwt != "" {
		return &LicenseTokenConfig{
			APIKey:  subnetAPIKey,
			License: licenseJwt,
		}, nil
	}
	return nil, errors.New("subnet api key not found")
}

func getClusterRegInfo(admInfo madmin.InfoMessage) mc.ClusterRegistrationInfo {
	noOfPools := 1
	noOfDrives := 0
	for _, srvr := range admInfo.Servers {
		if srvr.PoolNumber > noOfPools {
			noOfPools = srvr.PoolNumber
		}
		noOfDrives += len(srvr.Disks)
	}

	totalSpace, usedSpace := getDriveSpaceInfo(admInfo)

	return mc.ClusterRegistrationInfo{
		DeploymentID: admInfo.DeploymentID,
		ClusterName:  admInfo.DeploymentID,
		UsedCapacity: admInfo.Usage.Size,
		Info: mc.ClusterInfo{
			MinioVersion:    admInfo.Servers[0].Version,
			NoOfServerPools: noOfPools,
			NoOfServers:     len(admInfo.Servers),
			NoOfDrives:      noOfDrives,
			TotalDriveSpace: totalSpace,
			UsedDriveSpace:  usedSpace,
			NoOfBuckets:     admInfo.Buckets.Count,
			NoOfObjects:     admInfo.Objects.Count,
		},
	}
}

func getDriveSpaceInfo(admInfo madmin.InfoMessage) (uint64, uint64) {
	total := uint64(0)
	used := uint64(0)
	for _, srvr := range admInfo.Servers {
		for _, d := range srvr.Disks {
			total += d.TotalSpace
			used += d.UsedSpace
		}
	}
	return total, used
}

func generateRegToken(clusterRegInfo mc.ClusterRegistrationInfo) (string, error) {
	token, e := json.Marshal(clusterRegInfo)
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

// GetAPIKey returns api key from subnet
func GetAPIKey(token string) (string, error) {
	resp, err := subnetGetReq(KeyURL(), subnetAuthHeaders(token))
	if err != nil {
		return "", err
	}
	respJSON := gjson.Parse(resp)
	apiKey := respJSON.Get("api_key").String()
	return apiKey, nil
}

// GetSubnetKeyFromMinIOConfig return subnet config stored in minio
func GetSubnetKeyFromMinIOConfig(ctx context.Context, adminClient MinioAdmin) (*LicenseTokenConfig, error) {
	buf, err := adminClient.getConfigKV(ctx, "subnet")
	if err != nil {
		return nil, err
	}
	tgt, err := madmin.ParseServerConfigOutput(string(buf))
	if err != nil {
		return nil, err
	}
	res := LicenseTokenConfig{}
	for _, t := range tgt {
		for _, kv := range t.KV {
			switch kv.Key {
			case "api_key":
				res.APIKey = kv.Value
			case "license":
				res.License = kv.Value
			case "proxy":
				res.Proxy = kv.Value
			}
		}
	}
	return &res, nil
}

// Login helper function to login to subnet in a terminal
func Login() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("SUBNET username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if len(username) == 0 {
		return "", errors.New("username cannot be empty")
	}

	fmt.Print("Password: ")
	bytepw, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	loginReq := map[string]string{
		"username": username,
		"password": string(bytepw),
	}
	respStr, e := subnetPostReq(LoginURL(), loginReq, nil)
	if e != nil {
		return "", e
	}

	mfaRequired := gjson.Get(respStr, "mfa_required").Bool()
	if mfaRequired {
		mfaToken := gjson.Get(respStr, "mfa_token").String()
		fmt.Print("OTP received in email: ")
		byteotp, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()

		mfaLoginReq := subnetMFAReq{Username: username, OTP: string(byteotp), Token: mfaToken}
		respStr, e = subnetPostReq(MFAURL(), mfaLoginReq, nil)
		if e != nil {
			return "", e
		}
	}

	token := gjson.Get(respStr, "token_info.access_token")
	if token.Exists() {
		return token.String(), nil
	}
	return "", fmt.Errorf("access token not found in response")
}

func subnetAuthHeaders(authToken string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + authToken}
}

func subnetGetReq(reqURL string, headers map[string]string) (string, error) {
	r, e := http.NewRequest(http.MethodGet, reqURL, nil)
	if e != nil {
		return "", e
	}
	return subnetReqDo(r, headers)
}

func subnetPostReq(reqURL string, payload interface{}, headers map[string]string) (string, error) {
	body, e := json.Marshal(payload)
	if e != nil {
		return "", e
	}
	r, e := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if e != nil {
		return "", e
	}
	return subnetReqDo(r, headers)
}

func subnetReqDo(r *http.Request, headers map[string]string) (string, error) {
	for k, v := range headers {
		r.Header.Add(k, v)
	}

	ct := r.Header.Get("Content-Type")
	if len(ct) == 0 {
		r.Header.Add("Content-Type", "application/json")
	}

	resp, e := GetHTTPClient().Do(r)
	if e != nil {
		return "", e
	}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(io.LimitReader(resp.Body, subnetRespBodyLimit))
	if e != nil {
		return "", e
	}
	respStr := string(respBytes)

	if resp.StatusCode == http.StatusOK {
		return respStr, nil
	}
	return respStr, fmt.Errorf("request failed with code %d and error: %s", resp.StatusCode, respStr)
}
