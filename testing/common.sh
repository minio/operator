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

yell() { echo "$0: $*" >&2; }

die() {
  yell "$*"
  (kind delete cluster || true ) && exit 111
}

try() { "$@" || die "cannot $*"; }

function setup_kind() {
    try kind create cluster --config "${SCRIPT_DIR}/kind-config.yaml"
    echo "Kind is ready"
    try kubectl get nodes
}

function install_operator() {
    echo "Installing Current Operator"

    # TODO: Compile the current branch and create an overlay to use that image version
    try kubectl apply -k "${SCRIPT_DIR}/../resources"

    echo "Waiting for k8s api"
    sleep 10
    echo "Waiting for Operator Pods to come online (2m timeout)"

    try kubectl wait --namespace minio-operator \
	--for=condition=ready pod \
	--selector=name=minio-operator \
	--timeout=120s
}

function destroy_kind() {
    kind delete cluster
}

function check_tenant_status() {
    # Check MinIO is accessible

    waitdone=0
    totalwait=0
    while true; do
	waitdone=$(kubectl -n $1 get pods -l v1.min.io/tenant=$2 --no-headers | wc -l)
	if [ "$waitdone" -ne 0 ]; then
	    echo "Found $waitdone pods"
	    break
	fi
	sleep 5
	totalwait=$((totalwait + 5))
	if [ "$totalwait" -gt 305 ]; then
	    echo "Unable to create tenant after 5 minutes, exiting."
	    try false
	fi
    done

    echo "Tenant is created successfully, proceeding to validate 'mc admin info minio/'"

    kubectl run admin-mc -i --tty --image minio/mc --command -- bash -c "until (mc alias set minio/ https://minio.$1.svc.cluster.local $3 $4); do echo \"...waiting... for 5secs\" && sleep 5; done; mc admin info minio/;"

    echo "Done."
}