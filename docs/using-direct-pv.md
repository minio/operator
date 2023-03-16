# Using Direct-PV Driver

## Install Direct-PV Driver

Follow the instructions to install DirectPV [here](https://github.com/minio/directpv)

### Utilize the CSI with MinIO operator

```yaml
  ## This VolumeClaimTemplate is used across all the volumes provisioned for MinIO cluster.
  ## Please do not change the volumeClaimTemplate field while expanding the cluster, this may
  ## lead to unbound PVCs and missing data
  volumeClaimTemplate:
    metadata:
      name: data
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Ti
      storageClassName: directpv-min-io # This field references the existing StorageClass
```
