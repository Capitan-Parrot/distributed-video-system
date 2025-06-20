import boto3

from gateway.config import S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY

s3Client = boto3.client(
    "s3",
    endpoint_url=S3_ENDPOINT,
    aws_access_key_id=S3_ACCESS_KEY,
    aws_secret_access_key=S3_SECRET_KEY,
)
