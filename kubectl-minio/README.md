# MinIO Kubectl Plugin

## Prerequisites

- Kubernetes >= v1.19.0.
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.
- Create PVs.

## Install Plugin

Command: `kubectl krew install minio`

## Plugin Commands

### Operator Deployment

Command: `kubectl minio init [options]`

Creates MinIO Operator Deployment along with MinIO Tenant CRD, Service account, Cluster Role and Cluster Role Binding.

Options:

- `--image=minio/operator:v5.0.6`
- `--namespace=minio-operator`
- `--cluster-domain=cluster.local`
- `--namespace-to-watch=default`
- `--image-pull-secret=`
- `--output`

### Operator Deletion

Command: `kubectl minio delete [options]`

Deletes MinIO Operator Deployment along with MinIO Tenant CRD, Service account, Cluster Role and Cluster Role Binding.
It also removes all the Tenant instances.

Options:

- `--namespace=minio-operator`

### Tenant

#### MinIO Tenant Creation

Command: `kubectl minio tenant create TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Creates a MinIO Tenant based on the passed values. Please note that plugin adds `anti-affinity` rules to the MinIO
Tenant pods to ensure multiple pods don't end up on the same physical node. To disable this, use
the `-enable-host-sharing` flag during tenant creation.

example: `kubectl minio tenant create tenant1 --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--kes-config=kes-secret`
- `--output`

#### Add Tenant pools

Command: `kubectl minio tenant expand TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Add new volumes (and nodes) to existing MinIO Tenant.

example: `kubectl minio tenant expand tenant1 --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--output`

#### List Tenant pools

Command: `kubectl minio tenant info TENANT_NAME [options]`

List all existing MinIO pools in the given MinIO Tenant.

example: `kubectl minio tenant info tenant1`

Options:

- `--namespace=minio`

#### Upgrade Images

Command: `kubectl minio tenant upgrade TENANT_NAME --image IMAGE_TAG [options]`

Upgrade MinIO Docker image for the given MinIO Tenant.

example: `kubectl minio tenant upgrade tenant1 --image minio/minio:RELEASE.2023-06-23T20-26-00Z`

Options:

- `--namespace=minio`
- `--output`

#### Remove Tenant

Command: `kubectl minio tenant delete TENANT_NAME [options]`

Delete an existing MinIO Tenant.

example: `kubectl minio tenant delete tenant1`

Options:

- `--namespace=minio`
