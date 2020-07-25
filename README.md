# MinIO Operator Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

MinIO is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. MinIO is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect launchpad for MinIO to scale.

MinIO-Operator brings native MinIO, [MCS](https://github.com/minio/mcs), and [KES](https://github.com/minio/kes) support to Kubernetes. MinIO-Operator currently supports following features:

| Feature                 | Reference Document |
|-------------------------|--------------------|
| Create and delete highly available distributed MinIO clusters  | [Create a MinIO Instance](https://github.com/minio/operator#create-a-minio-instance). |
| TLS Configuration  | [TLS for MinIO Instance](https://github.com/minio/operator/blob/master/docs/tls.md). |
| Expand an existing MinIO cluster | [Expand a MinIO Cluster](https://github.com/minio/operator/blob/master/docs/adding-zones.md). |
| Use a custom template for hostname discovery | [Custom Hostname Discovery](https://github.com/minio/operator/blob/master/docs/custom-name-templates.md). |
| Use PodSecurityPolicy for MinIO Pods | [Apply PodSecurityPolicy](https://github.com/minio/operator/blob/master/docs/pod-security-policy.md). |
| Deploy MCS with MinIO cluster  | [Deploy MinIO Instance with MCS](https://github.com/minio/operator/blob/master/docs/mcs.md). |
| Deploy KES with MinIO cluster  | [Deploy MinIO Instance with KES](https://github.com/minio/operator/blob/master/docs/kes.md). |

## Getting Started

### Prerequisites

- Kubernetes version v1.17.0 and above for compatibility. MinIO Operator uses `k8s/client-go` v0.18.0.
- `kubectl` configured to refer to a Kubernetes cluster.
- Create the required PVs using [direct CSI driver](https://github.com/minio/operator/blob/master/docs/using-direct-csi.md).
- Optional: `kustomize` installed as [explained here](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md#installation).

### Create Operator Deployment

To start MinIO-Operator with default configuration, use the `operator.yaml` file.

```bash
kubectl apply -f https://raw.githubusercontent.com/minio/operator/master/operator.yaml
```

Advanced users can leverage [kustomize](https://github.com/kubernetes-sigs/kustomize) to customize operator configuration

```bash
git clone https://github.com/minio/operator
kustomize build | kubectl apply -f -
```

### Create a MinIO instance

Once MinIO-Operator deployment is running, you can create MinIO instances using the below command

```
kubectl apply -f https://raw.githubusercontent.com/minio/operator/master/examples/tenant.yaml
```

### Access Tenant via Service

Add an [external service](https://kubernetes.io/docs/concepts/services-networking/service/) in Tenant definition to enable Service based access to the Tenant pods. Refer [the example here](https://github.com/minio/operator/blob/master/examples/tenant.yaml?raw=true) for details on how to setup service based access for Tenant pods.

### Environment variables

These variables may be passed to operator Deployment in order to modify some of its parameters

| Name                | Default | Description                                                                                                                   |
| ---                 | ---     | ---                                                                                                                           |
| `CLUSTER_DOMAIN`    | `cluster.local` | Cluster Domain of the Kubernetes cluster |
| `WATCHED_NAMESPACE` | `-` | If set, the operator will watch MinIOInstance resources in specified namespace only. If empty, operator will watch all namespaces. |

## Explore Further

- [MinIO Erasure Code QuickStart Guide](https://docs.min.io/docs/minio-erasure-code-quickstart-guide)
- [Use `mc` with MinIO Server](https://docs.min.io/docs/minio-client-quickstart-guide)
- [Use `aws-cli` with MinIO Server](https://docs.min.io/docs/aws-cli-with-minio)
- [The MinIO documentation website](https://docs.min.io)
- Expose MinIO via Istio: Istio >= 1.4 has support for headless Services, so instead of creating an explicit `Service` for the created MinIO instance, you can also directly target the headless Service that is created by the operator. Use [Istio Ingress Gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/) to configure Istio to expose the MinIO service outside of the service mesh.
