FROM golang as golayer

RUN \
    apt-get update && \
    apt-get install -y curl ca-certificates golang-go

ADD go.mod /go/src/github.com/minio/operator/sts/example/go/go.mod
ADD go.sum /go/src/github.com/minio/operator/sts/example/go/go.sum
WORKDIR /go/src/github.com/minio/operator/sts/example/go/

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/operator/sts/example/go/

ENV CGO_ENABLED=0

RUN go build -o go-example .

FROM scratch

COPY --from=golayer /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=golayer /go/src/github.com/minio/operator/sts/example/go/go-example /usr/local/bin/

ENTRYPOINT ["go-example"]
