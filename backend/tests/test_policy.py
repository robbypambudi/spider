from __future__ import annotations

from pkg.apis.security.models import AggregatedDetectionResult, SecurityDecision
from pkg.security.policies.threshold import ThresholdPolicy


def test_threshold_policy_blocks_at_or_above() -> None:
    policy = ThresholdPolicy(threshold=0.5)
    blocked = policy.evaluate(
        AggregatedDetectionResult(score=0.5, is_injection=True, chunks_scanned=1)
    )
    allowed = policy.evaluate(
        AggregatedDetectionResult(score=0.49, is_injection=False, chunks_scanned=1)
    )
    assert blocked is SecurityDecision.BLOCK
    assert allowed is SecurityDecision.ALLOW


def test_threshold_not_owned_by_detector() -> None:
    policy = ThresholdPolicy(threshold=0.99)
    decision = policy.evaluate(
        AggregatedDetectionResult(score=0.8, is_injection=True, chunks_scanned=1)
    )
    assert decision is SecurityDecision.ALLOW
