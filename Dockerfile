FROM golang:1.10.1-alpine3.7

LABEL maintainer="Minio Inc <dev@minio.io>"

WORKDIR /go/src/github.com/minio/
COPY . /go/src/github.com/minio/minio-operator

RUN \
     apk add --no-cache ca-certificates 'curl>7.61.0' && \
     cd /go/src/github.com/minio/minio-operator && \
     go install

CMD ["minio-operator"]