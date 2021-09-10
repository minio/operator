#!/bin/bash

function _init() {
    ## All binaries are static make sure to disable CGO.
    export CGO_ENABLED=0

    ## List of architectures and OS to test coss compilation.
    SUPPORTED_OSARCH="linux/ppc64le linux/arm64 linux/s390x linux/amd64"
}

function _build() {
    local osarch=$1
    IFS=/ read -r -a arr <<<"$osarch"
    os="${arr[0]}"
    arch="${arr[1]}"
    package=$(go list -f '{{.ImportPath}}')
    printf -- "--> %15s:%s\n" "${osarch}" "${package}"

    # go build -trimpath to build the binary.
    export GOOS=$os
    export GOARCH=$arch
    export GO111MODULE=on
    go build --ldflags "-s -w" -trimpath -tags kqueue -o "logsearchapi_${arch}"
}

function main() {
    echo "Testing builds for OS/Arch: ${SUPPORTED_OSARCH}"
    for each_osarch in ${SUPPORTED_OSARCH}; do
        _build "${each_osarch}"
    done

    sudo sysctl net.ipv6.conf.wlp59s0.disable_ipv6=1

    release=$(git describe --abbrev=0 --tags)

    docker buildx build --push --no-cache -t "minio/logsearchapi:${release}" \
           --build-arg TAG="${release}" \
           --platform=linux/arm64,linux/amd64,linux/ppc64le,linux/s390x \
           -f Dockerfile .

    docker buildx prune -f

    docker buildx build --push --no-cache -t "quay.io/minio/logsearchapi:${release}" \
           --build-arg TAG="${release}" \
           --platform=linux/arm64,linux/amd64,linux/ppc64le,linux/s390x \
           -f Dockerfile .

    docker buildx prune -f

    sudo sysctl net.ipv6.conf.wlp59s0.disable_ipv6=0
}

_init && main "$@"
