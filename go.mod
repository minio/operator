module github.com/minio/operator

go 1.16

require (
	cloud.google.com/go v0.60.0 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/google/go-containerregistry v0.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-version v1.3.0
	github.com/minio/madmin-go v1.0.12
	github.com/minio/minio-go/v7 v7.0.11-0.20210302210017-6ae69c73ce78
	github.com/minio/pkg v1.0.11
	github.com/onsi/ginkgo v1.16.1 // indirect
	github.com/onsi/gomega v1.11.0 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.46.0
	github.com/shirou/gopsutil/v3 v3.21.5 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.2
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/controller-runtime v0.8.0
)
