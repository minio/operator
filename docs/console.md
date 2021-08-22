# Console Configuration [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to enable Console with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/operator#operator-setup).
- Install [`kubectl minio` plugin](https://github.com/minio/operator/tree/master/kubectl-minio#install-plugin).

### Create MinIO Tenant

Use `kubectl minio` plugin to create the MinIO tenant with console enabled:

```
kubectl create ns tenant1-ns
kubectl create secret generic tenant1-secret --from-literal=accesskey=YOUR-ACCESS-KEY --from-literal=secretkey=YOUR-SECRET-KEY --namespace tenant1-ns
kubectl create -f https://raw.githubusercontent.com/minio/operator/master/examples/console-secret.yaml --namespace tenant1-ns
kubectl minio tenant create --name tenant1 --secret tenant1-secret --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns --console-secret console-secret
```

A complete list of values is available [here](crd.adoc##consoleconfiguration) in the API reference.
