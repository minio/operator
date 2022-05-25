# General upgrade guide for MinIO Operator: `v3.x.x` to `v4.x.x` [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to upgrade your `MinIO tenants` and the `MinIO Operator` from an old version, such as `v3.x.x`,
to `4.x.x` and newer.

## Getting Started

### Prerequisites

- `kubectl` access to your `k8s` cluster
- MinIO Operator `v3.x.x` running on your cluster

### Prepare existing tenants to be migrated

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

---

# Upgrading from MinIO Operator `v4.2.2` to `v4.2.3` [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to upgrade your MinIO tenants and the MinIO Operator `v4.2.2`, to `v4.2.3` or newer.

## Getting Started

### Prerequisites

- `kubectl` access to your `k8s` cluster
- MinIO Operator `v4.2.2` running on your cluster

### Prepare existing tenants to be migrated

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