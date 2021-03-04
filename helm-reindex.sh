#!/bin/bash

helm package helm/minio-operator -d helm-releases/

helm repo index --merge index.yaml --url https://operator.min.io .
