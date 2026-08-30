package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func runBench(args []string) error {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	dataset := fs.String("dataset", "datasets/injection-bench-1000.jsonl", "path to labeled JSONL dataset")
	out := fs.String("out", "", "path to write the JSON result report (required)")
	detector := fs.String("detector", "rule-based", "detector: rule-based | noop | prompt-shield")
	threshold := fs.Float64("threshold", 0.5, "decision threshold (score >= threshold => predicted injection)")
	chunker := fs.String("chunker", "fixed", "chunker: fixed | token")
	chunkSize := fs.Int("chunk-size", 2048, "chunk size (chars for fixed, tokens for token chunker)")
	chunkOverlap := fs.Int("chunk-overlap", 128, "chunk overlap")
	failMode := fs.String("fail-mode", "closed", "closed | open, mirrors SPIDER_FAIL_MODE")
	promptShieldEndpoint := fs.String("prompt-shield-endpoint", "http://localhost:8081", "Prompt-Shield sidecar URL (only used when --detector=prompt-shield or --chunker=token)")

	nodes := fs.Int("nodes", 1, "number of simulated nodes (1 = single-node baseline)")
	strategy := fs.String("strategy", "least-loaded", "dispatch strategy when nodes > 1: least-loaded | round-robin")
	concurrency := fs.Int("concurrency-per-node", 4, "concurrent goroutines per node (simulates concurrent clients hitting that node)")
	repeat := fs.Int("repeat", 1, "replay the dataset this many times, to inflate a small dataset into a sustained-load run")
	targetFPRs := fs.String("target-fpr", "0.0005,0.001,0.005,0.01", "comma-separated target FPRs for TPR@FPR reporting")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		return fmt.Errorf("--out is required (path to write the JSON result)")
	}

	fprs, err := parseFloatList(*targetFPRs)
	if err != nil {
		return fmt.Errorf("--target-fpr: %w", err)
	}

	samples, err := loadDataset(*dataset)
	if err != nil {
		return err
	}

	cfg := BenchConfig{
		DatasetPath:          *dataset,
		Detector:             *detector,
		Threshold:            *threshold,
		Chunker:              *chunker,
		ChunkSize:            *chunkSize,
		ChunkOverlap:         *chunkOverlap,
		FailMode:             *failMode,
		PromptShieldEndpoint: *promptShieldEndpoint,
		Nodes:                *nodes,
		Strategy:             *strategy,
		ConcurrencyPerNode:   *concurrency,
		Repeat:               *repeat,
		TargetFPRs:           fprs,
	}

	result, err := runBenchmark(cfg, samples)
	if err != nil {
		return err
	}

	if err := saveResult(*out, result); err != nil {
		return err
	}
	printSummary(result)
	fmt.Printf("\nsaved: %s\n", *out)
	return nil
}

func parseFloatList(s string) ([]float64, error) {
	parts := strings.Split(s, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q: %w", p, err)
		}
		out = append(out, v)
	}
	return out, nil
}
