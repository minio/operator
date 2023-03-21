#!/bin/bash

set -e

get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

MINIO_RELEASE=$(get_latest_release minio/minio)
KES_RELEASE=$(get_latest_release minio/kes)

KES_CURRENT_RELEASE=$(sed -nr 's/.*(minio\/kes\:)([v]?.*)"/\2/p' pkg/apis/minio.min.io/v2/constants.go)

files=(
  "api/consts.go"
  "docs/tenant_crd.adoc"
  "docs/policybinding_crd.adoc"
  "docs/templates/asciidoctor/gv_list.tpl"
  "examples/kustomization/base/tenant.yaml"
  "examples/kustomization/tenant-certmanager-kes/tenant.yaml"
  "examples/kustomization/tenant-kes-encryption/tenant.yaml"
  "helm/operator/Chart.yaml"
  "helm/operator/values.yaml"
  "helm/tenant/Chart.yaml"
  "helm/tenant/values.yaml"
  "kubectl-minio/README.md"
  "kubectl-minio/cmd/helpers/constants.go"
  "kubectl-minio/cmd/tenant-upgrade.go"
  "pkg/apis/minio.min.io/v2/constants.go"
  "pkg/controller/operator.go"
  "resources/base/deployment.yaml"
  "resources/base/console-ui.yaml"
  "update-operator-krew.py"
  "testing/console-tenant+kes.sh"
  "web-app/src/screens/Console/Tenants/AddTenant/Steps/Images.tsx"
  "web-app/src/screens/Console/Tenants/TenantDetails/TenantEncryption.tsx")

CURRENT_RELEASE=$(get_latest_release minio/operator)
CURRENT_RELEASE="${CURRENT_RELEASE:1}"

echo "MinIO: $MINIO_RELEASE"
echo "Upgrade: $CURRENT_RELEASE => $RELEASE"
echo "KES: $KES_CURRENT_RELEASE => $KES_RELEASE"

if [ -z "$MINIO_RELEASE" ]; then
  echo "\$MINIO_RELEASE is empty"
  exit 0
fi

for file in "${files[@]}"; do
  sed -i -e "s/${CURRENT_RELEASE}/${RELEASE}/g" "$file"
  sed -i -e "s/RELEASE\.[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]-[0-9][0-9]-[0-9][0-9]Z/${MINIO_RELEASE}/g" "$file"
  sed -i -e "s/${KES_CURRENT_RELEASE}/${KES_RELEASE}/g" "$file"
done

echo "Re-indexing helm chart releases for $RELEASE"
./helm-reindex.sh

# Add all the generated files to git

echo "clean -e files"
rm -vf $(git ls-files --others | grep -e "-e$" | awk '{print $1}')
git add .
