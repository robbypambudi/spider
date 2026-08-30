# SPIDER Backend (Go)

Runtime defense framework backend written in Go (`cmd/`, `internal/`, `pkg/`, `dao/`).

## Quick start

```bash
cd backend
go mod tidy
go run ./cmd/api
```

## Prompt-Shield detector

Production ML detection uses fine-tuned Flan-T5 models from the
[Prompt-Shield collection](https://huggingface.co/collections/robbypambudi/prompt-shield):

| Model | Hugging Face ID |
| --- | --- |
| Small (default) | `robbypambudi/prompt-shield-flan-t5-small` |
| Base | `robbypambudi/prompt-shield-flan-t5-base` |

Start the inference sidecar:

```bash
cd backend/cmd/prompt-shield
uv sync
SPIDER_PROMPT_SHIELD_MODEL=robbypambudi/prompt-shield-flan-t5-small uv run python main.py
```

Configure the Go API:

```env
SPIDER_DEFAULT_DETECTOR=prompt-shield
SPIDER_PROMPT_SHIELD_ENDPOINT=http://localhost:8081
SPIDER_PROMPT_SHIELD_MODEL=robbypambudi/prompt-shield-flan-t5-small
```

For development without GPU/model download, use `SPIDER_DEFAULT_DETECTOR=rule-based`.

## Binaries

| Command | Role |
| --- | --- |
| `go run ./cmd/api` | HTTP control plane (:8000) |
| `go run ./cmd/controller` | Worker reconciliation loop |
| `go run ./cmd/worker` | Cluster worker agent |
| `go run ./cmd/cli` | CLI client |

## Tests

```bash
go test ./...
```
