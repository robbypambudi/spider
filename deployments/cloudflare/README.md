# SPIDER on Cloudflare

Deploy SPIDER ke Cloudflare dengan arsitektur hybrid edge + container.

## Apa yang cocok di Cloudflare

| Komponen | Cloudflare product | Catatan |
|----------|-------------------|---------|
| React UI | **Pages** | Static build dari `frontend/` |
| Go API | **Containers** + Worker router | Image dari `deployments/docker/Dockerfile.api` |
| Prompt-Shield (ML) | **Containers** (instance `standard`) | RAM besar, cold start 2–5 menit |
| PostgreSQL | **Neon** (+ optional **Hyperdrive**) | Postgres managed; Hyperdrive jika nanti ada Worker TS |
| Controller reconcile | **Cron Trigger** Worker | Panggil endpoint internal API (lihat Phase 3) |
| Worker agent | Di luar CF atau Cron mock | Butuh proses heartbeat kontinu |
| Redis | **KV** atau hapus | Belum dipakai di Go backend |
| Prometheus/Grafana | **Workers Observability** / Logpush | Opsional |

## Arsitektur target

```text
User
  │
  ▼
Cloudflare Pages (spider-ui)
  │  VITE_API_URL
  ▼
Worker "spider-api" ──► Container: spider-api (Go chi, :8000)
  │                           │
  │                           ├── Neon Postgres (DATABASE_URL)
  │                           └── Container: prompt-shield (:8081)
  │
Cron Worker (optional) ──► POST /internal/reconcile
```

## Prerequisites

1. [Cloudflare account](https://dash.cloudflare.com/sign-up/workers-and-pages) — **Workers Paid** direkomendasikan (Containers).
2. [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/install-and-update/): `npm i -g wrangler`
3. Docker Desktop (untuk `wrangler deploy` build image).
4. Database: buat project di [Neon](https://neon.tech) → copy connection string.
5. Hugging Face token (untuk Prompt-Shield): [huggingface.co/settings/tokens](https://huggingface.co/settings/tokens)

## Phase 1 — API container

```powershell
cd deployments/cloudflare/spider-api
npm install
wrangler login

# Secrets (production)
wrangler secret put DATABASE_URL
wrangler secret put SPIDER_JWT_SECRET
wrangler secret put SPIDER_WORKER_TOKEN

# Deploy (build Docker image + Worker router)
wrangler deploy
```

Set variabel container di dashboard **Workers → spider-api → Settings → Variables** atau tambah di `wrangler.jsonc` (`vars` untuk non-secret):

- `SPIDER_CORS_ORIGINS` — URL Pages Anda, contoh: `https://spider.pages.dev`
- `SPIDER_DEFAULT_DETECTOR` — `rule-based` (dev) atau `prompt-shield` (prod)
- `SPIDER_PROMPT_SHIELD_ENDPOINT` — URL Worker prompt-shield setelah Phase 2

**Catatan:** Deploy pertama container bisa 3–10 menit sebelum route stabil.

## Phase 2 — Prompt-Shield container

```powershell
cd deployments/cloudflare/prompt-shield
npm install
wrangler secret put HF_TOKEN
wrangler deploy
```

Update secret/var di spider-api:

```
SPIDER_PROMPT_SHIELD_ENDPOINT=https://spider-prompt-shield.<subdomain>.workers.dev
SPIDER_DEFAULT_DETECTOR=prompt-shield
```

Redeploy spider-api setelah endpoint prompt-shield tersedia.

## Phase 3 — Frontend (Pages)

```powershell
cd frontend
# Build dengan URL API production
$env:VITE_API_URL="https://spider-api.<subdomain>.workers.dev"
npm run build

# Deploy ke Pages (ganti project name)
npx wrangler pages deploy dist --project-name=spider-ui
```

Atau hubungkan repo Git di dashboard **Workers & Pages → Create → Pages**.

Tambahkan file `frontend/public/_redirects` untuk SPA routing (sudah disediakan).

## Phase 4 — Controller (Cron)

Saat ini `spider-controller` adalah proses loop 5 detik. Di Cloudflare, gunakan **Cron Trigger** yang memanggil endpoint reconcile di API (belum ada — ticket SPIDER-CF-003).

Alternatif sementara: jalankan controller lokal/container kecil yang hanya `DATABASE_URL`-nya sama dengan Neon.

## Phase 5 — Custom domain

1. Pages: `app.example.com` → spider-ui project
2. Worker: `api.example.com` → spider-api route
3. Update `SPIDER_CORS_ORIGINS` dan `VITE_API_URL`

## Biaya & limit

- **Containers**: billing vCPU + memory + egress ([pricing](https://developers.cloudflare.com/containers/platform/pricing/))
- **Prompt-Shield**: gunakan instance type `standard` (model Flan-T5 butuh RAM)
- **Neon**: free tier cukup untuk thesis/demo

## Troubleshooting

| Gejala | Penyebab | Solusi |
|--------|----------|--------|
| 502 setelah deploy | Container masih provisioning | Tunggu 5–10 menit, cek `wrangler containers list` |
| Prompt-Shield timeout | Cold start + model load | Naikkan `sleepAfter`, pre-warm dengan cron health check |
| CORS error | Origins tidak match | Set `SPIDER_CORS_ORIGINS` ke URL Pages exact |
| DB connection fail | Neon SSL / IP allow | Gunakan connection string `?sslmode=require` |

## Local dev vs Cloudflare

Docker Compose tetap dipakai untuk development lokal. Cloudflare deploy adalah path production/ demo publik.
