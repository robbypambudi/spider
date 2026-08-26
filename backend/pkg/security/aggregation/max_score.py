from __future__ import annotations

from pkg.apis.security.models import AggregatedDetectionResult, DetectionResult


class MaxScoreAggregator:
    """Research default: a long document is as risky as its highest-scoring chunk."""

    name = "max-score"

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult:
        if not results:
            return AggregatedDetectionResult(
                score=0.0,
                is_injection=False,
                chunks_scanned=0,
                detector_results=[],
                metadata={"strategy": self.name},
            )
        top = max(results, key=lambda item: item.score)
        return AggregatedDetectionResult(
            score=top.score,
            is_injection=any(item.is_injection for item in results),
            chunks_scanned=len(results),
            detector_results=results,
            metadata={"strategy": self.name, "top_detector": top.detector},
        )
