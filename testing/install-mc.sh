#!/usr/bin/env bash
# This script will run inside ubuntu-pod that is located at default namespace in the cluster
# This script will not and should not be executed in the self hosted runner

echo "install-mc.sh: Installing mc command"
apt-get update -y
apt-get upgrade -y
apt-get clean -y
apt-get install wget -y
apt-get install jq -y

echo "install-mc.sh: Install mc"
echo "MC_HOT_FIX_REL=$MC_HOT_FIX_REL, MC_VER=$MC_VER"

DOWNLOAD_NAME="archive/mc.${MC_VER}"
MC_RELEASE_TYPE="release"
if [ -n "${MC_HOT_FIX_REL}" ] && [ -n "${MC_VER}" ]; then
	MC_RELEASE_TYPE="hotfixes"
fi

if [ "${MC_VER}" == "latest" ] || [ -z "${MC_VER}" ]; then
	DOWNLOAD_NAME="mc"
fi

wget -O mc https://dl.min.io/client/mc/"${MC_RELEASE_TYPE}"/linux-amd64/"${DOWNLOAD_NAME}"

chmod +x mc
mv mc /usr/local/bin/mc

echo "install-mc.sh: we should see mc output if mc got installed:"
mc --version
RESULT=$(mc | grep -c GNU)
if [ "$RESULT" == "1" ]; then
	echo "script passed" >install-mc.log
else
	echo "mc not installed, install-mc.sh failed"
fi
