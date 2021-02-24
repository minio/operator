module github.com/minio/operator

go 1.15

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/google/go-containerregistry v0.1.2
	github.com/gorilla/mux v1.8.0
	github.com/minio/minio v0.0.0-20210128013121-e79829b5b368
	github.com/minio/minio-go/v7 v7.0.8-0.20210127003153-c40722862654
	github.com/secure-io/sio-go v0.3.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/controller-runtime v0.8.0
)
