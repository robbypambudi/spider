#!/usr/bin/env python3
"""Deprecated: use experiments/distributed-eval (spider-bench) instead.

This script targeted an old Python backend. Classification eval must run through
the Go harness with --detector prompt-shield and token chunking (lab defaults).
"""

from __future__ import annotations

import sys

if __name__ == "__main__":
    print(
        "experiments/scripts/evaluate_dataset.py is deprecated.\n"
        "Use: cd experiments/distributed-eval && go run . bench --dataset <PromptShield JSON> --out results/out.json",
        file=sys.stderr,
    )
    sys.exit(1)
