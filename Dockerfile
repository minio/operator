FROM golang:1.14.0

ADD go.mod /go/src/github.com/minio/minio-operator/go.mod
ADD go.sum /go/src/github.com/minio/minio-operator/go.sum
WORKDIR /go/src/github.com/minio/minio-operator/

# Get Certificates
RUN apt-get update -y && apt-get install -y ca-certificates

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download


ADD . /go/src/github.com/minio/minio-operator/
WORKDIR /go/src/github.com/minio/minio-operator/

ENV CGO_ENABLED=0

ARG ldflags
ENV env_ldflags=$ldflags

RUN go build -ldflags "$env_ldflags" -a -o minio-operator .

FROM scratch

LABEL maintainer="MinIO Inc <dev@min.io>"

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/minio/minio-operator/minio-operator .

CMD ["/minio-operator"]
