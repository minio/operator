# Configuring Sidecars for a Tenant

This document explains how to enable configure sidecars for your MinIO Tenant.

Sidecars are containers that run in the same pod as the MinIO container, this makes it so they run together on the same machine and have the ability to community with each other over `localhost`.

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://github.com/minio/operator#operator-setup).

## Sidecars Configuration

Sidecars Configuration is a part of Tenant yaml. 

The following example configures a warp container to run in the same pod as the MinIO pod.

```yaml
...
  sideCars:
    containers:
      - name: warp
        image: minio/warp:v0.3.21
        args:
          - client
        ports:
          - containerPort: 7761
            name: http
            protocol: TCP
```

**Note:** the MinIO Service for the tenant won't expose the ports added in the sidecar. It's up to the user to expose these ports with their own services.

### Sidecars Fields

| Field                         | Description                                                                                                                                    |
|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| spec.kes                      | Defines the Sidecars configuration.                                                                                                            |
| spec.kes.containers           | List of containers to run on the same pod as the MinIO tenant                                                                                  |
| spec.kes.volumeClaimtemplates | Since the MinIO pod is created by a statefulset if a sidecar needs a dynamically provisioned volume this is where it can insert it's template. |
| spec.kes.volumes              | List of volumes to add to the MinIO pod, useful if you need to mount configmaps or secrets to a sidecar.                                       |
