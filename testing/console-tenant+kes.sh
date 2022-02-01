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

function check_tenant_status_old() {
    # Check MinIO is accessible

    waitdone=0
    totalwait=0
    while true; do
	waitdone=$(kubectl get pods -l v1.min.io/tenant=kes-tenant --no-headers | wc -l)
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

    kubectl run admin-mc -i --tty --image minio/mc --command -- bash -c "until (mc alias set minio/ https://minio.tenant-lite.svc.cluster.local console console123); do echo \"...waiting... for 5secs\" && sleep 5; done; mc admin info minio/;"

    echo "Done."
}

function test_kes_tenant() {

    echo "Installing vault"

    try kubectl apply -f "${SCRIPT_DIR}/../examples/vault/deployment.yaml"

    kubectl get pods

    echo "Waiting for Vault Pods (10s)"
    sleep 10

    kubectl get pods


    echo "Waiting for Vault (2m timeout)"

    try kubectl wait --namespace default \
	--for=condition=ready pod \
	--selector=app=vault \
	--timeout=120s

    echo "Vault is ready. Bootstrapping Vault..."

    VAULT_ROOT_TOKEN=$(kubectl logs -l app=vault | grep "Root Token: " | sed -e "s/Root Token: //g")
    echo "Vault root: '$VAULT_ROOT_TOKEN'"


    try kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault auth enable approle'
    try kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault secrets enable kv'

    # copy kes file
    try kubectl cp "${SCRIPT_DIR}/../examples/vault/kes-policy.hcl" $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}'):/kes-policy.hcl

    try kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault policy write kes-policy /kes-policy.hcl'
    try kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault write auth/approle/role/kes-role token_num_uses=0 secret_id_num_uses=0 period=5m policies=kes-policy'


    ROLE_ID=$(kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault read auth/approle/role/kes-role/role-id' | grep "role_id    " | sed -e "s/role_id    //g")
    SECRET_ID=$(kubectl exec $(kubectl get pods -l app=vault  | grep -v NAME | awk '{print $1}') -- sh -c 'VAULT_TOKEN='$VAULT_ROOT_TOKEN' VAULT_ADDR="http://127.0.0.1:8200" vault write -f auth/approle/role/kes-role/secret-id' | grep "secret_id             " | sed -e "s/secret_id             //g")

    echo "Port Forwarding console"
    kubectl -n minio-operator port-forward svc/console 9090 &

    SA_TOKEN=$(kubectl -n minio-operator  get secret $(kubectl -n minio-operator get serviceaccount console-sa -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode)

    COOKIE=$(curl 'http://localhost:9090/api/v1/login/operator' -X POST \
		  -H 'Content-Type: application/json' \
		  --data-raw '{"jwt":"'$SA_TOKEN'"}' -i | grep "Set-Cookie: token=" | sed -e "s/Set-Cookie: token=//g" | awk -F ';' '{print $1}')

    echo "Creating Tenant"
    CREDENTIALS=$(curl 'http://localhost:9090/api/v1/tenants' \
                       -X POST \
                       -H 'Content-Type: application/json' \
                       -H 'Cookie: token='$COOKIE'' \
                       --data-raw '{"name":"kes-tenant","namespace":"default","access_key":"","secret_key":"","access_keys":[],"secret_keys":[],"enable_tls":true,"enable_console":true,"enable_prometheus":true,"service_name":"","image":"","expose_minio":true,"expose_console":true,"pools":[{"name":"pool-0","servers":4,"volumes_per_server":1,"volume_configuration":{"size":26843545600,"storage_class_name":"standard"},"securityContext":null,"affinity":{"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":[{"labelSelector":{"matchExpressions":[{"key":"v1.min.io/tenant","operator":"In","values":["kes-tenant"]},{"key":"v1.min.io/pool","operator":"In","values":["pool-0"]}]},"topologyKey":"kubernetes.io/hostname"}]}}}],"erasureCodingParity":2,"logSearchConfiguration":{"image":"","postgres_image":"","postgres_init_image":""},"prometheusConfiguration":{"image":"","sidecar_image":"","init_image":""},"tls":{"minio":[],"ca_certificates":[],"console_ca_certificates":[]},"encryption":{"replicas":"1","securityContext":{"runAsUser":"1000","runAsGroup":"1000","fsGroup":"1000","runAsNonRoot":true},"image":"","vault":{"endpoint":"http://vault.default.svc.cluster.local:8200","engine":"","namespace":"","prefix":"my-minio","approle":{"engine":"","id":"'$ROLE_ID'","secret":"'$SECRET_ID'","retry":0},"tls":{},"status":{"ping":0}}},"idp":{"keys":[{"access_key":"console","secret_key":"console123"}]}}')


    echo $CREDENTIALS

    check_tenant_status default kes-tenant

    echo "Port Forwarding tenant"
    try kubectl port-forward $(kubectl get pods -l v1.min.io/tenant=kes-tenant | grep -v NAME | awk '{print $1}' | head -1) 9000 &

    USER=$(kubectl -n default get secrets kes-tenant-env-configuration -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_USER="' | sed -e 's/export MINIO_ROOT_USER="//g' | sed -e 's/"//g')
    PASSWORD=$(kubectl -n default get secrets kes-tenant-env-configuration -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_PASSWORD="' | sed -e 's/export MINIO_ROOT_PASSWORD="//g' | sed -e 's/"//g')


    totalwait=0
    until (mc config host add kestest https://localhost:9000 $USER $PASSWORD --insecure); do
	echo "...waiting... for 5secs" && sleep 5

	totalwait=$((totalwait + 5))
    	if [ "$totalwait" -gt 305 ]; then
    	    echo "Unable to register mc tenant after 5 minutes, exiting."
    	    try false
    	fi

    done;
    try mc admin kms key status kestest --insecure
}

function main() {
    destroy_kind

    setup_kind

    install_operator

    test_kes_tenant

    destroy_kind
}

main "$@"
