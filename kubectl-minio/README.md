# MinIO Kubectl Plugin

## Prerequisites

- Kubernetes >= v1.17.0.
- Create PVs using [direct CSI driver](https://github.com/minio/operator/blob/master/docs/using-direct-csi.md).
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.

## Install Plugin

Download the Kubectl plugin binary from [plugin release page](https://github.com/minio/operator/releases), and place the binary in executable path on your machine (e.g. `/usr/local/bin`).

## Plugin Commands

### Operator Deployment

Command: `kubectl minio operator create`

Creates MinIO Operator Deployment along with MinIO Tenant CRD, Service account, Cluster Role and Cluster Role Binding.

Options:

- `--image=minio/k8s-operator:3.0.5`
- `--namespace=minio-operator`
- `--service-account=minio-operator`
- `--cluster-domain=cluster.local`
- `--namespace-to-watch=default`
- `--image-pull-secret=`
- `--output`

### Tenant

#### MinIO Tenant Creation

Command: `kubectl minio tenant create --name TENANT_NAME --secret SECRET_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Creates a MinIO Tenant based on the passed values.

example: `kubectl minio tenant create --name tenant1 --secret cred-secret --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--image=minio/minio:RELEASE.2020-09-10T22-02-45Z`
- `--storageClass=local`
- `--kms-secret=secret-name`
- `--console-secret=secret-name`
- `--cert-secret=secret-name`
- `--image-pull-secret=`
- `--output`

#### Add Tenant Zones

Command: `kubectl minio tenant volume add --name TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Add new volumes (and nodes) to existing MinIO Tenant.

example: `kubectl minio tenant volume add --name cluster1 --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--storageClass=local`
- `--image-pull-secret=`
- `--output`

#### List Tenant Zones

Command: `kubectl minio tenant volume list --name TENANT_NAME [options]`

List all existing MinIO Zones in the given MinIO Tenant.

example: `kubectl minio tenant volume list --name cluster1`

Options:

- `--namespace=minio`

#### Upgrade Images

Command: `kubectl minio tenant upgrade --name TENANT_NAME --image IMAGE_TAG [options]`

Upgrade MinIO Docker image for the given MinIO Tenant.

example: `kubectl minio tenant upgrade --name cluster1 --image minio/minio:edge`

Options:

- `--namespace=minio`
- `--image-pull-secret=`
- `--output`

#### Remove Tenant

Command: `kubectl minio tenant delete --name TENANT_NAME [options]`

Delete an existing MinIO Tenant.

example: `kubectl minio tenant delete --name tenant1`

Options:

- `--namespace=minio`
