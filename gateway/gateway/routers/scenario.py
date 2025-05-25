import os

from fastapi import APIRouter, HTTPException
from uuid import UUID
import httpx

router = APIRouter()

ORCHESTRATOR_URL = os.getenv("ORCHESTRATOR_URL", "http://localhost:8080")
client = httpx.AsyncClient(timeout=10.0)


@router.post("/scenario/")
async def initialize_scenario():
    try:
        resp = await client.post(f"{ORCHESTRATOR_URL}/scenario")
        resp.raise_for_status()
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    return resp.json()


@router.post("/scenario/{scenario_id}/")
async def change_scenario_status(scenario_id: UUID):
    try:
        resp = await client.post(f"{ORCHESTRATOR_URL}/scenario/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            raise HTTPException(status_code=404, detail="Scenario not found")
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    return resp.json()


@router.get("/scenario/{scenario_id}/")
async def get_scenario_status(scenario_id: UUID):
    try:
        resp = await client.get(f"{ORCHESTRATOR_URL}/scenario/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            raise HTTPException(status_code=404, detail="Scenario not found")
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    return resp.json()


@router.get("/prediction/{scenario_id}/")
async def get_predictions(scenario_id: UUID):
    try:
        resp = await client.get(f"{ORCHESTRATOR_URL}/prediction/{scenario_id}")
        resp.raise_for_status()
    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            raise HTTPException(status_code=404, detail="Scenario not found")
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    except httpx.HTTPError as e:
        raise HTTPException(status_code=500, detail=f"Orchestrator error: {e}")
    return resp.json()
