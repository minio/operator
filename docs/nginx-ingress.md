# Ingress Configuration [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

Ingress exposes HTTP and HTTPS routes from outside the cluster to services within the cluster. Traffic routing is controlled by rules defined on the Ingress resource. This document explains how to enable Ingress for a MinIO Tenant using the [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/).

## Getting Started

### Prerequisites

- MinIO Operator up and running as explained in the [document here](https://docs.min.io/minio/k8s/deployment/deploy-minio-operator.html).
- Nginx Ingress Controller installed and running as explained [here](https://kubernetes.github.io/ingress-nginx/deploy/).
- Network routing rules that enable external client access to Kubernetes worker nodes. For example, this tutorial assumes `minio.example.com` and `console.minio.example.com` as an externally resolvable URL.

### Create MinIO Tenant

Use the `kubectl minio` plugin to create the MinIO tenant if one does not already exist. See [Deploy a MinIO Tenant using the MinIO Plugin](https://docs.min.io/minio/k8s/tenant-management/deploy-minio-tenant.html) for more complete documentation. 

The following example deploys a MinIO Tenant with 4 servers and 16 volumes in total and a total capacity of 16 Terabytes into the `tenant1-ns` namespace using the default Kubernetes storage class. Change these values as appropriate for your requirements.

```sh
kubectl create ns tenant1-ns
kubectl minio tenant create --name tenant1 --servers 4 --volumes 16 --capacity 16Ti --namespace tenant1-ns --storage-class default
```

### TLS Certificate

To enable TLS termination at Ingress, we'll need to either acquire a CA certificate or create a self signed certificate. Either way, after acquiring the certificate, we'll need to create a secret with the certificate as its content. We'll then need to refer this secret from the Ingress rule.

The following example creates a self-signed certificate for `minio.example.com` and then adds it to a Kubernetes secret using the below commands.

- If you want to use a different hostname for your tenants, replace `minio.example.com` with the preferred hostname throughout this procedure.

- If specifying a certificate signed by your preferred CA, perform only the `kubectl create` command, replacing the values for `--key` and `-cert` with your TLS `.key` and `.cert` files respectively.

```sh
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.cert -subj "/CN=minio.example.com/O=minio.example.com"
kubectl create secret tls nginx-tls --key  tls.key --cert tls.cert -n tenant1-ns
```

*Note*: Using self-signed certificates may prevent client applications which require strict TLS validation and trust from connecting to the cluster. You may need to disable TLS validation / verification to allow connections to the Tenant.

### Create Ingress Rule

Use the `kubectl apply -f ingress.yaml -n tenant1-ns` using the example YAML file below to create the Ingress object in the `tenant1-ns` namespace. Once created successfully, you should be able to access the MinIO Tenant from clients outside the Kubernetes cluster using the specified hostname on the domain specified in the rule.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-minio
  namespace: tenant1-ns
  annotations:
    kubernetes.io/ingress.class: "nginx"
    ## Remove if using CA signed certificate
    nginx.ingress.kubernetes.io/proxy-ssl-verify: "off"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
    nginx.ingress.kubernetes.io/server-snippet: |
      client_max_body_size 0;
    nginx.ingress.kubernetes.io/configuration-snippet: |
      chunked_transfer_encoding off;
spec:
  tls:
  - hosts:
      - minio.example.com
    secretName: nginx-tls
  rules:
  - host: minio.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: minio
            port:
              number: 443
```

To enable Ingress route for the Tenant Console, we'll need to create a new Ingress rule. Note that this would require a separate TLS certificate with relevant domain and a secret with this TLS certificate as well (`nginx-tls-console` in below example).

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-console
  namespace: tenant1-ns
  annotations:
    kubernetes.io/ingress.class: "nginx"
    ## Remove if using CA signed certificate
    nginx.ingress.kubernetes.io/proxy-ssl-verify: "off"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
    nginx.ingress.kubernetes.io/server-snippet: |
      client_max_body_size 0;
    nginx.ingress.kubernetes.io/configuration-snippet: |
      chunked_transfer_encoding off;
spec:
  tls:
  - hosts:
      - console.minio.example.com
    secretName: nginx-tls-console
  rules:
  - host: console.minio.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-mgmt-console
            port:
              number: 9443
```
