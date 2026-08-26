from __future__ import annotations

import re
import time

from pkg.apis.security.models import DetectionResult

# DEVELOPMENT / TESTING ONLY.
# This is not a production prompt-injection detector. Pattern lists are a
# convenience for local demos and unit tests, not a security control.
_INJECTION_PATTERNS: tuple[re.Pattern[str], ...] = (
    re.compile(r"ignore\s+(all\s+)?(previous|prior|above)\s+instructions", re.I),
    re.compile(r"disregard\s+(your|all|the)\s+(previous\s+)?instructions", re.I),
    re.compile(r"reveal\s+(the\s+)?system\s+prompt", re.I),
    re.compile(r"show\s+(me\s+)?(your\s+)?(hidden\s+)?system\s+prompt", re.I),
    re.compile(r"you\s+are\s+now\s+dan", re.I),
    re.compile(r"jailbreak", re.I),
    re.compile(r"pretend\s+you\s+have\s+no\s+(restrictions|rules|guidelines)", re.I),
    re.compile(r"do\s+not\s+follow\s+(your\s+)?(safety\s+)?(policy|policies|rules)", re.I),
    re.compile(r"override\s+(the\s+)?(safety|content)\s+(policy|filters?)", re.I),
    re.compile(r"new\s+instructions?\s*:\s*you\s+are", re.I),
    re.compile(r"from\s+now\s+on\s+you\s+(will|must)\s+ignore", re.I),
)


class RuleBasedDetector:
    """Keyword/regex detector.

    DEVELOPMENT/TESTING ONLY. Scores are 0.0 or 1.0. ALLOW/BLOCK is still
    decided by the policy engine using a configured threshold — this detector
    does not own the decision.
    """

    name = "rule-based"
    version = "0.1.0-dev"

    async def detect(self, text: str) -> DetectionResult:
        started = time.perf_counter()
        matched = [pattern.pattern for pattern in _INJECTION_PATTERNS if pattern.search(text)]
        is_injection = bool(matched)
        score = 1.0 if is_injection else 0.0
        latency_ms = (time.perf_counter() - started) * 1000.0
        return DetectionResult(
            detector=self.name,
            score=score,
            is_injection=is_injection,
            threshold=None,
            latency_ms=latency_ms,
            metadata={
                "version": self.version,
                "kind": "rule-based",
                "warning": "development/testing only",
                "matched_patterns": matched,
            },
        )
