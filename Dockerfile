FROM alpine:3.7

LABEL maintainer="Minio Inc <dev@minio.io>"

COPY minio-operator /usr/bin/

RUN \
     apk add --no-cache ca-certificates 'curl>7.61.0' && \
     chmod +x /usr/bin/minio-operator

CMD ["minio-operator"]
