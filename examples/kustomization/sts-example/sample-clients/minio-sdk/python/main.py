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
import os
import ssl
import sys
from urllib.parse import urlparse

import certifi
import urllib3
from minio import Minio
from minio.credentials import IamAwsProvider

# import logging

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
tenant_region = os.getenv("MINIO_REGION")
bucket_name = os.getenv("BUCKET")
kubernetes_ca_file = os.getenv("KUBERNETES_CA_PATH")

# logging.basicConfig(format='%(message)s', level=logging.DEBUG)
# logger = logging.getLogger()
# logger.setLevel(logging.DEBUG)

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt == "" or sa_jwt is None:
    print("Token is empty")
    sys.exit(1)

# Load Kubernetes custom CA
ca_file = certifi.where()
try:
    with open(kubernetes_ca_file, 'rb') as infile:
        custom_ca = infile.read()

    # Append kubernetes custom CA
    with open(ca_file, 'ab') as outfile:
        outfile.write(custom_ca)
except Exception as e:
    print(e)

# Create a custom SSL context
custom_ssl_context = ssl.create_default_context(cafile=ca_file)

https_transport = urllib3.PoolManager(ssl_context=custom_ssl_context)

sts_url = urlparse(f"{sts_endpoint}/{tenant_namespace}")
provider = IamAwsProvider(sts_url.geturl(), http_client=https_transport)

credentials = provider.retrieve()

print(f"Access key: {credentials.access_key}")
print(f"Secret key: {credentials.secret_key}")
print(f"Session Token key: {credentials.session_token}")

tenant_url = urlparse(tenant_endpoint)
is_https = (tenant_url.scheme == "https")

client = Minio(
    f"{tenant_url.hostname}:{tenant_url.port}/{tenant_url.path}",
    credentials=provider,
    secure=is_https,
    http_client=https_transport,
    region=tenant_region
)

# list buckets
print("Listing Buckets:")
buckets = client.list_buckets()
for bucket in buckets:
    print(bucket.name, bucket.creation_date)

# list objects in a bucket
print(f"Listing Objects in bucket {bucket_name}:")
objects = client.list_objects(bucket_name, recursive=True)
for obj in objects:
    print(obj)
