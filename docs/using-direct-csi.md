# Using Direct CSI Driver [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) 

## Install Direct CSI Driver

[here](https://github.com/minio/direct-csi#quickstart)

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
      storageClassName: direct.csi.min.io # This field references the existing StorageClass
```
