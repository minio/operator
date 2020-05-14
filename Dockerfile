FROM scratch

LABEL maintainer="MinIO Inc <dev@min.io>"

COPY ca-certificates.crt /etc/ssl/certs/
COPY minio-operator /minio-operator

CMD ["/minio-operator"]
