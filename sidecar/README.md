# MinIO Operator Sidecar

This sidecar container is used to initialize the MinIO Tenants. It is responsible for retrieving and validating the
configuration for each tenant and creating the necessary resources locally in the pod. 