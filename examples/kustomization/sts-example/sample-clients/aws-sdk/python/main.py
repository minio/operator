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
import sys
from urllib.parse import urlparse

import boto3

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
bucket = os.getenv("BUCKET")
policy_path = os.getenv("STS_POLICY")

role_arn = "arn:aws:iam::111111111:dummyroot"
role_session_name = "optional-session-name"
os.environ.setdefault('AWS_ROLE_ARN', role_arn)  # In AWS SDK RoleArn parameter is mandatory

policy = None

if policy_path is not None:
    with open(policy_path, "r") as f:
        policy = f.read()

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt == "" or sa_jwt is None:
    print("Token is empty")
    sys.exit(1)

sts_url = urlparse(f"{sts_endpoint}/{tenant_namespace}")

sts = boto3.client('sts', endpoint_url=sts_url.geturl(), verify=False)
assumed_role_object = sts.assume_role_with_web_identity(
    RoleArn=role_arn,
    RoleSessionName=role_session_name,
    Policy=policy,
    DurationSeconds=25536,
    WebIdentityToken=sa_jwt
)

credentials = assumed_role_object['Credentials']
print(credentials)

tenant_url = urlparse(tenant_endpoint)
s3_client = boto3.resource('s3',
                           aws_access_key_id=credentials['AccessKeyId'],
                           aws_secret_access_key=credentials['SecretAccessKey'],
                           aws_session_token=credentials['SessionToken'],
                           endpoint_url=tenant_url.geturl(), verify=False)

my_bucket = s3_client.Bucket(bucket)
for my_bucket_object in my_bucket.objects.all():
    print(my_bucket_object)
