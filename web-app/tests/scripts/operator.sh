#!/bin/bash

# This file is part of MinIO Console Server
# Copyright (c) 2022 MinIO, Inc.
# # This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# # This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
# # You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

SCRIPT_DIR=$(dirname "$0")
export SCRIPT_DIR
source "${SCRIPT_DIR}/common.sh" # This is common.sh for TestCafe Tests
source "${GITHUB_WORKSPACE}/tests/common.sh" # This is common.sh for k8s tests.

## this enables :dev tag for minio/operator container image.
CI="true"
export CI

## Make sure to install things if not present already
sudo curl -#L "https://dl.k8s.io/release/v1.23.1/bin/linux/amd64/kubectl" -o /usr/local/bin/kubectl
sudo chmod +x /usr/local/bin/kubectl

sudo curl -#L "https://dl.min.io/client/mc/release/linux-amd64/mc" -o /usr/local/bin/mc
sudo chmod +x /usr/local/bin/mc

__init__() {
	TIMESTAMP=$(date "+%s")
	export TIMESTAMP
 	echo $TIMESTAMP > web-app/tests/constants/timestamp.txt
	GOPATH=/tmp/gopath
	export GOPATH
	PATH=${PATH}:${GOPATH}/bin
	export PATH
	destroy_kind
	setup_kind
	install_operator
	install_tenant
	echo "kubectl proxy"
	kubectl proxy &
#	echo "yarn start"
#	yarn start &
	echo "Start Operator UI"
	./minio-operator ui &
	echo "DONE with kind, yarn and console, next is testcafe"
	exit 0
}

( __init__ "$@")
