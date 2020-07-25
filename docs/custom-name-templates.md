# Custom Hostname Discovery

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to control the names used for host discovery.  This allows us to discover hosts using external name services, which is useful for serving with trusted certificates.

## Getting Started

Assuming you have a MinIO cluster with single zone, `zone-0` with 4 drives (as shown in [examples](https://github.com/minio/operator/tree/master/examples)). You can dd a new zone `zone-1` with 4 drives using `kubectl patch` command.

The example cluster is named minio, so the four servers will be called `minio-0`, `minio-1`, `minio-2`, and `minio-3`.  If all of your hosts are available at the domain `example.com` then you can use the `--hosts-template` flag to update discovery:

```
  containers:
  - command:
    - /operator
    - --hosts-template
    - '{{.StatefulSet}}-{{.Ellipsis}}.example.com'
```

This will generate the discovery string `minio-{0...3}.example.com`.  The following fields are available
| Field                 | Description |
|-----------------------|-------------|
| StatefulSet | The name of the instance StatefulSet (e.g. `minio`). |
| CIService | The name of the service provided in `spec.serviceName`. |
| HLService | The name of the headless service that is generated (e.g. `minio-hl-service`) |
| Ellipsis | `{0...N-1}` the per-zone host numbers. |
| Domain | The cluster domain, either `cluster.local` or the contents of the `CLUSTER_DOMAIN` environment variable. |
