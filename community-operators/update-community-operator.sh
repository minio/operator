#!/bin/bash

# How to run the script:
# cd ~/minio/olm-scripts
# source update-community-operator.sh
# As a result, the script will give you the URL to create the PR.

echo -n "Previous version (4.4.20): "
read -r PREVIOUSVERSION

echo -n "Enter authorized GitHub user, like (cniackz): "
read -r GITHUBUSER

echo -n "Enter RELEASE, like (4.4.20): "
read -r RELEASE
echo "RELEASE: ${RELEASE}"
VERSION="${RELEASE//./-}"
echo "VERSION: ${VERSION}"

echo "Remove old repositories"
rm -rf ~/operator
rm -rf ~/community-operators
cd ~/ || return
git clone git@github.com:$GITHUBUSER/operator.git
git clone git@github.com:$GITHUBUSER/community-operators.git

echo " "
echo "Update Forked Operator Repositories"
cd ~/operator || return
git checkout master
git remote add upstream git@github.com:minio/operator.git
git fetch upstream
git checkout master
git rebase upstream/master
git push
cd ~/community-operators || return
git checkout main
git remote add upstream git@github.com:k8s-operatorhub/community-operators.git
git fetch upstream
git checkout main
git rebase upstream/main
git push

echo " "
echo "Execute olm.sh"
echo "As a work around get working scripts from your repository"
cp ~/minio/olm-scripts/community-operators/olm.sh ~/operator/olm.sh
cd ~/operator || return
source olm.sh

echo " "
echo "Create the branch:"
rm -rf ~/community-operators
cd ~/ || return
git clone git@github.com:$GITHUBUSER/community-operators.git
cd ~/community-operators || return
git checkout main
git remote add upstream git@github.com:redhat-openshift-ecosystem/community-operators.git
git fetch upstream
git checkout main
git rebase upstream/main
git push
git checkout -b update-minio-operator-$VERSION

echo " "
echo "Copy the files from Operator Repo to Community Repo:"
cp -R ~/operator/bundles/community-operators/$RELEASE ~/community-operators/operators/minio-operator/$RELEASE
rm -rf ~/community-operators/operators/minio-operator/$PREVIOUSVERSION

echo " "
echo "Add files to Community Repo"
git add operators/minio-operator/$RELEASE
git restore --staged operators/minio-operator/$RELEASE/metadata/annotations.yaml-e
rm operators/minio-operator/$RELEASE/metadata/annotations.yaml-e
git rm operators/minio-operator/$PREVIOUSVERSION/manifests/minio-operator.clusterserviceversion.yaml
git rm operators/minio-operator/$PREVIOUSVERSION/manifests/minio.min.io_tenants.yaml
git rm operators/minio-operator/$PREVIOUSVERSION/manifests/operator_v1_service.yaml
git rm operators/minio-operator/$PREVIOUSVERSION/metadata/annotations.yaml
git rm operators/minio-operator/$PREVIOUSVERSION/metadata/annotations.yaml-e

echo " "
echo "Commit the changes"
git commit --signoff -m 'update minio operator'

echo " "
echo "Push the changes"
git push --set-upstream origin update-minio-operator-$VERSION
