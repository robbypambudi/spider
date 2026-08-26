# Scheduler

`LeastLoadedScheduler` selects among `ONLINE`/`BUSY` workers using running requests, optional GPU utilization/VRAM, and model locality (prefer workers that already report the requested model as `READY`).

Prepared: round-robin, GPU-aware, model-locality, fair-share, latency-aware, research hooks.
