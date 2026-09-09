from typing import Dict, Optional
from pydantic import BaseModel, ConfigDict, Field


class Transformation(BaseModel):
    model_config = ConfigDict(
        extra="forbid",
        populate_by_name=True,
    )

    type: str = Field(
        description="Type of conversion: 'datetime', 'uppercase', 'lowercase', 'integer', 'float', etc."
    )
    format: Optional[str] = Field(
        default=None,
        exclude_none=True,  # Omit field if value is None/null
        description="Format layout string if type is datetime, otherwise omit",
    )


class RuleGeneratorOutput(BaseModel):
    model_config = ConfigDict(extra="forbid")

    name: str = Field(
        description="Concise human-readable name for the parser, describing the log source or format"
    )
    pattern: str = Field(
        description="Named capture group regex matching the whole log string, e.g. ^(?<timestamp>\\S+ \\S+) (?<severity>\\w+) ...$"
    )
    mapping: Dict[str, str] = Field(
        description="Map of regex capture group names to canonical field names"
    )
    transformations: Dict[str, Transformation] = Field(
        default_factory=dict,
        description="Map of field names to their required transformations",
    )


class LogParseRequest(BaseModel):
    log: str = Field(
        ...,
        description="Raw log sample line to generate a parsing rule for",
        examples=["2026-09-05 18:30:21 ERROR payment-service server-01 Payment failed order_id=12345"]
    )