import boto3
import os
import sys
import logging
from urllib.parse import urlparse

logging.basicConfig(format='%(message)s', level=logging.DEBUG)
logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

sts_endpoint = os.getenv("STS_ENDPOINT")
tenant_endpoint = os.getenv("MINIO_ENDPOINT")
tenant_namespace = os.getenv("TENANT_NAMESPACE")
token_path = os.getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
bucket = os.getenv("BUCKET")
policy_path = os.getenv("STS_POLICY")

role_arn = "arn:aws:iam::111111111:dummyroot"
role_session_name = "optional-session-name"
os.environ.setdefault('AWS_ROLE_ARN', role_arn) #In AWS SDK RoleArn parameter is mandatory

policy = None

if policy_path is not None:
    with open(policy_path, "r") as f:
        policy = f.read()

with open(token_path, "r") as f:
    sa_jwt = f.read()

if sa_jwt == "" or sa_jwt == None:
    print("Token is empty")
    sys.exit(1)

stsUrl = urlparse(f"{sts_endpoint}/{tenant_namespace}")

sts = boto3.client('sts', endpoint_url=stsUrl.geturl(), verify=False)
assumed_role_object = sts.assume_role_with_web_identity(
    RoleArn=role_arn,
    RoleSessionName=role_session_name,
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
