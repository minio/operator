# MinIO Operator [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO Operator brings native support for [MinIO](https://github.com/minio/minio), [Graphical Console](https://github.com/minio/console), and [Encryption](https://github.com/minio/kes) to Kubernetes. This document explains how to get started with MinIO Operator using `kubectl minio` plugin.

## Prerequisites

- Kubernetes >= v1.17.0.
- Create PVs. We recommend [direct CSI driver](https://github.com/minio/operator/blob/master/docs/using-direct-csi.md) for PV creation.
- Install [`kubectl minio` plugin](https://github.com/minio/operator/tree/master/kubectl-minio#install-plugin).

## Operator Setup

MinIO Operator offers MinIO Tenant creation, management, upgrade, zone addition and more. Operator is meant to control and manage multiple MinIO Tenants.

To get started, create the MinIO Operator deployment. This is a _one time_ process.

```sh
kubectl minio operator create
```

Once the MinIO Operator is created, proceed with Tenant creation.

## Tenant Setup

A Tenant is a MinIO cluster created and managed by the Operator.

### Step 1: Create Tenant Namespace

Before creating a Tenant, please create the namespace where this Tenant will reside.

For logical isolation, Operator allows a single Tenant per Kubernetes Namespace.

```sh
kubectl create ns tenant1-ns
```

### Step 2: Create Secret for Tenant Credentials

Next, create the Kubernetes secret that encapsulates root credentials for MinIO Tenant. Please ensure to create secret object with literals `accesskey` and `secretkey`.

Remember to change `YOUR-ACCESS-KEY` and `YOUR-SECRET-KEY` to actual values.

```sh
kubectl create secret generic tenant1-secret --from-literal=accesskey=YOUR-ACCESS-KEY --from-literal=secretkey=YOUR-SECRET-KEY --namespace tenant1-ns
```

Note that the access key and secret key provided here is authorized to perform _all_ operations on the Tenant.

### Step 3: Create MinIO Tenant

We can create the Tenant now. Before that, please ensure you have requisite nodes and drives in place and relevant PVs are created. In below example, we ask MinIO Operator to create a Tenant with 4 nodes, 16 volumes, and 16 Ti total raw capacity (4 volumes of 1 Ti per node). This means you need to have 4 PVCs of 1 Ti each, per node, and total of 4 nodes, before attempting to create the MinIO Tenant.

We recommend [direct CSI driver](https://github.com/minio/operator/blob/master/docs/using-direct-csi.md) to create PVs.

```sh
kubectl minio tenant create --name tenant1 --secret tenant1-secret --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns --storageclass direct.csi.min.io
```

## Post Tenant Creation

### Expanding a Tenant

You can add capacity to the tenant using `kubectl minio` plugin, like this

```
kubectl minio tenant volume add --name tenant1 --servers 8 --volumes 32 --capacity 32Ti --namespace tenant1-ns
```

This will add 32 drives spread uniformly over 8 servers to the tenant `tenant1`, with additional capacity of 32Ti. Read more about [tenant expansion here](https://github.com/minio/operator/blob/master/docs/expansion.md).

## License

Use of MinIO Operator is governed by the GNU AGPLv3 or later, found in the [LICENSE](./LICENSE) file.

## Explore Further

- [Create a MinIO Tenant](https://github.com/minio/operator#create-a-minio-instance).
- [TLS for MinIO Tenant](https://github.com/minio/operator/blob/master/docs/tls.md).
- [Custom Hostname Discovery](https://github.com/minio/operator/blob/master/docs/custom-name-templates.md).
- [Apply PodSecurityPolicy](https://github.com/minio/operator/blob/master/docs/pod-security-policy.md).
- [Deploy MinIO Tenant with Console](https://github.com/minio/operator/blob/master/docs/console.md).
- [Deploy MinIO Tenant with KES](https://github.com/minio/operator/blob/master/docs/kes.md).
