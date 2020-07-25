# MinIO Operator KES Configuration

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to enable KES with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/operator#create-operator-and-related-resources).
- KES requires TLS enabled for MinIO. Make sure TLS is enabled as explained in the [document here](https://github.com/minio/operator/blob/master/docs/tls.md).
- KES requires a KMS backend in [configuration](https://raw.githubusercontent.com/minio/operator/master/examples/kes-config-secret.yaml). Currently KES supports [AWS Secrets Manager](https://github.com/minio/kes/wiki/AWS-SecretsManager) and [Hashicorp Vault](https://github.com/minio/kes/wiki/Hashicorp-Vault-Keystore) as KMS backend for production. We recommend setting up one of these as the KMS backend before setting up KES.

### Enable KES Configuration

KES Configuration is a part of Tenant yaml file. Check the sample file [available here](https://raw.githubusercontent.com/minio/operator/master/examples/tenant-kes.yaml). The config offers below options

#### KES Fields

| Field                 | Description |
|-----------------------|-------------|
| spec.kes | Defines the KES configuration. Refer [this](https://github.com/minio/kes) |
| spec.kes.replicas | Number of KES pods to be created. |
| spec.kes.image | Defines the KES image. |
| spec.kes.kesSecret | Secret to specify KES Configuration. This is a mandatory field. |
| spec.kes.metadata | This allows a way to map metadata to the KES pods. Internally `metadata` is a struct type as [explained here](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta). |

### Create MinIO Instance

Once you have updated the yaml file per your requirement, use `kubectl` to create the MinIO instance like

```
kubectl create -f examples/tenant-kes.yaml
```

Alternatively, you can deploy the example like this

```
kubectl create -f https://raw.githubusercontent.com/minio/operator/master/examples/tenant-kes.yaml
```

KES uses CSR for self signed certificate generation. KES requires three certificates/key pairs for working

- X.509 certificate for the KES server and the corresponding private key.
- X.509 certificate for the MinIO server and the corresponding private key.
- X.509 certificate for the KES client (MinIO is the KES client in this case) and the corresponding private key.

If `requestAutoCert` is enabled, Operator automatically creates the relevant CSRs and Certificates.
