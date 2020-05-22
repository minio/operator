# MinIO Operator Helm Chart Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains the steps to deploy MinIO Operator in a Kubernetes cluster via Helm Chart.

## Prerequisites

- Helm Installed as explained [here](https://helm.sh/docs/intro/install/).

## Deploy MinIO Operator

To Create a MinIO Operator deployment with Helm Chart, use the command

```
helm install https://raw.githubusercontent.com/minio/minio-operator/master/helm-chart/minio-operator-1.0.0.tgz --generate-name
```

This will create a MinIO Operator deployment with default configuration. To change the configuration per your requirements, take a look at [values.yaml](./minio-operator/values.yaml) file. You can pass custom values using

```
helm install -f myvalues.yaml https://raw.githubusercontent.com/minio/minio-operator/master/helm-chart/minio-operator-1.0.0.tgz --generate-name
```

