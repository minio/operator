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
FINAL_RESULT=1
export SCRIPT_DIR

source "${SCRIPT_DIR}/common.sh"

function perform_attempts_to_get_log_api_response() {
	# This function will perform some attempts to get the API response.
	for i in {1..1000}; do
		for i in {1..10}; do echo ""; done
		echo "kubectl get pods -n tenant-lite"
		kubectl get pods -n tenant-lite
		kubectl port-forward storage-lite-ss-0-0 9443 --namespace tenant-lite &
		process_id=$!
		echo "process_id: ${process_id}"
		echo 'Get token from MinIO Console'
		COOKIE=$(
			curl 'https://localhost:9443/api/v1/login' -vs \
			-H 'content-type: application/json' \
			--data-raw '{"accessKey":"minio","secretKey":"minio123"}' --insecure 2>&1 | \
			grep "set-cookie: token=" | sed -e "s/< set-cookie: token=//g" | \
			awk -F ';' '{print $1}'
		)
		echo "Cookie: ${COOKIE}"
		# If there is no cookie, there is no sense to proceed, so fail if no cookie
		if [ -z "$COOKIE" ]
		then
			echo "\$COOKIE is empty"
		fi
		echo 'Verify Logs via API'
		RESULT=$(
			curl 'https://localhost:9443/api/v1/logs/search?q=reqinfo&pageSize=100&pageNo=0&order=timeDesc' \
			-H 'cookie: token='$COOKIE'' \
			--compressed \
			--insecure | jq '.results[0].response_status'
		)
		echo $RESULT
		EXPECTED_RESULT='"OK"'
		echo $EXPECTED_RESULT
		if [ "$EXPECTED_RESULT" = "$RESULT" ]; then
			echo "Logs are present, no issue found"
			FINAL_RESULT=0
			break
		else
			echo "Logs are unreachable"
		fi
		sleep 30
	done
}

function main() {
    destroy_kind

    setup_kind

    install_operator

    install_tenant

	perform_attempts_to_get_log_api_response
	if [ $FINAL_RESULT = 1 ]; then
		echo "Test failed"
		exit 1
	fi

    destroy_kind
}

main "$@"
