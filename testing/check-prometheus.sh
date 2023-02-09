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

function wait_on_prometheus_pods() {
  echo "waiting for storage-lite-prometheus-0"
  i=0
  while [[ $(kubectl get pods -n tenant-lite --selector=statefulset.kubernetes.io/pod-name=storage-lite-prometheus-0 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
    ((i++))

    mod=$((i % 12))
    if [[ $mod -eq 0 ]]; then
      echo "waiting for storage-lite-prometheus-0"
    fi
    sleep 1
    if [[ $i -eq 300 ]]; then
      kubectl get pods -n tenant-lite -o wide
      kubectl describe pods -n tenant-lite
      kubectl logs -n minio-operator -l name=minio-operator
      break
    fi
  done
}

function main() {
  destroy_kind

  setup_kind

  install_operator

  install_tenant "prometheus"

  check_tenant_status tenant-lite storage-lite

  try kubectl get pods --namespace tenant-lite

  echo 'start - wait for prometheus to appear'
  wait_on_prometheus_pods
  echo 'end - wait for prometheus to appear'

  echo "make sure there's no rolling restart going"
  kubectl -n tenant-lite rollout status sts/storage-lite-pool-0

  port_forward tenant-lite storage-lite storage-lite-console 9443

  echo 'Get token from MinIO Console'
  COOKIE=$(
    curl 'https://localhost:9443/api/v1/login' -vs \
      -H 'content-type: application/json' \
      --data-raw '{"accessKey":"minio","secretKey":"minio123"}' --insecure 2>&1 |
      grep "set-cookie: token=" | sed -e "s/< set-cookie: token=//g" |
      awk -F ';' '{print $  1}'
  )
  echo $COOKIE

  echo 'start - wait for prometheus to be ready'

  wait_on_prometheus_pods

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
