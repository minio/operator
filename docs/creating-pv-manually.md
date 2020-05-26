# Creating PersistentVolumes

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

We recommend using [`local` Persistent Volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local) as the backing storage for MinIO StatefulSets. MinIO
creates a distributed, resilient, failure tolerant storage system on `local` PVs. There is no explicit dependency on external distributed file systems or CSI vendors.

This document explains how to create `local` PVs. Once you create these PVs, MinIOInstance PVCs should automatically bind to these PVs.

## Getting Started

A `local` volume represents a mounted local storage device such as a disk, partition or directory.

Currently, `local` volumes can only be used as a statically created PersistentVolume. Dynamic provisioning is not supported yet. Compared to `hostPath` volumes, `local` volumes can be used in a durable and portable manner without manually scheduling Pods to nodes, as the system is aware of the volume’s node constraints by looking at the node affinity on the PersistentVolume.

Total number of `local` PVs required for a MinIOInstance is equal to Total Servers (across all zones) x Total Volumes Per Server. For example, if you have 4 zones, with 4 servers per zone and each server has 4 volumes, then total PVs required is 64. Since each server has 4 volumes, you'll need to create 4 `local` PVs, on all 16 Servers.

## Sample PV

This is a sample PV configuration. Note that `storageClassName` and `capacity` specified here should match the `storageClassName` and `capacity` specified in MinIOInstance `volumeClaimTemplate` for the PV and PVC to bind. Additionally, please change `capacity`, `path`, and `nodeAffinity` as per your specific requirements.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: minio-pv
spec:
  capacity:
    storage: 1Ti
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: ""
  local:
    path: /mnt/disks/ssd1
  nodeAffinity:
    required:
    nodeSelectorTerms:
    - matchExpressions:
      - key: kubernetes.io/hostname
        operator: In
        values:
        - example-node
```

After configuring the PV details, use `kubectl` to create the requisite number of PVs.
