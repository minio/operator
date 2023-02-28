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

package cluster

import (
	"io/ioutil"
	"net"
	"strings"
	"time"

	xhttp "github.com/minio/console/pkg/http"
	"github.com/minio/console/restapi"

	"github.com/minio/console/pkg/utils"

	"github.com/minio/pkg/env"
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

// GetNsFromFile assumes console is running inside a k8s pod and extract the current namespace from the
// /var/run/secrets/kubernetes.io/serviceaccount/namespace file
func GetNsFromFile() string {
	dat, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "default"
	}
	return string(dat)
}

// GetMinioImage returns the image URL to be used when deploying a MinIO instance, if there is
// a preferred image to be used (configured via ENVIRONMENT VARIABLES) GetMinioImage will return that
// if not, GetMinioImage will try to obtain the image URL for the latest version of MinIO and return that
func GetMinioImage() (*string, error) {
	image := strings.TrimSpace(env.Get(MinioImage, ""))
	// if there is a preferred image configured by the user we'll always return that
	if image != "" {
		return &image, nil
	}
	client := restapi.GetConsoleHTTPClient("")
	client.Timeout = 5 * time.Second
	latestMinIOImage, errLatestMinIOImage := utils.GetLatestMinIOImage(
		&xhttp.Client{
			Client: client,
		})

	if errLatestMinIOImage != nil {
		return nil, errLatestMinIOImage
	}
	return latestMinIOImage, nil
}
