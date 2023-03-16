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
	"crypto/x509"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/minio/operator/pkg/auth/idp/oauth2"

	xcerts "github.com/minio/pkg/certs"

	"github.com/minio/pkg/env"
)

var (
	// Port console default port
	Port = "9090"

	// Hostname console hostname
	// avoid listening on 0.0.0.0 by default
	// instead listen on all IPv4 and IPv6
	// - Hostname should be empty.
	Hostname = ""

	// TLSPort console tls port
	TLSPort = "9443"

	// TLSRedirect console tls redirect rule
	TLSRedirect = "on"

	// SessionDuration cookie validity duration
	SessionDuration = 45 * time.Minute

	// LicenseKey in memory license key used by console ui
	LicenseKey = ""
	// GlobalRootCAs is CA root certificates, a nil value means system certs pool will be used
	GlobalRootCAs *x509.CertPool
	// GlobalPublicCerts has certificates Console will use to serve clients
	GlobalPublicCerts []*x509.Certificate
	// GlobalTLSCertsManager custom TLS Manager for SNI support
	GlobalTLSCertsManager *xcerts.Manager
)

// getK8sSAToken assumes the plugin is running inside a k8s pod and extract the current service account from the
// /var/run/secrets/kubernetes.io/serviceaccount/token file
func getK8sSAToken() string {
	dat, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return env.Get(OperatorSAToken, "")
	}
	return string(dat)
}

// Get Marketplace deployment platform
func getMarketplace() string {
	return env.Get(Marketplace, "")
}

// MinIOConfig represents application configuration passed in from the MinIO
// server to the console.
type MinIOConfig struct {
	OpenIDProviders oauth2.OpenIDPCfg
}

// GetHostname gets console hostname set on env variable,
// default one or defined on run command
func GetHostname() string {
	return strings.ToLower(env.Get(OperatorUIHostname, Hostname))
}

// GetPort gets console por set on env variable
// or default one
func GetPort() int {
	port, err := strconv.Atoi(env.Get(OperatorUIPort, Port))
	if err != nil {
		port = 9090
	}
	return port
}

// GetTLSPort gets console tls port set on env variable
// or default one
func GetTLSPort() int {
	port, err := strconv.Atoi(env.Get(OperatorUITLSPort, TLSPort))
	if err != nil {
		port = 9443
	}
	return port
}

// GetTLSRedirect if is set to true, then only allow HTTPS requests. Default is true.
func GetTLSRedirect() string {
	return strings.ToLower(env.Get(SecureTLSRedirect, TLSRedirect))
}

// GetSecureAllowedHosts secure middleware env variable configurations
func GetSecureAllowedHosts() []string {
	allowedHosts := env.Get(SecureAllowedHosts, "")
	if allowedHosts != "" {
		return strings.Split(allowedHosts, ",")
	}
	return []string{}
}

// GetSecureAllowedHostsAreRegex determines, if the provided AllowedHosts slice contains valid regular expressions. Default is false.
func GetSecureAllowedHostsAreRegex() bool {
	return strings.ToLower(env.Get(SecureAllowedHostsAreRegex, "off")) == "on"
}

// GetSecureFrameDeny If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is true.
func GetSecureFrameDeny() bool {
	return strings.ToLower(env.Get(SecureFrameDeny, "on")) == "on"
}

// GetSecureContentTypeNonSniff If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is true.
func GetSecureContentTypeNonSniff() bool {
	return strings.ToLower(env.Get(SecureContentTypeNoSniff, "on")) == "on"
}

// GetSecureBrowserXSSFilter If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is true.
func GetSecureBrowserXSSFilter() bool {
	return strings.ToLower(env.Get(SecureBrowserXSSFilter, "on")) == "on"
}

// GetSecureContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. Default is "".
// Passing a template string will replace `$NONCE` with a dynamic nonce value of 16 bytes for each request which can be
// later retrieved using the Nonce function.
func GetSecureContentSecurityPolicy() string {
	return env.Get(SecureContentSecurityPolicy, "")
}

// GetSecureContentSecurityPolicyReportOnly allows the Content-Security-Policy-Report-Only header value to be set with a custom value. Default is "".
func GetSecureContentSecurityPolicyReportOnly() string {
	return env.Get(SecureContentSecurityPolicyReportOnly, "")
}

// GetSecureHostsProxyHeaders is a set of header keys that may hold a proxied hostname value for the request.
func GetSecureHostsProxyHeaders() []string {
	allowedHosts := env.Get(SecureHostsProxyHeaders, "")
	if allowedHosts != "" {
		return strings.Split(allowedHosts, ",")
	}
	return []string{}
}

// GetSecureTLSHost is the host name that is used to redirect HTTP requests to HTTPS. Default is "", which indicates to use the same host.
func GetSecureTLSHost() string {
	tlsHost := env.Get(SecureTLSHost, "")
	if tlsHost == "" && Hostname != "" {
		return net.JoinHostPort(Hostname, TLSPort)
	}
	return ""
}

// GetSecureSTSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
func GetSecureSTSSeconds() int64 {
	seconds, err := strconv.Atoi(env.Get(SecureSTSSeconds, "0"))
	if err != nil {
		seconds = 0
	}
	return int64(seconds)
}

// GetSecureSTSIncludeSubdomains If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
func GetSecureSTSIncludeSubdomains() bool {
	return strings.ToLower(env.Get(SecureSTSIncludeSubdomains, "off")) == "on"
}

// GetSecureSTSPreload If STSPreload is set to true, the `preload` flag will be appended to the Strict-Transport-Security header. Default is false.
func GetSecureSTSPreload() bool {
	return strings.ToLower(env.Get(SecureSTSPreload, "off")) == "on"
}

// GetSecureTLSTemporaryRedirect If TLSTemporaryRedirect is true, the a 302 will be used while redirecting. Default is false (301).
func GetSecureTLSTemporaryRedirect() bool {
	return strings.ToLower(env.Get(SecureTLSTemporaryRedirect, "off")) == "on"
}

// GetSecureForceSTSHeader STS header is only included when the connection is HTTPS.
func GetSecureForceSTSHeader() bool {
	return strings.ToLower(env.Get(SecureForceSTSHeader, "off")) == "on"
}

// GetSecurePublicKey PublicKey implements HPKP to prevent MITM attacks with forged certificates. Default is "".
func GetSecurePublicKey() string {
	return env.Get(SecurePublicKey, "")
}

// GetSecureReferrerPolicy ReferrerPolicy allows the Referrer-Policy header with the value to be set with a custom value. Default is "".
func GetSecureReferrerPolicy() string {
	return env.Get(SecureReferrerPolicy, "")
}

// GetSecureFeaturePolicy FeaturePolicy allows the Feature-Policy header with the value to be set with a custom value. Default is "".
func GetSecureFeaturePolicy() string {
	return env.Get(SecureFeaturePolicy, "")
}

// GetSecureExpectCTHeader header
func GetSecureExpectCTHeader() string {
	return env.Get(SecureExpectCTHeader, "")
}
