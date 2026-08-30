# Worker protocol

```text
start → identity → hardware discovery → register → heartbeat → telemetry
```

* `POST /api/v1/workers/register`
* `POST /api/v1/workers/{id}/heartbeat`
* Header: `X-Spider-Worker-Token`

Statuses: `REGISTERING | ONLINE | BUSY | DRAINING | OFFLINE | ERROR`.

The controller's `WorkerReconciler` marks workers `OFFLINE` when `last_heartbeat_at` exceeds `SPIDER_WORKER_OFFLINE_TIMEOUT`.

CPU-only workers are valid. GPU discovery (`pkg/workerctl.DiscoverResources`) is currently CPU/memory only — NVML-based GPU inventory is not yet implemented in the Go worker.

## Joining a cluster

The register → heartbeat loop lives in `pkg/workerctl.Run` and is shared by two entrypoints:

* `cmd/worker` — the standalone worker process (used by `docker compose`, Kubernetes, etc). Identity, resources, and models are fixed at startup by its own `main.go`.
* `cmd/spider` (installed as the `spider` command via `go install ./cmd/spider`) — for joining ad hoc machines to a cluster without deploying the standalone binary:

  ```bash
  spider worker join --api http://localhost:8000 --worker-token <token> [--id id] [--site site] [--heartbeat-interval sec]
  ```

  Identity resolution order for `--id` (also used by `cmd/worker`): the flag/`SPIDER_WORKER_ID` env var, then a `.spider-worker-id` file in the working directory (created on first run), then `<hostname>-<random4hex>`.

### Background mode

`spider worker join --detach` re-execs the CLI as a separate, detached process (own process group; hidden console on Windows) instead of blocking the terminal, and writes:

* stdout/stderr to `~/.spider/worker/<id>.log` (override with `--log-file`)
* the child's PID to `~/.spider/worker/<id>.pid`

Companion commands read that same state directory:

* `spider worker ps` — lists locally-started workers with `running`/`stopped` status.
* `spider worker stop <id>` — kills the process and removes its PID file. Safe to call even if the process already died (e.g. crashed, or killed outside the CLI) — it just clears the stale entry.

This is local process management only, not a cluster "leave" call — there is no unregister endpoint. A stopped or killed worker is detected the same way as a crash: the reconciler marks it `OFFLINE` once `last_heartbeat_at` exceeds `SPIDER_WORKER_OFFLINE_TIMEOUT`. `--detach` also does not supervise or restart the process; for production reliability, run the worker under a real service manager (systemd, a Windows service, a Kubernetes DaemonSet) instead.
