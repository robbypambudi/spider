from __future__ import annotations

from typing import Protocol

from pkg.apis.security.models import DetectionResult


class PromptInjectionDetector(Protocol):
    name: str
    version: str

    async def detect(self, text: str) -> DetectionResult: ...
