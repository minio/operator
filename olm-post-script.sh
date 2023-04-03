#!/bin/bash

# get the minio version
minioVersionInExample=$(kustomize build examples/kustomization/tenant-lite | yq eval-all '.spec.image' | tail -1)
echo "minioVersionInExample: ${minioVersionInExample}"

# Get sha form of minio version
minioVersionDigest=$(docker pull $minioVersionInExample | grep Digest | awk -F ' ' '{print $2}')
minioVersionDigest="quay.io/minio/minio@${minioVersionDigest}"
echo "minioVersionDigest: ${minioVersionDigest}"

# There are 4 catalogs in Red Hat, we are interested in two of them:
# https://docs.openshift.com/container-platform/4.7/operators/understanding/olm-rh-catalogs.html
# 1. redhat-operators <------------ Supported by Red Hat.
# 2. certified-operators <--------- Supported by the ISV (independent software vendors) <------------- We want this!
# 3. redhat-marketplace <---------- an be purchased from Red Hat Marketplace. <----------------------- We want this!
# 4. community-operators <--------- No official support.

redhatCatalogs=("certified-operators" "redhat-marketplace" "community-operators")

for catalog in "${redhatCatalogs[@]}"; do
  echo " "
  echo $catalog
  package=minio-operator
  if [[ "$catalog" == "redhat-marketplace" ]]
  then
    package=minio-operator-rhmp
  fi
  echo "package: ${package}"

  # Avoid message: "There are unpinned images digests!" by using Digest Sha256:xxxx rather than vx.x.x
  containerImage="quay.io/minio/operator:v$RELEASE"
  echo "containerImage: ${containerImage}"
  digest=$(docker pull $containerImage | grep Digest | awk -F ' ' '{print $2}')
  operatorImageDigest="quay.io/minio/operator@${digest}"
  echo "operatorImageDigest: ${operatorImageDigest} @ ${digest}"
  yq -i ".metadata.annotations.containerImage |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # Console Image in Digested form: sha256:xxxx
  consoleImage=$(yq eval-all '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml)
  echo "consoleImage: ${consoleImage}"
  consoleImageDigest=$(docker pull ${consoleImage} | grep Digest | awk -F ' ' '{print $2}')
  echo "consoleImageDigest: ${consoleImageDigest}"
  consoleImageDigest="quay.io/minio/console@${consoleImageDigest}"
  yq -i ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image |= (\"${consoleImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # Operator Image in Digest mode: sha256:xxx
  yq -i ".spec.install.spec.deployments[1].spec.template.spec.containers[0].image |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq eval-all -i ". as \$item ireduce ({}; . * \$item )" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml resources/templates/olm-template.yaml

done
echo " "
