PWD := $(shell pwd)
ifeq '${CI}' 'true'
VERSION ?= dev
else
VERSION ?= $(shell git describe --tags)
VERSIONV ?= $(shell git describe --tags | sed 's,v,,g')
endif
TAG ?= "minio/operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"
GOPATH := $(shell go env GOPATH)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

HELM_HOME=helm/operator
HELM_TEMPLATES=$(HELM_HOME)/templates

KUSTOMIZE_HOME=resources
KUSTOMIZE_CRDS=$(KUSTOMIZE_HOME)/base/crds/

PLUGIN_HOME=kubectl-minio

LOGSEARCHAPI=logsearchapi

all: build

getdeps:
	@echo "Checking dependencies"
	@mkdir -p ${GOPATH}/bin
	@echo "Installing golangci-lint" && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0 && \
		echo "Installing govulncheck" && \
		go install golang.org/x/vuln/cmd/govulncheck@latest

verify: getdeps govet gotest lint

operator:
	@CGO_ENABLED=0 GOOS=linux go build -trimpath -o minio-operator

docker: operator logsearchapi
	@docker build --no-cache -t minio/operator:v4.5.9 .

build: regen-crd verify plugin logsearchapi operator docker

install: all

lint: getdeps
	@echo "Running $@ check"
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ./.golangci.yml

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

regen-crd:
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
	@${GOPATH}/bin/controller-gen crd:maxDescLen=0,generateEmbeddedObjectMeta=true paths="./..." output:crd:artifacts:config=$(KUSTOMIZE_CRDS)
	@sed 's#namespace: minio-operator#namespace: {{ .Release.Namespace }}#g' resources/base/crds/minio.min.io_tenants.yaml > $(HELM_TEMPLATES)/minio.min.io_tenants.yaml

regen-crd-docs:
	@which crd-ref-docs 1>/dev/null || (echo "Installing crd-ref-docs" && GO111MODULE=on go install -v github.com/elastic/crd-ref-docs@latest)
	@${GOPATH}/bin/crd-ref-docs --source-path=./pkg/apis/minio.min.io/v2 --config=docs/templates/config.yaml --renderer=asciidoctor --output-path=docs/crd.adoc --templates-dir=docs/templates/asciidoctor/

plugin: regen-crd
	@echo "Building 'kubectl-minio' binary"
	@(cd $(PLUGIN_HOME); \
		go vet ./... && \
		go test -race ./... && \
		GO111MODULE=on ${GOPATH}/bin/golangci-lint cache clean && \
		GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ../.golangci.yml)

.PHONY: logsearchapi
logsearchapi: getdeps
	@echo "Building 'logsearchapi' binary"
	@(cd $(LOGSEARCHAPI); \
		go vet ./... && \
		go test -race ./... && \
		GO111MODULE=on ${GOPATH}/bin/golangci-lint cache clean && \
		GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ../.golangci.yml && \
		CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -trimpath -o ../logsearchapi-bin )

getconsoleuiyaml:
	@echo "Getting the latest Console UI"
	@kustomize build github.com/minio/console/k8s/operator-console/base > resources/base/console-ui.yaml
	@echo "Done"

generate-code:
	@./k8s/update-codegen.sh

generate-openshift-manifests:
	@./olm.sh

release: generate-openshift-manifests
	@./release.sh
