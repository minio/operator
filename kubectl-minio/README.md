# MinIO Kubectl Plugin

## Prerequisites

- Kubernetes >= v1.17.0.
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.
- Create PVs.

## Install Plugin

Command: `kubectl krew install minio`

## Plugin Commands

### Operator Deployment

Command: `kubectl minio init [options]`

Creates MinIO Operator Deployment along with MinIO Tenant CRD, Service account, Cluster Role and Cluster Role Binding.

Options:

- `--image=minio/k8s-operator:3.0.5`
- `--namespace=minio-operator`
- `--cluster-domain=cluster.local`
- `--namespace-to-watch=default`
- `--image-pull-secret=`
- `--output`

### Operator Deletion

Command: `kubectl minio delete [options]`

Deletes MinIO Operator Deployment along with MinIO Tenant CRD, Service account, Cluster Role and Cluster Role Binding. It also removes all the Tenant instances.

Options:

- `--namespace=minio-operator`

### Tenant

#### MinIO Tenant Creation

Command: `kubectl minio tenant create --name TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Creates a MinIO Tenant based on the passed values.

example: `kubectl minio tenant create --name tenant1 --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--kes-config=kes-secret`
- `--output`

#### Add Tenant Zones

Command: `kubectl minio tenant expand --name TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY [options]`

Add new volumes (and nodes) to existing MinIO Tenant.

example: `kubectl minio tenant expand --name tenant1 --servers 4 --volumes 16 --capacity 16Ti`

Options:

- `--namespace=minio`
- `--output`

#### List Tenant Zones

Command: `kubectl minio tenant info --name TENANT_NAME [options]`

List all existing MinIO Zones in the given MinIO Tenant.

example: `kubectl minio tenant info --name tenant1`

Options:

- `--namespace=minio`

#### Upgrade Images

Command: `kubectl minio tenant upgrade --name TENANT_NAME --image IMAGE_TAG [options]`

Upgrade MinIO Docker image for the given MinIO Tenant.

example: `kubectl minio tenant upgrade --name tenant1 --image minio/minio:RELEASE.2020-10-09T22-55-05Z`

Options:

- `--namespace=minio`
- `--output`

#### Remove Tenant

Command: `kubectl minio tenant delete --name TENANT_NAME [options]`

Delete an existing MinIO Tenant.

example: `kubectl minio tenant delete --name tenant1`

Options:

- `--namespace=minio`
