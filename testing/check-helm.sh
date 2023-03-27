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

# This script requires: kubectl, kind, jq

SCRIPT_DIR=$(dirname "$0")
export SCRIPT_DIR
source "${SCRIPT_DIR}/common.sh"

function main() {
    
    echo "destroy kind cluster to start with"
    destroy_kind

    echo "setup kind right after it has been destroyed"
    setup_kind

    echo "Install operator with helm"
    install_operator "helm"

    echo "Install tenant with helm"
    install_tenant "helm"

    echo "check tenant status"
    check_tenant_status default myminio minio "helm"

    echo "destroy kind cluster to end with"
    destroy_kind
}

main "$@"
