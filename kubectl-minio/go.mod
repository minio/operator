module github.com/minio/kubectl-minio

go 1.13

replace github.com/minio/operator => ../

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/briandowns/spinner v1.12.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-openapi/errors v0.19.6 // indirect
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/minio/minio v0.0.0-20201203193910-919441d9c4d2
	github.com/minio/operator v0.3.23
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/montanaflynn/stats v0.6.3 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.0.0
	go.etcd.io/etcd/v3 v3.3.0-rc.0.0.20200707003333-58bb8ae09f8e // indirect
	go.mongodb.org/mongo-driver v1.3.5 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/cli-runtime v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
	sigs.k8s.io/yaml v1.2.0
)
