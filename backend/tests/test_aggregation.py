from __future__ import annotations

from pkg.apis.security.models import DetectionResult
from pkg.security.aggregation.max_score import MaxScoreAggregator


def test_max_score_aggregator() -> None:
    aggregator = MaxScoreAggregator()
    results = [
        DetectionResult(detector="rule-based", score=0.1, is_injection=False, latency_ms=1.0),
        DetectionResult(detector="rule-based", score=0.9, is_injection=True, latency_ms=1.2),
        DetectionResult(detector="rule-based", score=0.2, is_injection=False, latency_ms=0.8),
    ]
    aggregated = aggregator.aggregate(results)
    assert aggregated.score == 0.9
    assert aggregated.is_injection is True
    assert aggregated.chunks_scanned == 3


def test_max_score_aggregator_empty() -> None:
    aggregated = MaxScoreAggregator().aggregate([])
    assert aggregated.score == 0.0
    assert aggregated.chunks_scanned == 0
