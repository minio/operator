#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
ROOT_PKG=github.com/minio/operator

# Grab code-generator version from go.sum
CODEGEN_VERSION=$(grep 'k8s.io/code-generator' go.mod | awk '{print $2}' | sed 's/\/go.mod//g' | head -1)
GOPATH=$(go env GOPATH)
CODEGEN_PKG="${GOPATH}/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}"

if [[ ! -d ${CODEGEN_PKG} ]]; then
    echo "${CODEGEN_PKG} is missing. Running 'go mod download'."
    go mod download
fi

echo ">> Using ${CODEGEN_PKG}"

source ${CODEGEN_PKG}/kube_codegen.sh

kube::codegen::gen_helpers $SCRIPT_ROOT/pkg/apis \
    --boilerplate "k8s/boilerplate.go.txt"

kube::codegen::gen_client $SCRIPT_ROOT/pkg/apis \
    --with-watch \
    --with-applyconfig \
    --output-dir "./pkg/client" \
    --output-pkg "$ROOT_PKG/pkg/client" \
    --boilerplate "k8s/boilerplate.go.txt" || echo "Failed"