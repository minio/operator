#!/usr/bin/env bash
# Copyright (C) 2023, MinIO, Inc.
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

ARCH=$(go env GOARCH)
OS=$(uname | awk '{print tolower($0)}')
# shellcheck disable=SC2155
export TMP_BIN_DIR="$(mktemp -d)"

function install_binaries() {

	echo -e "\e[34mInstalling temporal binaries in $TMP_BIN_DIR\e[0m"

	echo "opm"
	curl -#L "https://mirror.openshift.com/pub/openshift-v4/$ARCH/clients/ocp/stable/opm-$OS.tar.gz" -o $TMP_BIN_DIR/opm-$OS.tar.gz
	tar -zxf $TMP_BIN_DIR/opm-$OS.tar.gz -C $TMP_BIN_DIR/
	chmod +x $TMP_BIN_DIR/opm

	echo "crc"
	curl -#L "https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/crc/latest/crc-$OS-$ARCH.tar.xz" -o $TMP_BIN_DIR/crc-$OS-$ARCH.tar.xz
	tar -xJf $TMP_BIN_DIR/crc-$OS-$ARCH.tar.xz -C $TMP_BIN_DIR/ --strip-components=1
	chmod +x $TMP_BIN_DIR/crc

}

function remove_temp_binaries() {
	echo -e "\e[34mRemoving temporary binaries in: $TMP_BIN_DIR\e[0m"
	rm -rf $TMP_BIN_DIR
}

yell() { echo "$0: $*" >&2; }

die() {
	yell "$*"
	exit 111
}

try() { "$@" || die "cannot $*"; }

function setup_crc() {
	# Set bundle
	# Example options
	# crc_libvirt_4.11.7_amd64
	# crc_libvirt_4.12.0_amd64
	# crc_libvirt_4.12.5_amd64
	# crc_libvirt_4.13.0_amd64
	# crc_libvirt_4.13.6_amd64

	bundle_version="$1"
	virtualization="libvirt"

	if [ -z "$bundle_version" ]; then
		die "missing bundle version"
	fi
	if [ "$OS" == "darwin" ]; then
	  virtualization="vfkit"
	fi
	bundle="crc_${virtualization}_${bundle_version}_${ARCH}.crcbundle"
	bundle_url="https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.14.1/${bundle}"
	echo -e "\e[34mConfiguring crc\e[0m"
	export PATH="$TMP_BIN_DIR:$PATH"
	crc config set consent-telemetry no
	crc config set skip-check-root-user true
	crc config set kubeadmin-password "crclocal"
	crc setup -b $bundle_url
	crc start -c 12 -m 20480
	eval $(crc oc-env)
	eval $(crc podman-env)
	# this creates a symlink "podman" from the "podman-remote", as a hack to solve the a issue with opm:
	# opm has hardcoded the command name "podman" causing the index creation to fail
	# https://github.com/operator-framework/operator-registry/blob/67e6777b5f5f9d337b94da98b8c550c231a8b47c/pkg/containertools/factory_podman.go#L32
	ocpath=$(dirname $(which podman-remote))
	ln -sf $ocpath/podman-remote $ocpath/podman
	try crc version
	echo "Waiting for podman vm come online (5m timeout)"
	timeout 600 bash -c -- 'while ! podman image ls 2> /dev/null; do sleep 1 && printf ".";done'
	oc login -u kubeadmin -p crclocal https://api.crc.testing:6443 --insecure-skip-tls-verify=true
}

function destroy_crc() {
	echo -e "\e[34mdestroy_crc\e[0m"

	# To allow the execution without killing the cluster at the end of the test
	# Use below statement to automatically test and kill cluster at the end:
	# `unset OPERATOR_DEV_TEST`
	# Use below statement to test and keep cluster alive at the end!:
	# `export OPERATOR_DEV_TEST="ON"`
	if [[ -z "${OPERATOR_DEV_TEST}" ]]; then
		# OPERATOR_DEV_TEST is not defined, hence destroy_kind
		echo "Cluster will be destroyed for automated testing"
		crc stop
		crc delete -f
		remove_temp_binaries
	else
		echo -e "\e[33mCluster will remain alive for manual testing\e[0m"
		echo "Use the following env varianbles setup"
		echo "export PATH=$TMP_BIN_DIR:\$PATH"
		echo "eval \$(crc oc-env)"
		echo "eval \$(crc podman-env)"
	fi
}

function create_marketplace_catalog() {
	# https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/openshift-deployment
	# https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/bundle-directory
	# https://operatorhub.io/preview

	# Obtain catalog
	catalog="$1"
	if [ -z "$catalog" ]; then
		die "missing catalog to install"
	fi

	echo "Create Marketplace for catalog '$catalog'"

	registry="default-route-openshift-image-registry.apps-crc.testing"
	operatorNamespace="openshift-operators"
	marketplaceNamespace="openshift-marketplace"
	operatorrepotag="operator:latest"
	operatorContainerImage="$registry/$operatorNamespace/$operatorrepotag"
	bundleContainerImage="$registry/$marketplaceNamespace/operator-bundle:latest"
	indexContainerImage="$registry/$marketplaceNamespace/minio-operator-index:latest"
	package="minio-operator"
	if [[ "$catalog" == "redhat-marketplace" ]]; then
		package=minio-operator-rhmp
	fi

	echo "Compiling operator in current branch"
	(cd "${SCRIPT_DIR}/.." && make operator && podman build --quiet --no-cache -t $operatorContainerImage .)

	echo "push operator image to crc registry"
	podman login -u $(oc whoami) -p $(oc whoami --show-token) $registry/$operatorNamespace --tls-verify=false
	podman push $operatorContainerImage --tls-verify=false

	echo "Image Stream for operator"
	oc get is -n $operatorNamespace operator
	try oc set image-lookup operator -n $operatorNamespace

	echo "Compiling operator bundle for $catalog"
	cp -r "${SCRIPT_DIR}/../$catalog/." ${SCRIPT_DIR}/openshift/bundle
	yq -i ".metadata.annotations.containerImage |= (\"${operatorrepotag}\")" ${SCRIPT_DIR}/openshift/bundle/manifests/$package.clusterserviceversion.yaml
	yq -i ".spec.install.spec.deployments.[0].spec.template.spec.containers.[0].image |= (\"${operatorrepotag}\")" ${SCRIPT_DIR}/openshift/bundle/manifests/$package.clusterserviceversion.yaml
	yq -i ".spec.install.spec.deployments.[1].spec.template.spec.containers.[0].image |= (\"${operatorrepotag}\")" ${SCRIPT_DIR}/openshift/bundle/manifests/$package.clusterserviceversion.yaml
	yq -i "del(.spec.replaces)" ${SCRIPT_DIR}/openshift/bundle/manifests/$package.clusterserviceversion.yaml
	yq -i ".annotations.\"operators.operatorframework.io.bundle.package.v1\" |= (\"${package}-noop\")" ${SCRIPT_DIR}/openshift/bundle/metadata/annotations.yaml
	(cd "${SCRIPT_DIR}/.." && podman build --quiet --no-cache -t $bundleContainerImage -f ${SCRIPT_DIR}/openshift/bundle.Dockerfile ${SCRIPT_DIR}/openshift)
	podman login -u $(oc whoami) -p $(oc whoami --show-token) $registry --tls-verify=false

	echo "push operator-bundle to crc registry"
	podman push $bundleContainerImage --tls-verify=false

	echo "Image Stream for operator-bundle"
	oc get is -n $marketplaceNamespace operator-bundle
	try oc set image-lookup -n $marketplaceNamespace operator-bundle

	echo "Compiling marketplace index"
	opm index add --bundles $bundleContainerImage --tag $indexContainerImage --skip-tls-verify=true --pull-tool podman

	echo "push minio-operator-index to crc registry"
	podman push $indexContainerImage --tls-verify=false
	echo "Image Stream for minio-operator-index"
	try oc set image-lookup -n $marketplaceNamespace minio-operator-index

	echo "Wait for ImageStream minio-operator-index to be local available"
	try oc wait -n $marketplaceNamespace is \
		--for=jsonpath='{.spec.lookupPolicy.local}'=true \
		--field-selector metadata.name=minio-operator-index \
		--timeout=300s

	echo "Create 'Test Minio Operators' marketplace catalog source"
	oc apply -f ${SCRIPT_DIR}/openshift/test-operator-catalogsource.yaml
	sleep 5
	echo "Catalog Source:"
	echo "Waiting for Package manifest to be ready (5m timeout)"
	try timeout 300 bash -c -- 'while ! oc get packagemanifests -n '"$marketplaceNamespace"' | grep "Test Minio Operators" 2> /dev/null; do sleep 1 && printf ".";done'
}

function install_operator() {

	# Obtain catalog
	catalog="$1"
	if [ -z "$catalog" ]; then
		catalog="certified-operators"
	fi

	#obtain subscription namespace
	snamespace="$2"
	if [ -z "$snamespace" ]; then
		snamespace="openshift-operators"
	fi

	echo -e "\e[34mInstalling Operator from catalog '$catalog'\e[0m"

	try oc apply -f ${SCRIPT_DIR}/openshift/${snamespace}-subscription.yaml

	echo "Subscription:"
	try oc get sub -n $snamespace test-subscription
	#we wait a moment for the resource to get a status field
	sleep 10s

	echo "Wait subscription to be ready (10m timeout)"
	try oc wait -n $snamespace \
		--for=jsonpath='{.status.state}'=AtLatestKnown subscription --field-selector metadata.name=$(oc get subscription -n $snamespace -o json | jq -r '.items[0] | .metadata.name') \
		--timeout=600s

	echo "Install plan:"
	try oc get installplan -n $snamespace

	echo "Waiting for install plan to be completed (10m timeout)"
	oc wait -n $snamespace \
		--for=jsonpath='{.status.phase}'=Complete installplan \
		--field-selector metadata.name=$(oc get installplan -n $snamespace -o json | jq -r '.items[0] | .metadata.name') \
		--timeout=600s

	echo "Deployment:"
	oc -n $snamespace get deployment minio-operator

	echo "Waiting for Operator Deployment to come online (5m timeout)"
	try oc wait -n $snamespace deployment \
		--for=condition=Available \
		--field-selector metadata.name=minio-operator \
		--timeout=300s

	echo "start - get data to verify proper image is being used"
	echo "Pods:"
	oc get pods -n $snamespace
	echo "Images:"
	oc describe pods -n $snamespace | grep Image
}
