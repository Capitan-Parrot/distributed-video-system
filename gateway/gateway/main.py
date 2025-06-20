import uvicorn
from fastapi import FastAPI

from gateway.routers import scenario

app = FastAPI(title="Scenario State Machine API")
app.include_router(scenario.router)

if __name__ == "__main__":
    uvicorn.run("main:app", reload=True, host="0.0.0.0", port=8080)
