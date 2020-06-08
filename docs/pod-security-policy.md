# Using PodSecurityPolicy for MinIO Pods

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to apply `PodSecurityPolicy` to MinIO Pods created by the MinIO Operator. A Pod Security Policy is a cluster-level resource that controls security sensitive aspects of the pod specification. Read more in [Kubernetes PodSecurityPolicy Documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).

## Getting Started

You can create a MinIO cluster with single zone, `zone-0` with 4 drives, with a custom `PodSecurityPolicy` applied to all the MinIO Pods created by the Operator.

```
kubectl create -f https://github.com/minio/minio-operator/tree/master/examples/minioinstance-pod-security-policy.yaml
```

This file creates a custom PodSecurityPolicy with these fields:

```yaml
privileged: false
allowPrivilegeEscalation: false
hostNetwork: true
seLinux:
  rule: RunAsAny
supplementalGroups:
  rule: RunAsAny
runAsUser:
  rule: RunAsAny
fsGroup:
  rule: RunAsAny
volumes:
  - '*'
```

Then it creates a `ClusterRole` attached to the `PodSecurityPolicy`. Finally a `ClusterRoleBinding` bounds the `ClusterRole` to a `ServiceAccount` which is added to all the MinIO Pods created by the MinIO Operator.
