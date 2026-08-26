from __future__ import annotations

from pkg.security.detectors import DETECTOR_REGISTRY


class DetectorCatalogService:
    def list_implemented(self) -> list[str]:
        return sorted(DETECTOR_REGISTRY.keys())
