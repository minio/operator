PWD := $(shell pwd)
VERSION ?= $(shell git describe --tags)
TAG ?= "minio/k8s-operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

all: build

verify: govet gotest

build:
	@CGO_ENABLED=0 go build --ldflags $(LDFLAGS) -o $(PWD)/minio-operator
	@docker build -t $(TAG) .

install: all
	@docker push $(TAG)

govet:
	@CGO_ENABLED=0 go vet ./...

gotest:
	@go test -race ./...
