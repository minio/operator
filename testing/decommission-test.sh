#!/usr/bin/env bash

echo "${GITHUB_WORKSPACE=$(dirname "$0")/../..}"
echo "GITHUB_WORKSPACE: ${GITHUB_WORKSPACE}"
DEBUG=false
POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
    case $1 in
    -d | --debug)
        DEBUG=true
        shift # past argument
        ;;
    *)
        POSITIONAL_ARGS+=("$1") # save positional arg
        shift                   # past argument
        ;;
    esac
done
set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters
echo "DEBUG: ${DEBUG}"

SCRIPT_DIR=$(dirname "$0")
export SCRIPT_DIR
source "${SCRIPT_DIR}/common.sh"

function main() {
    echo "main():"
    # Explanation: This test will start with 2 pools, one pool will be decommissioned.
    # Each pool contains 4 pods for a total of 8 pods per Tenant.
    # Then if decommission is successful, we should get 4 pods from one pool at the
    # end of the test and those pods should be running.
    # If no pod is running or if we don't have that exact number then test should fail.
    echo "setup_testbed distributed:"
    setup_testbed distributed

    TENANT_NAME="myminio"
    DEBUG_POD_NAME="ubuntu-pod"

    echo -e "\n\n"
    echo "decommission-test.sh: main(): Then we install the tenant with 2 pools"
    install_decommission_tenant "tenant-lite" "tenant-lite" "${TENANT_NAME}"
    try kubectl get pods --namespace "tenant-lite"

    echo -e "\n\n"
    echo "decommission-test.sh: main(): Once tenant is ready we upload some data to the tenant"
    execute_pod_script upload-data.sh "${DEBUG_POD_NAME}"
    RESULT=$(read_script_result default "${DEBUG_POD_NAME}" upload-data.log)
    if [ "$RESULT" != "script passed" ]; then
        teardown distributed
        exit 1
    fi

    echo -e "\n\n"
    echo "decommission-test.sh: main(): Execute decommission script"
    execute_pod_script decommission-script.sh "${DEBUG_POD_NAME}"
    check_script_result default "${DEBUG_POD_NAME}" decommission.log
    remove_decommissioned_pool "${TENANT_NAME}" "tenant-lite"
    wait_for_n_tenant_pods 4 "tenant-lite" "${TENANT_NAME}"

    echo -e "\n\n"
    echo "decommission-test.sh: main(): Verify data after decommission:"
    execute_pod_script verify-data.sh "${DEBUG_POD_NAME}"
    RESULT=$(read_script_result default "${DEBUG_POD_NAME}" verify-data.log)
    teardown distributed
    if [ "$RESULT" != "script passed" ]; then
        exit 1
    fi
}

main "$@"
