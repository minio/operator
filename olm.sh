#!/bin/bash

set -e

#binary versions
OPERATOR_SDK_VERSION=v1.34.1
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

RELEASE="$(sed -n 's/^.*DefaultOperatorImage = "minio\/operator:v\(.*\)"/\1/p' pkg/controller/operator.go)"

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

# This constants are supported Openshift versions
minOpenshiftVersion=4.8
maxOpenshiftVersion=4.15

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

  # https://connect.redhat.com/support/technology-partner/#/case/03206318
  # If no securityContext is specified, the OLM will choose one that fits within
  # the security context constraint either explicitly specified for the project under which the pod is run,
  # or the default. If the SCC specifies a value that doesn't match the specified value in our files,
  # the pods will not start properly and we can't be installed.
  # Let the user select their own securityContext and don't hardcode values that can affect the ability
  # to debug and deploy our Operator in OperatorHub.
  echo "Removing securityContext from CSV"
  yq -i eval 'del(.spec.install.spec.deployments[].spec.template.spec.containers[0].securityContext)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # Will query if a previous version of the CSV was published to the catalog of the latest supported Openshift version.
  # It will help to prevent add the `spec.replaces` annotation when there is no preexisting CSV in the catalog to replace.
  # See support case https://connect.redhat.com/support/technology-partner/#/case/03671253
  prev=$(curl -s "https://catalog.redhat.com/api/containers/v1/operators/bundles?channel_name=stable&package=${package}&organization=${catalog}&ocp_version=${maxOpenshiftVersion}&include=data.version,data.csv_name,data.ocp_version" | jq '.data | length' -r)

  # only add `spec.replaces` if at least one version have been published to the catalog
  if [ "$prev" -gt 0 ]; then
    # To provide channel for upgrade where we tell what versions can be replaced by the new version we offer
    # You can read the documentation at link below:
    # https://access.redhat.com/documentation/en-us/openshift_container_platform/4.2/html/operators/understanding-the-operator-lifecycle-manager-olm#olm-upgrades_olm-understanding-olm
    echo "To provide replacement for upgrading Operator..."
    PREV_VERSION=$(curl -s "https://catalog.redhat.com/api/containers/v1/operators/bundles?channel_name=stable&package=${package}&organization=${catalog}&ocp_version=${maxOpenshiftVersion}&include=data.version,data.csv_name,data.ocp_version" | jq '.data | max_by(.version).csv_name' -r)
    echo "replaces: $PREV_VERSION"
    yq -i e ".spec.replaces |= \"${PREV_VERSION}\"" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    # We need to remove "skips" and only use "replaces"
    yq -i "del(.spec.skips) " bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  else
    # This procedure is needed when a new Index is released
    # Having a new Index (ie Openshift 4.15) means that Operator haven't been published in it yet, so the "replaces" annotation fails.
    # To prevent it we reached out to RedHat support and told us to use "skips" instead, that way we can keep the update
    # graph and publish Operator in the new Index for the first time.
    # https://connect.redhat.com/support/technology-partner/#/case/03793912
    echo "no previous published in index ${maxOpenshiftVersion}, removing spec.replaces"
    yq -i "del(.spec.replaces) " bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    echo "adding spec.skips for new index"
    # Get the previous Openshift Index
    previousOpenshiftVersion=$(curl -s "https://catalog.redhat.com/api/containers/v1/operators/indices?ocp_versions_range=${minOpenshiftVersion}-${maxOpenshiftVersion}&organization=${catalog}" | yq  '.data.[].ocp_version' | sort -V | tail -n2 | head -n1)
    # Get the latest published operator in the previous Openshift Index
    PREV_VERSION=$(curl -s "https://catalog.redhat.com/api/containers/v1/operators/bundles?channel_name=stable&package=${package}&organization=${catalog}&ocp_version=${previousOpenshiftVersion}&include=data.version,data.csv_name,data.ocp_version" | jq '.data | max_by(.version).csv_name' -r)
    echo "skips: $PREV_VERSION"
    yq -i e ".spec.skips += [\"${PREV_VERSION}\"]" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  fi

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
    echo "  com.redhat.openshift.versions: v${minOpenshiftVersion}-v${maxOpenshiftVersion}"
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