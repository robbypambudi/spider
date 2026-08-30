package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func saveResult(path string, r *BenchResult) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}

func loadResult(path string) (*BenchResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read result: %w", err)
	}
	var r BenchResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &r, nil
}

func printSummary(r *BenchResult) {
	fmt.Printf("mode:              %s (nodes=%d", r.Mode, r.Nodes)
	if r.Strategy != "" {
		fmt.Printf(", strategy=%s", r.Strategy)
	}
	fmt.Printf(")\n")
	fmt.Printf("detector:          %s (threshold=%.3f, chunker=%s)\n", r.Detector, r.Threshold, r.Chunker)
	fmt.Printf("dataset:           %s (%d samples x %d repeat = %d requests)\n", r.DatasetPath, r.DatasetSize, r.Repeat, r.TotalRequests)
	fmt.Println()
	fmt.Printf("wall clock:        %.1f ms\n", r.WallClockMs)
	fmt.Printf("throughput:        %.1f req/s\n", r.ThroughputRPS)
	fmt.Printf("latency (ms):      mean=%.2f p50=%.2f p95=%.2f p99=%.2f max=%.2f\n",
		r.Latency.MeanMs, r.Latency.P50Ms, r.Latency.P95Ms, r.Latency.P99Ms, r.Latency.MaxMs)
	fmt.Println()
	c := r.Classification.Counts
	fmt.Printf("classification:    TP=%d FP=%d TN=%d FN=%d\n", c.TP, c.FP, c.TN, c.FN)
	fmt.Printf("                   TPR=%.4f FPR=%.4f Precision=%.4f F1=%.4f AUC=%.4f\n",
		r.Classification.TPR, r.Classification.FPR, r.Classification.Precision, r.Classification.F1, r.Classification.AUC)
	if len(r.NodeBreakdown) > 0 {
		fmt.Println()
		fmt.Println("per-node breakdown:")
		for _, n := range r.NodeBreakdown {
			fmt.Printf("  node %d: requests=%-6d busy=%.1fms\n", n.NodeID, n.Requests, n.BusyMs)
		}
		fmt.Printf("fairness (Jain):   by-requests=%.4f  by-busy-time=%.4f  (1.0 = perfectly even, 1/n = maximally skewed)\n",
			r.FairnessByRequests, r.FairnessByBusyTime)
	}
}
