PWD := $(shell pwd)
ifeq '${CI}' 'true'
VERSION ?= dev
else
VERSION ?= $(shell git describe --tags)
VERSIONV ?= $(shell git describe --tags | sed 's,v,,g')
endif
TAG ?= "minio/operator-sidecar:$(VERSION)"
SHA ?= $(shell git rev-parse --short HEAD)
LDFLAGS ?= "-s -w -X github.com/minio/operator/pkg.ReleaseTag=$(VERSIONV) -X github.com/minio/operator/sidecar/pkg.Version=$(VERSION) -X github.com/minio/operator/sidecar/pkg.ShortCommitID=$(SHA)"
GOPATH := $(shell go env GOPATH)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

HELM_HOME=helm/operator
HELM_TEMPLATES=$(HELM_HOME)/templates

KUSTOMIZE_HOME=resources
KUSTOMIZE_CRDS=$(KUSTOMIZE_HOME)/base/crds/


all: build

getdeps:
	@echo "Checking dependencies"
	@mkdir -p ${GOPATH}/bin
	@echo "Installing golangci-lint" && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2 && \
		echo "Installing govulncheck" && \
		go install golang.org/x/vuln/cmd/govulncheck@latest &&\
		echo "installng gopls" && \
		go install golang.org/x/tools/gopls@latest

verify: getdeps govet lint

binary:
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS) -o minio-operator-sidecar ./cmd/sidecar

sidecar: binary

docker: sidecar
	@docker buildx build --no-cache --load --platform linux/$(GOARCH) -t $(TAG) .

build: verify sidecar docker

install: all

lint: getdeps
	@echo "Running $@ check"
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ../.golangci.yml

govet:
	@go vet ./...

gotest:
	@go test -race ./...

vulncheck:
	@${GOPATH}/bin/govulncheck ./...

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@find . -name '*.zip' | xargs rm -fv
	@rm -rf dist/

