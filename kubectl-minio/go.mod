module github.com/minio/kubectl-minio

go 1.15

replace github.com/minio/operator => ../

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/google/uuid v1.1.2
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/minio/minio v0.0.0-20210128013121-e79829b5b368
	github.com/minio/operator v0.3.23
	github.com/montanaflynn/stats v0.6.3 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/yaml v1.2.0
)
