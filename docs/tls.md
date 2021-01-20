# MinIO Operator TLS Configuration [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to enable TLS on MinIO pods.

## Automatic TLS

This approach creates TLS certificates automatically using the Kubernetes cluster root Certificate Authority (CA) to establish trust. In this approach, MinIO Operator creates a private key, and a certificate signing request (CSR) which is submitted via the `certificates.k8s.io` API for signing. Automatic TLS approach creates other certificates required for KES as well as explained in [KES document](./kes.md).

To enable automatic CSR generation on Tenant, set `requestAutoCert` field in the config file to `true`. Optionally you can also pass additional configuration parameters to be used under `certConfig` section. The `certConfig` section currently supports below fields:

- CommonName: By default this is set to a wild card domain name as per [Kubernetes StatefulSet Pod Identity](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-identity). Set it any other value as per your requirements.

- Organization: By default set to `Acme Co`. Change it to the name of your organization.

- DNSNames: By default set to list of all pod DNS names that are part of current Tenant. Any value added under this section will be appended to the list of existing pod DNS names.

Once you enable `requestAutoCert` field and create the Tenant, MinIO Operator creates a CSR for this instance and sends to the Kubernetes API server. MinIO Operator will then approve the CSR. After the CSR is approved and Certificate available, MinIO operator downloads the certificate and then mounts the Private Key and Certificate within the Tenant pod.

## Pass Certificate Secret to Tenant

This approach involves acquiring a CA signed or self-signed certificate and use a Kubernetes Secret resource to store this information. Once you have the key and certificate file available, create a Kubernetes Secret using

```bash
kubectl create secret generic tls-ssl-minio --from-file=path/to/private.key --from-file=path/to/public.crt
```

Once created, set the name of Secret (here it is `tls-ssl-minio`) under `spec.externalCertSecret` field. Then create the Tenant. MinIO Operator will use this Secret to fetch key and certificate and mount it to relevant locations inside the Tenant pods. 

### Using Kubernetes TLS

Alternatively, it's possible to use a TLS secret. First, create the Kubernetes secret:

```bash
kubectl create secret tls tls-ssl-minio --key=private.key --cert=public.crt
```

Once created, set the name of the Secret (in this example `tls-ssl-minio`) under `spec.externalCertSecret[].name`. Also set the type under `spec.externalCertSecret[].type` to `kubernetes.io/tls`:

```yaml
  externalCertSecret:
    - name: tls-ssl-minio
      type: kubernetes.io/tls
```

## Using cert-manager

[Certificate Manager](https://cert-manager.io) is a Kubernetes Operator capable of automatically issuing certificates from multiple Issuers. Integration with MinIO is simple. First, create a new certificate issuer; for this demonstration the issuer certificate will be self-signed:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: selfsigning-issuer
spec:
  selfSigned: {}
```

Now it's possible to issue the MinIO certificate using the above issuer:

```yaml
---
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: tls-minio
spec:
  commonName: minio.example.com
  secretName: tls-minio
  dnsNames:
    # Ingress domain
    - minio.example.com
    # Internal domain
    - minio-hl-svc
    - minio-hl-svc.default.svc.cluster.local
  issuerRef:
    name: selfsigning-issuer
```

Finally configure MinIO to use the newly created TLS certificate:

```yaml
  externalCertSecret:
    - name: tls-minio
      type: cert-manager.io/v1alpha2
```
