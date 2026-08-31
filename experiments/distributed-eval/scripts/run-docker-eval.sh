#!/usr/bin/env bash
# Orchestrates a full Docker-based evaluation run: bring up the controlled-
# resource cluster (docker-compose.eval.yml), wait for health, run baseline
# (node-1 alone) then distributed (all 4 nodes) through spider-bench over
# real HTTP, snapshot real per-container CPU/memory via `docker stats`, and
# print a comparison. See ../README.md "Docker-based evaluation".
#
# --dataset must be a real spider-internal evaluation dataset (Prompt-Shield
# JSON or the camera-ready benchmark JSON) — this harness only supports the
# prompt-shield detector with token chunking (see labdefaults.go); there is
# no rule-based/synthetic fallback here. If the dataset is too large to run
# quickly, use --limit to take its first N rows rather than substituting a
# fake one.
#
# Usage:
#   ./run-docker-eval.sh --dataset /path/to/spider-internal/labs/PromptShield/test.json [--limit 2000] [--down]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$EVAL_DIR/docker-compose.eval.yml"
DATASET=""
LIMIT=0
TEAR_DOWN=false

while [ $# -gt 0 ]; do
  case "$1" in
    --dataset) DATASET="$2"; shift 2 ;;
    --limit) LIMIT="$2"; shift 2 ;;
    --down) TEAR_DOWN=true; shift ;;
    *) echo "unknown argument: $1" >&2; exit 1 ;;
  esac
done

if [ -z "$DATASET" ]; then
  echo "--dataset is required: a real spider-internal evaluation dataset, e.g." >&2
  echo "  --dataset \"/c/Users/DS - Research/Documents/spider-internal/labs/PromptShield/test.json\"" >&2
  exit 1
fi

NODE_PORTS=(9001 9002 9003 9004)

cleanup() {
  if [ "$TEAR_DOWN" = true ]; then
    echo "==> tearing down docker-compose.eval.yml"
    (cd "$EVAL_DIR" && docker compose -f "$COMPOSE_FILE" down)
  else
    echo "==> leaving the cluster running (pass --down to tear it down); stop manually with:"
    echo "    docker compose -f $COMPOSE_FILE down"
  fi
}
trap cleanup EXIT

echo "==> building and starting the cluster (prompt-shield sidecars + detector-node, 4 pairs)"
echo "    first run downloads the Flan-T5 model per shield container — can take several minutes"
(cd "$EVAL_DIR" && docker compose -f "$COMPOSE_FILE" up -d --build)

echo "==> waiting for all nodes to report healthy"
for port in "${NODE_PORTS[@]}"; do
  echo -n "   node on :$port "
  for i in $(seq 1 90); do
    if curl -sf "http://localhost:${port}/health" > /dev/null 2>&1; then
      echo "ok"
      break
    fi
    if [ "$i" -eq 90 ]; then
      echo "TIMED OUT"
      echo "check logs: docker compose -f $COMPOSE_FILE logs"
      exit 1
    fi
    sleep 3
  done
done

echo "==> building spider-bench"
(cd "$EVAL_DIR" && go build -o spider-bench.exe .)

BENCH="$EVAL_DIR/spider-bench.exe"
COMMON_FLAGS=(--dataset "$DATASET" --detector prompt-shield --chunker token --chunk-size 256 --chunk-overlap 0)
if [ "$LIMIT" -gt 0 ]; then
  COMMON_FLAGS+=(--limit "$LIMIT")
fi

echo "==> baseline (node-1 alone)"
(cd "$EVAL_DIR" && "$BENCH" bench "${COMMON_FLAGS[@]}" --node-endpoints "http://localhost:9001" --out results/docker_baseline.json)

echo "==> distributed (all 4 nodes, least-loaded)"
(cd "$EVAL_DIR" && "$BENCH" bench "${COMMON_FLAGS[@]}" --node-endpoints "http://localhost:9001,http://localhost:9002,http://localhost:9003,http://localhost:9004" --strategy least-loaded --out results/docker_distributed.json)

STATS_FILE="$EVAL_DIR/results/docker_stats_$(date +%Y%m%d_%H%M%S).txt"
echo "==> sampling real per-container CPU/memory (docker stats) -> $STATS_FILE"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" \
  shield-1 node-1 shield-2 node-2 shield-3 node-3 shield-4 node-4 | tee "$STATS_FILE"

echo
echo "==> compare"
(cd "$EVAL_DIR" && "$BENCH" compare --baseline results/docker_baseline.json --distributed results/docker_distributed.json)
