from __future__ import annotations

import os
from typing import Any

import psutil

from pkg.apis.worker.models import GPUResource, WorkerResources


def _memory_total_mb() -> int:
    return int(psutil.virtual_memory().total / (1024 * 1024))


def _discover_nvidia_gpus() -> list[GPUResource]:
    try:
        import pynvml
    except Exception:
        return []
    try:
        pynvml.nvmlInit()
    except Exception:
        return []
    gpus: list[GPUResource] = []
    try:
        count = pynvml.nvmlDeviceGetCount()
        for index in range(count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(index)
            name = pynvml.nvmlDeviceGetName(handle)
            if isinstance(name, bytes):
                name = name.decode("utf-8", errors="replace")
            try:
                mem = pynvml.nvmlDeviceGetMemoryInfo(handle)
                total_mb = int(mem.total / (1024 * 1024))
                used_mb = int(mem.used / (1024 * 1024))
            except Exception:
                total_mb = 0
                used_mb = 0
            try:
                util = pynvml.nvmlDeviceGetUtilizationRates(handle)
                utilization = int(util.gpu)
            except Exception:
                utilization = 0
            gpus.append(
                GPUResource(
                    index=index,
                    vendor="nvidia",
                    name=str(name),
                    memory_total_mb=total_mb,
                    memory_used_mb=used_mb,
                    utilization=utilization,
                )
            )
    finally:
        try:
            pynvml.nvmlShutdown()
        except Exception:
            pass
    return gpus


def discover_resources() -> WorkerResources:
    """CPU-only workers are first-class. GPUs are reported when NVML is present."""
    return WorkerResources(
        cpu_total=os.cpu_count() or 1,
        memory_total_mb=_memory_total_mb(),
        gpus=_discover_nvidia_gpus(),
    )


def cuda_driver_info() -> dict[str, Any]:
    info: dict[str, Any] = {"cuda_version": None, "driver_version": None}
    try:
        import pynvml

        pynvml.nvmlInit()
        try:
            info["driver_version"] = pynvml.nvmlSystemGetDriverVersion()
            if isinstance(info["driver_version"], bytes):
                info["driver_version"] = info["driver_version"].decode()
            try:
                info["cuda_version"] = pynvml.nvmlSystemGetCudaDriverVersion_v2()
            except Exception:
                pass
        finally:
            pynvml.nvmlShutdown()
    except Exception:
        return info
    return info
