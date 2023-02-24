package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	tenantEndpoint := os.Getenv("MINIO_ENDPOINT")
	stsEndpoint := os.Getenv("STS_ENDPOINT")
	sessionPolicyFile := os.Getenv("STS_POLICY")

	token, err := getToken()
	if err != nil {
		log.Fatalf("Could not get Service account JWT: %s", err)
		panic(1)
	}
	if token == "" {
		log.Fatalf("Service account JWT is empty", err)
		panic(1)
	}

	var stsOpts credentials.STSAssumeRoleOptions

	if sessionPolicyFile != "" {
		var policy string
		if f, err := os.Open(sessionPolicyFile); err != nil {
			log.Fatalf("Unable to open session policy file: %v", err)
		} else {
			bs, err := io.ReadAll(f)
			if err != nil {
				log.Fatalf("Error reading session policy file: %v", err)
			}
			policy = string(bs)
		}
		stsOpts.Policy = policy
	}

	stsEndpointURL, err := url.Parse(stsEndpoint)
	if err != nil {
		log.Fatalf("Error parsing sts endpoint: %v", err)
	}
	sts, err := credentials.NewSTSAssumeRole(stsEndpointURL.String(), stsOpts)
	if err != nil {
		log.Fatalf("Error initializing STS Identity: %v", err)
	}
	retrievedCredentials, err := sts.Get()
	if err != nil {
		log.Fatalf("Error retrieving STS credentials: %v", err)
	}
	fmt.Println("AccessKeyID:", retrievedCredentials.AccessKeyID)
	fmt.Println("SecretAccessKey:", retrievedCredentials.SecretAccessKey)
	fmt.Println("SessionToken:", retrievedCredentials.SessionToken)

	tenantEndpointURL, err := url.Parse(tenantEndpoint)
	if err != nil {
		log.Fatalf("Error parsing tenant endpoint: %v", err)
	}

	minioClient, err := minio.New(tenantEndpointURL.Host, &minio.Options{
		Creds:  sts,
		Secure: tenantEndpointURL.Scheme == "https",
	})

	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}

	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    "/",
		Recursive: true,
	}

	for object := range minioClient.ListObjects(context.Background(), "test-bucket", opts) {
		if object.Err != nil {
			fmt.Println(object.Err)
			panic(1)
		}
		fmt.Println(object)
	}

}

func getToken() (string, error) {
	tokenpath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	fileContent, err := ioutil.ReadFile(tokenpath)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

func getKubernetesCACertificates() (string, error) {
	kubeRootCApath := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	operatorSTSCApath := os.Getenv("STS_CA_PATH")

}
