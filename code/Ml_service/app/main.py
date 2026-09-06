import logging
from fastapi import FastAPI
from app.routers import parse
from dotenv import load_dotenv

load_dotenv()  # Reads the .env file automatically

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)

app = FastAPI(
    title="Log Parser Rule Synthesizer API",
    description="FastAPI service that generates Go-compatible regex parsing patterns, mappings, and transformations from sample log lines.",
    version="2.0.0",
)

# Register routes
app.include_router(parse.router)

@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "ok", "service": "rule-synthesizer"}