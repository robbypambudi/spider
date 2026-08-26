from __future__ import annotations

from pkg.apis.security.models import AggregatedDetectionResult, DetectionResult


class MeanScoreAggregator:
    name = "mean-score"

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult:
        raise NotImplementedError("MeanScoreAggregator is reserved for evaluation. See docs/aggregation.md.")


class TopKAggregator:
    name = "top-k"

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult:
        raise NotImplementedError("TopKAggregator is reserved for evaluation. See docs/aggregation.md.")


class VotingAggregator:
    name = "voting"

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult:
        raise NotImplementedError("VotingAggregator is reserved for evaluation. See docs/aggregation.md.")


class WeightedAggregator:
    name = "weighted"

    def aggregate(self, results: list[DetectionResult]) -> AggregatedDetectionResult:
        raise NotImplementedError(
            "WeightedAggregator is reserved for multi-detector experiments. See docs/aggregation.md."
        )
