# Operator Console SSO with Open ID

Operator Console support authentication via Kubernetes Service Account Json Web Token (JWT), which is the default
authentication method, as well Open ID. In this guide will be explained how to setup Operator Console to enable 
authentication via OpenID.

At this moment Operator Console only support Authentication with the Authorization Code Flow, see
https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth

Must be considered that only one authentication method: JWT or Open ID can be used at the time.

In this directory is a kustomization file, by executing it will install Operator and apply the basic configurations to
enable OpenID in Operator console. Make sure to Modify the values in the env variables and provide the CA certificate
in the files `console-deployment.yaml` and `console-tls-secret.yaml`

### IDP Server

Specify the OpenID server URL in the Operator Console Deployment, by setting the `CONSOLE_IDP_URL` environment variable.
The value should point to the open id Endpoint configuration, for example:
`https://your-extenal-idp.com/.well-known/openid-configuration`.

Also provide the Certificate Authority (CA) that signed the certificate the IDP server presents, a good way to present 
it is by mounting a secret containing the certificate `ca.crt`, here is an example secret to create:

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

And could mount the secret in the Deployment as follows:

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

Operator console is a standalone application, to identify himself with the OpenID server requires *client credentials*.
The client credentials are set in the Operator Console as environment variables in the deployment, 
`CONSOLE_IDP_CLIENT_ID` (client id) and `CONSOLE_IDP_SECRET` (client secret)

### Scopes:

In Oauth2 scopes defines the specific actions that an application (client) is allowed to perform, if the `Client` has 
assigned scopes to the OpenID server to allow login in Operator Console such scopes need to be set
to Operator Console in the `CONSOLE_IDP_SCOPES` environment variable, which is a comma delimited string, if blank or not
set the default value is `openid,profile,email`.

### Callback URL
Open ID will expect to be presented with a "call back" URL, where OpenID will redirect back once the authentication succeed
this callback URL is set in Operator Console with the `CONSOLE_IDP_CALLBACK` environment variable.

Callback url could also be build using on-the-fly, for that instead of set the `CONSOLE_IDP_CALBACK`, set the 
`CONSOLE_IDP_CALLBACK_DYNAMIC=on` environment variable.

The built URL will look like following: `$protocol://$host/oauth_callback`

The `$protocol` is deduced from whether if the Operator Console is running on TLS or not would be `https` or `http` if not.

The `$host` part will be deduced from the `HOST` header (URL) where the end user is sending the login request to Operator console.
For example, if the login page is being loaded in the browser on `https://operator.mydomain.com/login`,
then the `$host` would be deduced as `operator.mydomain.com`.

Setting the `CONSOLE_IDP_CALLBACK` could be useful when want to specify a custom domain for the Operator Console, Operator
Console is behind a reverse proxy or load balancer and the `HOST` header is not available.
The Operator Console page have a designated page to handle the redirect after the successful login, that is `/oauth_callback`.

Make sure that the `CONSOLE_IDP_CALLBACK` is presented with that path in the Url, ie: `https://minio-operator.mydomain.com/oauth_callback`.


### Token expiration

The default OpenID login token duration is 3600 seconds (1 hour), you can set a longer duration by setting the
`CONSOLE_IDP_TOKEN_EXPIRATION` env variable.
