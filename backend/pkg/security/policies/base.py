from __future__ import annotations

from typing import Protocol

from pkg.apis.security.models import AggregatedDetectionResult, SecurityDecision


class SecurityPolicy(Protocol):
    name: str
    threshold: float

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision: ...
