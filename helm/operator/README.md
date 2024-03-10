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

Deploy Operator with Ingress controller enabled
-----------------------------------------------

**Expose HTTP route to operator `console` service**

Install this chart using:

```bash
helm install \
  --namespace minio-operator \
  --create-namespace \
  --set console.ingress.enabled="true" \
  --set console.ingress.host="hostname" \
  --set console.ingress.ingressClassName="class-name" \
  minio-operator minio/operator
```

Replace `hostname` by the domain of your external load balancer to interface it with the ingress controller.

Replace `class-name` by the ingress class name of the ingress controller intended to perform the routing (for example `nginx`).

**Expose HTTPS route to operator `console` service**

Install this chart using:

```bash
helm install \
  --namespace minio-operator \
  --create-namespace \
  --set console.ingress.enabled="true" \
  --set console.ingress.host="hostname" \
  --set console.ingress.tls.enabled="true" \
  --set console.ingress.tls.secretName="certificate-tls-secret-name" \
  --set console.ingress.ingressClassName="class-name" \
  minio-operator minio/operator
```

Replace `hostname` by the domain of your external load balancer to interface it with the ingress controller. The `hostname` must match either the Common Name (CN) in your CA certificate or any of the Subject Alternative Names (SANs) for the certificate to be considered valid.

Replace `class-name` by the ingress class name of the ingress controller intended to perform the routing (for example `nginx`).

To mount the CA certificate as a secret use:

```bash
kubectl create secret tls certificate-tls-secret-name \
  --cert=/path/to/certificate.crt \
  --key=/path/to/certificatePrivate.key \
  --namespace minio-operator
```

Creating a Tenant
-----------------

Once the MinIO Operator Chart is successfully installed, create a MinIO Tenant using:

```bash
helm install --namespace tenant-ns \
  --create-namespace tenant minio/tenant
```

This creates a 4 Node MinIO Tenant (cluster). To change the default values, take a look at various [values.yaml](https://github.com/minio/operator/blob/master/helm/tenant/values.yaml).
