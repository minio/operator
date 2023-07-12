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

    install_operator "sts"

    install_tenant "policy-binding"

    check_tenant_status minio-tenant-1 myminio

    setup_sts_bucket

#    install_sts_client "minio-sdk-dotnet" &

    install_sts_client "minio-sdk-go" &

    #install_sts_client "minio-sdk-java" &

    # install_sts_client "minio-sdk-javascript" &

    install_sts_client "minio-sdk-python" &

    install_sts_client "aws-sdk-python" &

    wait

    destroy_kind
}

main "$@"
