import os

ORCHESTRATOR_URL = os.getenv("ORCHESTRATOR_URL", "http://localhost:8080")

S3_BUCKET = os.getenv("PREDICTIONS_BUCKET", "predictions")
S3_ENDPOINT = os.getenv("S3_ENDPOINT", "http://localhost:9000")
S3_ACCESS_KEY = os.getenv("S3_ACCESS_KEY", "minio-access-key")
S3_SECRET_KEY = os.getenv("S3_SECRET_KEY", "minio-secret-key")
