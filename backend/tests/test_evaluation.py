from __future__ import annotations

from pkg.security.evaluation import confusion_counts, evaluate_scores, security_overhead, tpr_at_fpr


def test_confusion_and_tpr_at_fpr() -> None:
    y_true = [True, True, False, False, False]
    scores = [0.99, 0.80, 0.10, 0.20, 0.01]
    counts = confusion_counts(y_true, [s >= 0.5 for s in scores])
    assert counts.tp == 2
    assert counts.fp == 0
    assert counts.tn == 3
    assert counts.tpr == 1.0
    assert tpr_at_fpr(y_true, scores, 0.0) == 1.0
    report = evaluate_scores(y_true, scores, threshold=0.5, target_fprs=[0.0005, 0.01])
    assert "0.0005" in report.tpr_at_target_fpr
    assert report.f1 > 0


def test_security_overhead() -> None:
    metrics = security_overhead(baseline_latency_ms=100.0, protected_latency_ms=112.0)
    assert metrics.security_overhead_ms == 12.0
    assert metrics.security_overhead_percent == 12.0
