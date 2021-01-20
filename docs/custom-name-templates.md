# Custom Hostname Discovery [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to control the names used for host discovery. This allows us to discover hosts using external name services, which is useful for serving with trusted certificates.

## Getting Started

If MinIO Tenant is named `tenant1`, then the four servers will be called `tenant1-pool-0-0`, `tenant1-pool-0-1`, `tenant1-pool-0-2`, and `tenant1-pool-0-3`.  If all of your hosts are available at the domain `example.com` then you can use the `--hosts-template` flag in [MinIO Operator Deployment yaml](https://github.com/minio/operator/blob/master/minio-operator.yaml) to update discovery. This will generate the discovery string `tenant1-pool-0-{0...3}.example.com`.

```yaml
  containers:
  - command:
    - /operator
    - --hosts-template
    - '{{.StatefulSet}}-{{.Ellipsis}}.example.com'
```

The following fields can be configured:

| Field       | Description                                                                                              |
|-------------|----------------------------------------------------------------------------------------------------------|
| StatefulSet | The name of the tenant StatefulSet (e.g. `minio`).                                                       |
| CIService   | The name of the service provided in `spec.serviceName`.                                                  |
| HLService   | The name of the headless service that is generated (e.g. `minio-hl-service`)                             |
| Ellipsis    | `{0...N-1}` the per-pool host numbers.                                                                   |
| Domain      | The cluster domain, either `cluster.local` or the contents of the `CLUSTER_DOMAIN` environment variable. |
