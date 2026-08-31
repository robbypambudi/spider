package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spider/spider/pkg/security/evaluation"
)

// remoteNodeState mirrors nodeState but dispatches to a real, separately
// running detector-node instance over HTTP instead of an in-process
// pipeline. Used when --node-endpoints is set: each endpoint is a genuinely
// separate OS process (typically a Docker container with its own --cpus/
// --memory limits), so latency/throughput/CPU here reflect real network and
// resource-isolation overhead instead of goroutine simulation.
type remoteNodeState struct {
	id       int
	endpoint string
	client   *http.Client
	jobs     chan job
	inflight int64

	mu             sync.Mutex
	requests       int
	failedRequests int
	busyMs         float64
}

type remoteDetectResponse struct {
	Score       float64 `json:"score"`
	IsInjection bool    `json:"is_injection"`
	Decision    string  `json:"decision"`
	LatencyMs   float64 `json:"latency_ms"`
}

func selectRemoteNode(nodes []*remoteNodeState, strategy string, idx int) *remoteNodeState {
	if len(nodes) == 1 {
		return nodes[0]
	}
	if strategy == "round-robin" {
		return nodes[idx%len(nodes)]
	}
	best := nodes[0]
	bestLoad := atomic.LoadInt64(&best.inflight)
	for _, n := range nodes[1:] {
		load := atomic.LoadInt64(&n.inflight)
		if load < bestLoad {
			best, bestLoad = n, load
		}
	}
	return best
}

// runRemoteBenchmark is the HTTP-dispatch counterpart to runBenchmark: same
// config/report shape, but every "node" is a real detector-node endpoint.
func runRemoteBenchmark(cfg BenchConfig, endpoints []string, samples []Sample) (*BenchResult, error) {
	if cfg.ConcurrencyPerNode < 1 {
		cfg.ConcurrencyPerNode = 1
	}
	if cfg.Repeat < 1 {
		cfg.Repeat = 1
	}

	timeout := time.Duration(cfg.HTTPTimeoutSeconds) * time.Second
	if timeout <= 0 {
		// Flan-T5 inference under a constrained container (e.g. 1-2 CPUs) can
		// legitimately take several seconds per request, more under
		// concurrent load. A short timeout doesn't fail loudly — it silently
		// turns slow-but-correct requests into "failed", which used to get
		// mis-scored as 0 (see filterFailed) and corrupted TPR/FPR. Default
		// generous; pass --http-timeout-seconds to tighten it deliberately.
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	nodes := make([]*remoteNodeState, len(endpoints))
	for i, ep := range endpoints {
		if err := pingHealth(client, ep); err != nil {
			return nil, fmt.Errorf("node %d (%s) not reachable: %w", i, ep, err)
		}
		nodes[i] = &remoteNodeState{id: i, endpoint: strings.TrimRight(ep, "/"), client: client, jobs: make(chan job, 64)}
	}

	total := len(samples) * cfg.Repeat
	labels := make([]bool, total)
	scores := make([]float64, total)
	latencies := make([]float64, total)
	failed := make([]bool, total)

	var wg sync.WaitGroup
	for _, n := range nodes {
		for c := 0; c < cfg.ConcurrencyPerNode; c++ {
			wg.Add(1)
			go func(n *remoteNodeState) {
				defer wg.Done()
				for j := range n.jobs {
					elapsed, resp, err := detectRemote(n.client, n.endpoint, j.text)
					latencies[j.idx] = elapsed

					// A failure here is a network/HTTP error, a timeout, OR the
					// pipeline itself reporting Decision=="ERROR" (e.g. detector-node's
					// own call to prompt-shield failed). None of these produced a
					// real detection result, so they must not enter classification
					// as an implicit "score 0 / not injection" — that previously
					// inflated FN and deflated TPR/AUC whenever timeouts clustered
					// under load (see README "Docker-based evaluation" for the
					// incident this fixed).
					isFailed := err != nil || resp.Decision == "ERROR"
					failed[j.idx] = isFailed
					if !isFailed {
						labels[j.idx] = j.label
						scores[j.idx] = resp.Score
					}

					n.mu.Lock()
					n.requests++
					n.busyMs += elapsed
					if isFailed {
						n.failedRequests++
					}
					n.mu.Unlock()
					atomic.AddInt64(&n.inflight, -1)
				}
			}(n)
		}
	}

	start := time.Now()
	idx := 0
	for r := 0; r < cfg.Repeat; r++ {
		for _, s := range samples {
			target := selectRemoteNode(nodes, cfg.Strategy, idx)
			atomic.AddInt64(&target.inflight, 1)
			target.jobs <- job{idx: idx, label: s.IsInjection, text: s.Text}
			idx++
		}
	}
	for _, n := range nodes {
		close(n.jobs)
	}
	wg.Wait()
	wallClock := time.Since(start)

	mode := "distributed-http"
	strategy := cfg.Strategy
	if len(nodes) == 1 {
		mode = "baseline-http"
		strategy = ""
	}

	okLabels, okScores, failedCount := filterFailed(labels, scores, failed)

	result := &BenchResult{
		Mode:                 mode,
		Timestamp:            time.Now().UTC(),
		Detector:             cfg.Detector,
		Threshold:            cfg.Threshold,
		Chunker:              cfg.Chunker,
		FailMode:             cfg.FailMode,
		PromptShieldEndpoint: cfg.PromptShieldEndpoint,
		PromptShieldModel:    cfg.PromptShieldModel,
		Nodes:                len(nodes),
		Strategy:             strategy,
		ConcurrencyPerNode:   cfg.ConcurrencyPerNode,
		DatasetPath:          cfg.DatasetPath,
		DatasetSize:          len(samples),
		Repeat:               cfg.Repeat,
		TotalRequests:        total,
		FailedRequests:       failedCount,
		WallClockMs:          float64(wallClock.Microseconds()) / 1000.0,
		ThroughputRPS:        float64(total) / wallClock.Seconds(),
		Latency:              computeLatencyStats(latencies),
		Classification:       evaluation.EvaluateScores(okLabels, okScores, cfg.Threshold, cfg.TargetFPRs),
	}

	breakdown := make([]NodeBreakdown, len(nodes))
	reqCounts := make([]float64, len(nodes))
	busyTimes := make([]float64, len(nodes))
	for i, n := range nodes {
		n.mu.Lock()
		breakdown[i] = NodeBreakdown{NodeID: n.id, Endpoint: n.endpoint, Requests: n.requests, Failed: n.failedRequests, BusyMs: n.busyMs}
		reqCounts[i] = float64(n.requests)
		busyTimes[i] = n.busyMs
		n.mu.Unlock()
	}
	result.NodeBreakdown = breakdown
	result.FairnessByRequests = jainsFairnessIndex(reqCounts)
	result.FairnessByBusyTime = jainsFairnessIndex(busyTimes)

	return result, nil
}

func detectRemote(client *http.Client, endpoint, text string) (float64, *remoteDetectResponse, error) {
	start := time.Now()
	body, _ := json.Marshal(detectRequest{Text: text})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint+"/detect", bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return elapsed, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return elapsed, nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var out remoteDetectResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return elapsed, nil, err
	}
	return elapsed, &out, nil
}

func pingHealth(client *http.Client, endpoint string) error {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(endpoint, "/")+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

// detectRequest mirrors cmd/detector-node's request shape (duplicated here
// deliberately: this module treats the HTTP contract as the integration
// point, not a shared Go type, so the two binaries can evolve independently).
type detectRequest struct {
	Text string `json:"text"`
}
