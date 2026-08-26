from pkg.security.detectors.base import PromptInjectionDetector
from pkg.security.detectors.noop import NoOpDetector
from pkg.security.detectors.prepared import (
    EnsembleDetector,
    FlanT5Detector,
    PromptShieldDetector,
    RemoteDetector,
    TransformerDetector,
)
from pkg.security.detectors.rule_based import RuleBasedDetector

DETECTOR_REGISTRY: dict[str, type] = {
    "noop": NoOpDetector,
    "rule-based": RuleBasedDetector,
}

__all__ = [
    "DETECTOR_REGISTRY",
    "EnsembleDetector",
    "FlanT5Detector",
    "NoOpDetector",
    "PromptInjectionDetector",
    "PromptShieldDetector",
    "RemoteDetector",
    "RuleBasedDetector",
    "TransformerDetector",
]
