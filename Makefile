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
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2 && \
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

generate-openshift-manifests:
	@./olm.sh

helm-reindex:
	@echo "Re-indexing helm chart release"
	@./helm-reindex.sh

update-versions:
	@./release.sh

release: update-versions generate-code regen-crd regen-crd-docs assets
	@git add .

apply-gofmt:
	@echo "Applying gofmt to all generated an existing files"
	@GO111MODULE=on gofmt -w .

clean-swagger:
	@echo "cleaning"
	@rm -rf models
	@rm -rf api/operations

install-go-swagger:
	@echo "installing latest go-swagger"
	@go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger-operator:
	@echo "Generating swagger server code from yaml"
	@swagger generate server -A operator --main-package=operator --server-package=api --exclude-main -P models.Principal -f ./swagger.yml -r NOTICE
	@echo "Generating typescript api"
	@npx swagger-typescript-api -p ./swagger.yml -o ./web-app/src/api -n operatorApi.ts
	@(cd web-app && npm install -g prettier && prettier -w .)

swagger-gen: install-go-swagger clean-swagger swagger-operator apply-gofmt
	@echo "Done Generating swagger server code from yaml"

assets:
	@(if [ -f "${NVM_DIR}/nvm.sh" ]; then \. "${NVM_DIR}/nvm.sh" && nvm install && nvm use && npm install -g yarn ; fi &&\
	  cd web-app; yarn install; make build-static; yarn prettier --write . --loglevel warn; cd ..)

test-unit-test-operator:
	@echo "execute unit test and get coverage for api"
	@(cd api && mkdir coverage && GO111MODULE=on go test -test.v -coverprofile=coverage/coverage-unit-test-operatorapi.out)

test-operator-integration:
	@(echo "Start cd operator-integration && go test:")
	@(pwd)
	@(cd operator-integration && go test -coverpkg=../api -c -tags testrunmain . && mkdir -p coverage && ./operator-integration.test -test.v -test.run "^Test*" -test.coverprofile=coverage/operator-api.out)

test-operator:
	@(env bash $(PWD)/web-app/tests/scripts/operator.sh)

models-gen-mac:
	@swagger generate client -f ./swagger.yml -m ./models
	@ls ./models | xargs -I {} gsed -i "2 a\
// This file is part of MinIO Operator\n\
// Copyright (c) 2023 MinIO, Inc.\n\
//\n\
// This program is free software: you can redistribute it and/or modify\n\
// it under the terms of the GNU Affero General Public License as published by\n\
// the Free Software Foundation, either version 3 of the License, or\n\
// (at your option) any later version.\n\
//\n\
// This program is distributed in the hope that it will be useful,\n\
// but WITHOUT ANY WARRANTY; without even the implied warranty of\n\
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n\
// GNU Affero General Public License for more details.\n\
//\n\
// You should have received a copy of the GNU Affero General Public License\n\
// along with this program.  If not, see <http://www.gnu.org/licenses/>.\n\
//\n\
" ./models/{}
	@rm -rf client

models-gen:
	@swagger generate client -f ./swagger.yml -m ./models
	@ls ./models | xargs -I {} sed -i "2 a\
// This file is part of MinIO Operator\n\
// Copyright (c) 2023 MinIO, Inc.\n\
//\n\
// This program is free software: you can redistribute it and/or modify\n\
// it under the terms of the GNU Affero General Public License as published by\n\
// the Free Software Foundation, either version 3 of the License, or\n\
// (at your option) any later version.\n\
//\n\
// This program is distributed in the hope that it will be useful,\n\
// but WITHOUT ANY WARRANTY; without even the implied warranty of\n\
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n\
// GNU Affero General Public License for more details.\n\
//\n\
// You should have received a copy of the GNU Affero General Public License\n\
// along with this program.  If not, see <http://www.gnu.org/licenses/>.\n\
//\n\
" ./models/{}
	@rm -rf client
