package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/evaluation"
	"github.com/spider/spider/pkg/security/pipeline"
)

// BenchConfig drives one run of the engine. Nodes=1 is the "baseline"
// (single-node) case; Nodes>1 simulates a distributed cluster of that many
// independent pipeline instances (see README's "What is and isn't
// simulated" section for exactly what this does and does not model).
type BenchConfig struct {
	DatasetPath          string
	Detector             string
	Threshold            float64
	Chunker              string
	ChunkSize            int
	ChunkOverlap         int
	FailMode             string
	PromptShieldEndpoint string
	PromptShieldModel    string

	Nodes              int
	Strategy           string // "round-robin" | "least-loaded"
	ConcurrencyPerNode int
	Repeat             int
	TargetFPRs         []float64
}

type nodeState struct {
	id       int
	pipeline *pipeline.Pipeline
	jobs     chan job
	inflight int64 // atomic; queued + in-progress, used by the least-loaded strategy

	mu       sync.Mutex
	requests int
	busyMs   float64
}

type job struct {
	idx   int
	label bool
	text  string
}

type NodeBreakdown struct {
	NodeID   int     `json:"node_id"`
	Endpoint string  `json:"endpoint,omitempty"` // set in remote-http mode (see remote_runner.go)
	Requests int     `json:"requests"`
	BusyMs   float64 `json:"busy_ms"`
}

// BenchResult is the full JSON report written by `bench` and read by `compare`.
type BenchResult struct {
	Mode      string    `json:"mode"`
	Timestamp time.Time `json:"timestamp"`

	Detector             string  `json:"detector"`
	Threshold            float64 `json:"threshold"`
	Chunker              string  `json:"chunker"`
	FailMode             string  `json:"fail_mode"`
	PromptShieldEndpoint string  `json:"prompt_shield_endpoint,omitempty"`
	PromptShieldModel    string  `json:"prompt_shield_model,omitempty"`

	Nodes              int    `json:"nodes"`
	Strategy           string `json:"strategy,omitempty"`
	ConcurrencyPerNode int    `json:"concurrency_per_node"`

	DatasetPath   string `json:"dataset"`
	DatasetSize   int    `json:"dataset_size"`
	Repeat        int    `json:"repeat"`
	TotalRequests int    `json:"total_requests"`

	WallClockMs   float64      `json:"wall_clock_ms"`
	ThroughputRPS float64      `json:"throughput_rps"`
	Latency       LatencyStats `json:"latency"`

	Classification apis.EvaluationReport `json:"classification"`

	NodeBreakdown       []NodeBreakdown `json:"node_breakdown,omitempty"`
	FairnessByRequests  float64         `json:"fairness_by_requests,omitempty"`
	FairnessByBusyTime  float64         `json:"fairness_by_busy_time,omitempty"`
}

func selectNode(nodes []*nodeState, strategy string, idx int) *nodeState {
	if len(nodes) == 1 {
		return nodes[0]
	}
	if strategy == "round-robin" {
		return nodes[idx%len(nodes)]
	}
	// least-loaded: mirrors pkg/scheduler.LeastLoadedScheduler's "fewest
	// running requests" dimension (GPU/site locality don't apply here).
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

func runBenchmark(cfg BenchConfig, samples []Sample) (*BenchResult, error) {
	if cfg.Nodes < 1 {
		cfg.Nodes = 1
	}
	if cfg.ConcurrencyPerNode < 1 {
		cfg.ConcurrencyPerNode = 1
	}
	if cfg.Repeat < 1 {
		cfg.Repeat = 1
	}

	nodes := make([]*nodeState, cfg.Nodes)
	for i := range nodes {
		p, err := pipeline.BuildDefault(pipeline.BuildOptions{
			DetectorName:         cfg.Detector,
			Threshold:            cfg.Threshold,
			Chunker:              cfg.Chunker,
			ChunkSize:            cfg.ChunkSize,
			ChunkOverlap:         cfg.ChunkOverlap,
			FailMode:             cfg.FailMode,
			PromptShieldEndpoint: cfg.PromptShieldEndpoint,
		})
		if err != nil {
			return nil, fmt.Errorf("build pipeline for node %d: %w", i, err)
		}
		nodes[i] = &nodeState{id: i, pipeline: p, jobs: make(chan job, 64)}
	}

	total := len(samples) * cfg.Repeat
	labels := make([]bool, total)
	scores := make([]float64, total)
	latencies := make([]float64, total)

	var wg sync.WaitGroup
	for _, n := range nodes {
		for c := 0; c < cfg.ConcurrencyPerNode; c++ {
			wg.Add(1)
			go func(n *nodeState) {
				defer wg.Done()
				for j := range n.jobs {
					reqStart := time.Now()
					res := n.pipeline.Inspect(context.Background(), apis.SecurityRequest{
						RequestID: fmt.Sprintf("bench-%d", j.idx),
						Text:      j.text,
					})
					elapsed := float64(time.Since(reqStart).Microseconds()) / 1000.0

					labels[j.idx] = j.label
					scores[j.idx] = res.Score
					latencies[j.idx] = elapsed

					n.mu.Lock()
					n.requests++
					n.busyMs += elapsed
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
			target := selectNode(nodes, cfg.Strategy, idx)
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

	mode := "distributed"
	strategy := cfg.Strategy
	if cfg.Nodes == 1 {
		mode = "baseline"
		strategy = ""
	}

	result := &BenchResult{
		Mode:                 mode,
		Timestamp:            time.Now().UTC(),
		Detector:             cfg.Detector,
		Threshold:            cfg.Threshold,
		Chunker:              cfg.Chunker,
		FailMode:             cfg.FailMode,
		PromptShieldEndpoint: cfg.PromptShieldEndpoint,
		PromptShieldModel:    cfg.PromptShieldModel,
		Nodes:                cfg.Nodes,
		Strategy:             strategy,
		ConcurrencyPerNode:   cfg.ConcurrencyPerNode,
		DatasetPath:          cfg.DatasetPath,
		DatasetSize:          len(samples),
		Repeat:               cfg.Repeat,
		TotalRequests:        total,
		WallClockMs:          float64(wallClock.Microseconds()) / 1000.0,
		ThroughputRPS:        float64(total) / wallClock.Seconds(),
		Latency:              computeLatencyStats(latencies),
		Classification:       evaluation.EvaluateScores(labels, scores, cfg.Threshold, cfg.TargetFPRs),
	}

	if cfg.Nodes > 1 {
		breakdown := make([]NodeBreakdown, len(nodes))
		reqCounts := make([]float64, len(nodes))
		busyTimes := make([]float64, len(nodes))
		for i, n := range nodes {
			n.mu.Lock()
			breakdown[i] = NodeBreakdown{NodeID: n.id, Requests: n.requests, BusyMs: n.busyMs}
			reqCounts[i] = float64(n.requests)
			busyTimes[i] = n.busyMs
			n.mu.Unlock()
		}
		result.NodeBreakdown = breakdown
		result.FairnessByRequests = jainsFairnessIndex(reqCounts)
		result.FairnessByBusyTime = jainsFairnessIndex(busyTimes)
	}

	return result, nil
}
