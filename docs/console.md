# MinIO Operator Console Configuration

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to enable Console with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/operator#create-operator-and-related-resources).

### Enable Console Configuration

Console Configuration is a part of Tenant yaml file. Check the sample file [available here](https://raw.githubusercontent.com/minio/operator/master/examples/tenant-console.yaml). The config offers below options

#### Console Fields

| Field                 | Description |
|-----------------------|-------------|
| spec.console | Defines the console configuration. console is a graphical user interface for MinIO. Refer [this](https://github.com/minio/console) |
| spec.console.image | Defines the MinIO Console image |
| spec.console.replicas | Number of MinIO Console pods to be created. |
| spec.console.consoleSecret | Use this secret to assign console credentials to Tenant. |
| spec.console.metadata | This allows a way to map metadata to the console container. Internally `metadata` is a struct type as [explained here](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta). |

### Create MinIO Instance

Once you have updated the yaml file per your requirement, use `kubectl` to create the MinIO instance like

```
kubectl create -f examples/tenant-console.yaml
```

Alternatively, you can deploy the example like this

```
kubectl create -f https://raw.githubusercontent.com/minio/operator/master/examples/tenant-console.yaml
```

Above example file uses CSR for self signed certificate generation. MinIO requires one certificates/key pair 

- X.509 certificate for the MinIO server and the corresponding private key.

Accordingly, you'll need to approve the CSR request, using below approach

```
kubectl get csr
kubectl certificate approve <csr-name>
```

Once all the CSRs are approved, MinIO Operator will deploy Console Pods and start MinIO Server with Console integration.
