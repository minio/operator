#!/usr/bin/env bash
# This script will run inside ubuntu-pod that is located at default namespace in the cluster
# This script will not and should not be executed in the self hosted runner

echo "create-alias.sh: sleep to wait for MinIO Server to be ready prior mc commands"
# https://github.com/minio/mc/issues/3599
sleep 60

echo "create-alias.sh: Get the ALIAS... after 60s"
ALIAS_CREATED=false
for i in {1..100}; do
    echo "attempt: ${i}/100"
    RESULT=$(mc alias set myminio/ "https://minio.tenant-lite.svc.cluster.local" minio minio123)
    if [ "$RESULT" == "Added \`myminio\` successfully." ]; then
        echo "create-alias.sh: ALIAS added, let's continue with decommission steps"
        ALIAS_CREATED=true
        break
    else
        echo "create-alias.sh: ALIAS not added, re-try!"
    fi
    sleep 10
done

echo "create-alias.sh: Verify ALIAS creation"
if [ "$ALIAS_CREATED" == "false" ]; then
    echo "create-alias.sh: ALIAS is not created hence we cannot continue any further"
    exit 1
elif [ "$ALIAS_CREATED" == "true" ]; then
    echo "script passed" >create-alias.log
fi
