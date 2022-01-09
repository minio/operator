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

    # To compile current branch
    echo "Compiling Current Branch Operator"
    (cd "${SCRIPT_DIR}/.." && make) # will not change your shell's current directory

    echo "print kustomize version"
    kustomize version

    echo 'start - load compiled image'
    kind load docker-image docker.io/minio/operator:dev
    echo 'end - load compiled image'

    echo "Installing Current Operator"
    # TODO: Compile the current branch and create an overlay to use that image version
    # dev contains the patch for dev image
    try kubectl apply -k "${SCRIPT_DIR}/../resources/dev"

    # I will remove this line, just want to print the image version we are using for debugging
    echo "pods images"
    kubectl describe pods -n minio-operator | grep Image

    echo "Waiting for k8s api"
    
    # This is just for testing, I want to see the Failed to pull image message
    sleep 130
    kubectl get pods --namespace minio-operator

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

    echo "Waiting for pods to be ready. (5m timeout)"

    USER=$(kubectl -n $1 get secrets $2-env-configuration -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_USER="' | sed -e 's/export MINIO_ROOT_USER="//g' | sed -e 's/"//g')
    PASSWORD=$(kubectl -n $1 get secrets $2-env-configuration -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_PASSWORD="' | sed -e 's/export MINIO_ROOT_PASSWORD="//g' | sed -e 's/"//g')

    try kubectl wait --namespace $1 \
        --for=condition=ready pod \
        --selector=v1.min.io/tenant=$2 \
        --timeout=300s

    echo "Tenant is created successfully, proceeding to validate 'mc admin info minio/'"

    kubectl run admin-mc -i --tty --image minio/mc --command -- bash -c "until (mc alias set minio/ https://minio.$1.svc.cluster.local $USER $PASSWORD); do echo \"...waiting... for 5secs\" && sleep 5; done; mc admin info minio/;"

    echo "Done."
}