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

# code-generator does work with go.mod but makes assumptions about
# the project living in `$GOPATH/src`. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
cd ${SCRIPT_ROOT}
${CODEGEN_PKG}/generate-groups.sh "all" \
              $ROOT_PKG/pkg/client $ROOT_PKG/pkg/apis \
              "minio.min.io:v2 sts.min.io:v1alpha1" \
              --output-base "${TEMP_DIR}" \
              --go-header-file "k8s/boilerplate.go.txt"

# Copy everything back.
cp -a "${TEMP_DIR}/${ROOT_PKG}/." "${SCRIPT_ROOT}/"
