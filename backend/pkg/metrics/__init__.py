"""Cluster/research metric helpers. Prometheus series live in internal.telemetry.metrics."""

from __future__ import annotations


def jains_fairness_index(loads: list[float]) -> float:
    """Jain's fairness index over worker load samples. 1.0 is perfectly fair."""
    if not loads:
        return 1.0
    total = sum(loads)
    squared = sum(value * value for value in loads)
    if squared == 0:
        return 1.0
    return (total * total) / (len(loads) * squared)
