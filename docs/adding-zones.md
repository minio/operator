# Adding Zones to a MinIO Cluster

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to add zones to an existing MinIO Cluster with Operator. This document is only applicable to a MinIO Cluster created by MinIO Operator.

Read more about MinIO Zones design in [MinIO Docs](https://github.com/minio/minio/blob/master/docs/distributed).

## Getting Started

Assuming you have a MinIO cluster with single zone, `zone-0` with 4 drives (as shown in [examples](https://github.com/minio/minio-operator/tree/master/examples)). You can dd a new zone `zone-1` with 4 drives using `kubectl patch` command.

```
kubectl patch tenants.minio.min.io minio --patch "$(cat examples/patch.yaml)" --type=merge
```

If you're using a custom configuration (e.g. multiple zones or higher number of drives per zone), make sure to change `patch.yaml` accordingly.

**NOTE**: Important points to consider _before_ using cluster expansion:

- During cluster expansion, MinIO Operator removes the existing StatefulSet and creates a new StatefulSet with required number of Pods. This means, there is a short downtime during expansion, as the pods are terminated and created again. As existing StatefulSet pods are terminated, its PVCs are also deleted. It is _very important_ to ensure PVs bound to MinIO StatefulSet PVCs are not deleted at this time to avoid data loss. We recommend configuring every PV with reclaim policy [`retain`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#retain), to ensure the PV is not deleted. If you attempt cluster expansion while the PV reclaim policy is set to something else, it may lead to data loss. If you have the reclaim policy set to something else, change it as explained in [Kubernetes documents](https://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/).

- MinIO server currently doesn't support zone removal. So, please ensure to not remove zones in patch.yaml file while applying the patch. It can have unintended consequences including missing data or failure of MinIO cluster.

## Rules of Adding Zones

Each zone is a self contained entity with same SLA's (read/write quorum) for each object as original cluster. By using the existing namespace for lookup validation MinIO ensures conflicting objects are not created. When no such object exists then MinIO simply uses the least used zone. There are no limits on how many zones can be combined.

There is only one requirement, i.e. based on initial zone's erasure set count (say `n`), new zones are expected to have a minimum of `n` drives to match the original cluster SLA or it should be in multiples of `n`. For example if initial set count is 4, new zones should have at least 4 or multiple of 4 drives.

Read more about MinIO Zones design in [MinIO Docs](https://github.com/minio/minio/blob/master/docs/distributed).

## Effects on KES/TLS Enabled Instance

If your MinIO Operator configuration has [KES](https://github.com/minio/minio-operator/blob/master/docs/kes.md) or [Automatic TLS](https://github.com/minio/minio-operator/blob/master/docs/tls.md#automatic-csr-generation) enabled, there are additional considerations:

- When new zones are added, Operator invalidates older self signed TLS certificates and the related secrets. Operator then creates new certificate signing requests (CSR). This is because there are new MinIO nodes that must be added in certificate DNS names. The administrator must approve these CSRs for MinIO server to be deployed again. Unless the CSR are approved, Operator will not create MinIO StatefulSet pods.

- If you're using your own certificates, as explained [here](https://github.com/minio/minio-operator/blob/master/docs/tls.md#pass-certificate-secret-to-tenant), please ensure to use/update proper certificates that allow older and new MinIO nodes.

## Downtime

Since Operator deletes existing StatefulSet and related CSR / Secrets (for TLS enabled setups), before re-creating a new StatefulSet and other resources, there is a downtime involved when adding zones. This downtime is generally few minutes, assuming CSRs are approved quickly and resources are available for new StatefulSet to be created.
