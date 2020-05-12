# MinIO Operator Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. MinIO is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for MinIO to scale. There are multiple options to deploy MinIO on Kubernetes:

- Helm Chart: MinIO Helm Chart offers customizable and easy MinIO deployment with a single command. Refer [MinIO Helm Chart repository documentation](https://github.com/helm/charts/tree/master/stable/minio) for more details.

- YAML File: MinIO can be deployed with yaml files via kubectl. Refer [MinIO yaml file documentation](https://docs.min.io/docs/deploy-minio-on-kubernetes.html) to deploy MinIO using yaml files.

- MinIO-Operator: Operator creates and manages distributed MinIO deployments running on Kubernetes, using CustomResourceDefinitions and Controller.

## Getting Started

### Prerequisites

- Kubernetes version v1.17.0 and above for compatibility. MinIO Operator uses `k8s/client-go` v0.18.0.
- `kubectl` configured to refer to a Kubernetes cluster.

### Create Operator and related resources

To start MinIO-Operator, use the `docs/minio-operator.yaml` file.

```
kubectl create -f https://raw.githubusercontent.com/minio/minio-operator/master/minio-operator.yaml
```

This will create all relevant resources required for the Operator to work. Here is a list of resources created by above `yaml` file:

- `Namespace`: Custom namespace for MinIO-Operator. By default it is named as `minio-operator-ns`.
- `CustomResourceDefinition`: Custom resource definition named as `minioinstances.operator.min.io`.
- `ClusterRole`: A cluster wide role for the controller. It is named as `minio-operator-role`. This is used for RBAC.
- `ServiceAccount`: Service account is used by the custom controller to access the cluster. Account name by default is `minio-operator-sa`.
- `ClusterRoleBinding`: This cluster wide binding binds the service account `minio-operator-sa` to cluster role `minio-operator-role`.
- `Deployment`: Deployment creates a pod using the MinIO-Operator Docker image. This is where the custom controller runs and looks after any changes in custom resource.

### Environment variables

These variables may be passed to operator Deployment in order to modify some of its parameters

| Name                | Default | Description                                                                                                                   |
| ---                 | ---     | ---                                                                                                                           |
| `WATCHED_NAMESPACE` |         | If set, the operator will watch only MinIO resources deployed in the specified namespace. All namespaces are watched if empty |
| `CLUSTER_DOMAIN`    | cluster.local | Cluster Domain of the Kubernetes cluster |

### Create a MinIO instance

Once MinIO-Operator deployment is running, you can create MinIO instances using the below command

```
kubectl create -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance.yaml
```

### Expand a MinIO cluster

After you have a distributed MinIO Cluster running (zones.server >= 4), you can expand the MinIO cluster using

```
kubectl patch minioinstances.operator.min.io minio --patch "$(cat examples/patch.yaml)" --type=merge
```

You can expand an existing cluster by adding new zones to the `patch.yaml` and run the above `kubectl-patch` command.

**NOTE**: Important point to consider _before_ using cluster expansion:

During cluster expansion, MinIO Operator removes the existing StatefulSet and creates a new StatefulSet with required number of Pods. This means, there is a short downtime during expansion, as the pods are terminated and created again.

As existing StatefulSet pods are terminated, its PVCs are also deleted. It is _very important_ to ensure PVs bound to MinIO StatefulSet PVCs are not deleted at this time to avoid data loss. We recommend configuring every PV with reclaim policy [`retain`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#retain), to ensure the PV is not deleted.

If you attempt cluster expansion while the PV reclaim policy is set to something else, it may lead to data loss. If you have the reclaim policy set to something else, change it as explained in [Kubernetes documents](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/).

### Expose MinIO via Istio

Istio >= 1.4 has support for headless Services, so instead of creating an explicit `Service` for the created MinIO instance, you can also directly target the headless Service that is created by the operator.

For example, to expose the created headless Service `minio-hl-svc` on http://minio.example.com:

```
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/expose-via-istio.yaml
```

## Features

MinIO-Operator currently supports following features:

- Create and delete highly available distributed MinIO clusters.
- Expand an existing MinIO cluster.
- Upgrading existing distributed MinIO clusters.

Refer [`minioinstance.yaml`](https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance.yaml) for details on how to pass supported fields to the operator.

## Upcoming features

- Continuous remote site mirroring with [`mc mirror`](https://docs.minio.io/docs/minio-client-complete-guide.html#mirror)

## Explore Further

- [MinIO Erasure Code QuickStart Guide](https://docs.min.io/docs/minio-erasure-code-quickstart-guide)
- [Use `mc` with MinIO Server](https://docs.min.io/docs/minio-client-quickstart-guide)
- [Use `aws-cli` with MinIO Server](https://docs.min.io/docs/aws-cli-with-minio)
- [The MinIO documentation website](https://docs.min.io)
