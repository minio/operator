module github.com/minio/operator

go 1.13

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/google/go-containerregistry v0.1.2
	github.com/gorilla/mux v1.8.0
	github.com/minio/minio v0.0.0-20201203193910-919441d9c4d2
	github.com/minio/minio-go/v7 v7.0.6
	github.com/secure-io/sio-go v0.3.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog/v2 v2.3.0
)
