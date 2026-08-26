from __future__ import annotations

import time
from typing import Any

from internal.exceptions import SecurityPipelineError

from pkg.apis.security.models import SecurityDecision, SecurityRequest, SecurityResult
from pkg.security.aggregation.base import DetectionAggregator
from pkg.security.chunking.base import Chunker
from pkg.security.detectors.base import PromptInjectionDetector
from pkg.security.enforcement import Enforcer
from pkg.security.policies.base import SecurityPolicy
from pkg.security.preprocessing import DefaultPreprocessor, Preprocessor


class SecurityPipeline:
    """Inspect a prompt before it is allowed to reach an LLM provider.

    Flow: preprocess → chunk → detect → aggregate → policy → result.
    Enforcement (forward/reject) is applied by callers using Enforcer so the
    pipeline remains a pure inspection stage.
    """

    def __init__(
        self,
        *,
        preprocessor: Preprocessor,
        chunker: Chunker,
        detector: PromptInjectionDetector,
        aggregator: DetectionAggregator,
        policy: SecurityPolicy,
        enforcer: Enforcer | None = None,
    ) -> None:
        self.preprocessor = preprocessor
        self.chunker = chunker
        self.detector = detector
        self.aggregator = aggregator
        self.policy = policy
        self.enforcer = enforcer or Enforcer()

    async def inspect(self, request: SecurityRequest) -> SecurityResult:
        started = time.perf_counter()
        try:
            processed = self.preprocessor.process(request.text)
            chunks = self.chunker.chunk(processed.text)
            detector_results = [await self.detector.detect(chunk.text) for chunk in chunks]
            aggregated = self.aggregator.aggregate(detector_results)
            decision = self.policy.evaluate(aggregated)
        except Exception as exc:  # noqa: BLE001 - convert to ERROR decision at the edge
            latency_ms = (time.perf_counter() - started) * 1000.0
            if isinstance(exc, SecurityPipelineError):
                raise
            return SecurityResult(
                request_id=request.request_id,
                decision=SecurityDecision.ERROR,
                score=0.0,
                detector_results=[],
                chunks_scanned=0,
                total_latency_ms=latency_ms,
                policy=getattr(self.policy, "name", "unknown"),
                metadata={"error": str(exc)},
            )

        latency_ms = (time.perf_counter() - started) * 1000.0
        threshold = getattr(self.policy, "threshold", None)
        metadata: dict[str, Any] = {
            "detector": getattr(self.detector, "name", "unknown"),
            "detector_version": getattr(self.detector, "version", "unknown"),
            "chunker": getattr(self.chunker, "name", "unknown"),
            "aggregator": getattr(self.aggregator, "name", "unknown"),
            "threshold": threshold,
            "source": request.source,
            "model": request.model,
            "original_length": processed.original_length,
            "preprocessed_length": len(processed.text),
        }
        metadata.update(aggregated.metadata)
        return SecurityResult(
            request_id=request.request_id,
            decision=decision,
            score=aggregated.score,
            detector_results=aggregated.detector_results,
            chunks_scanned=aggregated.chunks_scanned,
            total_latency_ms=latency_ms,
            policy=getattr(self.policy, "name", "unknown"),
            metadata=metadata,
        )


def build_default_pipeline(
    *,
    detector_name: str = "rule-based",
    threshold: float = 0.5,
    chunk_size: int = 2048,
    chunk_overlap: int = 128,
    fail_mode: str = "closed",
) -> SecurityPipeline:
    from pkg.security.aggregation.max_score import MaxScoreAggregator
    from pkg.security.chunking.fixed import FixedSizeChunker
    from pkg.security.detectors import DETECTOR_REGISTRY
    from pkg.security.policies.threshold import ThresholdPolicy

    detector_cls = DETECTOR_REGISTRY.get(detector_name)
    if detector_cls is None:
        supported = ", ".join(sorted(DETECTOR_REGISTRY))
        raise SecurityPipelineError(
            f"Unknown detector '{detector_name}'. Implemented detectors: {supported}"
        )
    return SecurityPipeline(
        preprocessor=DefaultPreprocessor(),
        chunker=FixedSizeChunker(size=chunk_size, overlap=chunk_overlap),
        detector=detector_cls(),
        aggregator=MaxScoreAggregator(),
        policy=ThresholdPolicy(threshold=threshold),
        enforcer=Enforcer(fail_mode=fail_mode),
    )
