from __future__ import annotations

from pkg.apis.security.models import DetectionResult


class PromptShieldDetector:
    """Placeholder for Azure AI Prompt Shield. Do not fake model scores."""

    name = "prompt-shield"
    version = "unconfigured"

    async def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError(
            "PromptShieldDetector requires Azure Prompt Shield credentials and API contract. "
            "See docs/detectors.md."
        )


class FlanT5Detector:
    """Placeholder for a Flan-T5 classifier detector."""

    name = "flan-t5"
    version = "unconfigured"

    async def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError(
            "FlanT5Detector requires a model artifact and inference interface. See docs/detectors.md."
        )


class TransformerDetector:
    """Placeholder for a local transformer classifier."""

    name = "transformer"
    version = "unconfigured"

    async def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError(
            "TransformerDetector requires a classifier checkpoint. See docs/detectors.md."
        )


class RemoteDetector:
    """Placeholder for an HTTP detector sidecar."""

    name = "remote"
    version = "unconfigured"

    async def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError(
            "RemoteDetector requires a remote scoring endpoint. See docs/detectors.md."
        )


class EnsembleDetector:
    """Placeholder for combining multiple production detectors."""

    name = "ensemble"
    version = "unconfigured"

    async def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError(
            "EnsembleDetector is reserved for multi-detector experiments. See docs/detectors.md."
        )
