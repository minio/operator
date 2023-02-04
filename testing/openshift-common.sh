#!/usr/bin/env bash
# Copyright (C) 2023, MinIO, Inc.
#
# This code is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License, version 3,
# as published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License, version 3,
# along with this program.  If not, see <http://www.gnu.org/licenses/>

OPERATOR_SDK_VERSION=v1.22.2
ARCH=`{ case "$(uname -m)" in "x86_64") echo -n "amd64";; "aarch64") echo -n "arm64";; *) echo -n "$(uname -m)";; esac; }`
MACHINE="$(uname -m)"
OS=$(uname | awk '{print tolower($0)}')
export TMP_BIN_DIR="$(mktemp -d)"

function install_binaries() {

  echo -e "\e[34mInstalling temporal binaries in $TMP_BIN_DIR\e[0m"
  
  echo "kubectl"
  curl -#L "https://dl.k8s.io/release/v1.23.1/bin/$OS/$ARCH/kubectl" -o $TMP_BIN_DIR/kubectl
  chmod +x $TMP_BIN_DIR/kubectl

  echo "mc"
  curl -#L "https://dl.min.io/client/mc/release/${OS}-${ARCH}/mc" -o $TMP_BIN_DIR/mc
  chmod +x $TMP_BIN_DIR/mc

  echo "yq"
  curl -#L "https://github.com/mikefarah/yq/releases/latest/download/yq_${OS}_${ARCH}" -o $TMP_BIN_DIR/yq
  chmod +x $TMP_BIN_DIR/yq

  # latest kubectl and oc
  # curl -#L "https://mirror.openshift.com/pub/openshift-v4/$MACHINE/clients/ocp/stable/openshift-client-$OS.tar.gz" -o $TMP_BIN_DIR/openshift-client-$OS.tar.gz
  # tar -zxvf openshift-client-$OS.tar.gz

  echo "opm"
  curl -#L "https://mirror.openshift.com/pub/openshift-v4/$MACHINE/clients/ocp/stable/opm-$OS.tar.gz" -o $TMP_BIN_DIR/opm-$OS.tar.gz
  tar -zxf $TMP_BIN_DIR/opm-$OS.tar.gz -C $TMP_BIN_DIR/
  chmod +x $TMP_BIN_DIR/opm

  echo "crc"
  curl -#L "https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/crc/latest/crc-$OS-$ARCH.tar.xz" -o $TMP_BIN_DIR/crc-$OS-$ARCH.tar.xz
  tar -xJf $TMP_BIN_DIR/crc-$OS-$ARCH.tar.xz -C $TMP_BIN_DIR/ --strip-components=1
  chmod +x $TMP_BIN_DIR/crc

  echo "operator-sdk"
  curl -#L "https://github.com/operator-framework/operator-sdk/releases/download/$OPERATOR_SDK_VERSION/operator-sdk_${OS}_${ARCH}" -o ${TMP_BIN_DIR}/operator-sdk
  chmod +x $TMP_BIN_DIR/operator-sdk
}

function setup_path(){
  export PATH="$TMP_BIN_DIR:$PATH"
}

function remove_temp_binaries() {
  echo -e "\e[34mRemoving temporary binaries in: $TMP_BIN_DIR\e[0m"
  rm -rf $TMP_BIN_DIR
}

yell() { echo "$0: $*" >&2; }

die() {
  yell "$*"
  destroy_crc && exit 111
}

try() { "$@" || die "cannot $*"; }

function setup_crc() {
  echo -e "\e[34mConfiguring crc\e[0m"
  setup_path
  crc config set consent-telemetry no
  crc config set skip-check-root-user true
  crc config set kubeadmin-password "crclocal"
  crc config set cpus 8
  crc setup
  eval $(crc oc-env)
  eval $(crc oc-podman)
  crc start -m 20480
  echo "crc is ready"
  try crc version
}

function destoy_crc() {
  echo -e "\e[34mdestroy_crc\e[0m"
  setup_path
  #crc stop
  #crc delete -f
  remove_temp_binaries
}

function create_marketplace_catalog(){

  echo -e "\e[34mCreate Marketplace catalog\e[0m"
  # To compile current branch
  echo "Compiling Current Branch Operator"
  (cd "${SCRIPT_DIR}/.." && make operator && make logsearchapi && podman build --no-cache -t quay.io/minio/operator:noop .)
 echo "Compiling operator bundle"
  # Compile bundle image https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/bundle-directory
  podman build --no-cache -t quay.io/minio/operator-bundle:noop -f openshift/bundle.Dockerfile .

  echo "Installing Current Operator"
  # https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/openshift-deployment

  #login in crc registry
  podman login -u `oc whoami` -p `oc whoami --show-token` default-route-openshift-image-registry.apps-crc.testing --tls-verify=false
  
  # create and push test marketplace listing to crc registry
  podman tag quay.io/minio/operator-index:noop default-route-openshift-image-registry.apps-crc.testing/openshift-marketplace/operator-bundle:noop
  podman push default-route-openshift-image-registry.apps-crc.testing/openshift-marketplace/operator-bundle:noop --tls-verify=false
  oc set image-lookup operator-bundle -n openshift-marketplace
  opm index add --bundles default-route-openshift-image-registry.apps-crc.testing/openshift-marketplace/operator-bundle:noop \
    --tag default-route-openshift-image-registry.apps-crc.testing/openshift-marketplace/minio-operator-index:latest \
    --skip-tls-verify=true
  podman push default-route-openshift-image-registry.apps-crc.testing/openshift-marketplace/minio-operator-index:latest --tls-verify=false
  oc set image-lookup minio-operator-index -n openshift-marketplace
  
  #push operator image to crc registry
  podman tag quay.io/minio/operator:noop default-route-openshift-image-registry.apps-crc.testing/openshift-operators/operator:noop
  podman push default-route-openshift-image-registry.apps-crc.testing/openshift-operators/operator:noop --tls-verify=false
  oc get is -n openshift-operators
  oc set image-lookup operator -n openshift-operators
  #oc registry login --insecure=true
  #oc image mirror quay.io/minio/operator-index:latest=default-route-openshift-image-registry.apps-crc.testing/openshift-operators/minio/operator-index:latest --insecure=true

  #create local marketplace listing
  oc apply -f ./openshift/test-operator-catalogsource.yaml
  # oc -n openshift-marketplace get catalogsource minio-test-operators
  # oc -n openshift-marketplace get pods
  # oc get packagemanifests | grep "Test Minio Operators"
  ## Create a operator group is not needed here because the openshift-operators namespace already have one
  # oc create -f ./openshift/test-operatorgroup.yaml
  # oc get og -n openshift-operators
}

function install_operator() {
  echo -e "\e[34mInstalling Operator from bundle\e[0m"

  oc apply -f ./openshift/test-subscription.yaml
  oc get sub -n openshift-operators
  oc get installplan -n openshift-operators
  oc get csv -n openshift-operators

  echo "key, value for pod selector in kustomize test"
  key=name
  value=minio-operator

  oc -n openshift-operators get deployments
  oc -n openshift-operators get pods

  echo "Waiting for Operator Pods to come online (5m timeout)"
  try oc wait -n openshift-operators \
    --for=condition=ready pod \
    --selector $key=$value \
    --timeout=300s

  echo "start - get data to verify proper image is being used"
  oc get pods --namespace openshift-operators
  oc describe pods -n openshift-operators | grep Image
  echo "end - get data to verify proper image is being used"
}