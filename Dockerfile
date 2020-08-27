FROM busybox:1.32

LABEL maintainer="MinIO Inc <dev@min.io>"

COPY CREDITS /third_party/
COPY ca-certificates.crt /etc/ssl/certs/
COPY minio-operator /minio-operator

CMD ["/minio-operator"]
