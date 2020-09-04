#!/bin/bash

minisign -s "/media/${USER}/minio/minisign.key" \
         -x "$(basename $(dirname ${1}))".minisig \
         -Sm "${1}" < "/media/${USER}/minio/minisign-passphrase"

if [ -f minio-operator_linux_amd64.minisig ]; then
    mv minio-operator_linux_amd64.minisig minio-operator.minisig
fi
