# MinIO Operator KES Configuration [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to enable KES with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/operator#operator-setup).
- KES requires a KMS backend
  in [configuration](https://raw.githubusercontent.com/minio/operator/master/examples/kes-secret.yaml). Currently KES
  supports [AWS Secrets Manager](https://github.com/minio/kes/wiki/AWS-SecretsManager)
  and [Hashicorp Vault](https://github.com/minio/kes/wiki/Hashicorp-Vault-Keystore) as KMS backend for production.S Set
  up one of these as the KMS backend before setting up KES.

### Create MinIO Tenant

We have an example Tenant with KES encryption available
at [examples/tenant-kes-encryption](../examples/tenant-kes-encryption).

You can install the example like:

```shell
kubectl apply -k github.com/minio/operator/examples/kustomization/tenant-kes-encryption
```

## KES Configuration

KES Configuration is a part of Tenant yaml file. Check the sample
file [available here](https://raw.githubusercontent.com/minio/operator/master/examples/kustomization/tenant-kes-encryption/tenant.yaml).
The config offers below options

### KES Fields

| Field              | Description                                                                                                                                                                       |
|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| spec.kes           | Defines the KES configuration. Refer [this](https://github.com/minio/kes)                                                                                                         |
| spec.kes.replicas  | Number of KES pods to be created.                                                                                                                                                 |
| spec.kes.image     | Defines the KES image.                                                                                                                                                            |
| spec.kes.kesSecret | Secret to specify KES Configuration. This is a mandatory field.                                                                                                                   |
| spec.kes.metadata  | This allows a way to map metadata to the KES pods. Internally `metadata` is a struct type as [explained here](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta). |

A complete list of values is available [here](tenant_crd.adoc#kesconfig) in the API reference.
