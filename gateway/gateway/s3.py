import boto3
import os

# Настройки S3
S3_ENDPOINT = os.getenv('S3_ENDPOINT', 'http://minio:9000')
S3_ACCESS_KEY = os.getenv('S3_ACCESS_KEY', 'minio-access-key')
S3_SECRET_KEY = os.getenv('S3_SECRET_KEY', 'minio-secret-key')

# Инициализация S3 клиента
s3Client = boto3.client(
    's3',
    endpoint_url=S3_ENDPOINT,
    aws_access_key_id=S3_ACCESS_KEY,
    aws_secret_access_key=S3_SECRET_KEY
)