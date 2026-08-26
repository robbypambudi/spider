from __future__ import annotations

import time

from pkg.apis.security.models import DetectionResult


class NoOpDetector:
    """Always returns score 0. Used as a baseline / passthrough for experiments."""

    name = "noop"
    version = "0.1.0"

    async def detect(self, text: str) -> DetectionResult:
        started = time.perf_counter()
        _ = text
        latency_ms = (time.perf_counter() - started) * 1000.0
        return DetectionResult(
            detector=self.name,
            score=0.0,
            is_injection=False,
            threshold=None,
            latency_ms=latency_ms,
            metadata={"version": self.version, "kind": "noop"},
        )
