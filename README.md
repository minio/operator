# Minio Operator Guide

[![Slack](https://slack.minio.io/slack?type=svg)](https://slack.minio.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

Minio is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. Minio is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for Minio to scale. There are multiple options to deploy Minio on Kubernetes:

- Helm Chart: Minio Helm Chart offers customizable and easy Minio deployment with a single command. Refer [Minio Helm Chart repository documentation](https://github.com/helm/charts/tree/master/stable/minio) for more details.

- YAML File: Minio can be deployed with yaml files via kubectl. Refer [Minio yaml file documentation](https://docs.minio.io/docs/deploy-minio-on-kubernetes.html) to deploy Minio using yaml files.

- Minio-Operator: Operator creates and manages distributed Minio deployments running on Kubernetes, using CustomResourceDefinitions and Controller.

## Getting Started

### Prerequisites

- Kubernetes cluster 1.8.0+.
- `kubectl` configured to refer to relevant Kubernetes cluster.

### Create Operator and related resources

To start Minio-Operator, use the `docs/minio-operator.yaml` file.

```
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-operator.yaml?raw=true
```

This will create all relevant resources required for the Operator to work. Here is a list of resources created by above `yaml` file:

- `Namespace`: Custom namespace for Minio-Operator. By default it is names as `minio-operator-ns`.
- `CustomResourceDefinition`: Custom resource definition named as `minioinstances.miniocontroller.minio.io`.
- `ClusterRole`: A cluster wide role for the controller. It is named as `minio-operator-role`. This is used for RBAC.
- `ServiceAccount`: Service account is used by the custom controller to access the cluster. Account name by default is `minio-operator-sa`.
- `ClusterRoleBinding`: This cluster wide binding binds the service account `minio-operator-sa` to cluster role `minio-operator-role`.
- `Deployment`: Deployment creates a pod using the Minio-Operator Docker image. This is where the custom controller runs and looks after any changes in custom resource.

### Create a Minio instance

Once Minio-Operator deployment is running, you can create Minio instances using the below command

```
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minio-secret.yaml?raw=true
kubectl create -f https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minioinstance.yaml?raw=true
```

## Features

Minio-Operator currently supports following features:

- Create and delete highly available distributed Minio clusters.
- Upgrading existing distributed Minio clusters.

Refer [`minioinstance.yaml`](https://github.com/minio/minio-operator/blob/master/docs/minio-examples/minioinstance.yaml?raw=true) for details on how to pass supported fields to 
the operator.

## Upcoming features

With next release, we'll add Minio cluster mirror option based on [mc mirror](https://docs.minio.io/docs/minio-client-complete-guide.html#mirror) command.

## Explore Further

- [Minio Erasure Code QuickStart Guide](https://docs.minio.io/docs/minio-erasure-code-quickstart-guide)
- [Use `mc` with Minio Server](https://docs.minio.io/docs/minio-client-quickstart-guide)
- [Use `aws-cli` with Minio Server](https://docs.minio.io/docs/aws-cli-with-minio)
- [The Minio documentation website](https://docs.minio.io)
