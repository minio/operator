# MinIO Operator Mirror Configuration

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to enable Mirror with MinIO Operator.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/minio-operator#create-operator-and-related-resources).

### Enable MCS Configuration

Mirror Configuration is a part of MirrorInstance yaml file. Check the sample file [available here](https://raw.githubusercontent.com/minio/minio-operator/master/examples/mirrorinstance.yaml). The config offers below options

#### Mirror Fields

| Field                 | Description |
|-----------------------|-------------|
| kind                  | This defines the resource type to be created. MinIO Operator CRD defines the `kind` for Mirror Operation as `MirrorInstance`.|
| metadata              | This field allows a way to assign metadata to a MirrorInstance. Internally `metadata` is a struct type as [explained here](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta).|
| spec.metadata         | Define the object metadata to be passed to all the members pods of this MirrorInstance. This allows adding annotations and labels.|
| spec.image            | Set the container registry and image tag for MinIO Client to be used in the MirrorInstance.|
| spec.env              | Add Mirror specific environment variables. There are two mandatory fields required for Mirror to work. `MC_HOST_source` is the environment variable to specify the source MinIO cluster for mirror operation. `MC_HOST_target` is the environment variable to specify the target MinIO cluster for mirror operation. The value of these environment variables in the format `https://<access_key>:<secret_key>@<minio_server_url>`. Refer [the document](https://github.com/minio/mc/blob/master/docs/minio-client-complete-guide.md#specify-host-configuration-through-environment-variable) for further details. |
| spec.args.source      | Specify the source location for mirror operation. This can be a top level alias (e.g `source`), a bucket (e.g `source/bucket`), or a prefix (e.g `source/bucket/prefix`.) |
| spec.args.target      | Specify the target location for mirror operation. This can be a top level alias (e.g `target`), a bucket (e.g `target/bucket`), or a prefix (e.g `target/bucket/prefix`.)  |
| spec.args.flags       | Specify the flags to fine tune the mirror operation. Refer the [mc mirror documentation](https://github.com/minio/mc/blob/master/docs/minio-client-complete-guide.md#mirror) for possible values for flags. |
| spec.resources | Specify CPU and Memory resources for each MirrorInstance Job container. Refer [this document](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-types) for details. |

### Create Mirror Instance

Once you have updated the yaml file per your requirement, use `kubectl` to create the Mirror instance like

```
kubectl create -f examples/mirrorinstance.yaml
```

Alternatively, you can deploy the example like this

```
kubectl create -f https://raw.githubusercontent.com/minio/minio-operator/master/examples/mirrorinstance.yaml
```
