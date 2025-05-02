PWD := $(shell pwd)
ifeq '${CI}' 'true'
VERSION ?= dev
else
VERSION ?= $(shell git describe --tags)
VERSIONV ?= $(shell git describe --tags | sed 's,v,,g')
endif
TAG ?= "minio/operator:$(VERSION)"
SHA ?= $(shell git rev-parse --short HEAD)
LDFLAGS ?= "-s -w -X github.com/minio/operator/pkg.ReleaseTag=$(VERSIONV) -X github.com/minio/operator/pkg.Version=$(VERSION) -X github.com/minio/operator/pkg.ShortCommitID=$(SHA)"
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
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6 && \
		echo "Installing govulncheck" && \
		go install golang.org/x/vuln/cmd/govulncheck@latest &&\
		echo "installing gopls" && \
		go install golang.org/x/tools/gopls@latest

verify: getdeps govet lint

binary:
	@CGO_ENABLED=0 GOOS=linux go build -trimpath --ldflags $(LDFLAGS) -o minio-operator ./cmd/operator

operator: binary

docker: operator
	@docker buildx build --no-cache --load --platform linux/$(GOARCH) -t $(TAG) .

build: regen-crd verify operator docker

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
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0
	@${GOPATH}/bin/controller-gen crd:maxDescLen=0,generateEmbeddedObjectMeta=true webhook paths="./..." output:crd:artifacts:config=$(KUSTOMIZE_CRDS)
	@sed 's#namespace: minio-operator#namespace: {{ .Release.Namespace }}#g' resources/base/crds/minio.min.io_tenants.yaml > $(HELM_TEMPLATES)/minio.min.io_tenants.yaml
	@sed 's#namespace: minio-operator#namespace: {{ .Release.Namespace }}#g' resources/base/crds/sts.min.io_policybindings.yaml > $(HELM_TEMPLATES)/sts.min.io_policybindings.yaml
	@sed 's#namespace: minio-operator#namespace: {{ .Release.Namespace }}#g' resources/base/crds/job.min.io_miniojobs.yaml > $(HELM_TEMPLATES)/job.min.io_jobs.yaml

regen-crd-docs:
	@echo "Installing crd-ref-docs" && GO111MODULE=on go install -v github.com/elastic/crd-ref-docs@latest
	@${GOPATH}/bin/crd-ref-docs --source-path=./pkg/apis/minio.min.io/v2  --config=docs/templates/config.yaml --renderer=asciidoctor --output-path=docs/tenant_crd.adoc --templates-dir=docs/templates/asciidoctor/
	@${GOPATH}/bin/crd-ref-docs --source-path=./pkg/apis/sts.min.io/v1beta1  --config=docs/templates/config.yaml --renderer=asciidoctor --output-path=docs/policybinding_crd.adoc --templates-dir=docs/templates/asciidoctor/
	@${GOPATH}/bin/crd-ref-docs --source-path=./pkg/apis/job.min.io/v1alpha1  --config=docs/templates/config.yaml --renderer=asciidoctor --output-path=docs/job_crd.adoc --templates-dir=docs/templates/asciidoctor/

generate-code:
	@./k8s/update-codegen.sh

helm-reindex:
	@echo "Re-indexing helm chart release"
	@./helm-reindex.sh

update-versions:
	@./release.sh --release-sidecar=$(RELEASE_SIDECAR)

release: update-versions generate-code regen-crd regen-crd-docs
	@git add .

apply-gofmt:
	@echo "Applying gofmt to all generated an existing files"
	@GO111MODULE=on gofmt -w .

