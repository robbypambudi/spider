# SPIDER Distributed Evaluation

Harness eksperimen untuk skenario **Baseline (single-node) vs Distributed Prompt-Shield** yang didiskusikan sebelumnya: mengukur FP/FN, throughput, latency, dan load fairness (Jain's Fairness Index), lalu membandingkan keduanya secara apples-to-apples.

Ini adalah Go module **terpisah** dari `backend/` (punya `go.mod` sendiri dengan `replace` ke `../../backend`), supaya bisa memakai langsung komponen asli SPIDER (`pkg/security/pipeline`, `pkg/security/detectors`, `pkg/security/evaluation`) tanpa menumpuk dependency baru ke module produksi.

## Kenapa "simulasi", bukan cluster sungguhan

Sebelum baca lebih jauh — ini penting untuk laporan/thesis kalian: **arsitektur chunk-ke-banyak-node yang literal belum ada di SPIDER** (lihat diskusi sebelumnya: `Pipeline.Inspect` di `backend/pkg/security/pipeline/pipeline.go` berjalan sekuensial, satu proses, satu detector). Tool ini **tidak** menunggu fitur itu dibangun — dia mensimulasikan cluster N-node di level benchmark:

- **Baseline** = 1 instance `*pipeline.Pipeline`, diakses oleh beberapa goroutine konkuren (mensimulasikan banyak client memukul satu node — persis perilaku API produksi sekarang).
- **Distributed** = N instance `*pipeline.Pipeline` independen (tiap "node" = goroutine pool + pipeline sendiri, tanpa shared state), dengan dispatcher yang membagi request pakai strategi `least-loaded` (meniru dimensi "running requests" dari `pkg/scheduler.LeastLoadedScheduler`) atau `round-robin`.

**Yang nyata diukur**: throughput, latency, distribusi beban antar node (Jain's index), dan kesetaraan hasil klasifikasi (baseline vs distributed harus identik — ini sanity check bawaan, lihat `compare`).

**Yang TIDAK nyata**: node-node ini bukan proses OS terpisah, apalagi mesin terpisah — semua goroutine dalam satu proses Go. Konsekuensinya:
- **CPU usage per node** dilaporkan sebagai `busy_ms` (total waktu wall-clock yang dihabiskan goroutine node itu memproses request) — ini proxy beban, **bukan** CPU% dari OS. Untuk CPU% asli per node, perlu varian multi-proses (lihat "Going further" di bawah).
- Kalau `--detector prompt-shield`, semua node simulasi memanggil **satu** `--prompt-shield-endpoint` yang sama — jadi ini mengukur overhead dispatch/fairness, **bukan** paralelisasi inferensi ML yang sesungguhnya (itu perlu N sidecar terpisah, belum diimplementasikan di sini).

Kalau targetnya klaim "N mesin fisik berbeda", perlakukan angka dari tool ini sebagai **upper-bound teoritis** dari benefit distribusi (best case, tanpa overhead jaringan sungguhan) — bukan pengganti eksperimen multi-mesin/multi-container yang sesungguhnya.

## Dataset

Ada dua sumber yang didukung, auto-detect dari isi file:

1. **Dataset asli Prompt-Shield** (`spider-internal/labs/PromptShield/{train,test,validation}.json`) — format `{"prompt": "...", "label": 0|1}` dalam satu JSON array, `label=1` berarti injeksi. Ini dataset fine-tuning model Prompt-Shield yang sebenarnya, jadi hasil evaluasinya representatif untuk riset. **Tidak disalin ke repo ini** (repo terpisah, hindari duplikasi data pihak ketiga) — panggil langsung pakai path absolut:

   ```bash
   go run . bench --dataset "/path/ke/spider-internal/labs/PromptShield/test.json" --out results/real_baseline.json --nodes 1
   ```

2. **Dataset sintetis bawaan** (`datasets/injection-bench-1000.jsonl`) — 1000 baris, format JSONL `{"text": "...", "is_injection": true|false}`, di-generate oleh `gendata` dari empat pool kalimat (lihat `gendata.go`): benign polos, benign yang secara literal memicu regex (penghasil FP jujur — mis. kalimat soal jailbreak penjara sungguhan), injection klasik (harus tertangkap), dan injection yang sengaja diformulasikan di luar pola regex (penghasil FN jujur). Berguna untuk uji cepat/CI tanpa dataset besar. Regenerasi/perbesar dengan:

   ```bash
   go run . gendata --out datasets/injection-bench-5000.jsonl --n 5000 --seed 42 --injection-ratio 0.3
   ```

## Cara pakai

### 1. Build (opsional, `go run .` juga bisa langsung)

```bash
cd experiments/distributed-eval
go build -o spider-bench.exe .
```

### 2. Baseline (single-node)

```bash
./spider-bench.exe bench \
  --dataset datasets/injection-bench-1000.jsonl \
  --detector rule-based \
  --nodes 1 \
  --out results/baseline.json
```

### 3. Distributed (N node simulasi)

```bash
./spider-bench.exe bench \
  --dataset datasets/injection-bench-1000.jsonl \
  --detector rule-based \
  --nodes 4 --strategy least-loaded --concurrency-per-node 8 \
  --out results/distributed_4.json
```

### 4. Bandingkan

```bash
./spider-bench.exe compare --baseline results/baseline.json --distributed results/distributed_4.json
```

Output `compare` mencakup: speedup throughput, delta latency p95, **cek kesetaraan FP/FN** (kalau beda padahal detector+threshold sama, itu bug di dispatch/aggregation — bukan trade-off performa), dan Jain's Fairness Index.

### Semua flag `bench`

| Flag | Default | Arti |
| --- | --- | --- |
| `--dataset` | `datasets/injection-bench-1000.jsonl` | JSONL sintetis atau JSON array Prompt-Shield (auto-detect) |
| `--out` | *(wajib)* | path output JSON report |
| `--detector` | `rule-based` | `rule-based` \| `noop` \| `prompt-shield` (butuh sidecar jalan di `--prompt-shield-endpoint`) |
| `--threshold` | `0.5` | ambang keputusan (score >= threshold → diprediksi injection) |
| `--chunker` | `fixed` | `fixed` (karakter) \| `token` (butuh sidecar) |
| `--chunk-size`, `--chunk-overlap` | `2048`, `128` | ukuran chunk |
| `--fail-mode` | `closed` | meniru `SPIDER_FAIL_MODE` |
| `--nodes` | `1` | jumlah node simulasi; `1` = baseline |
| `--strategy` | `least-loaded` | `least-loaded` \| `round-robin` (dipakai kalau `--nodes > 1`) |
| `--concurrency-per-node` | `4` | goroutine konkuren per node (mensimulasikan client concurrent) |
| `--repeat` | `1` | ulangi dataset R kali untuk memperbesar volume beban |
| `--target-fpr` | `0.0005,0.001,0.005,0.01` | titik operasi TPR@FPR yang dilaporkan |

## Contoh hasil nyata

`results/*.json` di-gitignore (output run itu personal/ad-hoc, bukan sesuatu yang perlu di-commit) — angka di bawah ini diambil dari run yang sudah dijalankan langsung selama pengembangan tool ini, supaya jejaknya tetap ada tanpa perlu commit file JSON mentah.

Dijalankan terhadap dataset asli `PromptShield/validation.json` (1000 sampel, `--detector rule-based`):

```
classification:    TP=4 FP=0 TN=497 FN=499
                   TPR=0.0080 FPR=0.0000 Precision=1.0000 F1=0.0158 AUC=0.5104
```

**Temuan penting**: `RuleBasedDetector` nyaris tidak menangkap apa pun dari serangan injeksi *nyata* (TPR 0,8%!) — jauh berbeda dari performanya di dataset sintetis buatan sendiri. Regex-nya ditulis untuk pola instruksional sederhana ("ignore previous instructions", "jailbreak", dst), sementara serangan nyata di dataset Prompt-Shield jauh lebih variatif (roleplay, encoding, multi-instruksi bersarang seperti contoh "PWNED" di dataset). Ini bukan bug — ini **bukti kuantitatif** kenapa `rule-based` cuma untuk dev/test ([SECURITY.md](../../SECURITY.md)) dan kenapa model ML Prompt-Shield diperlukan. Cocok dipakai sebagai motivasi/baseline pembanding di laporan.

Pada dataset penuh `test.json` (23.516 sampel) dengan 8 node simulasi: throughput naik dari ~36k → ~94k req/s (2,6x), FP/FN identik dengan baseline (sanity check lolos), Jain's index ~0,98 (beban tersebar cukup rata).

## Going further (belum diimplementasikan, kandidat lanjutan)

- **CPU% per node yang asli**: perlu proses OS terpisah per node (bukan goroutine), diukur lewat `/proc/[pid]/stat` (Linux) atau `GetProcessTimes` (Windows) — atau pakai `gopsutil` kalau mau cross-platform tanpa nulis sendiri.
- **Paralelisasi ML sungguhan**: dukung satu `--prompt-shield-endpoint` per node (bukan satu untuk semua), supaya distributed run benar-benar menyebar beban inferensi ke sidecar berbeda.
- **True chunk-sharding** (opsi A dari diskusi arsitektur): chunk satu dokumen benar-benar disebar ke proses/mesin berbeda lewat network, bukan disimulasikan in-process.
- **Failure-injection**: matikan satu node di tengah run, ukur dampaknya ke throughput/latency (SPIDER sudah punya `WorkerReconciler` yang relevan sebagai referensi perilaku).
