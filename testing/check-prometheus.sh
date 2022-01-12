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

# This script requires: kubectl, kind, jq

SCRIPT_DIR=$(dirname "$0")
export SCRIPT_DIR
source "${SCRIPT_DIR}/common.sh"

function main() {
    destroy_kind

    setup_kind

    install_operator

    install_tenant

    check_tenant_status tenant-lite storage-lite

    echo 'start - wait for prometheus to appears'
    i=0
    while [[ $(kubectl get pods -n tenant-lite --selector=statefulset.kubernetes.io/pod-name=storage-lite-prometheus-0 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]];
    do
      ((i++))
      echo "waiting for storage-lite-prometheus-0" && sleep 1;
      if [[ $i -eq 300 ]]; then
        break
      fi
    done
    echo 'end - wait for prometheus to appears'

    echo 'Wait for pod to be ready for port forward'
    try kubectl wait --namespace tenant-lite \
      --for=condition=ready pod \
      --selector=statefulset.kubernetes.io/pod-name=storage-lite-ss-0-0 \
      --timeout=120s

    echo 'port forward without the hop, directly from the tenant/pod'
    kubectl port-forward storage-lite-ss-0-0 9443 --namespace tenant-lite &

    echo 'start - wait for port-forward to be completed'
    sleep 15
    echo 'end - wait for port-forward to be completed'

    echo 'To display port connections'
    sudo netstat -tunlp # want to see if 9443 is LISTEN state to proceed

    echo 'start - open and allow port connection'
    sudo apt install ufw
    sudo ufw allow http
    sudo ufw allow https
    sudo ufw allow 9443/tcp
    echo 'end - open and allow port connection'

    echo 'Get token from MinIO Console'
    COOKIE=$(
      curl 'https://localhost:9443/api/v1/login' -vs \
      -H 'content-type: application/json' \
      --data-raw '{"accessKey":"minio","secretKey":"minio123"}' --insecure 2>&1 | \
      grep "set-cookie: token=" | sed -e "s/< set-cookie: token=//g" | \
      awk -F ';' '{print $1}'
    )
    echo $COOKIE

    echo 'start - wait for prometheus to be ready'
    try kubectl wait --namespace tenant-lite \
      --for=condition=ready pod \
      --selector=statefulset.kubernetes.io/pod-name=storage-lite-prometheus-0 \
      --timeout=300s
    echo 'end - wait for prometheus to be ready'

    echo 'start - print the entire output for debug'
    curl 'https://localhost:9443/api/v1/admin/info/widgets/66/?step=0&' \
      -H 'cookie: token='$COOKIE'' \
      --compressed \
      --insecure
    echo 'end - print the entire output for debug'

    echo 'Verify Prometheus via API'
    RESULT=$(
      curl 'https://localhost:9443/api/v1/admin/info/widgets/66/?step=0&' \
      -H 'cookie: token='$COOKIE'' \
      --compressed \
      --insecure | jq '.title'
    )
    echo $RESULT
    EXPECTED_RESULT='"Number of Buckets"'
    echo $EXPECTED_RESULT
    if [ "$EXPECTED_RESULT" = "$RESULT" ]; then
        echo "Prometheus is present, no issue found"
    else
        echo "Prometheus URL is unreachable"
        exit 111
    fi

    destroy_kind
}

main "$@"
