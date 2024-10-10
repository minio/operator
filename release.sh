#!/bin/bash

set -e

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --release-sidecar)
            RELEASE_SIDECAR="$2"
            shift 2
            ;;
        *)
            ;;
    esac
    shift
done

get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

MINIO_RELEASE=$(get_latest_release minio/minio)
KES_RELEASE=$(get_latest_release minio/kes)
MC_RELEASE=$(get_latest_release minio/mc)

MINIO_CURRENT_RELEASE=$(sed -nr 's/.*(minio\/minio\:)([v]?.*)"/\2/p' pkg/apis/minio.min.io/v2/constants.go)
KES_CURRENT_RELEASE=$(sed -nr 's/.*(minio\/kes\:)([v]?.*)"/\2/p' pkg/apis/minio.min.io/v2/constants.go)

files=(
  "README.md"
  "pkg/apis/job.min.io/v1alpha1/types.go"
  "docs/tenant_crd.adoc"
  "docs/policybinding_crd.adoc"
  "docs/job_crd.adoc"
  "docs/minio-job.md"
  "docs/templates/asciidoctor/gv_list.tpl"
  "examples/kustomization/base/tenant.yaml"
  "examples/kustomization/tenant-certmanager-kes/tenant.yaml"
  "examples/kustomization/tenant-kes-encryption/tenant.yaml"
  "helm/operator/Chart.yaml"
  "helm/operator/values.yaml"
  "helm/tenant/Chart.yaml"
  "helm/tenant/values.yaml"
  "pkg/apis/minio.min.io/v2/constants.go"
  "pkg/controller/operator.go"
  "resources/base/deployment.yaml"
  "testing/console-tenant+kes.sh"
)

CURRENT_RELEASE=$(get_latest_release minio/operator)
CURRENT_RELEASE="${CURRENT_RELEASE:1}"

echo "Upgrade: $CURRENT_RELEASE => $RELEASE"
echo "MinIO: $MINIO_RELEASE => $MINIO_RELEASE"
echo "KES: $KES_CURRENT_RELEASE => $KES_RELEASE"
echo "MC: $MC_CURRENT_RELEASE => $MC_RELEASE"

if [ -z "$MINIO_RELEASE" ]; then
  echo "\$MINIO_RELEASE is empty"
  exit 0
fi

for file in "${files[@]}"; do
	sed -i -e "s/${KES_CURRENT_RELEASE}/${KES_RELEASE}/g" "$file"
	sed -i -e "s/${MC_CURRENT_RELEASE}/${MC_RELEASE}/g" "$file"
  sed -i -e "s/${CURRENT_RELEASE}/${RELEASE}/g" "$file"
  sed -i -e "s/${MINIO_CURRENT_RELEASE}/${MINIO_RELEASE}/g" "$file"
done

annotations_files=(
  "pkg/apis/job.min.io/v1alpha1/types.go"
  "pkg/apis/minio.min.io/v2/types.go"
  "pkg/apis/sts.min.io/v1beta1/types.go"
)

for file in "${annotations_files[@]}"; do
  sed -i -e "s~operator.min.io/version=.*~operator.min.io/version=v${RELEASE}~g" "$file"
done

# Update annotation in kustomization yaml
sed -i -e "s~operator.min.io/version: .*~operator.min.io/version: v${RELEASE}~g" "resources/kustomization.yaml"

if [ "${RELEASE_SIDECAR}" = "true" ]; then

  sed -i -e 's~quay.io/minio/operator-sidecar:.*\"~quay.io/minio/operator-sidecar:v'$RELEASE'\"~g' "pkg/resources/statefulsets/minio-sidecar.go"
fi

# Add all the generated files to git

echo "clean -e files"
rm -vf $(git ls-files --others | grep -e "-e$" | awk '{print $1}')
