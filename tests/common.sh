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
	# TODO once feature is added: https://github.com/kubernetes-sigs/kind/issues/1300
	echo "kind: Cluster" > kind-config.yaml
	echo "apiVersion: kind.x-k8s.io/v1alpha4" >> kind-config.yaml
	echo "nodes:" >> kind-config.yaml
	echo "  - role: control-plane" >> kind-config.yaml
	echo "  - role: worker" >> kind-config.yaml
	echo "  - role: worker" >> kind-config.yaml
	echo "  - role: worker" >> kind-config.yaml
	echo "  - role: worker" >> kind-config.yaml
	try kind create cluster --config kind-config.yaml
	echo "Kind is ready"
	try kubectl get nodes
}

function get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" |
  grep '"tag_name":' |
  sed -E 's/.*"([^"]+)".*/\1/'
}

function install_operator() {

  OPR_LATEST=$(get_latest_release minio/operator)
    echo "  Load minio/operator image ($OPR_LATEST) to the cluster"
	try kubectl apply -k github.com/minio/operator/\?ref\=$OPR_LATEST
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

# Install tenant function is being used by deploy-tenant and check-prometheus
function install_tenant() {

	namespace=tenant-lite
	key=v1.min.io/tenant
	value=storage-lite
	echo "Installing lite tenant"

	try kubectl apply -k "${SCRIPT_DIR}/tenant-lite"

	echo "Waiting for the tenant statefulset, this indicates the tenant is being fulfilled"
	echo $namespace
	echo $value
	echo $key
	wait_for_resource $namespace $value $key

	echo "Waiting for tenant pods to come online (5m timeout)"
	try kubectl wait --namespace $namespace \
	--for=condition=ready pod \
	--selector $key=$value \
	--timeout=300s

	echo "Wait for Prometheus PVC to be bound"
	while [[ $(kubectl get pvc storage-lite-prometheus-storage-lite-prometheus-0 -n tenant-lite -o 'jsonpath={..status.phase}') != "Bound" ]]; do echo "waiting for PVC status" && sleep 1 && kubectl get pvc -A; done

	echo "Build passes basic tenant creation"

}
