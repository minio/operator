# MinIO Operator Sidecar

This document provides information on how to build and test the sidecar container.

# Testing

Build this project into a container image and run it with the following command:

```shell
TAG=miniodev/operator-sidecar:sc GOOS=linux  make docker
```

Patch the MinIO Operator deployment to include the sidecar container via the `OPERATOR_SIDECAR_IMAGE` environment
variable:

```shell
kubectl patch deployment minio-operator -n minio-operator --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/1", "value": {"name": "sidecar", "image": "miniodev/operator-sidecar:sc"}}]'
```