# MinIO tenant with Let's Encrypt [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

This document explains how to deploy a MinIO tenant using certificates generated
by [Let's Encrypt](https://letsencrypt.org/).

## Getting Started

### Prerequisites

- Kubernetes version `+v1.19`. While cert-manager
  supports [earlier K8s versions](https://cert-manager.io/docs/installation/supported-releases/), the MinIO Operator
  requires 1.19 or later.
- MinIO Operator installed
- `kubectl` access to your `k8s` cluster
- [cert-manager](https://cert-manager.io/docs/installation/) 1.7.X or later installed

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.2/cert-manager.yaml
```

- Support for assigning public IPs for `LoadBalancer` type services, if you are deploying `MinIO` on `GKE`, `EKS`, `AKS`
  or any other major public cloud provider this functionality is included out of the box, if you are deploying this on a
  bare metal `kubernetes` cluster you can use [metallb](https://metallb.universe.tf/), ie:

```bash
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-configmap.yaml
```

- [Nginx](https://docs.nginx.com/nginx-ingress-controller/) ingress controller installed

```bash
helm repo add nginx-stable https://helm.nginx.com/stable
helm repo update
helm install nginx-ingress  nginx-stable/nginx-ingress \
    --set rbac.create=true \
    --set controller.service.type=LoadBalancer \
    --set controller.service.externalTrafficPolicy=Local \
    --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-proxy-protocol"="*" \
    --set controller.config.use-proxy-protocol="true"
```

- [kustomize](https://kustomize.io/) installed
- Configure your DNS to route traffic from the MinIO Tenant S3 API hostname (e.g. minio.example.com) and the Tenant
  Console hostname(e.g. console.example.com) to the IP address of the worker node running ingress.

### Deploy tenant

In this example you are going to request a certificate valid for two domains, `minio.example.com`
and `console.example.com`, replace `example.com`
for the actual domain you want to use.

Create a new `ClusterIssuer` that will request a certificate from `Let's Encrypt`:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
  namespace: cert-manager
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: contact@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class:  nginx
EOF
```

Use the example to deploy a `MinIO` tenant, go into base folder of the operator project and run the following command.

```bash
kustomize build examples/kustomization/tenant-letsencrypt | kubectl apply -f -
```

This tenant was deployed without TLS on purpose (`requestAutoCert: false`), however if you look at ingress rule
on `examples/kustomization/tenant-letsencrypt/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tenant-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-body-size: 5t
spec:
  tls:
    - hosts:
        - minio.example.com
        - console.example.com
      secretName: tenant-tls
  rules:
    - host: minio.example.com
      http:
        paths:
          - pathType: Prefix
            path: "/"
            backend:
              service:
                name: minio
                port:
                  number: 80
    - host: console.example.com
      http:
        paths:
          - pathType: Prefix
            path: "/"
            backend:
              service:
                name: myminio-console
                port:
                  number: 9090
```

`cert-manager` will request a certificate for this tenant using `Let's Encrypt` and store the actual public and private
key on the `tenant-tls` secret.

Once all MinIO pods are up and running you can query your endpoints with curl to make sure the communication happens
over `TLS`.

```bash
curl -v https://minio.example.com
curl -v https://console.example.com
```

In this example the `nginx ingress controller` will do the TLS termination.