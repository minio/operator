PWD := $(shell pwd)
VERSION ?= $(shell git describe --tags)
TAG ?= "minio/k8s-operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

all: build

getdeps:
	@mkdir -p ${GOPATH}/bin
	@which golangci-lint 1>/dev/null || (echo "Installing golangci-lint" && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.27.0)

verify: govet gotest lint

build: verify
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS)
	@docker build -t $(TAG) .

install: all
	@docker push $(TAG)

lint:
	@echo "Running $@ check"
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint cache clean
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ./.golangci.yml

govet:
	@go vet ./...

gotest:
	@go test -race ./...
