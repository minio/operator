#!/bin/bash

# How to run the script:
# cd ~/minio/olm-scripts
# source update-redhat-marketplace-operator.sh
# As a result, the script will give you the URL to create the PR.

echo -n "Enter authorized GitHub user, like (cniackz): "
read -r GITHUBUSER

echo -n "Enter RELEASE, like (4.4.20): "
read -r RELEASE
echo "RELEASE: ${RELEASE}"
VERSION="${RELEASE//./-}"
echo "VERSION: ${VERSION}"

echo "Remove old repository"
rm -rf ~/operator
cd ~/ || return
git clone git@github.com:$GITHUBUSER/operator.git

echo " "
echo "Update Forked Operator Repository"
cd ~/operator || return
git checkout master
git remote add upstream git@github.com:minio/operator.git
git fetch upstream
git checkout master
git rebase upstream/master
git push

echo " "
echo "Execute olm.sh and then olm-post-script.sh"
echo "As a work around get working scripts from your repository"
cp ~/minio/olm-scripts/olm.sh ~/operator/olm.sh
cp ~/minio/olm-scripts/olm-post-script.sh ~/operator/olm-post-script.sh
cd ~/operator || return
source olm.sh
source olm-post-script.sh

echo " "
echo "Create the branch:"
rm -rf ~/redhat-marketplace-operators
cd ~/ || return
git clone git@github.com:$GITHUBUSER/redhat-marketplace-operators.git
cd ~/redhat-marketplace-operators || return
git checkout main
git remote add upstream git@github.com:redhat-openshift-ecosystem/redhat-marketplace-operators.git
git fetch upstream
git checkout main
git rebase upstream/main
git push
git checkout -b update-minio-operator-$VERSION

echo " "
echo "Copy the files from Operator Repo to RHMP Repo:"
cp -R ~/operator/bundles/redhat-marketplace/$RELEASE ~/redhat-marketplace-operators/operators/minio-operator-rhmp/$RELEASE

echo " "
echo "Add files to RHMP Repo"
git add operators/minio-operator-rhmp/$RELEASE
git restore --staged operators/minio-operator-rhmp/$RELEASE/manifests/minio-operator-rhmp.clusterserviceversion.yaml-e
git restore --staged operators/minio-operator-rhmp/4.4.20/metadata/annotations.yaml-e
rm operators/minio-operator-rhmp/4.4.20/manifests/minio-operator-rhmp.clusterserviceversion.yaml-e
rm operators/minio-operator-rhmp/4.4.20/metadata/annotations.yaml-e

echo " "
echo "Commit the changes"
git commit -m "operator minio-operator-rhmp (${RELEASE})"

echo " "
echo "Push the changes"
git push --set-upstream origin update-minio-operator-$VERSION
