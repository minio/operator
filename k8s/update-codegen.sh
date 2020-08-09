#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GO111MODULE=off go get -d k8s.io/code-generator/...

REPOSITORY=github.com/minio/operator
$(go env GOPATH)/src/k8s.io/code-generator/generate-groups.sh all \
                $REPOSITORY/pkg/client $REPOSITORY/pkg/apis \
                "minio.min.io:v1" \
                --go-header-file "boilerplate.go.txt"
