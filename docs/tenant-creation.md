# Tenant Creation

The basic recommended tenant creation method is
using [kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) to create a tenant. The
following steps will guide you through the process.

In this document we'll use the basic tenant, but there are other examples available in the `examples/kustomization`
folder in this repository.

## Using Kustomize

We have a base tenant example that can be used to create a tenant. The base tenant example is available in
the `examples/kustomization/base` directory and can be used like:

```shell
kubectl apply -k github.com/minio/operator/examples/kustomization/base
```

This will create a tenant with the name `myminio` in the namespace `minio-tenant`. The tenant will have 4 servers and 4
drives per server (16 drives in total) with a total capacity of 16Ti. The tenant will use the default storage class for
storage.

It is possible to customize the tenant by creating a `kustomization.yaml` and using the example tenant as a base.

For example, to use a particular version of MinIO in your tenant and not the one the Operator is defaulting to, you can
create a `kustomization.yaml` like:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: minio-tenant

resources:
  - github.com/minio/operator/examples/kustomization/base

patches:
  - path: tenant.yaml
```

and an overlay `tenant.yaml` like:

```yaml
apiVersion: minio.min.io/v2
kind: Tenant
metadata:
  name: myminio
  namespace: minio-tenant
spec:
  image: quay.io/minio/minio:RELEASE.2024-03-15T01-07-19Z 
```

This will create a tenant with the name `myminio` in the namespace `minio-tenant` with the MinIO version specified in
your overlay.

Assuming you placed the `kustomization.yaml` and `tenant.yaml` in the same directory, you can create the tenant like:

```shell
kubectl apply -k .
```

## Using YAML Manifests

You can create a single static YAML file containing an example tenant as shown below:

```yaml
kubectl kustomize github.com/minio/operator/examples/kustomization/base > minio-tenant.yaml
```

The YAML will have all the necessary fields to create a tenant. You can then apply the YAML to create the tenant:

```shell
kubectl apply -f minio-tenant.yaml
```