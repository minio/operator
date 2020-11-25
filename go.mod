module github.com/minio/operator

go 1.13

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/google/go-containerregistry v0.1.2
	github.com/gorilla/mux v1.8.0
	github.com/minio/console v0.4.5
	github.com/minio/minio v0.0.0-20200927172404-27d9bd04e544
	github.com/minio/minio-go/v7 v7.0.6-0.20200923173112-bc846cb9b089
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog/v2 v2.3.0
)
