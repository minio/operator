# Upgrading from an old version of MinIO Operator [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to upgrade your `MinIO tenants` and the `MinIO Operator` from an old version, such as `v3.x.x`,
to `4.4.13` and newer.

## Getting Started

### Prerequisites

- `kubectl` access to your `k8s` cluster
- MinIO Operator `v3.x.x` running on your cluster

### Prepare existing tenants to be migrated

Before upgrading the `MinIO Operator` you need to make the following changes to all the existing `MinIO Tenants`.

- Make sure every `zone` in `tenant.spec.zones` explicitly set a zone `name` if not configured already.
- Make sure every `zone` in `tenant.spec.zones` explicitly set a `securityContext` if not configured already.

### Example

```yaml
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
existing volumes (this happens if you don't add the right `securityContext`).

### Upgrade MinIO Operator and Upgrade tenants

Once all your tenants are prepared for the upgrade it's time to upgrade Operator:

```bash
kubectl apply -k github.com/minio/operator/\?ref\=v4.4.3
```

The above command will update the MinIO `Tenant CRD`, update the MinIO Operator `image` and trigger the upgrade for each
existing tenant.
