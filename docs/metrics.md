# Metrics

`GET /metrics` (Prometheus) and `GET /api/v1/metrics/summary` (dashboard JSON).

Security: `spider_security_scans_total`, `allowed/blocked/review` counters, detection and pipeline latency histograms, chunk counters, detector scores.

Inference: request totals, e2e latency, security overhead.

Cluster: worker/GPU gauges, per-GPU utilization and memory.

Research helpers: Jain's fairness index in `pkg/metrics`.
