from __future__ import annotations

import os
import platform
import socket
import uuid
from pathlib import Path

from pydantic import BaseModel


class WorkerIdentity(BaseModel):
    worker_id: str
    hostname: str
    site: str | None = None
    version: str = "0.1.0"


def load_identity(*, site: str | None = None, version: str = "0.1.0") -> WorkerIdentity:
    hostname = socket.gethostname()
    configured = os.environ.get("SPIDER_WORKER_ID")
    if configured:
        worker_id = configured
    else:
        path = Path(os.environ.get("SPIDER_WORKER_ID_FILE", ".spider-worker-id"))
        if path.exists():
            worker_id = path.read_text(encoding="utf-8").strip()
        else:
            worker_id = f"{hostname}-{uuid.uuid4().hex[:8]}"
            try:
                path.write_text(worker_id, encoding="utf-8")
            except OSError:
                pass
    return WorkerIdentity(
        worker_id=worker_id,
        hostname=hostname,
        site=site or os.environ.get("SPIDER_SITE"),
        version=version,
    )


def platform_summary() -> dict[str, str]:
    return {
        "system": platform.system(),
        "release": platform.release(),
        "architecture": platform.machine(),
        "python": platform.python_version(),
    }
