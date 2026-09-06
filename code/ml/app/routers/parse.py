import os
from fastapi import APIRouter, HTTPException, Depends, status
from groq import Groq
from app.schemas import LogParseRequest, RuleGeneratorOutput
from app.models.rule_synthesizer import RuleSynthesizer

router = APIRouter(prefix="/api/v1", tags=["Parser Rule Generation"])

def get_groq_client() -> Groq:
    api_key = os.environ.get("GROQ_API_KEY")
    if not api_key:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="GROQ_API_KEY environment variable is missing"
        )
    return Groq(api_key=api_key)

@router.post(
    "/generate-parser-rule",
    response_model=RuleGeneratorOutput,
    status_code=status.HTTP_200_OK,
    summary="Synthesize a Go-compatible log parsing rule using Groq Qwen"
)
async def generate_parser_rule(
    payload: LogParseRequest,
    client: Groq = Depends(get_groq_client)
):
    raw_log = payload.log.strip()
    if not raw_log:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Log string cannot be empty"
        )

    try:
        synthesizer = RuleSynthesizer(groq_client=client)
        rule = synthesizer.generate_rule(raw_log)
        return rule
    except ValueError as ve:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Rule synthesis failed validation: {str(ve)}"
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Internal error during rule generation: {str(e)}"
        )