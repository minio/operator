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
