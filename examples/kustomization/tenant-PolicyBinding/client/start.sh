#!/bin/sh

cp /var/run/secrets/kubernetes.io/serviceaccount/ca.crt /usr/local/share/ca-certificates/minio-ca.pem
chmod 644 /usr/local/share/ca-certificates/minio-ca.pem
#chmod 644 /usr/local/share/ca-certificates/minio-ca.pem
update-ca-certificates
/app/client