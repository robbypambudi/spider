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

**Git Bash / Linux / macOS:**

```bash
./spider-bench.exe bench \
  --dataset "/c/Users/DS - Research/Documents/spider-internal/labs/PromptShield/validation.json" \
  --out results/promptshield_baseline.json \
  --nodes 1
```

**PowerShell:**

```powershell
.\spider-bench.exe bench `
  --dataset "C:\Users\DS - Research\Documents\spider-internal\labs\PromptShield\validation.json" `
  --out results\promptshield_baseline.json `
  --nodes 1
```

**One line (any shell):**

```bash
./spider-bench.exe bench --dataset "/c/Users/DS - Research/Documents/spider-internal/labs/PromptShield/validation.json" --out results/promptshield_baseline.json --nodes 1
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
| `--node-endpoints` | *(kosong)* | comma-separated `http://host:port` — kalau diisi, request di-dispatch lewat HTTP asli ke instance `detector-node` sungguhan (lihat bagian Docker di bawah), **bukan** simulasi in-process. `--nodes` diabaikan; jumlah node = jumlah endpoint. |
| `--limit` | `0` (semua) | pakai N baris pertama saja dari `--dataset` — buat memotong dataset **asli** yang kebesaran, bukan pengganti data asli |
| `--http-timeout-seconds` | `60` | timeout per request `--node-endpoints`. Terlalu pendek bikin request lambat-tapi-benar dihitung "gagal" — lihat catatan bug di bawah |

`rule-based` dan `fixed` chunker **sengaja ditolak** — gunakan pipeline ML yang sama dengan lab.

## Bug yang sudah diperbaiki: request timeout mencemari angka klasifikasi

**Kejadian**: run pertama `docker_baseline.json` vs `docker_distributed.json` (lihat `results/`) menunjukkan AUC turun dari 0,945 → 0,885 dan TPR 83,4% → 77,6% di sisi distributed — kelihatan seperti "distribusi bikin deteksi lebih buruk". **Itu bukan temuan asli** — itu bug di harness ini.

**Akar masalah**: `remote_runner.go` sebelumnya pakai timeout HTTP hardcode 15 detik, dan request yang gagal/timeout otomatis diberi `score=0` (dianggap "diprediksi ALLOW") alih-alih dikeluarkan dari perhitungan. `max_ms` di kedua file persis mentok di ~15000ms — bukti langsung timeout inilah yang kepotong. Saya buktikan lewat hitung selisih TP/FP/TN/FN antar dua run: **persis 35 sampel** yang tadinya TP di baseline jadi FN di distributed, dan **persis 10 sampel** yang tadinya FP jadi TN — pola yang cuma mungkin terjadi kalau sejumlah skor "dipaksa" jadi 0 oleh timeout, bukan oleh model.

**Perbaikan**:
- Timeout sekarang **configurable** (`--http-timeout-seconds`, default dinaikkan jadi 60 detik — Flan-T5 di container ber-CPU terbatas bisa butuh waktu lebih dari 15 detik, apalagi di bawah beban konkuren).
- Request yang gagal/timeout (HTTP error, atau respons dengan `decision:"ERROR"`) sekarang **dikeluarkan** dari `classification`, dihitung terpisah sebagai `failed_requests` (total dan per-node). `spider-bench bench` akan mencetak peringatan mencolok kalau `failed_requests > 0`.
- Perbaikan yang sama juga diterapkan ke mode in-process (`runner.go`): request dengan `Decision=="ERROR"` juga dikeluarkan, bukan cuma di mode Docker/HTTP.

**Penting**: `results/docker_baseline.json` dan `results/docker_distributed.json` yang sudah ada **direkam sebelum perbaikan ini** — field `failed_requests` di file lama itu tidak ada (bukan nol yang valid, cuma tidak pernah dicatat). Angka AUC/TPR di kedua file itu **tidak bisa dipercaya** dan tidak bisa diperbaiki secara retroaktif — harus **dijalankan ulang** dengan versi kode yang sudah diperbaiki ini untuk dapat angka yang valid.

## Simulasi in-process vs Docker (CPU/memory terkontrol nyata)

Mode default (`--nodes N`, tanpa `--node-endpoints`) tetap simulasi goroutine in-process — cepat untuk iterasi, tapi CPU/memory yang dilaporkan cuma proxy busy-time, bukan angka OS asli (lihat catatan lama di bawah).

Untuk evaluasi yang bisa dipertanggungjawabkan sebagai "distributed" — CPU, memory, dan jaringan **sungguhan terkontrol** — pakai `cmd/detector-node` + `docker-compose.eval.yml`:

- **`cmd/detector-node`**: HTTP server tipis yang membungkus pipeline SPIDER asli (`pkg/security/pipeline`). Endpoint: `POST /detect`, `GET /health`, `GET /stats`.
- **`docker-compose.eval.yml`**: 4 pasang container `shield-N` (sidecar Prompt-Shield asli) + `node-N` (detector-node), masing-masing dengan `cpus`/`mem_limit` eksplisit dan **identik** (default 1 CPU / 1GB per node, 2 CPU / 3GB per sidecar — sesuaikan di file, yang penting semua node-N sama supaya perbandingan adil). "Baseline" = `node-1` sendirian; "distributed" = keempatnya. Detector/chunker **dikunci** ke `prompt-shield` + `token` (256/0) di compose file — tidak ada mode rule-based/sintetis di jalur ini, sesuai metodologi lab.
- **`scripts/run-docker-eval.sh`**: orkestrasi penuh — build+up compose, tunggu health check tiap node, jalankan `spider-bench bench` baseline lalu distributed lewat `--node-endpoints` asli, ambil snapshot `docker stats` (CPU%/MemUsage **nyata**, cgroup-enforced) per container, lalu `compare`.

```bash
cd experiments/distributed-eval
./scripts/run-docker-eval.sh --dataset "/path/ke/spider-internal/labs/PromptShield/test.json"
```

Dataset **wajib** dari `spider-internal/labs` yang sama dipakai untuk testing (`PromptShield/test.json` atau camera-ready benchmark) — bukan `datasets/injection-bench-*.jsonl` sintetis (itu hanya untuk load/smoke, lihat komentar di `dataset.go`). Kalau dataset asli terlalu besar untuk satu run cepat, potong dengan `--limit N` di script (ambil N baris pertama dari data **asli**), bukan ganti ke data buatan sendiri:

```bash
./scripts/run-docker-eval.sh --dataset "/path/ke/PromptShield/test.json" --limit 3000
```

Sudah divalidasi end-to-end (build image, jalankan container dengan `--cpus`/`--memory` sungguhan, `/health` `/detect` `/stats` semua benar, dan dispatch HTTP dari `spider-bench` ke container asli — latency network asli ikut terukur, bukan cuma waktu proses). Yang **belum** saya jalankan penuh dalam sesi ini: pull image Python + download model Flan-T5 untuk `shield-N` (butuh waktu/bandwidth signifikan) — jalankan sendiri saat setup, first build akan makan beberapa menit, run berikutnya cepat karena Docker build cache.

## Simulasi vs cluster nyata (mode in-process, tanpa Docker)

Goroutine in-process, bukan mesin terpisah; satu sidecar endpoint untuk semua node simulasi. Throughput distributed = upper-bound dispatch, bukan klaim multi-GPU ML sungguhan. Mode Docker di atas mengatasi sebagian besar keterbatasan ini (proses OS asli, resource limit asli, network asli) — kecuali paralelisasi ML sungguhan (tiap node Docker sudah punya sidecar sendiri, jadi ini justru **sudah** teratasi juga dibanding mode in-process).

## Going further

- True chunk-sharding satu dokumen ke banyak node lewat network (bukan request independen ke node berbeda)
- Failure-injection: matikan satu container `node-N` di tengah run, ukur dampaknya
