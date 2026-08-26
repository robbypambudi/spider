from __future__ import annotations

from typing import Protocol

from pkg.apis.security.models import AggregatedDetectionResult, DetectionResult


class DetectionAggregator(Protocol):
    name: str

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult: ...
