import os
from uuid import UUID

from fastapi import APIRouter, HTTPException, File, UploadFile
import httpx

router = APIRouter()

ORCHESTRATOR_URL = os.getenv("ORCHESTRATOR_URL", "http://localhost:8080")
print("Using orchestrator:", ORCHESTRATOR_URL)

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
        resp = await client.get(f"{ORCHESTRATOR_URL}/prediction/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        raise parse_httpx_error(e)
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator HTTPError: {e}")
    return resp.json()
