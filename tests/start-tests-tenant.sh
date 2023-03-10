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

function install_tenants() {
	echo "Installing tenants"

	# Install lite & kes tenants
	try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/tenant-lite"
	try kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/tenant-kes-encryption"

	echo "Waiting for the tenant statefulset, this indicates the tenant is being fulfilled"
	waitdone=0
	totalwait=0
	while true; do
	waitdone=$(kubectl -n tenant-lite get pods -l v1.min.io/tenant=myminio --no-headers | wc -l)
	if [ "$waitdone" -ne 0 ]; then
	    echo "Found $waitdone pods"
	    break
	fi
	sleep 5
	totalwait=$((totalwait + 5))
	if [ "$totalwait" -gt 300 ]; then
		echo "Tenant never created statefulset after 5 minutes"
		try false
	fi
	done

	echo "Waiting for tenant pods to come online (5m timeout)"
	try kubectl wait --namespace tenant-lite \
	--for=condition=ready pod \
	--selector="v1.min.io/tenant=myminio" \
	--timeout=300s

	echo "Build passes basic tenant creation"
}


function main() {
	destroy_kind
	setup_kind
	install_operator
	install_tenants
	check_tenant_status tenant-lite myminio
	kubectl proxy &
	# Beginning  Kubernetes 1.24 ----> Service Account Token Secrets are not 
	# automatically generated, to generate them manually, users must manually
	# create the secret, for our examples where we lead people to get the JWT
	# from the console-sa service account, they additionally need to manually
	# generate the secret via
	kubectl apply -f "${SCRIPT_DIR}/console-sa-secret.yaml"
}

main "$@"
