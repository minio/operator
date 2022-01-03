#!/bin/bash

# This script requires: kubectl, kind, jq

#set -e

yell() { echo "$0: $*" >&2; }

die() {
  yell "$*"
  (kind delete cluster || true ) && exit 111
}

try() { "$@" || die "cannot $*"; }

try kind create cluster --config kind-config.yaml

echo "Installing Current Operator"

# TODO: Compile the current branch and create an overlay to use that image version
try kubectl apply -k ../resources

echo "Waiting for Operator Pods to come online"

try kubectl wait --namespace minio-operator \
  --for=condition=ready pod \
  --selector=name=minio-operator \
  --timeout=90s

echo "Installing lite tenant"

try kubectl apply -k ../examples/kustomization/tenant-lite


echo "Waiting for the tenant statefulset, this indicates the tenant is being fulfilled"
waitdone=0
totalwait=0
while true; do
  waitdone=$(kubectl -n tenant-lite get pods -o json | jq '.items | length')
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
  --selector="v1.min.io/tenant=tenant-lite" \
  --timeout=300s

echo "Build passes basic tenant creation"

# Check MinIO is accessible

waitdone=0
totalwait=0
while true; do
  waitdone=$(kubectl -n tenant-lite get pods -l v1.min.io/tenant=tenant-lite --no-headers | wc -l)
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

# clean up
kind delete cluster
