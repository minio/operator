from minio import Minio
from minio.credentials import IamAwsProvider
from urllib.parse import urlparse
import os
import sys

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
bucket = os.getenv("BUCKET")

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt is "" or sa_jwt is None:
    print("Token is empty")
    sys.exit(1)

stsUrl = urlparse(tenant_endpoint)
stsUrl.path = stsUrl.path + f"/{tenant_namespace}"

provider = IamAwsProvider(stsUrl.geturl())

tenantUrl = urlparse(tenant_endpoint)
client = Minio(f"{tenantUrl.hostname}:{tenantUrl.port}/{tenantUrl.path}", credentials=provider, secure=tenantUrl.scheme == "https")

# Get information of an object.
stat = client.list_objects(bucket)
print(stat)
