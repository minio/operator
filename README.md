# MinIO Operator Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. MinIO is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for MinIO to scale.

MinIO-Operator brings native MinIO, [MCS](https://github.com/minio/mcs), and [KES](https://github.com/minio/kes) support to Kubernetes. MinIO-Operator currently supports following features:

| Feature                 | Reference Document |
|-------------------------|--------------------|
| Create and delete highly available distributed MinIO clusters  | [Create a MinIO Instance](https://github.com/minio/minio-operator#create-a-minio-instance). |
| Automatic TLS for MinIO                                        | [Automatic TLS for MinIO Instance](https://github.com/minio/minio-operator/blob/master/docs/tls.md#automatic-csr-generation). |
| Expand an existing MinIO cluster                               | [Expand a MinIO Cluster](https://github.com/minio/minio-operator/blob/master/docs/adding-zones.md). |
| Use a custom template for hostname discovery                   | [Custom Hostname Discovery](https://github.com/minio/minio-operator/blob/master/docs/custom-name-templates.md). |
| Use PodSecurityPolicy for MinIO Pods | [Apply PodSecurityPolicy](https://github.com/minio/minio-operator/blob/master/docs/pod-security-policy.md). |
| Deploy MCS with MinIO cluster  | [Deploy MinIO Instance with MCS](https://github.com/minio/minio-operator/blob/master/docs/mcs.md). |
| Deploy KES with MinIO cluster  | [Deploy MinIO Instance with KES](https://github.com/minio/minio-operator/blob/master/docs/kes.md). |

## Getting Started

### Prerequisites

- Kubernetes version v1.17.0 and above for compatibility. MinIO Operator uses `k8s/client-go` v0.18.0.
- `kubectl` configured to refer to a Kubernetes cluster.
- Create the required PVs as [explained here](https://github.com/minio/minio-operator/blob/master/docs/creating-pv-manually.md).
- Optional: `kustomize` installed as [explained here](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md#installation).

### Create Operator Deployment

To start MinIO-Operator with default configuration, use the `minio-operator.yaml` file.

```bash
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/minio-operator.yaml
```

Advanced users can leverage [kustomize](https://github.com/kubernetes-sigs/kustomize) to customize operator configuration

```bash
git clone https://github.com/minio/minio-operator
cd operator-deployment
kustomize build | kubectl apply -f -
```

### Create a MinIO instance

Once MinIO-Operator deployment is running, you can create MinIO instances using the below command

```
kubectl apply -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/minioinstance.yaml
```

### Access MinIOInstance via Service

Add an [external service](https://kubernetes.io/docs/concepts/services-networking/service/) in MinIOInstance definition to enable Service based access to the MinIOInstance pods. Refer [the example here](https://github.com/minio/minio-operator/blob/master/examples/minioinstance.yaml?raw=true) for details on how to setup service based access for MinIOInstance pods.

### Advanced: Expose MinIO via Istio

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
