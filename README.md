# MinIO Operator [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO Operator brings native support for [MinIO](https://github.com/minio/minio), [Graphical Console](https://github.com/minio/console), and [Encryption](https://github.com/minio/kes) to Kubernetes. This document explains how to get started with MinIO Operator using `kubectl minio` plugin.

## Prerequisites

- Kubernetes >= v1.17.0.
- Create PVs.
- Install [`kubectl minio` plugin using `krew install minio`.

## Operator Setup

MinIO Operator offers MinIO Tenant creation, management, upgrade, zone addition and more. Operator is meant to control and manage multiple MinIO Tenants.

To get started, initialize the MinIO Operator deployment. This is a _one time_ process.

```sh
kubectl minio init
```

Once the MinIO Operator is created, proceed with Tenant creation.

## Tenant Setup

A Tenant is a MinIO cluster created and managed by the Operator. Before creating tenant, please ensure you have requisite nodes and drives in place and relevant PVs are created.

In below example, we ask MinIO Operator to create a Tenant yaml with 4 nodes, 16 volumes, and 16 Ti total raw capacity (4 volumes of 1 Ti per node). This means you need to have 4 PVs of 1Ti each per node, with a total of 4 nodes, before attempting to create the MinIO tenant.

We recommend [direct CSI driver](https://github.com/minio/operator/blob/master/docs/using-direct-csi.md) to create PVs.

```sh
kubectl minio tenant create --name tenant1 --secret tenant1-secret --servers 4 --volumes 16 --capacity 16Ti
```

Optionally, you can generate a yaml file with the `-o` flag in above command and modify the yaml file as per your specific requirements. Once you verify and optionally add any other relevant fields to the file, create the tenant

```sh
kubectl minio tenant create --name tenant1 --secret tenant1-secret --servers 4 --volumes 16 --capacity 16Ti -o > tenant.yaml
kubectl apply -f tenant.yaml
```

## Post Tenant Creation

### Expanding a Tenant

You can add capacity to the tenant using `kubectl minio` plugin, like this

```
kubectl minio tenant expand --name tenant1 --servers 8 --volumes 32 --capacity 32Ti
```

This will add 32 drives spread uniformly over 8 servers to the tenant `tenant1`, with additional capacity of 32Ti. Read more about [tenant expansion here](https://github.com/minio/operator/blob/master/docs/expansion.md).

## License

Use of MinIO Operator is governed by the GNU AGPLv3 or later, found in the [LICENSE](./LICENSE) file.

## Explore Further

- [Create a MinIO Tenant](https://github.com/minio/operator#create-a-minio-instance).
- [TLS for MinIO Tenant](https://github.com/minio/operator/blob/master/docs/tls.md).
- [Examples for MinIO Tenant Settings](https://github.com/minio/operator/blob/master/docs/examples.md)
- [Custom Hostname Discovery](https://github.com/minio/operator/blob/master/docs/custom-name-templates.md).
- [Apply PodSecurityPolicy](https://github.com/minio/operator/blob/master/docs/pod-security-policy.md).
- [Deploy MinIO Tenant with Console](https://github.com/minio/operator/blob/master/docs/console.md).
- [Deploy MinIO Tenant with KES](https://github.com/minio/operator/blob/master/docs/kes.md).
