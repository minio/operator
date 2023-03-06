# -*- coding: utf-8 -*-
# This file is part of MinIO Operator
# Copyright (c) 2023 MinIO, Inc.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

from minio import Minio
from minio.credentials import IamAwsProvider
from urllib.parse import urlparse
import urllib3
import os
import sys
# import logging

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
bucketName = os.getenv("BUCKET")
kubernetes_ca_file = os.getenv("KUBERNETES_CA_PATH")

# logging.basicConfig(format='%(message)s', level=logging.DEBUG)
# logger = logging.getLogger()
# logger.setLevel(logging.DEBUG)

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt == "" or sa_jwt == None:
    print("Token is empty")
    sys.exit(1)

https_transport = urllib3.PoolManager(
    cert_reqs='REQUIRED',
    ca_certs=kubernetes_ca_file,
    retries=urllib3.Retry(
                total=5,
                backoff_factor=0.2,
                status_forcelist=[500, 502, 503, 504],
            )
    )

stsUrl = urlparse(f"{sts_endpoint}/{tenant_namespace}")
provider = IamAwsProvider(stsUrl.geturl(), http_client=https_transport)

credentials = provider.retrieve()

print(f"Access key: {credentials.access_key}")
print(f"Secret key: {credentials.secret_key}")
print(f"Session Token key: {credentials.session_token}")

tenantUrl = urlparse(tenant_endpoint)
isHttps = (tenantUrl.scheme == "https")

client = Minio(
    f"{tenantUrl.hostname}:{tenantUrl.port}/{tenantUrl.path}",
    credentials=provider,
    secure=isHttps,
    http_client=https_transport
    )

# list buckets
print("Listing Buckets:")
buckets = client.list_buckets()
for bucket in buckets:
    print(bucket.name, bucket.creation_date)

# list objects in a bucket
print(f"Listing Objects in bucket {bucketName}:")
objects = client.list_objects(bucketName, recursive=True)
for obj in objects:
    print(obj)
