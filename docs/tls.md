# MinIO Operator TLS Configuration

[![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/k8s-operator.svg?maxAge=604800)](https://hub.docker.com/r/minio/k8s-operator)

This document explains how to enable TLS on MinIOInstance pods. There are two approaches to enable TLS via MinIO Operator. First approach is using Kubernetes cluster root Certificate Authority (CA) to establish trust. In this approach MinIO Operator creates a private key and a certificate signing request (CSR) and submits it to `certificates.k8s.io` API for signing. Approving the CSR is either done by an automated approval process or on a one off basis by a Kubernetes cluster administrator.

Second approach is to acquire a self-signed or CA signed certificate and use Kubernetes Secret to store this information. Once the Secret is created, pass the Secret name to MinIO Operator, which then uses the Secret to extract certificate data and mount it at relevant locations within MinIOInstance pods.

## Automatic CSR Generation

To enable automatic CSR generation on MinIOInstance, set `requestAutoCert` field in the config file to `true`. Optionally you can also pass additional configuration parameters to be used under `certConfig` section. The `certConfig` section currently supports below fields:

- CommonName: By default this is set to a wild card domain name as per [Kubernetes StatefulSet Pod Identity](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-identity). Set it any other value as per your requirements.

- Organization: By default set to `Acme Co`. Change it to the name of your organization.

- DNSNames: By default set to list of all pod DNS names that are part of current MinIOInstance cluster. Any value added under this section will be appended to the list of existing pod DNS names.

Once you enable `requestAutoCert` field and create the MinIOInstance, MinIO Operator creates a CSR for this instance and sends to the Kubernetes API server. MinIO Operator will then wait for the CSR to be approved (wait timeout is 20 minutes). CSR can be approved manually via `kubectl` using below steps

- Get the CSR

```
kubectl get csr
```

- Approve the CSR

```
kubectl certificate approve <CSR_Name>
```

Once the CSR is approved and Certificate available, MinIO operator downloads the certificate and then mounts the Private Key and Certificate within the MinIOInstance pod.

## Pass Certificate Secret to MinIOInstance

Follow this approach if you plan to use CA signed or self-signed certificates for MinIOInstance pods. Once you have the key and certificate file available, create a Kubernetes Secret using

```
kubectl create secret generic tls-ssl-minio --from-file=path/to/private.key --from-file=path/to/public.crt
```

Once created, set the name of Secret (here it is `tls-ssl-minio`) under `spec.externalCertSecret` field. Then create the MinIOInstance. MinIO Operator will use this Secret to fetch key and certificate and mount it to relevant locations inside the MinIOInstance pods.
