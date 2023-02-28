# Passing custom Certs/CAs to Operator

To configure MinIO Operator to trust custom certificates, create a secret with the certificate.

```shell
kubectl create secret generic my-custom-tls --from-file=path/to/public.crt
```

then add the following volume to the `minio-operator` deployment under .spec.template.spec

```yaml
      volumes:
        - name: tls-certificates
          projected:
            defaultMode: 420
            sources:
              - secret:
                  items:
                    - key: public.crt
                      path: CAs/custom-public.crt
                  name: my-custom-tls
```

and for the `.spec.temaplte.spec.container[0]`

```yaml
        volumeMounts:
          - mountPath: /tmp/certs
            name: tls-certificates
```