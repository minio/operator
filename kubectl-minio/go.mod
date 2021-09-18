module github.com/minio/kubectl-minio

go 1.16

replace github.com/minio/operator => ../

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/google/uuid v1.1.2
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/minio/operator v0.0.0-00010101000000-000000000000
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/kustomize/api v0.8.5
	sigs.k8s.io/yaml v1.2.0
)
