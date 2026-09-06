import json
import logging
import re
from typing import Tuple
from groq import Groq
from app.schemas import RuleGeneratorOutput

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """
You are an expert log-parsing rule synthesizer.
Given a raw sample log line, return a JSON object matching Go / Oniguruma named capture group syntax.

You MUST respond strictly with a valid JSON object matching this structure:
{
  "pattern": "^(?<timestamp>\\\\S+ \\\\S+) (?<severity>\\\\w+) (?<service>\\\\S+) (?<host>\\\\S+) (?<message>.*)$",
  "mapping": {
    "timestamp": "timestamp",
    "severity": "severity",
    "service": "service",
    "host": "host",
    "message": "message"
  },
  "transformations": {
    "timestamp": { "type": "datetime", "format": "2006-01-02 15:04:05" },
    "severity": { "type": "uppercase" }
  }
}

CRITICAL RULES:
1. "pattern" MUST be a valid regex using named capture groups: `(?<field_name>regex_pattern)`.
2. Anchor the regex with `^` at the start and `$` at the end.
3. Every named group in `pattern` MUST have an entry in `mapping` mapping the captured group name to its canonical output field name.
4. For `transformations`:
   - If a field is a timestamp/date/time, set `type` to "datetime" and provide a Go layout format string in `format` (e.g. "2006-01-02 15:04:05", "Jan _2 15:04:05", RFC3339).
   - If a field is severity or level, set `type` to "uppercase".
   - If a field is numeric, set `type` to "integer" or "float".
5. Do NOT hardcode dynamic values inside the regex; capture them generically.
6. Return ONLY raw JSON. No markdown backticks (no ```json).
"""

class RuleSynthesizer:
    def __init__(self, groq_client: Groq, model_name: str = "openai/gpt-oss-120b", timeout_sec: float = 5.0):
        self.client = groq_client
        self.model_name = model_name
        self.timeout_sec = timeout_sec

    def generate_rule(self, raw_log: str) -> RuleGeneratorOutput:
        """
        Synthesizes a Go-compatible log parsing rule from a raw log line using Groq.
        """
        response = self.client.chat.completions.create(
            model=self.model_name,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": f"Synthesize a parser rule for this log line:\n\n{raw_log}"},
            ],
            response_format={"type": "json_object"},
            temperature=0.1,
            timeout=self.timeout_sec,
        )

        content = response.choices[0].message.content
        if not content:
            raise ValueError("Groq returned empty response")

        # Parse JSON into Pydantic model
        try:
            data = json.loads(content)
            rule = RuleGeneratorOutput(**data)
        except Exception as err:
            logger.error(f"Failed to parse LLM output into schema: {content}")
            raise ValueError(f"Invalid JSON format returned by LLM: {err}")

        # Local regex validation check
        is_valid, err_msg = self._validate_regex(rule.pattern)
        if not is_valid:
            logger.error(f"Generated regex failed validation: {err_msg}")
            raise ValueError(f"Synthesized invalid regex pattern: {err_msg}")

        return rule

    def _validate_regex(self, pattern: str) -> Tuple[bool, str]:
        try:
            # Convert Go syntax (?<name>...) to Python syntax (?P<name>...) for local check
            py_pattern = re.sub(r"\(\?<([a-zA-Z0-9_]+)>", r"(?P<\1>", pattern)
            re.compile(py_pattern)
            return True, "OK"
        except re.error as e:
            return False, str(e)