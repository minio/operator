#!/bin/bash

# Objective:
# To put the image in a Public Registry for testing Operator's deployment via OLM
# Once image is up, we can use operator-sdk tool to deploy Operator in a given cluster,
# Example:
#   $ operator-sdk run bundle quay.io/cniackz4/minio-operator:v5.0.4 --index-image=quay.io/operator-framework/opm:v1.23.0
#   INFO[0013] Created registry pod: quay-io-cniackz4-minio-operator-v5-0-4 
#   INFO[0013] Created CatalogSource: minio-operator-catalog 
#   INFO[0013] OperatorGroup "operator-sdk-og" created      
#   INFO[0013] Created Subscription: minio-operator-v5-0-4-sub 
#   INFO[0015] Approved InstallPlan install-smxqt for the Subscription: minio-operator-v5-0-4-sub 
#   INFO[0015] Waiting for ClusterServiceVersion "default/minio-operator.v5.0.4" to reach 'Succeeded' phase 
#   INFO[0015]   Waiting for ClusterServiceVersion "default/minio-operator.v5.0.4" to appear 
#   INFO[0026]   Found ClusterServiceVersion "default/minio-operator.v5.0.4" phase: Pending 
#   INFO[0028]   Found ClusterServiceVersion "default/minio-operator.v5.0.4" phase: Installing 
#   INFO[0036]   Found ClusterServiceVersion "default/minio-operator.v5.0.4" phase: Succeeded 
#   INFO[0036] OLM has successfully installed "minio-operator.v5.0.4"

function get_latest_release() {
curl --silent "https://api.github.com/repos/minio/operator/releases/latest" |
  grep '"tag_name":' |
  sed -E 's/.*"([^"]+)".*/\1/'
}
RELEASE=$(get_latest_release)
podman build -f bundle.Dockerfile -t quay.io/cniackz4/minio-operator:$RELEASE .
podman push quay.io/cniackz4/minio-operator:$RELEASE
