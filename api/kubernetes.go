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
	"net"
	"strings"

	operator "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/minio/pkg/env"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
)

// GetK8sAPIServer returns the URL to use for the k8s api
func GetK8sAPIServer() string {
	// if console is running inside a k8s pod KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT will contain the k8s api server apiServerAddress
	// if console is not running inside k8s by default will look for the k8s api server on localhost:8001 (kubectl proxy)
	// NOTE: using kubectl proxy is for local development only, since every request send to localhost:8001 will bypass service account authentication
	// more info here: https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#directly-accessing-the-rest-api
	// you can override this using CONSOLE_K8S_API_SERVER, ie use the k8s cluster from `kubectl config view`
	host, port := env.Get("KUBERNETES_SERVICE_HOST", ""), env.Get("KUBERNETES_SERVICE_PORT", "")
	apiServerAddress := "http://localhost:8001"
	if host != "" && port != "" {
		apiServerAddress = "https://" + net.JoinHostPort(host, port)
	}
	return env.Get(K8sAPIServer, apiServerAddress)
}

// If CONSOLE_K8S_API_SERVER_TLS_ROOT_CA is true console will load the certificate into the
// http.client rootCAs pool, this is useful for testing an k8s ApiServer or when working with self-signed certificates
func getK8sAPIServerTLSRootCA() string {
	return strings.TrimSpace(env.Get(K8SAPIServerTLSRootCA, ""))
}

// getTLSClientConfig will return the right TLS configuration for the K8S client based on the configured TLS certificate
func getTLSClientConfig() rest.TLSClientConfig {
	defaultRootCAFile := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	customRootCAFile := getK8sAPIServerTLSRootCA()
	tlsClientConfig := rest.TLSClientConfig{}
	// if console is running inside k8s by default he will have access to the CA Cert from the k8s local authority
	if _, err := certutil.NewPool(defaultRootCAFile); err == nil {
		tlsClientConfig.CAFile = defaultRootCAFile
	}
	// if the user explicitly define a custom CA certificate, instead, we will use that
	if customRootCAFile != "" {
		if _, err := certutil.NewPool(customRootCAFile); err == nil {
			tlsClientConfig.CAFile = customRootCAFile
		}
	}
	return tlsClientConfig
}

// This operation will run only once at console startup
var tlsClientConfig = getTLSClientConfig()

// GetK8sConfig returns the config for k8s api
func GetK8sConfig(token string) *rest.Config {
	config := &rest.Config{
		Host:            GetK8sAPIServer(),
		TLSClientConfig: tlsClientConfig,
		APIPath:         "/",
		BearerToken:     token,
	}
	return config
}

// GetOperatorClient returns an operator client using GetK8sConfig for its config
func GetOperatorClient(token string) (*operator.Clientset, error) {
	return operator.NewForConfig(GetK8sConfig(token))
}

// K8sClient returns kubernetes client using GetK8sConfig for its config
func K8sClient(token string) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(GetK8sConfig(token))
}
