# Using Direct CSI Driver [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

## Install Direct CSI Driver

### Set the environment variables

```sh
cat << EOF > default.env
DIRECT_CSI_DRIVES=data{1...4}
DIRECT_CSI_DRIVES_DIR=/mnt
KUBELET_DIR_PATH=/var/lib/kubelet
EOF

export $(cat default.env)
```

If you are using microk8s `KUBELET_DIR_PATH` should be changed to `/var/snap/microk8s/common/var/lib/kubelet`

### Create the namespace for the driver

```
kubectl apply -k github.com/minio/direct-csi
```

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
