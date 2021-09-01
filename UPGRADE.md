Upgrades
====

In this document we will try to document relevant upgrade notes for the MinIO Operator.

v4.2.3 - v4.2.4
---
In this version we started running the MinIO pods as `non-root` to increase security in the MinIO deployment, however this has the implication that older tenants that were not sepcifying a [securityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) on a per-pool basis may suddenly stop starting due to file-ownership problems.

This problem may be identified on the MinIO logs by seeing a log line like:

```
Unable to read 'format.json' from https://production-storage-pool-0-1.production-storage-hl.ns-3.svc.cluster
.local:9000/export3: file access denied      
```

The solution for an existing tenant is to add a `securityContext` to each pool in the Tenant's `.spec.pools[*].securityContext` field with the following imlpicit default:

```
securityContext:
  fsGroup: 0
  runAsGroup: 0
  runAsNonRoot: false
  runAsUser: 0
```

This scenario is automatically handled by the operator, however if the tenant is updated from a pre-stored source (i.e: a yaml file) which is missing the added `securityContext` this problem may arise again, so update your stored yamls respectively. 

