#!/usr/bin/env bash
# Copyright (C) 2023, MinIO, Inc.
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

function main() {
    destroy_kind

    setup_kind

    install_operator

    install_tenant "policyBinding-cm"

    check_tenant_status minio-tenant-1 storage-policyBinding

    setup_sts_bucket

    # install_sts_client "minio-dotnet" "cm"

    install_sts_client "minio-go" "cm"
    
    install_sts_client "minio-java" "cm"

    # install_sts_client "minio-javascript"

    install_sts_client "minio-python"

    install_sts_client "aws-python"

    destroy_kind
}

main "$@"
