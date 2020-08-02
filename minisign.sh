#!/bin/bash

minisign -s "/media/${USER}/minio/minisign.key" -Sm "${1}" < "/media/${USER}/minio/minisign-passphrase"
cp -a "${1}".minisig "$(basename ${1})".minisig
