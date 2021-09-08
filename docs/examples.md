# Tenant deployment examples with kustomize

This document explains various yaml files listed in the [examples directory](https://github.com/minio/operator/tree/master/examples/kustomization) used to deploy a Tenant using MinIO Operator.

### Prerequisites

- kustomize/v4.3.0 https://kubectl.docs.kubernetes.io/installation/kustomize/

## MinIO Tenant with AutoCert TLS

MinIO Operator can automatically generate TLS secrets and mount these secrets to the MinIO, Console, and/or KES pods (enabled by default). To disable this, set the `requestAutoCert` field to `false`.

You can deploy the pre-configured example by running the following command:

```$xslt
kustomize examples/kustomization/base | kubectl apply -f -
```

## MinIO Tenant with Encryption enabled using Vault KMS

This example will deploy a MinIO tenant with Server Side Encryption using KES and Hashicorp Vault.

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

- Open `example/kustomization/tenant-kes-encryption/kes-configuration-secret.yaml`
- In the  `Vault` configuration replace `<PUT YOUR VAULT ENDPOINT HERE>`
 for `http://vault.default.svc.cluster.local:8200`, `<PUT YOUR APPROLE ID HERE>` for your `app-role-id`, `<PUT YOUR APPROLE SECRET ID HERE>` for your `app-role-secret-id` and
 `<PUT YOUR KEY PREFIX HERE>` for `my-minio`.

You can deploy a preconfigured example by running the following command:

```$xslt
kustomize build example/kustomization/tenant-kes-encryption | kubectl apply -f -
```

Verify data is encrypted by connecting directly to MinIO via `ingress controller` or using port-forward:

```$xslt
kubectl port-forward svc/minio 9000:443 -n tenant-kms-encrypted
mc alias set alias https://127.0.0.1:9000 minio minio123 --insecure
mc admin kms key status alias --insecure
Key: my-minio-key
   - Encryption ✔
   - Decryption ✔
```

## MinIO Tenant with TLS via customer provided certificates

This example will deploy a MinIO tenant with TLS using certificates provided by the user.

### Prerequisites

- You can generate certificates using `Vault CA`, `Openssl` or `Mkcert`, for this example we will use https://github.com/FiloSottile/mkcert
- Assuming your Tenant name will be `minio` you should generate the following certificate keypairs:

  ```sh
    mkcert "*.minio-tenant.svc.cluster.local"
    mkcert "*.storage.minio-tenant.svc.cluster.local"
    mkcert "*.storage-hl.minio-tenant.svc.cluster.local"
  ```
  
`MinIO` will use `*.minio-tenant.svc.cluster.local`, `*.storage.minio-tenant.svc.cluster.local` and `*.storage-hl.minio-tenant.svc.cluster.local` certificates for
inter-node communication.

Create `kubernetes secrets`  based on the previous certificates

```$xslt
kubectl create secret tls minio-tls-cert --key="_wildcard.minio-tenant.svc.cluster.local-key.pem" --cert="_wildcard.minio-tenant.svc.cluster.local.pem" -n minio-tenant
kubectl create secret tls minio-buckets-cert --key="_wildcard.storage.minio-tenant.svc.cluster.local-key.pem" --cert="_wildcard.storage.minio-tenant.svc.cluster.local.pem" -n minio-tenant
kubectl create secret tls minio-hl-cert --key="_wildcard.storage-hl.minio-tenant.svc.cluster.local-key.pem" --cert="_wildcard.storage-hl.minio-tenant.svc.cluster.local.pem" -n minio-tenant
```

You need to provide those `kubernetes secrets` in your Tenant `YAML` overlay using the `externalCertSecret` fields, ie:

```$xslt
  externalCertSecret:
    - name: minio-tls-cert
      type: kubernetes.io/tls
    - name: minio-buckets-cert
      type: kubernetes.io/tls
    - name: minio-hl-cert
      type: kubernetes.io/tls
```

You can deploy a preconfigured example by running the following command:

```$xslt
kustomize build examples/kustomization/base | kubectl apply -f -
```

## MinIO Tenant with TLS via customer provided certificates and Encryption enabled via Vault KMS

This example will deploy a minio tenant using mTLS certificates (authentication between `MinIO` and `KES`) provided by the user, the data will be encrypted at rest

### Prerequisites

- Configure `Vault` the same way as in the first example
- Set the `app-role-id`, the `app-role-secret-id` and `key-prefix` in your KES configuration `YAML` file
- Assuming your Tenant name is `storage-kms-encrypted` and namespace is `tenant-kms-encrypted` create all the certificates and secrets as in the previous step
- Generate new `KES` identity keypair (https://github.com/minio/kes), this is needed it for the authentication, `mTLS` between `MinIO` and `KES`:

  ```sh
    kes tool identity new --key="./app.key" --cert="app.cert" app
  ```

- Using the generated `app.key` and `app.cert` create a new kubernetes secret: `kubectl create secret tls minio-kes-mtls --key="app.key" --cert="app.cert"` -n tenant-kms-encrypted
  and provide that secret in the `externalClientCertSecret` field of your tenant `YAML` overlay (if the field doesn't exist add it)

  ```$xslt
  spec:
  ...
    externalClientCertSecret:
      name: minio-kes-mtls
      type: kubernetes.io/tls
  ```

- Calculate the `app.cert` identity using `KES`: `kes tool identity of app.cert`, copy the resulting hash and open your
  KES configuration `YAML` (`kes-configuration-secret.yaml`) file and replace `${MINIO_KES_IDENTITY}` for the `bda5d8b6531d2f3bcd64e5ec73841bcb23ecb57b19c5f814e491ea2b2088995c` string, you can
  add additional identities using this array, ie:

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

You can deploy a pre-configured example by running the following command:

```$xslt
kustomize build examples/kustomization/tenant-kes-encryption | kubectl apply -f -
```

### Additional Examples

For additional examples on how to deploy a tenant with [LDAP](https://docs.min.io/minio/baremetal/security/ad-ldap-external-identity-management/configure-ad-ldap-external-identity-management.html) or [OIDC](https://docs.min.io/minio/baremetal/security/openid-external-identity-management/configure-openid-external-identity-management.html) you can look at the [examples directory](https://github.com/minio/operator/tree/master/examples/kustomization)