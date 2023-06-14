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

    kubectl run admin-mc -i --tty --image quay.io/minio/mc \
	    --env="MC_HOST_minio=https://console:console123@minio.tenant-lite.svc.cluster.local" \
	    --command -- bash -c "until (mc admin info minio/); do echo 'waiting... for 5secs' && sleep 5; done"

    echo "Done."
}

function test_kes_tenant() {

    echo "Installing vault"

    try kubectl apply -f "${SCRIPT_DIR}/../examples/vault/deployment.yaml"

    try kubectl get pods

    echo "Waiting for Vault Pods (10s)"
    sleep 10

    try kubectl get pods

    echo "Waiting for Vault (2m timeout)"

    try kubectl wait --namespace default \
        --for=condition=Available deploy \
        --field-selector="metadata.name=vault" \
        --timeout=120s

    echo "Vault is ready. Bootstrapping Vault..."

    end_time=$((SECONDS + 120)) #timeout the log search 120 seconds
    while [ $SECONDS -lt $end_time ]; do
      if kubectl logs -l app=vault | grep -q "Root Token: "; then
        # shellcheck disable=SC2155
        export VAULT_ROOT_TOKEN=$(kubectl logs -l app=vault | grep "Root Token: " | sed -e "s/Root Token: //g")
        break
      fi
    done
    
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

    # Beginning  Kubernetes 1.24 ----> Service Account Token Secrets are not
    # automatically generated, to generate them manually, users must manually
    # create the secret, for our examples where we lead people to get the JWT
    # from the console-sa service account, they additionally need to manually
    # generate the secret via
    # Don't apply the entire file: kubectl apply -f "${SCRIPT_DIR}/../resources/base/console-ui.yaml"
    # Because you will get 500 due to:
    # CREDENTIALS: {"code":500,"detailedMessage":"secrets is forbidden: User \"system:serviceaccount:minio-operator:console-sa\"
    # cannot create resource \"secrets\" in API group \"\" in the namespace \"default\"","message":"an errors occurred, please try again"}
    RESOURCE=$(yq e 'select(.kind == "Secret")' "${SCRIPT_DIR}/../resources/base/console-ui.yaml")
    echo $RESOURCE | kubectl apply -f -
    SA_TOKEN=$(kubectl -n minio-operator  get secret console-sa-secret -o jsonpath="{.data.token}" | base64 --decode)
    echo "SA_TOKEN: ${SA_TOKEN}"
    if [ -z "$SA_TOKEN" ]
    then
      echo "\$SA_TOKEN is empty and it cannot be empty!"
      return 1
    fi

	echo "Creating Tenant"
	sed -i -e 's/ROLE_ID/'"$ROLE_ID"'/g' "${SCRIPT_DIR}/kes-config.yaml"
	sed -i -e 's/SECRET_ID/'"$SECRET_ID"'/g' "${SCRIPT_DIR}/kes-config.yaml"
	cp "${SCRIPT_DIR}/kes-config.yaml" "${SCRIPT_DIR}/../examples/kustomization/tenant-kes-encryption/kes-configuration-secret.yaml"
	yq e -i '.spec.kes.image = "minio/kes:2023-05-02T22-48-10Z"' "${SCRIPT_DIR}/../examples/kustomization/tenant-kes-encryption/tenant.yaml"
	kubectl apply -k "${SCRIPT_DIR}/../examples/kustomization/tenant-kes-encryption"

    echo "Check Tenant Status in tenant-kms-encrypted namespace for myminio:"
    check_tenant_status tenant-kms-encrypted myminio

    echo "Port Forwarding tenant"
    try kubectl port-forward $(kubectl get pods -l v1.min.io/tenant=myminio -n tenant-kms-encrypted | grep -v NAME | awk '{print $1}' | head -1) 9000 -n tenant-kms-encrypted &

    TENANT_CONFIG_SECRET=$(kubectl -n tenant-kms-encrypted get tenants.minio.min.io myminio -o jsonpath="{.spec.configuration.name}")
    # kes-tenant-env-configuration
    USER=$(kubectl -n tenant-kms-encrypted get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_USER="' | sed -e 's/export MINIO_ROOT_USER="//g' | sed -e 's/"//g')
    PASSWORD=$(kubectl -n tenant-kms-encrypted get secrets "$TENANT_CONFIG_SECRET" -o go-template='{{index .data "config.env"|base64decode }}' | grep 'export MINIO_ROOT_PASSWORD="' | sed -e 's/export MINIO_ROOT_PASSWORD="//g' | sed -e 's/"//g')

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
