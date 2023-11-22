FROM registry.access.redhat.com/ubi9/ubi-minimal:latest as build

RUN microdnf update -y --nodocs && microdnf install ca-certificates -y --nodocs

FROM registry.access.redhat.com/ubi9/ubi-micro:latest

ARG TAG

LABEL name="MinIO" \
      vendor="MinIO Inc <dev@min.io>" \
      maintainer="MinIO Inc <dev@min.io>" \
      version="${TAG}" \
      release="${TAG}" \
      summary="MinIO Operator brings native support for MinIO, Console, and Encryption to Kubernetes." \
      description="MinIO object storage is fundamentally different. Designed for performance and the S3 API, it is 100% open-source. MinIO is ideal for large, private cloud environments with stringent security requirements and delivers mission-critical availability across a diverse range of workloads."

# On RHEL the certificate bundle is located at:
# - /etc/pki/tls/certs/ca-bundle.crt (RHEL 6)
# - /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem (RHEL 7)
COPY --from=build /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/pki/ca-trust/extracted/pem/

COPY CREDITS /licenses/CREDITS
COPY LICENSE /licenses/LICENSE

COPY minio-operator /minio-operator

ENTRYPOINT ["/minio-operator"]
