"""Worker control-plane protocol constants."""

from __future__ import annotations

REGISTER_PATH = "/api/v1/workers/register"
HEARTBEAT_PATH = "/api/v1/workers/{worker_id}/heartbeat"
WORKER_TOKEN_HEADER = "X-Spider-Worker-Token"
