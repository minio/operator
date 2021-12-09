# MinIO ![license](https://img.shields.io/badge/license-AGPL%20V3-blue)

[MinIO](https://min.io) is a High Performance Object Storage released under GNU AGPLv3 or later. It is API compatible
with Amazon S3 cloud storage service. Use MinIO to build high performance infrastructure for machine learning, analytics
and application data workloads.

For more detailed documentation please visit [here](https://docs.minio.io/)

Introduction
------------

This chart bootstraps MinIO Operator on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

Configure MinIO Helm repo
--------------------

```bash
helm repo add minio https://operator.min.io/
```

Installing the Chart
--------------------

Install this chart using:

```bash
helm install \
  --namespace minio-operator \
  --create-namespace \
  minio-operator minio/operator
```

The command deploys MinIO Operator on the Kubernetes cluster in the default configuration.

Creating a Tenant
-----------------

Once the MinIO Operator Chart is successfully installed, create a MinIO Tenant using:

```bash
helm install --namespace tenant-ns \
  --create-namespace tenant minio/tenant
```

This creates a 4 Node MinIO Tenant (cluster). To change the default values, take a look at various [values.yaml](https://github.com/minio/operator/blob/master/helm/tenant/values.yaml).
