#!/bin/bash

RELEASE=4.4.7
EXAMPLE=$(kustomize build examples/kustomization/tenant-lite | yq eval-all '. | [.]' | yq -o json | jq -c )

operator-sdk generate bundle \
  --package minio-operator \
  --version 4.4.7 \
  --deploy-dir resources/base \
  --crds-dir resources/base/crds \
  --manifests \
  --metadata \
  --output-dir bundles/$RELEASE \
  --channels stable

myenv=$EXAMPLE yq -i e ".metadata.annotations.alm-examples |= (\"\${myenv}\" | envsubst)" bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml

miniocontainer="quay.io/minio/operator:v$RELEASE" yq -i e '.metadata.annotations.containerImage |= env(miniocontainer)' bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml

yq eval-all -i ". as \$item ireduce ({}; . * \$item )" bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml resources/templates/olm-template.yaml
