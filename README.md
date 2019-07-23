# MinIO Operator Guide

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

MinIO is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. MinIO is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for MinIO to scale. There are multiple options to deploy MinIO on Kubernetes:

- Helm Chart: MinIO Helm Chart offers customizable and easy MinIO deployment with a single command. Refer [MinIO Helm Chart repository documentation](https://github.com/helm/charts/tree/master/stable/minio) for more details.

- YAML File: MinIO can be deployed with yaml files via kubectl. Refer [MinIO yaml file documentation](https://docs.min.io/docs/deploy-minio-on-kubernetes.html) to deploy MinIO using yaml files.

- MinIO-Operator: Operator creates and manages distributed MinIO deployments running on Kubernetes, using CustomResourceDefinitions and Controller.

## Getting Started

### Prerequisites

- Kubernetes API v1.15 and above.
- `kubectl` configured to refer to relevant Kubernetes cluster.

### Create Operator and related resources

To start MinIO-Operator, use the `docs/minio-operator.yaml` file.

```
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-operator.yaml?raw=true
```

This will create all relevant resources required for the Operator to work. Here is a list of resources created by above `yaml` file:

- `Namespace`: Custom namespace for MinIO-Operator. By default it is names as `minio-operator-ns`.
- `CustomResourceDefinition`: Custom resource definition named as `minioinstances.miniocontroller.min.io`.
- `ClusterRole`: A cluster wide role for the controller. It is named as `minio-operator-role`. This is used for RBAC.
- `ServiceAccount`: Service account is used by the custom controller to access the cluster. Account name by default is `minio-operator-sa`.
- `ClusterRoleBinding`: This cluster wide binding binds the service account `minio-operator-sa` to cluster role `minio-operator-role`.
- `Deployment`: Deployment creates a pod using the MinIO-Operator Docker image. This is where the custom controller runs and looks after any changes in custom resource.

### Create a MinIO instance

Once MinIO-Operator deployment is running, you can create MinIO instances using the below command

```
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minio-secret.yaml?raw=true
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minioinstance.yaml?raw=true
```

## Features

MinIO-Operator currently supports following features:

- Create and delete highly available distributed MinIO clusters.
- Upgrading existing distributed MinIO clusters.

Refer [`minioinstance.yaml`](https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minioinstance.yaml?raw=true) for details on how to pass supported fields to 
the operator.

## Upcoming features
- Federation and CoreDNS
- Continuous remote site mirroring with [`mc mirror`](https://docs.minio.io/docs/minio-client-complete-guide.html#mirror)

## Explore Further

- [MinIO Erasure Code QuickStart Guide](https://docs.min.io/docs/minio-erasure-code-quickstart-guide)
- [Use `mc` with MinIO Server](https://docs.min.io/docs/minio-client-quickstart-guide)
- [Use `aws-cli` with MinIO Server](https://docs.min.io/docs/aws-cli-with-minio)
- [The MinIO documentation website](https://docs.min.io)
