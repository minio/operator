PWD := $(shell pwd)
TAG ?= "minio/k8s-operator"

all: build

build:
	@CGO_ENABLED=0 go build --ldflags "-s -w" -o $(PWD)/minio-operator
	@docker build -t $(TAG) .

install: all
	@docker push $(TAG)
