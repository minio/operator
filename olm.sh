#!/bin/bash

set -e

#binary versions
OPERATOR_SDK_VERSION=v1.22.2
TMP_BIN_DIR="$(mktemp -d)"

function install_binaries() {

  if ! operator-sdk version; then
    echo "Installing temporary Binaries into: $TMP_BIN_DIR";
    echo "Installing temporary operator-sdk binary: $OPERATOR_SDK_VERSION"
    ARCH=`{ case "$(uname -m)" in "x86_64") echo -n "amd64";; "aarch64") echo -n "arm64";; *) echo -n "$(uname -m)";; esac; }`
    OS=$(uname | awk '{print tolower($0)}')
    OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/$OPERATOR_SDK_VERSION
    curl -L ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH} -o ${TMP_BIN_DIR}/operator-sdk
    OPERATOR_SDK_BIN=${TMP_BIN_DIR}/operator-sdk
    chmod +x $OPERATOR_SDK_BIN
  else
    OPERATOR_SDK_BIN="$(which operator-sdk)"
  fi
}

install_binaries

# get the minio version
minioVersionInExample=$(kustomize build examples/kustomization/tenant-openshift | yq eval-all '.spec.image' | tail -1)
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
  kustomize build config/manifests | $OPERATOR_SDK_BIN generate bundle \
    --package $package \
    --version $RELEASE \
    --manifests \
    --metadata \
    --output-dir bundles/$catalog/$RELEASE \
    --channels stable \
    --overwrite \
    --use-image-digests \
    --kustomize-dir config/manifests

  # Set the version, later in olm-post-script.sh we change for Digest form.
  containerImage="quay.io/minio/operator:v${RELEASE}"
  echo "containerImage: ${containerImage}"
  operatorImageDigest="quay.io/minio/operator:v${RELEASE}"
  yq -i ".metadata.annotations.containerImage |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i ".spec.install.spec.deployments[1].spec.template.spec.containers[0].image |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # To provide channel for upgrade where we tell what versions can be replaced by the new version we offer
  # You can read the documentation at link below:
  # https://access.redhat.com/documentation/en-us/openshift_container_platform/4.2/html/operators/understanding-the-operator-lifecycle-manager-olm#olm-upgrades_olm-understanding-olm
  echo "To provide replacement for upgrading Operator..."
  PREV_VERSION=$(curl -s "https://catalog.redhat.com/api/containers/v1/operators/bundles?channel_name=stable&package=${package}&organization=${catalog}&include=data.version,data.csv_name,data.ocp_version" | jq '.data |  max_by(.version | split(".") | map(tonumber)).csv_name' -r)
  echo "replaces: $PREV_VERSION"
  yq -i e ".spec.replaces |= \"${PREV_VERSION}\"" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # Now promote the latest release to the root of the repository
  rm -Rf manifests
  rm -Rf metadata

  mkdir -p $catalog
  cp -R bundles/$catalog/$RELEASE/manifests $catalog
  cp -R bundles/$catalog/$RELEASE/metadata $catalog

  sed -i -e '/metrics/d' bundle.Dockerfile
  sed -i -e '/scorecard/d' bundle.Dockerfile
  sed -i -e '/testing/d' bundle.Dockerfile

  echo "clean released annotations"
  sed -i -e '/metrics/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml
  sed -i -e '/scorecard/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml
  sed -i -e '/testing/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml

  # Add openshift supported version & default channel
  # It needs to be added, since you have to declare both the potential eligible
  # channels for an operator via operators.operatorframework.io.bundle.channels
  # as well as the default.
  {
    echo "  # Annotations to specify OCP versions compatibility."
    echo "  com.redhat.openshift.versions: v4.8-v4.14"
    echo "  # Annotation to add default bundle channel as potential is declared"
    echo "  operators.operatorframework.io.bundle.channel.default.v1: stable"
    echo "  operatorframework.io/suggested-namespace: minio-operator"
  } >> bundles/$catalog/$RELEASE/metadata/annotations.yaml

  echo "clean root level annotations.yaml"
  sed -i -e '/metrics/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml
  sed -i -e '/scorecard/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml
  sed -i -e '/testing/d' bundles/$catalog/$RELEASE/metadata/annotations.yaml
done
echo " "
echo "clean -e files"
rm -vf $(git ls-files --others | grep -e "-e$" | awk '{print $1}')

echo "Copying latest bundle to root"
cp -R bundles/redhat-marketplace/$RELEASE/manifests manifests
cp -R bundles/redhat-marketplace/$RELEASE/metadata metadata

echo "Commit all assets"
#git add -u
#git add bundles
#git add community-operators
#git add helm-releases

echo "Removing temporary binaries in: $TMP_BIN_DIR"
rm -rf $TMP_BIN_DIR