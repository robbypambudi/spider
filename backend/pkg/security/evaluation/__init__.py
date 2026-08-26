from __future__ import annotations

from collections.abc import Sequence

from pydantic import BaseModel, Field


class ConfusionCounts(BaseModel):
    tp: int
    fp: int
    tn: int
    fn: int

    @property
    def tpr(self) -> float:
        denom = self.tp + self.fn
        return self.tp / denom if denom else 0.0

    @property
    def fpr(self) -> float:
        denom = self.fp + self.tn
        return self.fp / denom if denom else 0.0

    @property
    def tnr(self) -> float:
        denom = self.tn + self.fp
        return self.tn / denom if denom else 0.0

    @property
    def fnr(self) -> float:
        denom = self.fn + self.tp
        return self.fn / denom if denom else 0.0

    @property
    def precision(self) -> float:
        denom = self.tp + self.fp
        return self.tp / denom if denom else 0.0

    @property
    def recall(self) -> float:
        return self.tpr

    @property
    def f1(self) -> float:
        p, r = self.precision, self.recall
        return 2 * p * r / (p + r) if (p + r) else 0.0


class OverheadMetrics(BaseModel):
    baseline_latency_ms: float
    protected_latency_ms: float
    security_overhead_ms: float
    security_overhead_percent: float


def confusion_counts(y_true: Sequence[bool], y_pred: Sequence[bool]) -> ConfusionCounts:
    if len(y_true) != len(y_pred):
        raise ValueError("y_true and y_pred must have the same length")
    tp = fp = tn = fn = 0
    for truth, pred in zip(y_true, y_pred, strict=True):
        if truth and pred:
            tp += 1
        elif not truth and pred:
            fp += 1
        elif not truth and not pred:
            tn += 1
        else:
            fn += 1
    return ConfusionCounts(tp=tp, fp=fp, tn=tn, fn=fn)


def predict_at_threshold(scores: Sequence[float], threshold: float) -> list[bool]:
    return [score >= threshold for score in scores]


def roc_points(
    y_true: Sequence[bool],
    scores: Sequence[float],
) -> list[tuple[float, float, float]]:
    """Return (threshold, fpr, tpr) points from a score sweep, high threshold first."""
    unique_scores = sorted(set(scores), reverse=True)
    thresholds = [1.0] + unique_scores + [0.0]
    points: list[tuple[float, float, float]] = []
    seen: set[tuple[float, float]] = set()
    for threshold in thresholds:
        counts = confusion_counts(y_true, predict_at_threshold(scores, threshold))
        key = (round(counts.fpr, 10), round(counts.tpr, 10))
        if key in seen:
            continue
        seen.add(key)
        points.append((threshold, counts.fpr, counts.tpr))
    return points


def auc_roc(y_true: Sequence[bool], scores: Sequence[float]) -> float:
    points = roc_points(y_true, scores)
    if len(points) < 2:
        return 0.0
    area = 0.0
    for (_, fpr_a, tpr_a), (_, fpr_b, tpr_b) in zip(points, points[1:], strict=False):
        area += (fpr_b - fpr_a) * (tpr_a + tpr_b) / 2.0
    return abs(area)


def tpr_at_fpr(y_true: Sequence[bool], scores: Sequence[float], target_fpr: float) -> float:
    """Highest TPR achievable with FPR <= target_fpr."""
    if target_fpr < 0:
        raise ValueError("target_fpr must be >= 0")
    best = 0.0
    for _threshold, fpr, tpr in roc_points(y_true, scores):
        if fpr <= target_fpr:
            best = max(best, tpr)
    return best


def tpr_at_target_fprs(
    y_true: Sequence[bool],
    scores: Sequence[float],
    targets: Sequence[float],
) -> dict[str, float]:
    return {f"{target:g}": tpr_at_fpr(y_true, scores, target) for target in targets}


def security_overhead(
    *,
    baseline_latency_ms: float,
    protected_latency_ms: float,
) -> OverheadMetrics:
    overhead_ms = protected_latency_ms - baseline_latency_ms
    percent = (overhead_ms / baseline_latency_ms * 100.0) if baseline_latency_ms else 0.0
    return OverheadMetrics(
        baseline_latency_ms=baseline_latency_ms,
        protected_latency_ms=protected_latency_ms,
        security_overhead_ms=overhead_ms,
        security_overhead_percent=percent,
    )


class EvaluationReport(BaseModel):
    counts: ConfusionCounts
    tpr: float
    fpr: float
    tnr: float
    fnr: float
    precision: float
    recall: float
    f1: float
    auc: float
    tpr_at_target_fpr: dict[str, float] = Field(default_factory=dict)
    threshold: float
    samples: int


def evaluate_scores(
    y_true: Sequence[bool],
    scores: Sequence[float],
    *,
    threshold: float,
    target_fprs: Sequence[float] = (0.0005, 0.001, 0.005, 0.01),
) -> EvaluationReport:
    preds = predict_at_threshold(scores, threshold)
    counts = confusion_counts(y_true, preds)
    return EvaluationReport(
        counts=counts,
        tpr=counts.tpr,
        fpr=counts.fpr,
        tnr=counts.tnr,
        fnr=counts.fnr,
        precision=counts.precision,
        recall=counts.recall,
        f1=counts.f1,
        auc=auc_roc(y_true, scores),
        tpr_at_target_fpr=tpr_at_target_fprs(y_true, scores, target_fprs),
        threshold=threshold,
        samples=len(y_true),
    )
