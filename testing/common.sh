#!/usr/bin/env bash
# Copyright (C) 2022, MinIO, Inc.
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

source "${GITHUB_WORKSPACE}/shared-functions/shared-code.sh" # This is common.sh for k8s tests across multiple repos.

## this enables :dev tag for minio/operator container image.
CI="true"
export CI
ARCH=`{ case "$(uname -m)" in "x86_64") echo -n "amd64";; "aarch64") echo -n "arm64";; *) echo -n "$(uname -m)";; esac; }`
OS=$(uname | awk '{print tolower($0)}')

DEV_TEST=$OPERATOR_DEV_TEST

# Set OPERATOR_DEV_TEST to skip downloading these dependencies
if [[ -z "${DEV_TEST}" ]]; then
  ## Make sure to install things if not present already
  sudo curl -#L "https://dl.k8s.io/release/v1.23.1/bin/$OS/$ARCH/kubectl" -o /usr/local/bin/kubectl
  sudo chmod +x /usr/local/bin/kubectl

  sudo curl -#L "https://dl.min.io/client/mc/release/${OS}-${ARCH}/mc" -o /usr/local/bin/mc
  sudo chmod +x /usr/local/bin/mc

  ## Install yq
  sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_${OS}_${ARCH}
  sudo chmod a+x /usr/local/bin/yq
fi

yell() { echo "$0: $*" >&2; }

die() {
  yell "$*"
  (kind delete cluster || true) && exit 111
}

try() { "$@" || die "cannot $*"; }

function setup_kind() {
  echo "setup_kind():"
  if [ "$TEST_FLOOR" = "true" ]; then
    try kind create cluster --config "${SCRIPT_DIR}/kind-config-floor.yaml"
  else
    try kind create cluster --config "${SCRIPT_DIR}/kind-config.yaml"
  fi
  echo "Kind is ready"
  try kubectl get nodes
}

# Function Intended to Test cert-manager for Tenant's certificate.
function install_cert_manager() {
    # https://cert-manager.io/docs/installation/
    try kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml

    echo "Wait until cert-manager pods are running:"
    try kubectl wait -n cert-manager --for=condition=ready pod -l app=cert-manager --timeout=120s
    try kubectl wait -n cert-manager --for=condition=ready pod -l app=cainjector --timeout=120s
    try kubectl wait -n cert-manager --for=condition=ready pod -l app=webhook --timeout=120s
}

# removes the pool from the provided tenant.
# usage: remove_decommissioned_pool <tenant-name> <tenant-ns>
function remove_decommissioned_pool() {

    TENANT_NAME=$1
    NAMESPACE=$2

    # While there is a conflict, let's retry
    RESULT=1
    while [ "$RESULT" == "1" ]; do
        sleep 10 # wait for new tenant spec to be ready
        get_tenant_spec "$TENANT_NAME" "$NAMESPACE"
        # modify it
        yq eval 'del(.spec.pools[0])' ~/tenant.yaml >~/new-tenant.yaml
        # replace it
        RESULT=$(kubectl_replace ~/new-tenant.yaml)
    done

}

# waits for n resources to come up.
# usage: wait_for_n_resources <namespace> <label-selector> <no-of-resources>
function wait_for_n_resources() {
    NUMBER_OF_RESOURCES=$3
    waitdone=0
    totalwait=0
    DEFAULT_WAIT_TIME=300 # 300 SECONDS
    if [ -z "$4" ]; then
        echo "No requested waiting time hence using default value"
    else
        echo "Requested specific waiting time"
        DEFAULT_WAIT_TIME=$4
    fi
    echo "* Waiting for ${NUMBER_OF_RESOURCES} pods to come up; wait_time: ${DEFAULT_WAIT_TIME};"
    command_to_wait="kubectl -n $1 get pods --field-selector=status.phase=Running --no-headers --ignore-not-found=true"
    if [ "$2" ]; then
        command_to_wait="kubectl -n $1 get pods --field-selector=status.phase=Running --no-headers --ignore-not-found=true -l $2"
    fi
    echo "* Waiting on: $command_to_wait"

    while true; do
        # xargs to trim whitespaces from bash variable.
        waitdone=$($command_to_wait | wc -l | xargs)

        echo " "
        echo " "
        echo "##############################"
        echo "To show visibility in all pods"
        echo "##############################"
        kubectl get pods -A
        echo " "
        echo " "
        echo " "

        if [ "$waitdone" -ne 0 ]; then
            if [ "$NUMBER_OF_RESOURCES" == "$waitdone" ]; then
                break
            fi
        fi
        sleep 5
        totalwait=$((totalwait + 10))
        if [ "$totalwait" -gt "$DEFAULT_WAIT_TIME" ]; then
            echo "* Unable to get resource after 10 minutes, exiting."
            try false
        fi
    done
}

# waits for n tenant pods to come up.
# usage: wait_for_n_tenant_pods <no-of-resources> <tenant-ns> <tenant-name>
function wait_for_n_tenant_pods() {
    NUMBER_OF_RESOURCES=$1
    NAMESPACE=$2
    TENANT_NAME=$3
    echo "wait_for_n_tenant_pods(): * Waiting for ${NUMBER_OF_RESOURCES} '${TENANT_NAME}' tenant pods in ${NAMESPACE} namespace"
    wait_for_n_resources "$NAMESPACE" "v1.min.io/tenant=$TENANT_NAME" "$NUMBER_OF_RESOURCES" 600
    echo "Waiting for the tenant pods to be ready (5m timeout)"
    try kubectl wait --namespace "$NAMESPACE" \
        --for=condition=ready pod \
        --selector v1.min.io/tenant="$TENANT_NAME" \
        --timeout=600s
}

# copies the script to the pod.
# usage: copy_script_to_pod <script> <pod-name>
function copy_script_to_pod() {
    SCRIPT=$1   # Example: mirror-script.sh
    POD_NAME=$2 # Example: ubuntu-pod
    echo -e "\n\n"
    echo "Copy script to pod"
    echo "From:"
    echo "${GITHUB_WORKSPACE}/testing/$SCRIPT"
    echo "To:"
    echo "/root/$SCRIPT"
    echo "In pod:"
    echo "$POD_NAME"
    echo -e "\n\n"
    kubectl cp "${GITHUB_WORKSPACE}/testing/$SCRIPT" "$POD_NAME":"/root/$SCRIPT"
}

# executes the provided script inside the pod.
# usage: execute_pod_script <script> <pod-name>
function execute_pod_script() {
    # Objective: Execute a script in a pod.
    POD_NAME=$2
    SCRIPT_NAME=$1
    echo -e "\n\n"
    echo "##############################"
    echo "#                            #"
    echo "# Execute pod's script       #"
    echo "#                            #"
    echo "##############################"
    echo "* File: common.sh"
    echo "* Function: execute_pod_script()"
    echo "* Script: ${SCRIPT_NAME}"
    echo "* Pod: ${POD_NAME}"
    copy_script_to_pod "$SCRIPT_NAME" "$POD_NAME"
    kubectl exec "$POD_NAME" -- /bin/bash -c "chmod 755 /root/$SCRIPT_NAME; export MC_HOT_FIX_REL=${MC_ENTERPRISE_TEST_WITH_HOTFIX_VERSION}; export MC_VER=${MC_VERSION}; source /root/$SCRIPT_NAME"
    echo -e "\n\n"
}

# usage: install_minio
function install_minio() {
    echo -e "\n\n"
    echo "################"
    echo "Installing MinIO"
    echo "################"
    echo ""
    rm -rf /usr/local/bin/minio
    rm -rf ~/minio-installation
    mkdir ~/minio-installation
    cd ~/minio-installation || exit

    DOWNLOAD_NAME="archive/minio.${MINIO_VERSION}"
    MINIO_RELEASE_TYPE="release"
    if [ -n "${MINIO_ENTERPRISE_TEST_WITH_HOTFIX_VERSION}" ] && [ -n "${MINIO_VERSION}" ]; then
        MINIO_RELEASE_TYPE="hotfixes"
    fi

    if [ "${MINIO_VERSION}" == "latest" ] || [ -z "${MINIO_VERSION}" ]; then
        DOWNLOAD_NAME="minio"
    fi

    RESULT=$(uname -a | grep Darwin | grep -c arm64 | awk -F' ' '{print $1}')
    if [ "$RESULT" == "1" ]; then
        curl --progress-bar -o minio https://dl.min.io/server/minio/"${MINIO_RELEASE_TYPE}"/darwin-amd64/"${DOWNLOAD_NAME}"
    fi
    RESULT=$(uname -a | grep Linux | grep -c x86_64 | awk -F' ' '{print $1}')
    if [ "$RESULT" == "1" ]; then
        wget -O minio https://dl.min.io/server/minio/"${MINIO_RELEASE_TYPE}"/linux-amd64/"${DOWNLOAD_NAME}"
    fi
    chmod +x minio
    mv minio /usr/local/bin/minio || sudo mv minio /usr/local/bin/minio
    minio --version
}

# usage: install_mc
function install_mc() {
    rm -rf /usr/local/bin/mc
    rm -rf ~/mc-installation
    mkdir ~/mc-installation
    cd ~/mc-installation || exit

    DOWNLOAD_NAME="archive/mc.${MC_VERSION}"
    MC_RELEASE_TYPE="release"
    if [ -n "${MC_ENTERPRISE_TEST_WITH_HOTFIX_VERSION}" ] && [ -n "${MC_VERSION}" ]; then
        MC_RELEASE_TYPE="hotfixes"
    fi

    if [ "${MC_VERSION}" == "latest" ] || [ -z "${MC_VERSION}" ]; then
        DOWNLOAD_NAME="mc"
    fi

    RESULT=$(uname -a | grep Darwin | grep -c arm64 | awk -F' ' '{print $1}')
    if [ "$RESULT" == "1" ]; then
        curl --progress-bar -o mc https://dl.min.io/client/mc/"${MC_RELEASE_TYPE}"/darwin-amd64/"${DOWNLOAD_NAME}"
    fi
    RESULT=$(uname -a | grep Linux | grep -c x86_64 | awk -F' ' '{print $1}')
    if [ "$RESULT" == "1" ]; then
        wget -O mc https://dl.min.io/client/mc/"${MC_RELEASE_TYPE}"/linux-amd64/"${DOWNLOAD_NAME}"
    fi
    chmod +x mc
    mv mc /usr/local/bin/mc || sudo mv mc /usr/local/bin/mc
    mc --version
}

# usage: get_minio_image_name <version>
function get_minio_image_name() {
  ### NOTE: DON'T PUT ECHO IN BETWEEN BECAUSE THAT IS WHAT WE RETURN AT THE END OF THE FUNCTION
  VERSION=$1
  IMG="quay.io/minio/minio:${VERSION}"
  if [[ "${VERSION}" == *"hotfix"* ]]; then
    IMG="docker.io/minio/minio:${VERSION}"
  fi
  echo "${IMG}"
}

# usage: setup_testbed
function setup_testbed() {
    echo "setup_testbed():"
    DEPLOYMENT_TYPE=$1
    case ${DEPLOYMENT_TYPE} in
    baremetal)
        install_minio
        install_mc
        ;;
    distributed)
        destroy_kind
        setup_kind
        install_operator ""
        load_kind_images
        deploy_debug_pod
        ;;
    *)
        echo "unknown"
        ;;
    esac
}

# usage: get_latest_minio_version
function get_latest_minio_version() {
    version=$(curl -sL https://api.github.com/repos/minio/minio/tags | jq -r '.[1].name')
    echo "$version"
}

# installs the tenant with the provided parameter values.
# usage: install_tenant_with_minio_version <tenant-type> <tenant-name> <tenant-ns> <tenant-storage-class> <tenant-per-volume-size>
function install_tenant_with_minio_version() {
    TENANT_TYPE=$1
    TENANT_NAME=$2
    NS=$3
    TENANT_SC=$4
    echo "TENANT_SC: ${TENANT_SC}"
    TENANT_VOLUME_SIZE=$5
    echo "TENANT_VOLUME_SIZE: ${TENANT_VOLUME_SIZE}"
    TENANT_YAML="${TENANT_NAME}".yaml
    echo "TENANT_YAML: ${TENANT_YAML}"
    VER=$(get_latest_minio_version)
    echo "VER: ${VER}"
    IMG=$(get_minio_image_name "${VER}")
    echo "IMG: ${IMG}"

    echo "================================================================"
    echo "Installing MinIO Tenant"
    echo "================================================================"
    echo
    echo "TYPE: ${TENANT_TYPE}"
    echo "NAME: ${TENANT_NAME}"
    echo "NS:   ${NS}"

    echo "install_tenant_with_minio_version(): kustomize build github..."
    kustomize build github.com/minio/operator/examples/kustomization/"${TENANT_TYPE}" >"${TENANT_YAML}"
    sed -i "s/tenant-lite/${NS}/g" "${TENANT_YAML}"
    sed -i "s/tenant-tiny/${NS}/g" "${TENANT_YAML}"
    sed -i "s/myminio/${TENANT_NAME}/g" "${TENANT_YAML}"
    yq -i e "select(.kind == \"Tenant\").spec.image = \"${IMG}\"" "${TENANT_YAML}"

    echo " "
    echo " "
    echo " "
    echo "######################"
    echo "To display tenant spec"
    echo "######################"
    yq '.spec' "${TENANT_YAML}"
    echo " "
    echo " "
    echo " "

    if [ "${TENANT_SC}" ]; then
        echo "SC:   ${TENANT_SC}"
        yq -i e "select(.kind == \"Tenant\").spec.pools[].volumeClaimTemplate.spec.storageClassName = \"${TENANT_SC}\"" "${TENANT_YAML}"
    fi

    if [ "${TENANT_VOLUME_SIZE}" ]; then
        echo "PER_VOLUME_SIZE: ${TENANT_VOLUME_SIZE}"
        yq -i e "select(.kind == \"Tenant\").spec.pools[].volumeClaimTemplate.spec.resources.requests.storage = \"${TENANT_VOLUME_SIZE}\"" "${TENANT_YAML}"
    fi

    echo "install_tenant_with_minio_version(): kubectl apply -f..."
    kubectl apply -f "${TENANT_YAML}"
    rm -f "${TENANT_NAME}".yaml

    echo "install_tenant_with_minio_version(): sleep 10..."
    sleep 10
}

# checks if the tenant is ready by creating an alias.
# usage: verify_tenant_is_ready
function verify_tenant_is_ready() {
    echo -e "\n\n"
    echo "* Checking if the tenant is ready"
    execute_pod_script create-alias.sh ubuntu-pod
    check_script_result "default" "ubuntu-pod" "create-alias.log"
    echo -e "\n\n"
}

# To get the Tenant Spec from a particular namespace
# usage: get_tenant_spec <tenant-name> <tenant-ns>
function get_tenant_spec() {
    kubectl get tenants.minio.min.io "$1" -n "$2" -o yaml >~/tenant.yaml
}

# To replace an object and return 1 if it failed due to a conflict otherwise returns zero.
# usage: kubectl_replace <file>
function kubectl_replace() {
    # if the object has been modified then
    # please apply your changes to the latest version
    kubectl replace -f "$1" >>~/kubectl-replace.log 2>&1
    # if Conflict in kubectl-replace.log then return error for code to perform change again with new tenant.
    RESULT=$(grep -c Conflict ~/kubectl-replace.log | awk -F' ' '{print $1}')
    echo "$RESULT"
}

# installs the tenant with 2 pools.
# usage: install_decommission_tenant <tenant-type> <tenant-ns> <tenant-name>
function install_decommission_tenant() {
    echo "* Installing tenant to test decommission"
    TENANT_TYPE=$1
    NAMESPACE=$2
    TENANT_NAME=$3
    NUMBER_OF_RESOURCES=4
    echo "install_decommission_tenant(): Calling install_tenant_with_minio_version"
    install_tenant_with_minio_version "$TENANT_TYPE" "$TENANT_NAME" "$NAMESPACE" "" ""

    echo "install_decommission_tenant(): Calling wait_for_n_tenant_pods"
    wait_for_n_tenant_pods $NUMBER_OF_RESOURCES "$NAMESPACE" "$TENANT_NAME"

    echo "install_decommission_tenant(): Calling verify_tenant_is_ready"
    verify_tenant_is_ready

    NAMESPACE=$2 # Re-set NAMESPACE as script execution within debug pod would have set it to default
    echo "* Adding another pool to ${TENANT_NAME} tenant to test decommissioning"
    # While there is a conflict, let's retry
    RESULT=1
    while [ "$RESULT" == "1" ]; do
        sleep 10 # wait for new tenant spec to be ready
        get_tenant_spec "$TENANT_NAME" "${NAMESPACE}"
        # modify it
        yq '.spec.pools += {"name": "pool-1", "servers": 4, "volumesPerServer": 2, "volumeClaimTemplate": { "metadata": { "name": "data"  }, "spec": { "accessModes": ["ReadWriteOnce"], "resources": { "requests": { "storage": "2Gi"  } }  } }}' ~/tenant.yaml >~/new-tenant.yaml
        # replace it
        RESULT=$(kubectl_replace ~/new-tenant.yaml)
    done

    NUMBER_OF_RESOURCES=8
    wait_for_n_tenant_pods $NUMBER_OF_RESOURCES "$NAMESPACE" "$TENANT_NAME"

    verify_tenant_is_ready
}

# usage: read_script_result <namespace> <pod-name> >file>
function read_script_result() {
    NAMESPACE=$1
    POD_NAME=$2
    FILE=$3
    rm -f "$FILE"
    kubectl cp "$NAMESPACE"/"$POD_NAME":"$FILE" "$FILE"
    grep "script passed" <"$FILE"
}

# usage: teardown <deployment-type> <data-dir> <minio-server-pid>
function teardown() {
    echo "Cleanup..."
    DEPLOYMENT_TYPE=$1
    DATA_DIR=$2
    case ${DEPLOYMENT_TYPE} in
    baremetal)
        # Stop minio server instance
        pkill -9 minio
        # Remove data dir (if any)
        rm -fr "${DATA_DIR}"
        ;;
    distributed)
        # Destroy kind cluster
        destroy_kind
        ;;
    *)
        echo "unknown"
        ;;
    esac
}

# usage: check_script_result <namespace> <pod-name> <log-file-name>
function check_script_result() {
    # Manual testing: read_script_result default ubuntu-pod decommission.log
    NAMESPACE=$1 # Example: default
    POD_NAME=$2  # Example: ubuntu-pod
    LOG_NAME=$3  # Example: decommission.log
    echo "NAMESPACE: ${NAMESPACE}"
    echo "POD_NAME: ${POD_NAME}"
    echo "LOG_NAME: ${LOG_NAME}"
    RESULT=$(read_script_result "$NAMESPACE" "$POD_NAME" "$LOG_NAME")
    echo "RESULT: $RESULT"
    if [ "$RESULT" == "script passed" ]; then
        echo "Test Passed"
    else
        echo "#####################################################################"
        echo " "
        echo " "
        echo "Test Failed"
        echo " "
        echo " "
        echo "#####################################################################"
        exit 1
    fi
}

# deploys a debug pod in default namespace.
# usage: deploy_debug_pod
function deploy_debug_pod() {
    # https://downey.io/notes/dev/ubuntu-sleep-pod-yaml/
    echo "* Creating a Simple Kubernetes Debug Pod"
    kubectl apply -f "${GITHUB_WORKSPACE}/testing/configurations/debug-pod.yaml"
    wait_for_resource default ubuntu app
    kubectl -n default get pods -l app=ubuntu --no-headers
    echo "* Waiting for the ubuntu pod"
    sleep 20
    try kubectl wait --namespace default \
        --for=condition=ready pod \
        --selector=app=ubuntu \
        --timeout=60s
    execute_pod_script install-mc.sh ubuntu-pod
    check_script_result default ubuntu-pod install-mc.log
}

# usage: get_latest_operator_version
function get_latest_operator_version() {
  ### NOTE: DON'T PUT ECHO IN BETWEEN BECAUSE THAT IS WHAT WE RETURN AT THE END OF THE FUNCTION
  version=$(curl -sL https://api.github.com/repos/minio/operator/tags | jq -r '.[0].name')
  echo "$version"
}

# usage: get_console_image
function get_console_image() {
  ### NOTE: DON'T PUT ECHO IN BETWEEN BECAUSE THAT IS WHAT WE RETURN AT THE END OF THE FUNCTION
  if [ -z "$OPERATOR_VERSION" ] || [ "$OPERATOR_VERSION" == "latest" ]; then
    operator_tag=$(get_latest_operator_version)
  else
    operator_tag=$OPERATOR_VERSION
  fi
  echo "minio/operator:$operator_tag"
}

# usage: load_kind_image <image>
function load_kind_image() {
  echo "load_kind_image():"
  echo "* Loading image ${1}"
  try docker pull "$1"
  try kind load docker-image "$1"
}

# usage: load_kind_images
function load_kind_images() {
    echo "load_kind_images():"
    MINIO_RELEASE=$(get_minio_image_name "latest")
    CONSOLE_RELEASE=$(get_console_image)
    echo "MINIO_RELEASE: ${MINIO_RELEASE}"
    echo "CONSOLE_RELEASE: ${CONSOLE_RELEASE}"
    load_kind_image "$MINIO_RELEASE"
    load_kind_image "$CONSOLE_RELEASE"
}

function create_restricted_namespace() {
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: "$1"
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/audit-version: latest
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/warn-version: latest
EOF
}

function install_operator() {
  # It requires compiled binary in minio-operator folder in order for docker build to work when copying this folder.
  # For that in the github actions you need to wait for operator test/step to get the binary.
  echo "install_operator():"
  # To compile current branch
  echo "Compiling Current Branch Operator"
  TAG=minio/operator:noop
  (cd "${SCRIPT_DIR}/.." && try docker build -t $TAG .) # will not change your shell's current directory

  echo 'start - load compiled image so we can use it later on'
  try kind load docker-image $TAG
  echo 'end - load compiled image so we can use it later on'
  # To compile current branch
  echo "Compiling Current Branch Sidecar"
  SIDECAR_TAG=minio/operator-sidecar:noop
  (cd "${SCRIPT_DIR}/../" && try docker build -t $SIDECAR_TAG -f sidecar/Dockerfile .) # will not change your shell's current directory

  echo 'start - load compiled sidecar image so we can use it later on'
  try kind load docker-image $SIDECAR_TAG
  echo 'end - load compiled sidecar image so we can use it later on'

  if [ "$1" = "helm" ]; then

    echo "Change the version accordingly for image to be found within the cluster"
    yq -i '.operator.image.repository = "minio/operator"' "${SCRIPT_DIR}/../helm/operator/values.yaml"
    yq -i '.operator.image.tag = "noop"' "${SCRIPT_DIR}/../helm/operator/values.yaml"
    yq -i '.console.image.repository = "minio/operator"' "${SCRIPT_DIR}/../helm/operator/values.yaml"
    yq -i '.console.image.tag = "noop"' "${SCRIPT_DIR}/../helm/operator/values.yaml"
    yq -i '.operator.env += [{"name": "OPERATOR_SIDECAR_IMAGE", "value": "'$SIDECAR_TAG'"}]' "${SCRIPT_DIR}/../helm/operator/values.yaml"
    echo "Installing Current Operator via HELM"
    create_restricted_namespace minio-operator
    helm install \
      --namespace minio-operator \
      minio-operator ./helm/operator

    echo "key, value for pod selector in helm test"
    key=app.kubernetes.io/name
    value=operator
  elif [ "$1" = "sts" ]; then
    echo "Installing Current Operator with sts enabled"
    try kubectl apply -k "${SCRIPT_DIR}/../testing/sts/operator"
    try kubectl -n minio-operator set env deployment/minio-operator OPERATOR_SIDECAR_IMAGE="$SIDECAR_TAG"
    echo "key, value for pod selector in kustomize test"
    key=name
    value=minio-operator
  elif [ "$1" = "certmanager" ]; then
    echo "Installing Current Operator with certmanager"
    try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/operator-certmanager"
    echo "key, value for pod selector in kustomize test"
    key=name
    value=minio-operator
  else
    echo "Installing Current Operator"
    echo "When coming from the upgrade test, the operator is already installed."
    operator_deployment=$(kubectl get deployment minio-operator -n minio-operator | grep -c minio-operator)
    echo "Check whether we should reinstall it or simply change the images."
    if [ "$operator_deployment" != 1 ]; then
        echo "operator is not installed, will be installed"
        try kubectl apply -k "${SCRIPT_DIR}/../resources" # we maintain just one spot, no new overlays
    else
        echo "Operator is not re-installed as is already installed"
    fi

    # and then we change the images, no need to have more overlays in different folders.
    echo "changing images for console and minio-operator deployments"
    try kubectl -n minio-operator set image deployment/minio-operator minio-operator="$TAG"
    try kubectl -n minio-operator set image deployment/console console="$TAG"
    try kubectl -n minio-operator set env deployment/minio-operator OPERATOR_SIDECAR_IMAGE="$SIDECAR_TAG"

    echo "key, value for pod selector in kustomize test"
    key=name
    value=minio-operator
  fi

  # Reusing the wait for both, Kustomize and Helm
  echo "Waiting for k8s api"
  sleep 10

  kubectl get ns

  kubectl -n minio-operator get deployments
  kubectl -n minio-operator get pods

  echo "Waiting for Operator Pods to come online (2m timeout)"
  try kubectl wait --namespace minio-operator \
    --for=condition=ready pod \
    --selector $key=$value \
    --timeout=120s

  echo "start - get data to verify proper image is being used"
  kubectl get pods --namespace minio-operator
  kubectl describe pods -n minio-operator | grep Image
  echo "end - get data to verify proper image is being used"
}

function install_operator_version() {
  # Obtain release
  version="$1"
  if [ -z "$version" ]; then
    version=$(curl https://api.github.com/repos/minio/operator/releases/latest | jq --raw-output '.tag_name | "\(.[1:])"')
  fi
  echo "Target operator release: $version"

  # Initialize the MinIO Kubernetes Operator
  kubectl apply -k github.com/minio/operator/resources/\?ref=v"$version"


  if [ "$1" = "helm" ]; then
    echo "key, value for pod selector in helm test"
    key=app.kubernetes.io/name
    value=operator
  else
    echo "key, value for pod selector in kustomize test"
    key=name
    value=minio-operator
  fi

  # Reusing the wait for both, Kustomize and Helm
  echo "Waiting for k8s api"
  sleep 10

  kubectl get ns

  kubectl -n minio-operator get deployments
  kubectl -n minio-operator get pods

  echo "Waiting for Operator Pods to come online (2m timeout)"
  try kubectl wait --namespace minio-operator \
    --for=condition=ready pod \
    --selector $key=$value \
    --timeout=120s

  echo "start - get data to verify proper image is being used"
  kubectl get pods --namespace minio-operator
  kubectl describe pods -n minio-operator | grep Image
  echo "end - get data to verify proper image is being used"
}

function destroy_kind() {
  # To allow the execution without killing the cluster at the end of the test
  # Use below statement to automatically test and kill cluster at the end:
  # `unset OPERATOR_DEV_TEST`
  # Use below statement to test and keep cluster alive at the end!:
  # `export OPERATOR_DEV_TEST="ON"`
  echo "destroy_kind():"
  if [[ -z "${DEV_TEST}" ]]; then
    echo "Cluster not destroyed due to manual testing"
  else
    kind delete cluster
  fi
}

function check_tenant_status() {
  # Check MinIO is accessible
  # $1 namespace
  # $2 tenant name
  # $3 metadata.app field value (optional)
  # $4 "helm", means it's testing helm tenant (optional)

  key=v1.min.io/tenant
  value=$2
  if [ $# -ge 3 ]; then
    echo "Third argument provided, then set key value"
    key=app
    value=$3
  else
    echo "No third argument provided, using default key"
  fi

  wait_for_resource $1 $value $key

  echo "Waiting for tenant to be Initialized"

  condition=jsonpath='{.status.currentState}'=Initialized
  selector="metadata.name=$2"
  try wait_for_resource_field_selector "$1" tenant $condition "$selector" 600s

  if [ $# -ge 4 ]; then
    echo "Fourth argument provided, then get secrets from helm"
    USER=$(kubectl get secret myminio-secret -o jsonpath="{.data.accesskey}" | base64 --decode)
    PASSWORD=$(kubectl get secret myminio-secret -o jsonpath="{.data.secretkey}" | base64 --decode)
  else
    echo "No fourth argument provided, using default USER and PASSWORD"
    TENANT_CONFIG_SECRET=$(kubectl -n $1 get tenants.minio.min.io $2 -o jsonpath="{.spec.configuration.name}")
    USER=$(kubectl -n $1 get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_USER="' | sed -e 's/export MINIO_ROOT_USER="//g' | sed -e 's/"//g')
    PASSWORD=$(kubectl -n $1 get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_PASSWORD="' | sed -e 's/export MINIO_ROOT_PASSWORD="//g' | sed -e 's/"//g')
  fi

  if [ $# -ge 4 ]; then
    # make sure no rollout is happening
    try kubectl -n $1 rollout status sts/myminio-pool-0
  else
    # make sure no rollout is happening
    try kubectl -n $1 rollout status sts/$2-pool-0
  fi

  echo "Tenant is created successfully, proceeding to validate 'mc admin info minio/'"

  try kubectl get pods --namespace $1

  if [ "$4" = "helm" ]; then
    # File: operator/helm/tenant/values.yaml
    # Content: s3.bucketDNS: false
    echo "In helm values by default bucketDNS.s3 is disabled, skipping mc validation on helm test"
  else
    kubectl run --restart=Never admin-mc --image quay.io/minio/mc \
      --env="MC_HOST_minio=https://${USER}:${PASSWORD}@minio.${1}.svc.cluster.local" \
      --command -- bash -c "until (mc admin info minio/); do echo 'waiting... for 5secs' && sleep 5; done"

    echo "Wait to mc admin info minio/"
    try kubectl wait --for=jsonpath="{.status.phase}=Succeeded" --timeout=10m pod/admin-mc --namespace default

    # Retrieve the logs
    kubectl logs admin-mc
  fi

  echo "Done."
}

# To install tenant with cert-manager from our example provided.
function install_cert_manager_tenant() {

    echo "Install cert-manager tenant from our example:"
    try kubectl apply -k github.com/minio/operator/examples/kustomization/tenant-certmanager

    echo "Wait until tenant-certmanager-tls secret is generated by cert-manager..."
    while ! kubectl get secret tenant-certmanager-tls --namespace tenant-certmanager
    do
      echo "Waiting for my secret. Current secrets are:"
      kubectl get secrets -n tenant-certmanager
      sleep 1
    done

    # https://github.com/minio/operator/blob/master/docs/cert-manager.md
    echo "# Pass the CA cert to our Operator to trust the tenant:"
    echo "## First get the CA from cert-manager secret..."
    try kubectl get secrets -n tenant-certmanager tenant-certmanager-tls -o=jsonpath='{.data.ca\.crt}' | base64 -d > public.crt
    echo "## Then create the secret in operator's namespace..."
    try kubectl create secret generic operator-ca-tls-tenant-certmanager --from-file=public.crt -n minio-operator
    echo "## Finally restart minio operator pods to catch up and trust tenant..."
    try kubectl rollout restart deployment.apps/minio-operator -n minio-operator

}

# Install tenant function is being used by deploy-tenant and check-prometheus
function install_tenant() {
  # Check if we are going to install helm, latest in this branch or a particular version
  if [ "$1" = "helm" ]; then
    echo "Installing tenant from Helm"
    echo "This test is intended for helm only not for KES, there is another kes test, so let's remove KES here"
    yq -i eval 'del(.tenant.kes)' "${SCRIPT_DIR}/../helm/tenant/values.yaml"

    try helm lint "${SCRIPT_DIR}/../helm/tenant" --quiet

    namespace=default
    key=v1.min.io/tenant
    value=myminio
    create_restricted_namespace $namespace
    try helm install --namespace $namespace \
      tenant ./helm/tenant
  elif [ "$1" = "logs" ]; then
    namespace="tenant-lite"
    key=v1.min.io/tenant
    value=myminio
    echo "Installing lite tenant from current branch"

    try kubectl apply -k "${SCRIPT_DIR}/../testing/tenant-logs"
  elif [ "$1" = "prometheus" ]; then
    namespace="tenant-lite"
    key=v1.min.io/tenant
    value=myminio
    echo "Installing lite tenant from current branch"

    try kubectl apply -k "${SCRIPT_DIR}/../testing/tenant-prometheus"
  elif [ "$1" = "policy-binding" ]; then
    namespace="minio-tenant-1"
    key=v1.min.io/tenant
    value=myminio
    echo "Installing policyBinding tenant from current branch"

    try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/sts-example/tenant"
  elif [ -e $1 ]; then
    namespace="tenant-lite"
    key=v1.min.io/tenant
    value=myminio
    echo "Installing lite tenant from current branch"

    try kubectl apply -k "${SCRIPT_DIR}/../testing/tenant"
  else
    namespace="tenant-lite"
    key=v1.min.io/tenant
    value=myminio
    echo "Installing lite tenant for version $1"

    try kubectl apply -k "github.com/minio/operator/testing/tenant\?ref\=$1"
  fi

  echo "Waiting for the tenant statefulset, this indicates the tenant is being fulfilled"
  echo $namespace
  echo $value
  echo $key
  wait_for_resource $namespace $value $key

  echo "Waiting for tenant pods to come online (5m timeout)"
  try kubectl wait --namespace $namespace \
    --for=condition=ready pod \
    --selector $key=$value \
    --timeout=300s

  echo "Build passes basic tenant creation"

}

function setup_sts_bucket() {
  echo "Installing setup bucket job"
  try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/sts-example/sample-data"
  namespace="minio-tenant-1"
  condition="condition=Complete"
  selector="metadata.name=setup-bucket"
  try wait_for_resource_field_selector $namespace job $condition $selector "10m"
  echo "Installing setup bucket job: DONE"
}

function install_sts_client() {
  echo "Installing sts client job for $1"
  # Definition of the sdk and client to test

  OLDIFS=$IFS
  IFS="-"; declare -a CLIENTARR=($1)
  sdk="${CLIENTARR[0]}-${CLIENTARR[1]}"
  lang="${CLIENTARR[2]}"
  makefiletarget="${CLIENTARR[0]}${CLIENTARR[1]}$lang"
  IFS=$OLDIFS

  # Build and load client images
  echo "Building docker image for miniodev/operator-sts-example:$1"
  (cd "${SCRIPT_DIR}/../examples/kustomization/sts-example/sample-clients" && try make "${makefiletarget}")
  try kind load docker-image "miniodev/operator-sts-example:$1"

  client_namespace="sts-client"
  tenant_namespace="minio-tenant-1"

  if [ $# -ge 2 ]; then
    if [ "$2" = "cm" ]; then
      echo "Setting up certmanager CA secret"
      # When certmanager issues the certificates, we copy the certificate to a secret in the client namespace
      try kubectl get secrets -n $tenant_namespace tenant-certmanager-tls -o=jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
      try kubectl create secret generic tenant-certmanager-tls --from-file=ca.crt -n $client_namespace
    fi
  fi

  echo "creating client $1"
  try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/sts-example/sample-clients/$sdk/$lang"
  condition="condition=Complete"
  selector="metadata.name=sts-client-example-$sdk-$lang-job"
  try wait_for_resource_field_selector $client_namespace job $condition $selector 600s
  echo "Installing sts client job for $1: DONE"
}

# Port forward
function port_forward() {
  namespace=$1
  tenant=$2
  svc=$3
  localport=$4

  totalwait=0
  echo 'Validating tenant pods are ready to serve'
  for pod in `kubectl --namespace $namespace --selector=v1.min.io/tenant=$tenant get pod -o json |  jq '.items[] | select(.metadata.name|contains("'$tenant'"))| .metadata.name' | sed 's/"//g'`; do
    while true; do
      if kubectl --namespace $namespace -c minio logs pod/$pod | grep --quiet 'All MinIO sub-systems initialized successfully'; then
        echo "$pod is ready to serve" && break
      fi
      sleep 5
      totalwait=$((totalwait + 5))
      if [ "$totalwait" -gt 305 ]; then
        echo "Unable to validate pod $pod after 5 minutes, exiting."
        try false
      fi
    done
  done

  echo "Killing any current port-forward"
  for pid in $(lsof -i :$localport | awk '{print $2}' | uniq | grep -o '[0-9]*')
  do
    if [ -n "$pid" ]
    then
      kill -9 $pid
      echo "Killed previous port-forward process using port $localport: $pid"
    fi
  done

  echo "Establishing port-forward"
  kubectl port-forward service/$svc -n $namespace $localport &

  echo 'start - wait for port-forward to be completed'
  sleep 15
  echo 'end - wait for port-forward to be completed'
}
