module github.com/minio/operator

go 1.13

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/georgysavva/scany v0.2.7 // indirect
	github.com/google/go-containerregistry v0.1.2
	github.com/gorilla/mux v1.8.0
	github.com/jackc/pgx/v4 v4.10.0 // indirect
	github.com/minio/minio v0.0.0-20201203193910-919441d9c4d2
	github.com/minio/minio-go/v7 v7.0.6
	github.com/secure-io/sio-go v0.3.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.6
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/controller-tools v0.4.1 // indirect
	sigs.k8s.io/kind v0.9.0 // indirect
)
