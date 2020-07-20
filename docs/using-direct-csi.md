# Using Direct CSI Driver [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

## Install Direct CSI Driver

### Set the environment variables

```sh
cat << EOF > default.env
DIRECT_CSI_DRIVER_PATHS=/var/lib/direct-csi-driver/data{1...4}
DIRECT_CSI_DRIVER_COMMON_CONTAINER_ROOT=/var/lib/direct-csi-driver
DIRECT_CSI_DRIVER_COMMON_HOST_ROOT=/var/lib/direct-csi-driver
EOF

export $(cat default.env)
```

### Create the namespace for the driver
```
kubectl apply -k github.com/minio/direct-csi-driver
```

### Utilize the CSI with MinIO operator

```yaml
  ## This VolumeClaimTemplate is used across all the volumes provisioned for MinIO cluster.
  ## Please do not change the volumeClaimTemplate field while expanding the cluster, this may
  ## lead to unbound PVCs and missing data
  volumeClaimTemplate:
    metadata:
      name: direct-csi-driver-min-io-volume
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Ti
      storageClassName: direct.csi.driver.min.io # This field references the existing StorageClass
```
