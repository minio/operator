FROM alpine:3.10

LABEL maintainer="MinIO Inc <dev@min.io>"

COPY minio-operator /usr/bin/

RUN \
     apk add --no-cache ca-certificates 'curl>7.61.0' && \
     chmod +x /usr/bin/minio-operator

CMD ["minio-operator"]
