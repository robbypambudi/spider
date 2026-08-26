# Worker protocol

```text
start → identity → hardware discovery → register → heartbeat → telemetry
```

* `POST /api/v1/workers/register`
* `POST /api/v1/workers/{id}/heartbeat`
* Header: `X-Spider-Worker-Token`

Statuses: `REGISTERING | ONLINE | BUSY | DRAINING | OFFLINE | ERROR`.

The controller's `WorkerReconciler` marks workers `OFFLINE` when `last_heartbeat_at` exceeds `SPIDER_WORKER_OFFLINE_TIMEOUT`.

CPU-only workers are valid. NVML (`nvidia-ml-py`) is used when present.
