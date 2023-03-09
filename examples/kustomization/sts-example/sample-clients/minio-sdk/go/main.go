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

package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	tenantEndpoint := os.Getenv("MINIO_ENDPOINT")
	stsEndpoint := os.Getenv("STS_ENDPOINT")
	tenantNamespace := os.Getenv("TENANT_NAMESPACE")
	bucketName := os.Getenv("BUCKET")
	kubeRootCApath := os.Getenv("KUBERNETES_CA_PATH")

	token, err := getToken()
	if err != nil {
		log.Fatalf("Could not get Service account JWT: %s", err)
		panic(1)
	}
	if token == "" {
		log.Fatal("Service account JWT is empty")
		panic(1)
	}
	// Declare a custom transport to trust custom CA's, in this case we will trust
	// Kubernete's Internal CA or Cert Manager's CA
	httpsTransport, err := getHttpsTransportWithCACert(kubeRootCApath)
	if err != nil {
		log.Fatalf("Error Creating https transport: %s", err)
		panic(1)
	}

	stsEndpointURL, err := url.Parse(stsEndpoint)
	stsEndpointURL.Path = path.Join(stsEndpointURL.Path, tenantNamespace)
	if err != nil {
		log.Fatalf("Error parsing sts endpoint: %v", err)
		panic(1)
	}

	sts := credentials.New(&credentials.IAM{
		Client: &http.Client{
			Transport: httpsTransport,
		},
		Endpoint: stsEndpointURL.String(),
	})

	retrievedCredentials, err := sts.Get()
	if err != nil {
		log.Fatalf("Error retrieving STS credentials: %v", err)
		panic(1)
	}
	fmt.Println("AccessKeyID:", retrievedCredentials.AccessKeyID)
	fmt.Println("SecretAccessKey:", retrievedCredentials.SecretAccessKey)
	fmt.Println("SessionToken:", retrievedCredentials.SessionToken)

	tenantEndpointURL, err := url.Parse(tenantEndpoint)
	if err != nil {
		log.Fatalf("Error parsing tenant endpoint: %s", err)
		panic(1)
	}

	minioClient, err := minio.New(tenantEndpointURL.Host, &minio.Options{
		Creds:     sts,
		Secure:    tenantEndpointURL.Scheme == "https",
		Transport: httpsTransport,
	})
	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
		panic(1)
	}

	fmt.Print("List Buckets:")
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, bucket := range buckets {
		log.Println(bucket)
	}

	fmt.Printf("List Objects in bucket %s", bucketName)
	opts := minio.ListObjectsOptions{
		Prefix:    "/",
		Recursive: true,
	}
	for object := range minioClient.ListObjects(context.Background(), bucketName, opts) {
		if object.Err != nil {
			fmt.Println(object.Err)
			panic(1)
		}
		fmt.Println(object)
	}
	return
}

func getToken() (string, error) {
	tokenpath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	fileContent, err := ioutil.ReadFile(tokenpath)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

func getFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func getHttpsTransportWithCACert(cacertpath string) (*http.Transport, error) {
	caCertificate, err := getFile(cacertpath)
	if err != nil {
		return nil, fmt.Errorf("Error loading CA Certifiate : %s", err)
	}

	transport, err := minio.DefaultTransport(true)
	if err != nil {
		return nil, fmt.Errorf("Error creating default transport : %s", err)
	}

	if transport.TLSClientConfig.RootCAs == nil {
		pool, err := x509.SystemCertPool()
		if err != nil {
			log.Fatalf("Error initializing TLS Pool: %s", err)
			transport.TLSClientConfig.RootCAs = x509.NewCertPool()
		} else {
			transport.TLSClientConfig.RootCAs = pool
		}
	}

	if ok := transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(caCertificate); !ok {
		return nil, fmt.Errorf("Error parsing CA Certifiate : %s", err)
	}
	return transport, nil
}
