from __future__ import annotations

from prometheus_client import (
    CONTENT_TYPE_LATEST,
    CollectorRegistry,
    Counter,
    Gauge,
    Histogram,
    generate_latest,
)


class SpiderMetrics:
    """Prometheus metrics for security runtime and serving evaluation."""

    def __init__(self, registry: CollectorRegistry | None = None) -> None:
        self.registry = registry or CollectorRegistry()

        self.security_scans_total = Counter(
            "spider_security_scans_total",
            "Total security scans",
            ["decision"],
            registry=self.registry,
        )
        self.security_allowed_total = Counter(
            "spider_security_allowed_total",
            "Requests allowed by the security pipeline",
            registry=self.registry,
        )
        self.security_blocked_total = Counter(
            "spider_security_blocked_total",
            "Requests blocked by the security pipeline",
            registry=self.registry,
        )
        self.security_review_total = Counter(
            "spider_security_review_total",
            "Requests marked for review",
            registry=self.registry,
        )
        self.security_detection_latency_seconds = Histogram(
            "spider_security_detection_latency_seconds",
            "Detector latency in seconds",
            ["detector"],
            registry=self.registry,
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5),
        )
        self.security_pipeline_latency_seconds = Histogram(
            "spider_security_pipeline_latency_seconds",
            "End-to-end security pipeline latency in seconds",
            registry=self.registry,
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5),
        )
        self.security_chunks_total = Counter(
            "spider_security_chunks_total",
            "Chunks scanned by the security pipeline",
            registry=self.registry,
        )
        self.security_detector_score = Histogram(
            "spider_security_detector_score",
            "Detector score distribution",
            ["detector"],
            registry=self.registry,
            buckets=(0.0, 0.1, 0.25, 0.5, 0.75, 0.9, 0.95, 0.99, 0.999, 1.0),
        )
        self.inference_requests_total = Counter(
            "spider_inference_requests_total",
            "Inference requests",
            ["status"],
            registry=self.registry,
        )
        self.inference_latency_seconds = Histogram(
            "spider_inference_latency_seconds",
            "End-to-end inference latency in seconds",
            registry=self.registry,
            buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
        )
        self.inference_security_overhead_seconds = Histogram(
            "spider_inference_security_overhead_seconds",
            "Security pipeline overhead on inference path in seconds",
            registry=self.registry,
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0),
        )
        self.workers_total = Gauge(
            "spider_workers_total",
            "Registered workers",
            registry=self.registry,
        )
        self.workers_online = Gauge(
            "spider_workers_online",
            "Workers in ONLINE status",
            registry=self.registry,
        )
        self.gpus_total = Gauge(
            "spider_gpus_total",
            "GPUs reported by workers",
            registry=self.registry,
        )
        self.gpus_available = Gauge(
            "spider_gpus_available",
            "GPUs currently considered available",
            registry=self.registry,
        )
        self.worker_gpu_utilization = Gauge(
            "spider_worker_gpu_utilization",
            "GPU utilization percent",
            ["worker_id", "gpu_index"],
            registry=self.registry,
        )
        self.worker_gpu_memory_used_bytes = Gauge(
            "spider_worker_gpu_memory_used_bytes",
            "GPU memory used in bytes",
            ["worker_id", "gpu_index"],
            registry=self.registry,
        )

    def observe_scan(
        self,
        *,
        decision: str,
        pipeline_latency_ms: float,
        chunks: int,
        detector_name: str,
        detector_score: float,
        detector_latency_ms: float,
    ) -> None:
        self.security_scans_total.labels(decision=decision).inc()
        if decision == "ALLOW":
            self.security_allowed_total.inc()
        elif decision == "BLOCK":
            self.security_blocked_total.inc()
        elif decision == "REVIEW":
            self.security_review_total.inc()
        self.security_pipeline_latency_seconds.observe(pipeline_latency_ms / 1000.0)
        self.security_chunks_total.inc(chunks)
        self.security_detector_score.labels(detector=detector_name).observe(detector_score)
        self.security_detection_latency_seconds.labels(detector=detector_name).observe(
            detector_latency_ms / 1000.0
        )

    def observe_inference(
        self,
        *,
        status: str,
        e2e_latency_ms: float,
        security_overhead_ms: float,
    ) -> None:
        self.inference_requests_total.labels(status=status).inc()
        self.inference_latency_seconds.observe(e2e_latency_ms / 1000.0)
        self.inference_security_overhead_seconds.observe(security_overhead_ms / 1000.0)

    def render(self) -> tuple[bytes, str]:
        return generate_latest(self.registry), CONTENT_TYPE_LATEST


_metrics: SpiderMetrics | None = None


def get_metrics() -> SpiderMetrics:
    global _metrics
    if _metrics is None:
        _metrics = SpiderMetrics()
    return _metrics


def reset_metrics() -> SpiderMetrics:
    global _metrics
    _metrics = SpiderMetrics()
    return _metrics
