// Command detector-node is a thin HTTP wrapper around SPIDER's real security
// pipeline (pkg/security/pipeline), meant to run as one container in
// docker-compose.eval.yml. Unlike spider-bench's in-process simulation, a
// detector-node instance is a genuinely separate OS process — when run in
// Docker with --cpus/--memory (or the compose `cpus:`/`mem_limit:` keys),
// its CPU and memory usage are real, cgroup-enforced, and independently
// observable via `docker stats`.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/security/detectors"
	"github.com/spider/spider/pkg/security/pipeline"
)

// Defaults mirror experiments/distributed-eval/labdefaults.go (spider-internal/labs
// wave-1 token-chunking methodology) — keep these two in sync deliberately;
// detector-node is a separate `main` package so it can't import the root
// module's unexported constants directly.
const (
	defaultDetector             = "prompt-shield"
	defaultChunker              = "token"
	defaultChunkSize            = 256
	defaultChunkOverlap         = 0
	defaultPromptShieldEndpoint = "http://localhost:8081"
)

type detectRequest struct {
	Text string `json:"text"`
}

type detectResponse struct {
	Score       float64 `json:"score"`
	IsInjection bool    `json:"is_injection"`
	Decision    string  `json:"decision"`
	LatencyMs   float64 `json:"latency_ms"`
}

type statsResponse struct {
	NodeID   string  `json:"node_id"`
	Requests int64   `json:"requests"`
	BusyMs   float64 `json:"busy_ms"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	nodeID := getEnv("NODE_ID", "node")
	port := getEnv("PORT", "9000")
	promptShieldEndpoint := getEnv("PROMPT_SHIELD_ENDPOINT", defaultPromptShieldEndpoint)
	promptShieldModel := getEnv("PROMPT_SHIELD_MODEL", config.PromptShieldSmall)

	// Registers the "prompt-shield" detector factory into detectors.Registry —
	// required before BuildDefault can look it up (only rule-based/noop are
	// registered by default). Harmless to call even if DETECTOR != prompt-shield.
	detectors.RegisterPromptShield(promptShieldEndpoint, promptShieldModel)

	p, err := pipeline.BuildDefault(pipeline.BuildOptions{
		DetectorName:         getEnv("DETECTOR", defaultDetector),
		Threshold:            getEnvFloat("THRESHOLD", 0.5),
		Chunker:              getEnv("CHUNKER", defaultChunker),
		ChunkSize:            getEnvInt("CHUNK_SIZE", defaultChunkSize),
		ChunkOverlap:         getEnvInt("CHUNK_OVERLAP", defaultChunkOverlap),
		FailMode:             getEnv("FAIL_MODE", "closed"),
		PromptShieldEndpoint: promptShieldEndpoint,
	})
	if err != nil {
		slog.Error("build pipeline", "error", err)
		os.Exit(1)
	}

	var requests int64
	var busyMu sync.Mutex
	var busyMs float64

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "node_id": nodeID})
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		busyMu.Lock()
		b := busyMs
		busyMu.Unlock()
		writeJSON(w, http.StatusOK, statsResponse{
			NodeID:   nodeID,
			Requests: atomic.LoadInt64(&requests),
			BusyMs:   b,
		})
	})
	mux.HandleFunc("/detect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body detectRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		start := time.Now()
		res := p.Inspect(r.Context(), apis.SecurityRequest{RequestID: "detector-node", Text: body.Text})
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0

		atomic.AddInt64(&requests, 1)
		busyMu.Lock()
		busyMs += elapsed
		busyMu.Unlock()

		writeJSON(w, http.StatusOK, detectResponse{
			Score:       res.Score,
			IsInjection: res.Score >= p.Policy.Threshold(),
			Decision:    string(res.Decision),
			LatencyMs:   elapsed,
		})
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	slog.Info("detector_node_started", "node_id", nodeID, "port", port, "detector", p.Detector.Name())
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
