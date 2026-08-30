# SPIDER Distributed Evaluation

Harness eksperimen untuk **Baseline (single-node) vs Distributed Prompt-Shield**, selaras dengan metodologi `spider-internal/labs/`:

- Detector: **`prompt-shield`** (Flan-T5 fine-tuned) — bukan `rule-based`
- Chunking: **token W=256, O=0** via sidecar `/chunk` (wave-1 lab)
- Metrik: FP/FN, TPR@FPR (0.05%, 0.1%, 0.5%, 1%), throughput, latency, Jain's fairness

Module Go terpisah (`replace` → `../../backend`), memakai pipeline SPIDER asli.

## Prasyarat

1. Sidecar jalan: `docker compose up -d prompt-shield`
2. Dataset Prompt-Shield dari `spider-internal/labs/` (path absolut)

## Dataset (wajib `--dataset`)

Auto-detect dari isi file:

| Format | Contoh path | Catatan |
|--------|-------------|---------|
| Camera-ready benchmark | `.../camera_ready_datasets/.../2024-11-28_evaluation_benchmark_en.json` | `instruction` + `input` → `extractPrompt` (sama lab) |
| Prompt-Shield JSON | `.../PromptShield/test.json` | `prompt` + `label` |
| JSONL | `datasets/load-smoke.jsonl` | Hanya smoke/load — **bukan** untuk angka paper |

## Cara pakai

### Build

```powershell
cd experiments\distributed-eval
go build -o spider-bench.exe .
```

### Baseline (single-node)

```powershell
.\spider-bench.exe bench `
  --dataset "C:\Users\DS - Research\Documents\spider-internal\labs\PromptShield\validation.json" `
  --out results\promptshield_baseline.json `
  --nodes 1
```

Flag default sudah lab-aligned (`prompt-shield`, `token`, 256/0). Sidecar harus jalan di `http://localhost:8081`.

### Distributed (N node simulasi)

```powershell
.\spider-bench.exe bench `
  --dataset "C:\Users\DS - Research\Documents\spider-internal\labs\PromptShield\test.json" `
  --nodes 4 --strategy least-loaded --concurrency-per-node 8 `
  --out results\promptshield_distributed_4.json
```

### Bandingkan

```powershell
.\spider-bench.exe compare --baseline results\promptshield_baseline.json --distributed results\promptshield_distributed_4.json
```

### Benchmark penuh (camera-ready, ~23k sampel)

```powershell
.\spider-bench.exe bench `
  --dataset "C:\Users\DS - Research\Documents\spider-internal\labs\PromptShieldCode\camera_ready_datasets\en_dataset_no_dups\2024-11-28_evaluation_benchmark_en.json" `
  --prompt-shield-model robbypambudi/prompt-shield-flan-t5-small `
  --out results\camera_ready_baseline.json `
  --nodes 1
```

Angka TPR@FPR bandingkan dengan tabel paper (FLAN-T5-small: AUC ~0.942, TPR@FPR 1% ~7.56%).

## Flag `bench`

| Flag | Default | Arti |
| --- | --- | --- |
| `--dataset` | *(wajib)* | JSON benchmark Prompt-Shield |
| `--out` | *(wajib)* | path output JSON |
| `--detector` | `prompt-shield` | satu-satunya detector yang didukung |
| `--prompt-shield-endpoint` | `http://localhost:8081` | sidecar URL |
| `--prompt-shield-model` | `robbypambudi/prompt-shield-flan-t5-small` | model HF |
| `--threshold` | `0.5` | untuk confusion @ τ; TPR@FPR dari ROC sweep |
| `--chunker` | `token` | satu-satunya chunker yang didukung |
| `--chunk-size` | `256` | token window (lab) |
| `--chunk-overlap` | `0` | overlap (lab wave-1) |
| `--target-fpr` | `0.0005,0.001,0.005,0.01` | titik TPR@FPR |
| `--nodes` | `1` | 1 = baseline |
| `--strategy` | `least-loaded` | `least-loaded` \| `round-robin` |
| `--concurrency-per-node` | `4` | goroutine per node |
| `--repeat` | `1` | ulangi dataset untuk load test |

`rule-based` dan `fixed` chunker **sengaja ditolak** — gunakan pipeline ML yang sama dengan lab.

## Simulasi vs cluster nyata

Lihat bagian sebelumnya di README asli: goroutine in-process, bukan mesin terpisah; satu sidecar endpoint untuk semua node simulasi. Throughput distributed = upper-bound dispatch, bukan klaim multi-GPU ML.

## Going further

- Multi sidecar endpoint per node
- True chunk-sharding over network
- CPU% per OS process
