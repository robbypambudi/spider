#!/usr/bin/env python3
"""Evaluate a labeled JSONL file through the in-process security pipeline."""

from __future__ import annotations

import argparse
import asyncio
import json
from pathlib import Path

from pkg.security.evaluation import evaluate_scores
from pkg.security.pipeline import build_default_pipeline
from pkg.apis.security.models import SecurityRequest


async def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset", default="experiments/datasets/test.jsonl")
    parser.add_argument("--detector", default="rule-based")
    parser.add_argument("--threshold", type=float, default=0.5)
    args = parser.parse_args()

    pipeline = build_default_pipeline(detector_name=args.detector, threshold=args.threshold)
    labels: list[bool] = []
    scores: list[float] = []
    for line in Path(args.dataset).read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        row = json.loads(line)
        result = await pipeline.inspect(SecurityRequest(text=row["text"]))
        labels.append(bool(row["is_injection"]))
        scores.append(result.score)
    report = evaluate_scores(labels, scores, threshold=args.threshold)
    print(report.model_dump_json(indent=2))


if __name__ == "__main__":
    asyncio.run(main())
