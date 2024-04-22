#!/usr/bin/env bash
# This script will run inside ubuntu-pod that is located at default namespace in the cluster
# This script will not and should not be executed in the self hosted runner

# Test Variables:
ALIAS_NAME="myminio"
POOL="https://${ALIAS_NAME}-pool-0-{0...3}.${ALIAS_NAME}-hl.tenant-lite.svc.cluster.local/export{0...1}"

echo "script failed" >decommission.log # assume initial state
echo "Starting decommission log file"

# https://docs.min.io/minio/baremetal/installation/decommission-pool.html
echo "1) Review the MinIO Deployment Topology"
mc admin decommission status "${ALIAS_NAME}"

echo "2) Start the Decommission Process:"
mc admin decommission start "${ALIAS_NAME}"/ "${POOL}"

echo "3) Monitor the Decommission Process:"
DECOMMISSION_COMPLETED=false
for i in {1..1000}; do
    echo "attempt ${i}/1000"
    echo "mc admin decommission status ${ALIAS_NAME}/ ${POOL} --json"
    RESULT=$(mc admin decommission status ${ALIAS_NAME}/ ${POOL} --json | jq '.decommissionInfo.complete')
    if [ "$RESULT" == "true" ]; then
        mc admin decommission status ${ALIAS_NAME}
        echo "decommission has been completed!"
        DECOMMISSION_COMPLETED=true
        break
    else
        echo "decommission is still work in progress"
    fi
    sleep 10
done

if [ "$DECOMMISSION_COMPLETED" == "false" ]; then
    echo "couldn't complete decommission during this time..."
    echo "script failed" >decommission.log
    exit 1
else
    echo "script passed"
    echo "script passed" >decommission.log
fi
