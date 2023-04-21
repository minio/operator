#!/bin/bash

get_latest_release() {
curl --silent "https://api.github.com/repos/minio/operator/releases/latest" |
  grep '"tag_name":' |
  sed -E 's/.*"([^"]+)".*/\1/' |
  sed 's/v//'
}
RELEASE=$(get_latest_release)
echo "RELEASE=${RELEASE}"
echo "ls"
ls
echo "pwd .................................."
pwd
echo "source olm.sh"
source olm.sh