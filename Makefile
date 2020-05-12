PWD := $(shell pwd)
VERSION ?= $(shell git describe --tags)
TAG ?= "minio/k8s-operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

all: build

verify: govet gotest

build:
	@docker build -t $(TAG) --build-arg ldflags=$(LDFLAGS) .

install: all
	@docker push $(TAG)

govet:
	@CGO_ENABLED=0 go vet ./...

gotest:
	@go test -race ./...
