# MinIO Kubectl Plugin

This is a `kubectl` plugin to interact with MinIO Operator on Kubernetes.

## Prerequisites

- Kubernetes cluster with storage nodes earmarked.
- Required number of drives attached to each storage node. All drives mounted and formatted.
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.

Common Flags

- `--namespace=minio-operator`

## Operator Deployment

Command:

`kubectl minio init [options] [flags]`

Options:

- `--image=minio/k8s-operator:2.0.6`
- `--namespace=minio-operator`
- `--service-account=minio-operator`
- `--cluster-domain=cluster.local`
- `--namespace-to-watch=default`

Flags:

- `--o`

## Tenant

### MinIO Tenant Creation

Command:

`kubectl minio tenant create NAME ACCESS_KEY SECRET_KEY rack1:4 4 1Ti [options] [flags]`

Options:

- `--namespace=minio`
- `--image=minio/minio:RELEASE.2020-07-12T19-14-17Z`
- `--storageClass=local`
- `--kms-secret=secret-name`
- `--console-secret=secret-name`
- `--cert-secret=secret-name`

Flags:

- `--disable-tls`
- `-o`

### Scale Tenant Zones

Command:

`kubectl minio tenant scale NAME rack2:4 [options]`

Options:

- `--namespace=minio`

### Update Images

Command:

`kubectl minio tenant update NAME minio/minio:RELEASE.2020-06-12T00-06-19Z [options]`

Options:

- `--namespace=minio`

### Remove Tenant

Command:

`kubectl minio tenant delete NAME [options]`

Options:

- `--namespace=minio`
