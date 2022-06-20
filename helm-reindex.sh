#!/bin/bash

set -e

helm package helm/operator -d helm-releases/
helm package helm/tenant -d helm-releases/
helm repo index --merge index.yaml --url https://operator.min.io .
