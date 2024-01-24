# Operator Console SSO with OpenID

Operator Console supports authentication with a Kubernetes Service Account Json Web Token (JWT) or OpenID. This guide explains how to configure OpenID authentication for Operator Console using the [OpenID Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth).

Note: only one authentication method can be enabled at the same time, either JWT or OpenID.

The `kustomization.yaml` file provided in this directory installs Operator and applies the basic configurations to enable OpenID authentication for Operator Console. Modify its environment variable values as needed for your deployment and provide the CA certificate in `console-deployment.yaml` and `console-tls-secret.yaml`.

```shell
kubectl apply -k examples/kustomization/operator-external-idp-oid/
```

### IDP Server

Specify the OpenID server URL in the Operator Console Deployment by setting the `CONSOLE_IDP_URL` environment variable. This value should point to the appropriate OpenID Endpoint configuration, for example: `https://your-extenal-idp.com/.well-known/openid-configuration`.

Also provide the Certificate Authority (CA) that signed the certificate the IDP server presents. You can do this by mounting a secret containing the certificate `ca.crt`. For example:

For a CA certificate resembling the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: idp-ca-tls
  namespace: minio-operator
type: Opaque
stringData:
  ca.crt: |
    <CA public certificate content in plain text here> 
```

Mount the secret in the Deployment as follows:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: console
  namespace: minio-operator
spec:
  template:
    spec:
      containers:
        - name: console
          volumeMounts:
            - mountPath: /tmp/certs/CAs
              name: idp-certificate
      volumes:
        - name: idp-certificate
          projected:
            sources:
              - secret:
                  items:
                    - key: ca.crt
                      path: idp.crt
                  name: idp-ca-tls
...
```

### Client credentials

Operator Console is a standalone application that identifies itself to the OpenID server using *client credentials*. The client credentials are set in the Operator Console with the following environment variables: 
- `CONSOLE_IDP_CLIENT_ID` (client id)
- `CONSOLE_IDP_SECRET` (client secret)

### Scopes:

In OAuth2, scopes defines the specific actions that an application (client) is allowed to perform. If the `Client` has assigned scopes to the OpenID server to allow login in Operator Console, such scopes need to be set to Operator Console in the `CONSOLE_IDP_SCOPES` environment variable. This value should be a comma delimited string. If no value is provided, the default is `openid,profile,email`.

### Callback URL
OpenID uses a "call back" URL to redirect back to the application once the authentication succeeds. This callback URL is set in Operator Console with the `CONSOLE_IDP_CALLBACK` environment variable.

A Callback URL can also be constructed dynamically. To do this, set `CONSOLE_IDP_CALLBACK_DYNAMIC` environment variable to `on` instead of setting a `CONSOLE_IDP_CALBACK`.

The constructed URL resembles following: `$protocol://$host/oauth_callback`

- `$protocol` is either `https` or `http`, depending on whether the Operator Console has TLS enabled.
- `$host` is determined from the `HOST` header (URL) where the end user is sending the login request to Operator Console. For example, for the login URL `https://operator.mydomain.com/login`, `$host` is `operator.mydomain.com`. 

Setting `CONSOLE_IDP_CALLBACK` can be useful if you need to specify a custom domain for the Operator Console, or if the Operator Console is behind a reverse proxy or load balancer and the `HOST` header is not available.
The page located at `/oauth_callback` handles the redirect after a successful login.

Make sure the `CONSOLE_IDP_CALLBACK` URL contains the correct path, for example `https://minio-operator.mydomain.com/oauth_callback`.

### Token expiration

The default OpenID login token duration is 3600 seconds (1 hour). You can set a longer duration with the
`CONSOLE_IDP_TOKEN_EXPIRATION` environment variable.
