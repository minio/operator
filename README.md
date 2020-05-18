# MinIO Operator Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. MinIO is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for MinIO to scale.

MinIO-Operator brings native MinIO, [MCS](https://github.com/minio/mcs), [KES](https://github.com/minio/kes) and [mc Mirror](https://docs.minio.io/docs/minio-client-complete-guide.html#mirror) support to Kubernetes. MinIO-Operator currently supports following features:

| Feature                 | Reference Document |
|-------------------------|--------------------|
| Create and delete highly available distributed MinIO clusters  | [Create a MinIO Instance](https://github.com/minio/minio-operator#create-a-minio-instance). |
| Expand an existing MinIO cluster                               | [Expand a MinIO Cluster](https://github.com/minio/minio-operator#expand-a-minio-cluster). |
| Automatic TLS for MinIO                                        | [Automatic TLS for MinIO Instance](https://github.com/minio/minio-operator/blob/master/docs/tls.md#automatic-csr-generation). |
| Deploy MCS with MinIO cluster  | [Deploy MinIO Instance with MCS](https://github.com/minio/minio-operator/blob/master/docs/mcs.md). |
| Deploy KES with MinIO cluster  | [Deploy MinIO Instance with KES](https://github.com/minio/minio-operator/blob/master/docs/kes.md). |
| Deploy mc mirror  | [Deploy Mirror Instance](https://github.com/minio/minio-operator/blob/master/docs/mirror.md). |

## Getting Started

### Prerequisites

- Kubernetes version v1.17.0 and above for compatibility. MinIO Operator uses `k8s/client-go` v0.18.0.
- `kubectl` configured to refer to a Kubernetes cluster.

### Create Operator and related resources

To start MinIO-Operator, use the `minio-operator.yaml` file.

```bash
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/minio-operator.yaml
```

This will create all relevant resources required for the Operator to work.

You could install the MinIO Operator a custom namespace by passing the `-n` flag

```bash
kubectl create namespace minio-operator-ns
kubectl apply -n minio-operator-ns -f https://raw.githubusercontent.com/minio/minio-operator/master/minio-operator.yaml
```

### Create a MinIO instance

Once MinIO-Operator deployment is running, you can create MinIO instances using the below command

```
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance.yaml
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

### Access MinIOInstance via Service

Add an [external service](https://kubernetes.io/docs/concepts/services-networking/service/) in MinIOInstance definition to enable Service based access to the MinIOInstance pods. Refer [the example here](https://github.com/minio/minio-operator/blob/master/examples/minioinstance.yaml?raw=true) for details on how to setup service based access for MinIOInstance pods.

### Expose MinIO via Istio

Istio >= 1.4 has support for headless Services, so instead of creating an explicit `Service` for the created MinIO instance, you can also directly target the headless Service that is created by the operator.

For example, to expose the created headless Service `minio-hl-svc` on http://minio.example.com:

```
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/expose-via-istio.yaml
```

### Environment variables

These variables may be passed to operator Deployment in order to modify some of its parameters

| Name                | Default | Description                                                                                                                   |
| ---                 | ---     | ---                                                                                                                           |
| `WATCHED_NAMESPACE` |         | If set, the operator will watch only MinIO resources deployed in the specified namespace. All namespaces are watched if empty |
| `CLUSTER_DOMAIN`    | cluster.local | Cluster Domain of the Kubernetes cluster |

## Explore Further

- [MinIO Erasure Code QuickStart Guide](https://docs.min.io/docs/minio-erasure-code-quickstart-guide)
- [Use `mc` with MinIO Server](https://docs.min.io/docs/minio-client-quickstart-guide)
- [Use `aws-cli` with MinIO Server](https://docs.min.io/docs/aws-cli-with-minio)
- [The MinIO documentation website](https://docs.min.io)
