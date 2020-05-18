# MinIO Operator MCS Configuration

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to enable MCS with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/minio-operator#create-operator-and-related-resources).

### Enable MCS Configuration

MCS Configuration is a part of MinIOInstance yaml file. Check the sample file [available here](https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance-mcs.yaml). The config offers below options

#### MCS Fields

| Field                 | Description |
|-----------------------|-------------|
| spec.mcs | Defines the mcs configuration. mcs is a graphical user interface for MinIO. Refer [this](https://github.com/minio/mcs) |
| spec.mcs.image | Defines the mcs image |
| spec.mcs.replicas | Number of MCS pods to be created. |
| spec.mcs.mcsSecret | Use this secret to assign mcs credentials to MinIOInstance. |
| spec.mcs.selector | Add a selector for the mcs. Which will be used by the mcs container for grouping. (Note: Should not match the labels provided in `spec.selector`) |
| spec.mcs.metadata | This allows a way to map metadata to the mcs container. Internally `metadata` is a struct type as [explained here](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta). [Note: Should match the labels in `spec.mcs.selector`] |

### Create MinIO Instance

Once you have updated the yaml file per your requirement, use `kubectl` to create the MinIO instance like

```
kubectl create -f examples/minioinstance-mcs.yaml
```

Alternatively, you can deploy the example like this

```
kubectl create -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance-mcs.yaml
```

Above example file uses CSR for self signed certificate generation. MinIO requires one certificates/key pair 

- X.509 certificate for the MinIO server and the corresponding private key.

Accordingly, you'll need to approve the CSR request, using below approach

```
kubectl get csr
kubectl certificate approve <csr-name>
```

Once all the CSRs are approved, MinIO Operator will deploy MCS Pods and start MinIO Server with MCS integration.
