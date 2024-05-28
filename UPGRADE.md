Upgrades
====

In this document we will try to document relevant upgrade notes for the MinIO Operator.

v5.0.0
---

Automatic Tenant migrations only start from tenants previously migrated to `v4.2.0` or newer, users coming from older
version are recommended to upgrade to `v4.5.8` before upgrading to `v5.0.0`.

The `Operator UI` is now bundled on the same container as operator.

The `.spec.S3` field was removed in favor of `.spec.features`.

Field `.spec.credsSecret` was removed in favor of `.spec.configuration`, this secret should hold all the environment
variables for the MinIO deployment that contain sensitive information and should not be shown on `.spec.env`.

Both `Log Search API` (`.spec.log`) and `Prometheus` (`.spec.prometheus`) deployments were removed, however they will be
left running as stand-alone
deployments/statefulset with no connection to the Tenant CR itself, this means that if the Tenant CR is deleted, this
will not cascade to these deployments.

> ⚠️ It is recommended to create a yaml file to manage these deployments subsequently.

To back up these deployments:

```shell
export TENANT_NAME=myminio
export NAMESPACE=mynamespace
kubectl -n $NAMESPACE get secret $TENANT_NAME-log-secret -o yaml > $TENANT_NAME-log-secret.yaml
kubectl -n $NAMESPACE get cm $TENANT_NAME-prometheus-config-map -o yaml > $TENANT_NAME-prometheus-config-map.yaml
kubectl -n $NAMESPACE get sts $TENANT_NAME-prometheus -o yaml > $TENANT_NAME-prometheus.yaml
kubectl -n $NAMESPACE get sts $TENANT_NAME-log -o yaml > $TENANT_NAME-log.yaml
kubectl -n $NAMESPACE get deployment $TENANT_NAME-log-search-api -o yaml > $TENANT_NAME-log-search-api.yaml
kubectl -n $NAMESPACE get svc $TENANT_NAME-log-hl-svc -o yaml > $TENANT_NAME-log-hl-svc.yaml
kubectl -n $NAMESPACE get svc $TENANT_NAME-log-search-api -o yaml > $TENANT_NAME-log-search-api-svc.yaml
kubectl -n $NAMESPACE get svc $TENANT_NAME-prometheus-hl-svc -o yaml > $TENANT_NAME-prometheus-hl-svc.yaml
```

After exporting these objects, remove `.metadata.ownerReferences` for all these files.

After upgrading, to have the MinIO Tenant keep using these services, just add the following environment variables to `.spec.env`

```yaml
- name: MINIO_LOG_QUERY_AUTH_TOKEN
  valueFrom:
    secretKeyRef:
      key: MINIO_LOG_QUERY_AUTH_TOKEN
      name: <TENANT_NAME>-log-secret
- name: MINIO_LOG_QUERY_URL
  value: http://<TENANT_NAME>-log-search-api:8080
- name: MINIO_PROMETHEUS_JOB_ID
  value: minio-job
- name: MINIO_PROMETHEUS_URL
  value: http://<TENANT_NAME>-prometheus-hl-svc:9090
```


v4.4.5
---

The Operator and Logsearch API container have been merged, no new `minio/logsearchapi` images will be built, if your
tenant has a specific MinIO Image specified in `.spec.log.image` you need to update it to use either the upstream `
minio/operator image or your private registry image.


v4.3.9 - v4.4.0
---
Support for Prometheus ServiceMonitor is removed. Using ServiceMonitor to configure prometheus endpoints will lead to
duplicate metrics. The alternate approach is to use
Prometheus [AdditionalScrapeConfigs](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/additional-scrape-config.md).
This can be enabled by setting `prometheusOperator: true` on the tenant.
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

v4.2.2 to v4.2.3
---

Before upgrading the `MinIO Operator` you need to make the following changes to all the existing `MinIO Tenants`.

- Update your current `MinIO image` to the latest version in the tenant spec.
- Make sure every `pool` in `tenant.spec.pools` explicitly set a `securityContext` if not configured already, if this is
  the first time you are configuring a `securityContext` then your `MinIO` pods are running as root, and you need to use:

```yaml
      securityContext:
        runAsUser: 0
        runAsGroup: 0
        runAsNonRoot: false
        fsGroup: 0
```

### Example

```yaml
  image: "minio/minio:$(LATEST-VERSION)"
  ...
  pools:
    - servers: 4
      name: "pool-0"
      volumesPerServer: 4
      volumeClaimTemplate:
        metadata:
          name: data
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Ti
      securityContext:
        runAsUser: 0
        runAsGroup: 0
        runAsNonRoot: false
        fsGroup: 0
```

You can make all the changes directly via `kubectl edit tenants $(TENANT-NAME) -n $(NAMESPACE)` or edit your
`tenant.yaml` and apply the changes: `kubectl apply -f tenant.yaml`.

Failing to apply this changes will cause some issues during the upgrade such as `MinIO` pods not able to `read/write` on
existing volumes (this happens if you don't add the right `securityContext`) or they will take too long to start.

### Upgrade MinIO Operator and Upgrade tenants

Once all your tenants are prepared for the upgrade it's time to upgrade Operator:

```bash
kubectl apply -k github.com/minio/operator/\?ref\=v4.2.3
```

The above command will update the MinIO `Tenant CRD`, update the MinIO Operator `image` and trigger the upgrade for each
existing tenant.


v3.x.x to v4.x.x
---

Before upgrading the `MinIO Operator` you need to make the following changes to all the existing `MinIO Tenants`.

- Update your current `MinIO image` to the latest version in the tenant spec.
- Make sure every `zone` in `tenant.spec.zones` explicitly set a zone `name` if not configured already.
- Make sure every `zone` in `tenant.spec.zones` explicitly set a `securityContext` if not configured already.

### Example

```yaml
  image: "minio/minio:$(LATEST-VERSION)"
  ...
  zones:
    - servers: 4
      name: "zone-0"
      volumesPerServer: 4
      volumeClaimTemplate:
        metadata:
          name: data
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Ti
      securityContext:
        runAsUser: 0
        runAsGroup: 0
        runAsNonRoot: false
        fsGroup: 0
  - servers: 4
      name: "zone-1"
      volumesPerServer: 4
      volumeClaimTemplate:
        metadata:
          name: data
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Ti
      securityContext:
        runAsUser: 0
        runAsGroup: 0
        runAsNonRoot: false
        fsGroup: 0
```

You can make all the changes directly via `kubectl edit tenants $(TENANT-NAME) -n $(NAMESPACE)` or edit your
`tenant.yaml` and apply the changes: `kubectl apply -f tenant.yaml`.

Failing to apply this changes will cause some issues during the upgrade such as the tenants not able to provision because
of wrong `persistent volume claims` (this happens if you don't add the zone name) or MinIO not able to `read/write` on
existing volumes (this happens if you don't add the right `securityContext`) or they will take too long to start.

### Upgrade MinIO Operator and Upgrade tenants

Once all your tenants are prepared for the upgrade it's time to upgrade Operator:

```bash
kubectl apply -k github.com/minio/operator/\?ref\=v4.4.18
```

The above command will update the MinIO `Tenant CRD`, update the MinIO Operator `image` and trigger the upgrade for each
existing tenant.

---

# Upgrade MinIO Operator via Helm Charts

Make sure your current version of the `tenants.minio.min.io` `CRD` includes the necessary `labels` and `annotations` for `Helm`
to perform the upgrade:

```bash
kubectl label crd tenants.minio.min.io app.kubernetes.io/managed-by=Helm --overwrite
kubectl annotate crd tenants.minio.min.io meta.helm.sh/release-name=minio-operator meta.helm.sh/release-namespace=minio-operator --overwrite
```

Run the `helm upgrade` command:

```bash
helm upgrade -n minio-operator [RELEASE] [CHART] [flags]
```

or

```bash
helm upgrade -n minio-operator minio-operator [RELEASE-FOLDER]
```

