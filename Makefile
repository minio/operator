PWD := $(shell pwd)
VERSION ?= $(shell git describe --tags)
TAG ?= "minio/k8s-operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

all: build

verify: govet gotest

build: verify
	@CGO_ENABLED=0 go build -trimpath --ldflags $(LDFLAGS)
	@docker build -t $(TAG) .

install: all
	@docker push $(TAG)

govet:
	@go vet ./...

gotest:
	@go test -race ./...
