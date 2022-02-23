#!/bin/bash

RELEASE=4.4.9
EXAMPLE=$(kustomize build examples/kustomization/tenant-lite| yq eval-all '. | [.]' | yq  'del( .[] | select(.kind == "Namespace") )'| yq  'del( .[] | select(.kind == "Secret") )' | yq -o json | jq -c )

operator-sdk generate bundle \
  --package minio-operator \
  --version $RELEASE \
  --deploy-dir resources/base \
  --crds-dir resources/base/crds \
  --manifests \
  --metadata \
  --output-dir bundles/$RELEASE \
  --channels stable

myenv=$EXAMPLE yq -i e ".metadata.annotations.alm-examples |= (\"\${myenv}\" | envsubst)" bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml

miniocontainer="quay.io/minio/operator:v$RELEASE" yq -i e '.metadata.annotations.containerImage |= env(miniocontainer)' bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml

yq eval-all -i ". as \$item ireduce ({}; . * \$item )" bundles/$RELEASE/manifests/minio-operator.clusterserviceversion.yaml resources/templates/olm-template.yaml

# Now promote the latest release to the root of the repository

rm -Rf manifests
rm -Rf metadata

cp -R bundles/$RELEASE/manifests manifests
cp -R bundles/$RELEASE/metadata metadata

sed -i -e '/metrics/d' bundle.Dockerfile
sed -i -e '/scorecard/d' bundle.Dockerfile
sed -i -e '/testing/d' bundle.Dockerfile

# clean released annotations
sed -i -e '/metrics/d' bundles/$RELEASE/metadata/annotations.yaml
sed -i -e '/scorecard/d' bundles/$RELEASE/metadata/annotations.yaml
sed -i -e '/testing/d' bundles/$RELEASE/metadata/annotations.yaml

# clean root level annotations.yaml
sed -i -e '/metrics/d' metadata/annotations.yaml
sed -i -e '/scorecard/d' metadata/annotations.yaml
sed -i -e '/testing/d' metadata/annotations.yaml