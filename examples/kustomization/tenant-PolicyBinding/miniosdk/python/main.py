from minio import Minio
from minio.credentials import AssumeRoleProvider
import os
from urllib.parse import urlparse

# STS endpoint usually point to MinIO server.
sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
bucket = os.getenv("BUCKET")
# Policy if available.
policy_path = os.getenv("STS_POLICY")
policy = ""

if policy_path is not None:
    f = open("demofile.txt", "r")
    policy =f.read()

provider = AssumeRoleProvider(
    sts_endpoint,
    policy=policy
)
tenantUrl = urlparse(tenant_endpoint)
client = Minio(f"{tenantUrl.hostname}:{tenantUrl.port}", credentials=provider, secure=tenantUrl.scheme == "https")

# Get information of an object.
stat = client.list_objects(bucket)
print(stat)
