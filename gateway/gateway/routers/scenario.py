import json
import os

from botocore.exceptions import ClientError
from fastapi import APIRouter, HTTPException, UploadFile, File
from uuid import UUID
import httpx

from gateway.s3 import s3Client


ORCHESTRATOR_URL = os.getenv("ORCHESTRATOR_URL", "http://localhost:8080")
S3_BUCKET = os.getenv("PREDICTIONS_BUCKET", "predictions")

router = APIRouter()
client = httpx.AsyncClient(timeout=10.0)


def parse_httpx_error(e: httpx.HTTPStatusError) -> HTTPException:
    try:
        detail = e.response.json()
    except Exception:
        detail = e.response.text
    return HTTPException(status_code=e.response.status_code, detail=f"Orchestrator error: {detail}")


@router.post("/scenario/")
async def initialize_scenario(video: UploadFile = File(...)):
    try:
        file_bytes = await video.read()
        files = {
            "video": (video.filename, file_bytes, video.content_type)
        }

        resp = await client.post(
            f"{ORCHESTRATOR_URL}/scenario",
            files=files
        )
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        raise parse_httpx_error(e)
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator HTTPError: {e}")
    return resp.json()


@router.get("/scenario/{scenario_id}/")
async def get_scenario_status(scenario_id: UUID):
    try:
        resp = await client.get(f"{ORCHESTRATOR_URL}/scenario/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        raise parse_httpx_error(e)
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator HTTPError: {e}")
    return resp.json()


@router.post("/scenario/{scenario_id}/")
async def change_scenario_status(scenario_id: UUID):
    try:
        resp = await client.post(f"{ORCHESTRATOR_URL}/scenario/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        raise parse_httpx_error(e)
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator HTTPError: {e}")
    return resp.json()


@router.get("/prediction/{scenario_id}/")
async def get_predictions(scenario_id: UUID):
    try:
        folder_prefix = f"{str(scenario_id)}/"
        results = []

        # Получаем список объектов в папке
        objects = s3Client.list_objects_v2(
            Bucket=S3_BUCKET,
            Prefix=folder_prefix,
            Delimiter='/'
        )

        if 'Contents' not in objects:
            raise HTTPException(status_code=404, detail="No predictions found for this scenario")

        # Обрабатываем каждый файл
        for obj in objects['Contents']:
            if obj['Key'].endswith('/'):
                continue

            file_key = obj['Key']
            try:
                response = s3Client.get_object(Bucket=S3_BUCKET, Key=file_key)
                file_content = response['Body'].read().decode('utf-8')
                results.append({
                    "frame": file_key.split('/')[1],
                    "predictions": json.loads(file_content),
                })
            except ClientError:
                continue
            except json.JSONDecodeError:
                continue

        if not results:
            raise HTTPException(status_code=404, detail="No valid prediction files found")

        results.sort(key=lambda name: int(name['frame'].split('.')[0]))

        return {
            "scenario_id": str(scenario_id),
            "results": results
        }

    except ClientError as e:
        if e.response['Error']['Code'] == 'NoSuchBucket':
            raise HTTPException(status_code=404, detail="Predictions bucket not found")
        raise HTTPException(status_code=500, detail=f"S3 error: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")