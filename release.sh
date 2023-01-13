#!/bin/bash

set -e

get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

MINIO_RELEASE=$(get_latest_release minio/minio)
CONSOLE_RELEASE=$(get_latest_release minio/console)
CONSOLE_RELEASE="${CONSOLE_RELEASE:1}"

# Figure out the FROM console release we are updating from
CONSOLE_CURRENT_RELEASE=$(grep -Eo 'minio\/console:v([0-9]?[0-9].[0-9]?[0-9].[0-9]?[0-9])' resources/base/console-ui.yaml | grep -Eo '([0-9]?[0-9].[0-9]?[0-9].[0-9]?[0-9])')

files=("docs/crd.adoc" "docs/templates/asciidoctor/gv_list.tpl" "examples/kustomization/base/tenant.yaml" "helm/operator/Chart.yaml" "helm/operator/values.yaml" "helm/tenant/Chart.yaml" "kubectl-minio/README.md" "kubectl-minio/cmd/helpers/constants.go" "kubectl-minio/cmd/tenant-upgrade.go" "pkg/apis/minio.min.io/v2/constants.go" "resources/base/deployment.yaml" "update-operator-krew.py" "resources/base/console-ui.yaml")

CURRENT_RELEASE=$(get_latest_release minio/operator)
CURRENT_RELEASE="${CURRENT_RELEASE:1}"

echo "MinIO: $MINIO_RELEASE"
echo "Upgrade: $CURRENT_RELEASE => $RELEASE"
echo "Console: $CONSOLE_CURRENT_RELEASE => $CONSOLE_RELEASE"

if [ -z "$MINIO_RELEASE" ]; then
  echo "\$MINIO_RELEASE is empty"
  exit 0
fi

for file in "${files[@]}"; do
  sed -i -e "s/${CURRENT_RELEASE}/${RELEASE}/g" "$file"
  sed -i -e "s/RELEASE\.[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]-[0-9][0-9]-[0-9][0-9]Z/${MINIO_RELEASE}/g" "$file"
  sed -i -e "s/${CONSOLE_CURRENT_RELEASE}/${CONSOLE_RELEASE}/g" "$file"
done

echo "Re-indexing helm chart releases for $RELEASE"
./helm-reindex.sh

# Add all the generated files to git

echo "clean -e files"
rm -vf $(git ls-files --others | grep -e "-e$" | awk '{print $1}')
git add .
