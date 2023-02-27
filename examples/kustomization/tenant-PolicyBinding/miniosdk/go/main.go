/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2023 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
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
	// certManagerCAPath := os.Getenv("STS_CA_PATH")

	token, err := getToken()
	if err != nil {
		log.Fatalf("Could not get Service account JWT: %s", err)
		panic(1)
	}
	if token == "" {
		log.Fatal("Service account JWT is empty")
		panic(1)
	}

	stsEndpointURL, err := url.Parse(stsEndpoint)
	stsEndpointURL.Path = path.Join(stsEndpointURL.Path, tenantNamespace)
	if err != nil {
		log.Fatalf("Error parsing sts endpoint: %v", err)
	}
	sts := credentials.NewIAM(stsEndpointURL.String())

	if err != nil {
		log.Fatalf("Error initializing STS Identity: %v", err)
		panic(1)
	}
	// This might fail for https with self-signed certificates,
	// need to find a way  to set trust CA certificate to credentials.Credentials.Get()
	// retrievedCredentials, err := sts.Get()
	// if err != nil {
	// 	log.Fatalf("Error retrieving STS credentials: %v", err)
	// 	panic(1)
	// }
	// fmt.Println("AccessKeyID:", retrievedCredentials.AccessKeyID)
	// fmt.Println("SecretAccessKey:", retrievedCredentials.SecretAccessKey)
	// fmt.Println("SessionToken:", retrievedCredentials.SessionToken)

	tenantEndpointURL, err := url.Parse(tenantEndpoint)
	if err != nil {
		log.Fatalf("Error parsing tenant endpoint: %s", err)
		panic(1)
	}

	caCertificate, err := getFile(kubeRootCApath)
	if err != nil {
		log.Fatalf("Error loading CA Certifiate : %s", err)
		panic(1)
	}

	transport, err := minio.DefaultTransport(true)
	if err != nil {
		log.Fatalf("Error creating default transport : %s", err)
		panic(1)
	}

	if ok := transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(caCertificate); !ok {
		log.Fatalf("Error parsing CA Certifiate : %s", err)
		panic(1)
	}

	minioClient, err := minio.New(tenantEndpointURL.Host, &minio.Options{
		Creds:     sts,
		Secure:    true,
		Transport: transport,
	})

	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}

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
