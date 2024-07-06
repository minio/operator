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

function wait_for_resource() {
	waitdone=0
	totalwait=0
	echo "command to wait on:"
	command_to_wait="kubectl -n $1 get pods -l $3=$2 --no-headers"
	echo $command_to_wait

	while true; do
	waitdone=$($command_to_wait | wc -l)
	if [ "$waitdone" -ne 0 ]; then
		echo "Found $waitdone pods"
			break
	fi
	sleep 5
	totalwait=$((totalwait + 5))
	if [ "$totalwait" -gt 305 ]; then
			echo "Unable to get resource after 5 minutes, exiting."
			try false
	fi
	done
}

function wait_for_resource_field_selector() {
  # example 1 job:
  # namespace="minio-tenant-1"
  # codition="condition=Complete"
  # selector="metadata.name=setup-bucket"
  # wait_for_resource_field_selector $namespace job $condition $selector
  #
  # example 2 tenant:
  # wait_for_resource_field_selector $namespace job $condition $selector
  # condition=jsonpath='{.status.currentState}'=Initialized
  # selector="metadata.name=storage-policy-binding"
  # wait_for_resource_field_selector $namespace tenant $condition $selector 900s

  namespace=$1
  resourcetype=$2
  condition=$3
  fieldselector=$4
  if [ $# -ge 5 ]; then
    timeout="$5"
  else
    timeout="600s"
  fi

  echo "Waiting for $resourcetype \"$fieldselector\" for \"$condition\" ($timeout timeout)"
  echo "namespace: ${namespace}"
  kubectl wait -n "$namespace" "$resourcetype" \
    --for=$condition \
    --field-selector $fieldselector \
    --timeout="$timeout"
}

# usage: wait_resource_status <namespace> <resourcetype> <resourcename> [<timeout>]
function wait_resource_status() {
	# This function will wait until the resource has a status field
	# wait_resource_status ns-1 Tenant my-tenant
	# wait_resource_status ns-1 Tenant my-tenant 1200

	local namespace=$1
	local resourcetype=$2
	local resourcename=$3

	# Timeout in seconds
	if [ $# -ge 4 ]; then
		TIMEOUT=$4
	else
		TIMEOUT=600
	fi

	START_TIME=$(date +%s)
	# Interval in seconds between checks
	INTERVAL=10

	echo -e "${BLUE}Waiting for $resourcetype status to be created (${TIMEOUT}s timeout)${NC}"
	while true; do
		if kubectl get "$resourcetype" -n "$namespace" "$resourcename" -ojson | jq -e '.status' >/dev/null 2>&1; then
			echo -e "${CHECK}$resourcetype/$resourcename status found."
			break
		else
			echo "$resourcetype/$resourcename status not found, waiting for $INTERVAL seconds."
			sleep $INTERVAL
		fi

		# Check timeout
		CURRENT_TIME=$(date +%s)
		# shellcheck disable=SC2004
		ELAPSED_TIME=$(($CURRENT_TIME - $START_TIME))

		if [ "$ELAPSED_TIME" -ge "$TIMEOUT" ]; then
			echo "Timeout waiting for $resourcetype/$resourcename to be created."
			exit 1
		fi
	done
}