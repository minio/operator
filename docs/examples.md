# Tenant deployment examples

This document explains various yaml files listed in the [examples directory](https://github.com/minio/operator/tree/master/examples) used to deploy a Tenant using MinIO Operator.

## MinIO Tenant with AutoCert TLS

MinIO Operator can automatically generate TLS secrets and mount these secrets to the MinIO, Console, and/or KES pods (if enabled). To enable this, set the `requestAutoCert` field to `true`.

You can deploy the pre-configured example by running the following command:

```$xslt
kubectl apply -f examples/tenant-with-autocert-encryption-disabled.yaml
```

## MinIO Tenant with AutoCert TLS and Encryption enabled with Vault KMS

This example will deploy a MinIO tenant with TLS and Server Side Encryption.

### Prerequisites

- Deploy `Vault` KMS in your cluster: `kubectl apply -f examples/vault/deployment.yaml`
- Expose vault via k8s-portforward: `kubectl port-forward svc/vault 8200` on a terminal
- Obtain the `Vault` Root token from the pod logs: `kubectl logs -l app=vault`
- Set the `Vault` token and address in the client:

  ```sh
  export VAULT_ADDR=http://localhost:8200
  export VAULT_TOKEN=TOKEN
  ```

- Enable role auth: `vault auth enable approle`
- Enable secrets k/v: `vault secrets enable kv`
- Create a new `KES` policy: `vault policy write kes-policy examples/vault/kes-policy.hcl`
- Create a new `KES` role based on the `KES` policy: `vault write auth/approle/role/kes-role token_num_uses=0 secret_id_num_uses=0 period=5m policies=kes-policy`
- Get the `app-role-id` and write it down: `vault read auth/approle/role/kes-role/role-id`
- Get the `app-role-secret-id` and write it down: `vault write -f auth/approle/role/kes-role/secret-id`

### Getting Started

- Open `example/tenant-with-autocert-encryption-enabled.yaml`
- Look for the `minio-autocert-encryption-kes-config` configmap and in the `Vault` configuratiton replace `<PUT YOUR VAULT ENDPOINT HERE>`
 for `http://127.0.0.1:8200`, `<PUT YOUR APPROLE ID HERE>` for your `app-role-id` and `<PUT YOUR APPROLE SECRET ID HERE>` for your `app-role-secret-id` 

You can deploy a preconfigured example by running the following command:

```$xslt
kubectl apply -f examples/tenant-with-autocert-encryption-enabled.yaml
```

Verify data is encrypted by connecting directly to MinIO via `ingress controller` or using port-forward:

```$xslt
kubectl port-forward svc/minio 9000:443
mc config host add miniok8s https://127.0.0.1:9000 --insecure
mc mb miniok8s/my-bucket --insecure
mc encrypt set sse-s3 miniok8s/my-bucket --insecure
./mc cp file miniok8s/my-bucket --insecure
./mc stat miniok8s/my-bucket/file --insecure
```

## MinIO Tenant with TLS via customer provided certificates

This example will deploy a MinIO tenant with TLS using certificates provided by the user.

### Prerequisites

- You can generate certificates using `Vault CA`, `Openssl` or `Mkcert`, for this example we will use https://github.com/FiloSottile/mkcert
- Assuming your Tenant name will be `minio` you should generate the following certificate keypairs:

  ```sh
    mkcert "minio.default.svc.cluster.local"
    mkcert "minio-console.default.svc.cluster.local"
    mkcert "*.minio.default.svc.cluster.local"
    mkcert "*.minio-hl.default.svc.cluster.local"
  ```
  
`MinIO` will use `minio.default.svc.cluster.local`, `*.minio.default.svc.cluster.local` and `*.minio-hl.default.svc.cluster.local` certificates,
`Console` will use `minio-console.default.svc.cluster.local` only.

Create `kubernetes secrets`  based on the previous certificates

```$xslt
kubectl create secret tls minio-tls-cert --key="minio.default.svc.cluster.local-key.pem" --cert="minio.default.svc.cluster.local.pem"
kubectl create secret tls minio-buckets-cert --key="_wildcard.minio.default.svc.cluster.local-key.pem" --cert="_wildcard.minio.default.svc.cluster.local.pem"
kubectl create secret tls minio-hl-cert --key="_wildcard.minio-hl.default.svc.cluster.local-key.pem" --cert="_wildcard.minio-hl.default.svc.cluster.local.pem"
kubectl create secret tls console-tls-cert --key="minio-console.default.svc.cluster.local-key.pem" --cert="minio-console.default.svc.cluster.local.pem"
```

You need to provide those `kubernetes secrets` in your Tenant `YAML` file using the `externalCertSecret` fields, ie:

```$xslt
  externalCertSecret:
    - name: minio-tls-cert
      type: kubernetes.io/tls
    - name: minio-buckets-cert
      type: kubernetes.io/tls
    - name: minio-hl-cert
      type: kubernetes.io/tls
  ...
  console:
    image: minio/console:v0.4.6
    ...
    externalCertSecret:
      name: console-tls-cert
      type: kubernetes.io/tls
```

You can deploy a preconfigured example by running the following command:

```$xslt
kubectl apply -f examples/tenant-with-custom-cert-encryption-disabled.yaml
```

## MinIO Tenant with TLS via customer provided certificates and Encryption enabled via Vault KMS

This example will deploy a minio tenant with TLS using certificates provided by the user, the data will be encrypted at rest

### Prerequisites

- Configure `Vault` the same way as in the first example
- Set the `app-role-id` and the `app-role-secret-id` in your Tenant `YAML` file
- Assuming your Tenant name is `minio` create all the certificates and secrets as in the previous step
- Generate certificate for `KES`:

  ```sh
    mkcert "minio-kes-hl-svc.default.svc.cluster.local"
  ```

- Create secret for `KES` certificate:

  ```sh
    kubectl create secret tls console-tls-cert --key="minio-kes-hl-svc.default.svc.cluster.local-key.pem" --cert="minio-kes-hl-svc.default.svc.cluster.local.pem"
  ```

- Generate new `KES` identity keypair (https://github.com/minio/kes), this is need it for the authentication, `mTLS` between MinIO and `KES`:

  ```sh
    kes tool identity new --key="./app.key" --cert="app.cert" app
  ```

- Using the generated `app.key` and `app.cert` create a new kubernetes secret: `kubectl create secret tls minio-kes-mtls --key="app.key" --cert="app.cert"`
  and provide that secret in the `externalClientCertSecret` field

  ```$xslt
  spec:
  ...
    externalClientCertSecret:
      name: minio-custom-cert-encryption-mtls-app-cert
      type: kubernetes.io/tls
  ```

- Calculate the app.cert identity using `KES`: `kes tool identity of app.cert`, copy the resulting hash and open your
  Tenant `YAML` file and replace it for the `bda5d8b6531d2f3bcd64e5ec73841bcb23ecb57b19c5f814e491ea2b2088995c` string, you can
  add aditional identities using this array, ie:

  ```$xslt
    policy:
      my-policy:
        paths:
        - /v1/key/create/*
        - /v1/key/generate/*
        - /v1/key/decrypt/*
        identities:
        - bda5d8b6531d2f3bcd64e5ec73841bcb23ecb57b19c5f814e491ea2b2088995c
  ```
  
### Getting Started

You can deploy a pre configured example by running the following command:

```$xslt
kubectl apply -f examples/tenant-with-custom-cert-encryption-enabled.yaml
```
