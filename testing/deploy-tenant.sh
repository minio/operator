#!/usr/bin/env bash
# Copyright (C) 2022, MinIO, Inc.
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

    install_tenant

    check_tenant_status tenant-lite storage-lite

    # To allow the execution without killing the cluster at the end of the test
    # Use below statement to automatically test and kill cluster at the end:
    # `unset OPERATOR_ENABLE_MANUAL_TESTING`
    # Use below statement to test and keep cluster alive at the end!:
    # `export OPERATOR_ENABLE_MANUAL_TESTING="ON"`
    if [[ -z "${OPERATOR_ENABLE_MANUAL_TESTING}" ]]; then
        # OPERATOR_ENABLE_MANUAL_TESTING is not defined, hence destroy_kind
        echo "Cluster will be destroyed for automated testing"
        destroy_kind
    else
        echo "Cluster will remain alive for manual testing"
    fi
}

main "$@"
