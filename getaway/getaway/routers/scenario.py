from fastapi import APIRouter
from uuid import UUID, uuid4
from typing import Dict

router = APIRouter()

# In-memory store for scenarios (for stub purposes)
fake_db: Dict[UUID, str] = {}


@router.post("/scenario/")
def initialize_scenario():
    scenario_id = uuid4()
    fake_db[scenario_id] = "initialized"
    return {"scenario_id": scenario_id, "status": "initialized"}


@router.post("/scenario/{scenario_id}/")
def change_scenario_status(scenario_id: UUID):
    if scenario_id not in fake_db:
        return {"error": "Scenario not found"}
    fake_db[scenario_id] = "updated"
    return {"scenario_id": scenario_id, "status": "updated"}


@router.get("/scenario/{scenario_id}/")
def get_scenario_status(scenario_id: UUID):
    status = fake_db.get(scenario_id)
    if not status:
        return {"error": "Scenario not found"}
    return {"scenario_id": scenario_id, "status": status}


@router.get("/prediction/{scenario_id}/")
def get_predictions(scenario_id: UUID):
    return {
        "scenario_id": scenario_id,
        "predictions": [
            {"value": 0.8, "label": "success"},
            {"value": 0.2, "label": "failure"},
        ]
    }
