# MinIO Operator STS: Native IAM Authentication for Kubernetes

Each example in this folder contains an example using a different SDK on how to adopt Operator's STS.

> ⚠️ This feature is an alpha release and is subject to breaking changes in future releases.

# Requirements

## Enabling STS functionality

At the moment, the STS feature ships off by default, to turn it on switch `OPERATOR_STS_ENABLED` to `on` on
the `minio-operator` deployment.

## TLS

The STS functionality works only with TLS configured. We can request certificates automatically, but additional you can
use `cert-manager` or bring your own certificates.

# Installation

To install the example, you need an existing tenant, optionally, you can install the `tenant-lite` example, or
the `tenant-certmanager` example

# 0. Enable STS Functionality

If you haven't done so, enable the STS feature on operator by turning setting the feature flag `OPERATOR_STS_ENABLED=on`

```shell
kubectl -n minio-operator set env deployment/minio-operator OPERATOR_STS_ENABLED=on
```

# 1. Install Tenant (Optional)

```shell
kubectl apply -k examples/kustomization/sts-example/tenant
```

For an example with Cert Manager

```shell
kubectl apply -k examples/kustomization/sts-example/tenant-certmanager
```

# 2. Create a bucket and a policy (Optional)

We will set up some sample buckets to access from our sample application

```shell
kubectl apply -k examples/kustomization/sts-example/sample-data
```

# 3. Install sample application

The sample application will install to `sts-client` namespace and grant access to the job called `sts-example-job` to
access `tenant` with the MinIO Policy called `test-bucket-rw` that we created in the previous step on
namespace `minio-tenant-1` by installing a `PolicyBinding` on the `minio-tenant-1` namespace.

Example policy binding (see CRD documentation in [policybinding_crd.adoc](../../../docs/policybinding_crd.adoc) )

```yaml
apiVersion: sts.min.io/v1alpha1
kind: PolicyBinding
metadata:
  name: binding-1
  namespace: minio-tenant-1
spec:
  application:
    namespace: sts-client
    serviceaccount: stsclient-sa
  policies:
    - test-bucket-rw

```

To install the sample application, which uses the Go SDK, run:

```shell
kubectl apply -k examples/kustomization/sts-example/
```

To use a specfic SDK, use any of the following:

### Go

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/minio-sdk/go
```

### Java

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/minio-sdk/java
```

### Python

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/minio-sdk/python
```

### Python: AWS Boto3 SDK

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/aws-sdk/python
```

### Javascript

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/minio-sdk/javascript
```

### .NET

```shell
kubectl apply -k examples/kustomization/sts-example/sample-clients/minio-sdk/dotnet
```