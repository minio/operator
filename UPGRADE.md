Upgrades
====

In this document we will try to document relevant upgrade notes for the MinIO Operator.

v4.4.5
---

The Operator and Logsearch API container have been merged, no new `minio/logsearchapi` images will be built, if your
tenant has a specific MinIO Image specified in `.spec.log.image` you need to update it to use either the upstream `
minio/operator image or your private registry image.


v4.3.9 - v4.4.0
---
Support for Prometheus ServiceMonitor is removed. Using ServiceMonitor to configure prometheus endpoints will lead to duplicate metrics. The alternate approach is to use Prometheus [AdditionalScrapeConfigs] https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/additional-scrape-config.md. This can be enabled by setting `prometheusOperator: true` on the tenant.
Once this is configured, MinIO Operator will create the additional configuration for the tenant.
If the prometheus is running on a particular namespace, `PROMETHEUS_NAMESPACE` can be set accordingly.

v4.2.3 - v4.2.4
---
In this version we started running the MinIO pods as `non-root` to increase security in the MinIO deployment, however
this has the implication that older tenants that were not sepcifying
a [securityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) on a per-pool basis may
suddenly stop starting due to file-ownership problems.

This problem may be identified on the MinIO logs by seeing a log line like:

```
Unable to read 'format.json' from https://production-storage-pool-0-1.production-storage-hl.ns-3.svc.cluster
.local:9000/export3: file access denied      
```

The solution for an existing tenant is to add a `securityContext` to each pool in the
Tenant's `.spec.pools[*].securityContext` field with the following imlpicit default:

```
securityContext:
  fsGroup: 0
  runAsGroup: 0
  runAsNonRoot: false
  runAsUser: 0
```

This scenario is automatically handled by the operator, however if the tenant is updated from a pre-stored source (i.e:
a yaml file) which is missing the added `securityContext` this problem may arise again, so update your stored yamls
respectively. 


