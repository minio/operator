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

source "${SCRIPT_DIR}/openshift-common.sh"

function main() {

	install_binaries

	setup_crc "4.14.1"

	create_marketplace_catalog "certified-operators"

	install_operator "certified-operators"

	#destroy_crc
}

time main "$@"
