#!/bin/bash

set -e

#binary versions
OPERATOR_SDK_VERSION=v1.22.2
TMP_BIN_DIR="$(mktemp -d)"

function install_binaries() {
  
  echo "Installing temporary Binaries into: $TMP_BIN_DIR";
  echo "Installing temporary operator-sdk binary: $OPERATOR_SDK_VERSION"
  ARCH=`{ case "$(uname -m)" in "x86_64") echo -n "amd64";; "aarch64") echo -n "arm64";; *) echo -n "$(uname -m)";; esac; }`
  OS=$(uname | awk '{print tolower($0)}')
  OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/$OPERATOR_SDK_VERSION
  curl -L ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH} -o ${TMP_BIN_DIR}/operator-sdk
  OPERATOR_SDK_BIN=${TMP_BIN_DIR}/operator-sdk
  chmod +x $OPERATOR_SDK_BIN
}

install_binaries

# get the minio version
minioVersionInExample=$(kustomize build examples/kustomization/tenant-openshift | yq eval-all '.spec.image' | tail -1)
echo "minioVersionInExample: ${minioVersionInExample}"

# Get sha form of minio version
minioVersionDigest=$(docker pull $minioVersionInExample | grep Digest | awk -F ' ' '{print $2}')
minioVersionDigest="quay.io/minio/minio@${minioVersionDigest}"
echo "minioVersionDigest: ${minioVersionDigest}"

# Generate the alm-examples
EXAMPLE=$(kustomize build examples/kustomization/tenant-openshift | yq ".spec.image = \"${minioVersionDigest}\"" | yq eval-all '. | [.]' | yq  'del( .[] | select(.kind == "Namespace") )'| yq  'del( .[] | select(.kind == "Secret") )' | yq -o json | jq -c )

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
  $OPERATOR_SDK_BIN generate bundle \
    --package $package \
    --version $RELEASE \
    --deploy-dir resources/base \
    --crds-dir resources/base/crds \
    --manifests \
    --metadata \
    --output-dir bundles/$catalog/$RELEASE \
    --channels stable \
    --overwrite \
    --use-image-digests

  # deploymentName has to be minio-operator, the reason is in case/03206318 or redhat support.
  # the deployment name you set is "operator", and in CSV, there are two deployments 'console' and 'minio-operator'
  # but there is no 'operator' option, without this change error is: "calculated deployment install is bad"
  yq -i 'del(.spec.webhookdefinitions)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml 

  # I get the update from engineering team, typically no user/group specification is made in a container image.
  # Rather, the user spec (if there is one) is placed in the clusterserviceversion.yaml file as a RunAsUser clause.
  # If no userid/groupid is specified, the OLM will choose one that fits within the security context constraint either
  # explicitly specified for the project under which the pod is run, or the default. If the SCC specifies a userid range
  # that doesn't include the specified value, the pod will not start properly. So you need to remove folowing items in securityContext
  yq -i eval 'del(.spec.install.spec.deployments[0].spec.template.spec.containers[0].securityContext.runAsGroup)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i eval 'del(.spec.install.spec.deployments[0].spec.template.spec.containers[0].securityContext.runAsUser)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i eval 'del(.spec.install.spec.deployments[1].spec.template.spec.containers[0].securityContext.runAsGroup)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i eval 'del(.spec.install.spec.deployments[1].spec.template.spec.containers[0].securityContext.runAsUser)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # To support different architectures.
  echo "Supporting different architectures..."
  yq -i e '.spec.labels."operatorframework.io/arch.386" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.amd64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.amd64p32" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.arm" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.armbe" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.arm64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.arm64be" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.loong64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mips" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mipsle" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mips64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mips64le" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mips64p32" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.mips64p32le" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.ppc" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.ppc64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.ppc64le" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.riscv" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.riscv64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.s390" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.s390x" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.sparc" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.sparc64" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/arch.wasm" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # To support different Operating Systems
  echo "Supporting different Operating Systems..."
  yq -i e '.spec.labels."operatorframework.io/os.aix" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.android" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.darwin" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.dragonfly" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.freebsd" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.hurd" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.illumos" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.ios" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.js" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.linux" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.nacl" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.netbsd" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.openbsd" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.plan9" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.solaris" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.windows" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i e '.spec.labels."operatorframework.io/os.zos" |= "supported"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # To provide channel for upgrade where we tell what versions can be replaced by the new version we offer
  # You can read the documentation at link below:
  # https://access.redhat.com/documentation/en-us/openshift_container_platform/4.2/html/operators/understanding-the-operator-lifecycle-manager-olm#olm-upgrades_olm-understanding-olm
  echo "To provide replacement for upgrading Operator..."
  echo "WARNING: This need to be updated manually prior to publish in redhat catalogs"
  yq -i e ".spec.\"replaces\" |= \"${package}.v${RELEASE}\"" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # In order to deploy via OLM, we should let OLM to decide on the security
  # context; otherwise deploy will fail and operator update will not be possible
  # already tested this manually to prove the point:
  # https://github.com/k8s-operatorhub/community-operators/pull/1212
  # It only applies for community operators
  if [[ "$catalog" == "community-operators" ]]
  then
    yq -i eval 'del(.spec.install.spec.deployments[0].spec.template.spec.containers[0].securityContext.runAsNonRoot)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    yq -i eval 'del(.spec.install.spec.deployments[1].spec.template.spec.containers[0].securityContext.runAsNonRoot)' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  fi

  # annotations-validation: To fix this issue define the annotation in 'manifests/*.clusterserviceversion.yaml' file.
  # [annotations-validation : bundle-parse] + EXPECTED_MARKETPLACE_SUPPORT_WORKFLOW='https://marketplace.redhat.com/en-us/operators/minio-operator-rhmp/support?utm_source=openshift_console'
  # [annotations-validation : bundle-parse] + EXPECTED_MARKETPLACE_REMOTE_WORKFLOW='https://marketplace.redhat.com/en-us/operators/minio-operator-rhmp/pricing?utm_source=openshift_console'
  if [[ "$catalog" == "redhat-marketplace" ]]
  then
    yq -i '.metadata.annotations.replaceitone = "https://marketplace.redhat.com/en-us/operators/minio-operator-rhmp/pricing?utm_source=openshift_console"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    yq -i '.metadata.annotations.replaceittwo = "https://marketplace.redhat.com/en-us/operators/minio-operator-rhmp/support?utm_source=openshift_console"' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    sed -i -e "s/replaceitone/marketplace.openshift.io\/remote-workflow/" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
    sed -i -e "s/replaceittwo/marketplace.openshift.io\/support-workflow/" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  fi

  myenv=$EXAMPLE yq -i e ".metadata.annotations.alm-examples |= (\"\${myenv}\" | envsubst)" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

  # Set the version, later in olm-post-script.sh we change for Digest form.
  containerImage="quay.io/minio/operator:v${RELEASE}"
  echo "containerImage: ${containerImage}"
  operatorImageDigest="quay.io/minio/operator:v${RELEASE}"
  yq -i ".metadata.annotations.containerImage |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq -i ".spec.install.spec.deployments[1].spec.template.spec.containers[0].image |= (\"${operatorImageDigest}\")" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml
  yq eval-all -i ". as \$item ireduce ({}; . * \$item )" bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml resources/templates/olm-template.yaml

  echo "One replica only for Code Ready Containers with one node"
  yq -i e '.spec.install.spec.deployments[1].spec.replicas |= 1' bundles/$catalog/$RELEASE/manifests/$package.clusterserviceversion.yaml

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
    echo "  com.redhat.openshift.versions: v4.8-v4.12"
    echo "  # Annotation to add default bundle channel as potential is declared"
    echo "  operators.operatorframework.io.bundle.channel.default.v1: stable"
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