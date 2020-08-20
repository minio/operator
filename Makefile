PWD := $(shell pwd)
VERSION ?= $(shell git describe --tags)
TAG ?= "minio/k8s-operator:$(VERSION)"
LDFLAGS ?= "-s -w -X main.Version=$(VERSION)"

GOPATH := $(shell go env GOPATH)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

CRD_GEN_PATH=operator-kustomize/crds/
OPERATOR_YAML=./operator-kustomize/
TENANT_CRD=$(OPERATOR_YAML)crds/minio.min.io_tenants.yaml
PLUGIN_HOME=./kubectl-minio
PLUGIN_OPERATOR_YAML=$(PLUGIN_HOME)/static
PLUGIN_CRD_COPY=$(PLUGIN_OPERATOR_YAML)/crd.yaml
all: build

getdeps:
	@mkdir -p ${GOPATH}/bin
	@which golangci-lint 1>/dev/null || (echo "Installing golangci-lint" && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.27.0)

verify: govet gotest lint

build: verify
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS) -o minio-operator
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

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv

regen-crd:
	@controller-gen crd:trivialVersions=true paths="./..." output:crd:artifacts:config=$(CRD_GEN_PATH)


plugin: regen-crd
	@rm -rf $(PLUGIN_OPERATOR_YAML)
	@mkdir -p $(PLUGIN_OPERATOR_YAML)
	@cp $(TENANT_CRD) $(PLUGIN_CRD_COPY)
	@cp  $(OPERATOR_YAML)/*.yaml $(PLUGIN_OPERATOR_YAML)
	@cd $(PLUGIN_HOME); go build -o kubectl-minio main.go
