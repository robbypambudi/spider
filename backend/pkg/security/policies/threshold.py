from __future__ import annotations

from pkg.apis.security.models import AggregatedDetectionResult, SecurityDecision


class ThresholdPolicy:
    """score >= threshold → BLOCK, otherwise ALLOW.

    Threshold is injected from configuration so detector implementations stay
    threshold-agnostic. This is required for TPR @ target-FPR sweeps.
    """

    name = "threshold"

    def __init__(self, threshold: float, *, action_on_detection: str = "block") -> None:
        if not 0.0 <= threshold <= 1.0:
            raise ValueError("threshold must be between 0.0 and 1.0")
        self.threshold = threshold
        self.action_on_detection = action_on_detection.lower()

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision:
        if result.score >= self.threshold:
            if self.action_on_detection == "review":
                return SecurityDecision.REVIEW
            return SecurityDecision.BLOCK
        return SecurityDecision.ALLOW
