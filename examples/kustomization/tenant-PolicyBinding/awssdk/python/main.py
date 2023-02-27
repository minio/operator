import boto3
import os
import sys
from urllib.parse import urlparse

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
bucket = os.getenv("BUCKET")
policy_path = os.getenv("STS_POLICY")

policy = None

if policy_path is not None:
    with open(policy_path, "r") as f:
        policy = f.read()

stsUrl = urlparse(tenant_endpoint)
stsUrl.path = stsUrl.path + f"/{tenant_namespace}"

sts = boto3.client('sts', endpoint_url=stsUrl.geturl(), verify=False)

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt is "" or sa_jwt is None:
    print("Token is empty")
    sys.exit(1)

assumed_role_object = sts.assume_role_with_web_identity(
    RoleArn='arn:aws:iam::111111111:root', #In AWS SDK RoleArn parameter is mandatory
    RoleSessionName='optional-session-name',
    Policy=policy,
    DurationSeconds=25536,
    WebIdentityToken=sa_jwt
)

credentials = assumed_role_object['Credentials']
print(credentials)

tenantUrl = urlparse(tenant_endpoint)
s3_client = boto3.resource('s3',
    aws_access_key_id=credentials['AccessKeyId'],
    aws_secret_access_key=credentials['SecretAccessKey'],
    aws_session_token=credentials['SessionToken'],
    endpoint_url=tenantUrl.geturl(), verify=False)

my_bucket = s3_client.Bucket(bucket)
for my_bucket_object in my_bucket.objects.all():
    print(my_bucket_object)
