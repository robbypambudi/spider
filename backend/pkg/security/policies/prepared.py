from __future__ import annotations

from pkg.apis.security.models import AggregatedDetectionResult, SecurityDecision


class AdaptiveThresholdPolicy:
    name = "adaptive"
    threshold = 0.0

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision:
        raise NotImplementedError(
            "AdaptiveThresholdPolicy requires online calibration. See docs/policies.md."
        )


class DetectorSpecificPolicy:
    name = "detector-specific"
    threshold = 0.0

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision:
        raise NotImplementedError(
            "DetectorSpecificPolicy requires per-detector threshold maps. See docs/policies.md."
        )


class TenantPolicy:
    name = "tenant"
    threshold = 0.0

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision:
        raise NotImplementedError("TenantPolicy requires tenant configuration. See docs/policies.md.")


class RiskBasedPolicy:
    name = "risk-based"
    threshold = 0.0

    def evaluate(self, result: AggregatedDetectionResult) -> SecurityDecision:
        raise NotImplementedError(
            "RiskBasedPolicy requires risk features beyond score. See docs/policies.md."
        )
