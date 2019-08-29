PWD := $(shell pwd)
TAG ?= "minio/k8s-operator"

all: build

verify: govet gotest

build:
	@CGO_ENABLED=0 GOPROXY=https://proxy.golang.org go build --ldflags "-s -w" -o $(PWD)/minio-operator
	@docker build -t $(TAG) .

install: all
	@docker push $(TAG)

govet:
	@CGO_ENABLED=0 GOPROXY=https://proxy.golang.org go vet ./...

gotest:
	@GOPROXY=https://proxy.golang.org go test -race ./...
