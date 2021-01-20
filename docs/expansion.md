# Adding capacity to a MinIO Tenant [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to expand an existing MinIO Tenant with Operator. This is only applicable to a Tenant (MinIO Deployment) created by MinIO Operator.

MinIO expansion is done in terms of MinIO pools, read more about the design in [MinIO Docs](https://github.com/minio/minio/blob/master/docs/distributed).

## Getting Started

You can add capacity to the tenant using `kubectl minio` plugin.

```
kubectl minio tenant volume add --name TENANT_NAME --servers SERVERS --volumes TOTAL_VOLUMES --capacity TOTAL_RAW_CAPACITY
```

Remember to replace `TENANT_NAME` with tenant name where you want to add volumes, `SERVERS` with new servers to be added to the tenant, `TOTAL_VOLUMES` with total new volumes to be added to tenant and `TOTAL_RAW_CAPACITY` with total new raw capacity to be added to the tenant.

**NOTE**: Important points to consider _before_ using Tenant expansion:

- During Tenant expansion, MinIO Operator removes the existing StatefulSet and creates a new StatefulSet with required number of Pods. This means, there is a  downtime during expansion, as the pods are terminated and created again. As existing StatefulSet pods are terminated, its PVCs are also deleted. It is _very important_ to ensure PVs bound to MinIO StatefulSet PVCs are not deleted at this time to avoid data loss. We recommend configuring every PV with reclaim policy [`retain`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#retain), to ensure the PV is not deleted. If you attempt Tenant expansion while the PV reclaim policy is set to something else, it may lead to data loss. If you have the reclaim policy set to something else, change it as explained in [Kubernetes documents](https://kubernetes.io/docs/tasks/administer-Tenant/change-pv-reclaim-policy/).

- MinIO server currently doesn't support reducing storage capacity.

## Underlying Details in Tenant Expansion

### What are MinIO pools

A MinIO pool is a self contained entity with same SLA's (read/write quorum) for each object. There are no limits on how many pools can be combined. After adding of a pool, MinIO simply uses the least used pool. All pools are for all purposes are invisible to an any application, and MinIO handles the pools internally. 

### Rules of Adding pools

There is only one requirement, i.e. based on initial pool's erasure set count (say `n`), new pools are expected to have a minimum of `n` drives to match the original Tenant SLA or it should be in multiples of `n`. For example if initial set count is 4, new pools should have at least 4 or multiple of 4 drives.

### Effects on KES/TLS Enabled Instance

If your MinIO Operator configuration has [KES](https://github.com/minio/operator/blob/master/docs/kes.md) or [Automatic TLS](https://github.com/minio/operator/blob/master/docs/tls.md#automatic-csr-generation) enabled, there are additional considerations:

- When new pools are added, Operator invalidates older self signed TLS certificates and the related secrets. Operator then creates new certificate signing requests (CSR). This is because there are new MinIO nodes that must be added in certificate DNS names. The administrator must approve these CSRs for MinIO server to be deployed again. Unless the CSR are approved, Operator will not create MinIO StatefulSet pods.

- If you're using your own certificates, as explained [here](https://github.com/minio/operator/blob/master/docs/tls.md#pass-certificate-secret-to-tenant), please ensure to use/update proper certificates that allow older and new MinIO nodes.

## Downtime

The Tenant expansion process requires removing the existing StatefulSet and creating a new StatefulSet with the required number of pods. Kubernetes automatically terminates and re-creates pods and PVCs during this process. Since MinIO requires at least (Volumes/2)+1 volumes to support regular read and write operations, the expansion process may result in a period of downtime where MinIO returns errors for read and write operations.
