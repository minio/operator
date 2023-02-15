FROM registry.access.redhat.com/ubi8/ubi-minimal:8.7

ARG TAG

RUN \
    microdnf update --nodocs && \
    microdnf install curl ca-certificates shadow-utils --nodocs

COPY create-secret /create-secret

CMD ["/create-secret"]
