#!/usr/bin/env bash
# Copyright (C) 2024, MinIO, Inc.
#
# This code is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License, version 3,
# as published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License, version 3,
# along with this program.  If not, see <http://www.gnu.org/licenses/>

# This script requires: kubectl, kind

SCRIPT_DIR=$(dirname "$0")

export SCRIPT_DIR

source "${SCRIPT_DIR}/common.sh"

lower_version="$1"
hotfix_version="$2"
namespace="tenant-lite"
tenant="myminio"
bucket="data"
dummy="dummy.data"
localport="9000"
alias="minios3"

# Announce test
function announce_test() {
  local lower_text
  local hotfix_version_text
  if [ -n "$lower_version" ] 
  then 
    lower_text=$lower_version; 
  else 
    echo "missing MinIO version"
    exit 1
  fi

  if [ -n "$hotfix_version" ] 
  then 
    hotfix_version_text=$hotfix_version;
  else 
    echo "missing lower version"
		exit 1
  fi

  echo "## Testing upgrade of Tenant from version $lower_text to $hotfix_version_text ##"
}

# Preparing tenant for bucket manipulation
# shellcheck disable=SC2317
function bootstrap_tenant() {
  port_forward $namespace $tenant minio $localport

  # Obtain root credentials
  TENANT_CONFIG_SECRET=$(kubectl -n $namespace get tenants $tenant -o jsonpath="{.spec.configuration.name}")
  USER=$(kubectl -n $namespace get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_USER="' | sed -e 's/export MINIO_ROOT_USER="//g' | sed -e 's/"//g')
  PASSWORD=$(kubectl -n $namespace get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_PASSWORD="' | sed -e 's/export MINIO_ROOT_PASSWORD="//g' | sed -e 's/"//g')

  echo "Creating alias with user ${USER}"
  mc alias set $alias https://localhost:$localport ${USER} ${PASSWORD} --insecure

  echo "Creating bucket on tenant"
  mc mb $alias/$bucket --insecure
}

# Upload dummy data to tenant bucket
function upload_dummy_data() {
  port_forward $namespace $tenant minio $localport

  echo "Uploading dummy data to tenant bucket"
  cp ${SCRIPT_DIR}/tenant-hotfix-update.sh ${SCRIPT_DIR}/$dummy
  mc cp ${SCRIPT_DIR}/$dummy $alias/$bucket/$dummy --insecure
}

# Download dummy data from tenant bucket
function download_dummy_data() {
  port_forward $namespace $tenant minio $localport

  echo "Download dummy data from tenant bucket"
  mc cp $alias/$bucket/$dummy ${SCRIPT_DIR}/$dummy --insecure

  if cmp "${SCRIPT_DIR}/tenant-hotfix-update.sh" "${SCRIPT_DIR}/$dummy"; then
    echo "Tenant hotfix update test complete; no issue found"
  else
    echo "Tenant hotfix update failed"
    try false
  fi
}

function install_tenant_with_image() {
	minio_image="$1"
	if [ -z "$1" ]
	then
		echo "MinIO version is not set"
		exit 1
	fi
	kustomize build "${SCRIPT_DIR}/../examples/kustomization/tenant-lite" > tenant-lite.yaml
	yq -i e "select(.kind == \"Tenant\").spec.image = \"${minio_image}\"" tenant-lite.yaml

	try kubectl apply -f tenant-lite.yaml
}

function main() {
  announce_test

  destroy_kind

  setup_kind

  install_operator

  echo "Installing tenant with Image: $lower_version"

  install_tenant_with_image "$lower_version"

  check_tenant_status tenant-lite myminio

  bootstrap_tenant

  upload_dummy_data

  install_tenant_with_image "$hotfix_version"
  
  check_tenant_status tenant-lite myminio

  download_dummy_data

  destroy_kind
}

main "$@"
