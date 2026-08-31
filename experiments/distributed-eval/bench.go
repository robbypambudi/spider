package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/spider/spider/pkg/security/detectors"
)

func runBench(args []string) error {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	dataset := fs.String("dataset", "", "path to Prompt-Shield JSON array or JSONL (required)")
	out := fs.String("out", "", "path to write the JSON result report (required)")
	detector := fs.String("detector", defaultDetector, "detector (must be prompt-shield)")
	threshold := fs.Float64("threshold", 0.5, "decision threshold for reporting (score >= threshold => predicted injection); TPR@FPR uses ROC sweep")
	chunker := fs.String("chunker", defaultChunker, "chunker (must be token)")
	chunkSize := fs.Int("chunk-size", defaultChunkSize, "token window size W (lab default 256)")
	chunkOverlap := fs.Int("chunk-overlap", defaultChunkOverlap, "token overlap O (lab wave-1 default 0)")
	failMode := fs.String("fail-mode", "closed", "closed | open, mirrors SPIDER_FAIL_MODE")
	promptShieldEndpoint := fs.String("prompt-shield-endpoint", defaultPromptShieldEndpoint, "Prompt-Shield sidecar URL")
	promptShieldModel := fs.String("prompt-shield-model", defaultPromptShieldModel, "Hugging Face model id (e.g. robbypambudi/prompt-shield-flan-t5-small)")

	nodes := fs.Int("nodes", 1, "number of simulated nodes (1 = single-node baseline); ignored when --node-endpoints is set")
	strategy := fs.String("strategy", "least-loaded", "dispatch strategy when >1 node: least-loaded | round-robin")
	concurrency := fs.Int("concurrency-per-node", 4, "concurrent goroutines per node (simulates concurrent clients hitting that node)")
	repeat := fs.Int("repeat", 1, "replay the dataset this many times, to inflate a small dataset into a sustained-load run")
	targetFPRs := fs.String("target-fpr", defaultTargetFPR, "comma-separated target FPRs for TPR@FPR reporting")
	nodeEndpoints := fs.String("node-endpoints", "", "comma-separated http://host:port list of real detector-node instances (e.g. Docker containers with their own --cpus/--memory). When set, requests are dispatched over real HTTP instead of the in-process simulation — see docker-compose.eval.yml")
	limit := fs.Int("limit", 0, "use only the first N rows of --dataset (0 = use all). For trimming a large real dataset down to a manageable run — never a substitute for real data.")
	httpTimeout := fs.Int("http-timeout-seconds", 60, "per-request timeout for --node-endpoints HTTP calls. Too short silently turns slow-but-correct detections into 'failed' (excluded from classification, see --out's failed_requests) — raise this if you see failed_requests > 0 and node CPU limits are tight")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dataset == "" {
		return fmt.Errorf("--dataset is required (Prompt-Shield JSON from spider-internal/labs, e.g. PromptShield/test.json or camera_ready .../evaluation_benchmark_en.json)")
	}
	if *out == "" {
		return fmt.Errorf("--out is required (path to write the JSON result)")
	}
	if err := validateLabConfig(*detector, *chunker); err != nil {
		return err
	}
	if *chunkOverlap >= *chunkSize {
		return fmt.Errorf("--chunk-overlap must be smaller than --chunk-size (lab constraint)")
	}

	detectors.RegisterPromptShield(*promptShieldEndpoint, *promptShieldModel)

	fprs, err := parseFloatList(*targetFPRs)
	if err != nil {
		return fmt.Errorf("--target-fpr: %w", err)
	}

	samples, err := loadDataset(*dataset)
	if err != nil {
		return err
	}
	if *limit > 0 && *limit < len(samples) {
		samples = samples[:*limit]
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
		PromptShieldModel:    *promptShieldModel,
		Nodes:                *nodes,
		Strategy:             *strategy,
		ConcurrencyPerNode:   *concurrency,
		Repeat:               *repeat,
		TargetFPRs:           fprs,
		HTTPTimeoutSeconds:   *httpTimeout,
	}

	var result *BenchResult
	if *nodeEndpoints != "" {
		endpoints := parseStringList(*nodeEndpoints)
		result, err = runRemoteBenchmark(cfg, endpoints, samples)
	} else {
		result, err = runBenchmark(cfg, samples)
	}
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

func parseStringList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
