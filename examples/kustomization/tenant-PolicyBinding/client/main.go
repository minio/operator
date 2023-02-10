package main

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "minio.minio-tenant-1.svc.cluster"
	operatorEndpoint := "http://sts.minio-operator.svc.cluster.local:4222/sts/"
	useSSL := true

	sts, err := credentials.NewSTSWebIdentity(operatorEndpoint, getWebTokenExpiry)
	if err != nil {
		fmt.Println(fmt.Errorf("Could not get STS credentials: %s", err))
		return
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  sts,
		Secure: useSSL,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    "/",
		Recursive: true,
	}

	for object := range minioClient.ListObjects(context.Background(), "test-bucket", opts) {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}
	return
}
