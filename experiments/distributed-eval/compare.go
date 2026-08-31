package main

import (
	"flag"
	"fmt"
	"math"
)

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	baselinePath := fs.String("baseline", "", "path to a baseline (nodes=1) result JSON (required)")
	distPath := fs.String("distributed", "", "path to a distributed (nodes>1) result JSON (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *baselinePath == "" || *distPath == "" {
		return fmt.Errorf("both --baseline and --distributed are required")
	}

	b, err := loadResult(*baselinePath)
	if err != nil {
		return fmt.Errorf("baseline: %w", err)
	}
	d, err := loadResult(*distPath)
	if err != nil {
		return fmt.Errorf("distributed: %w", err)
	}

	fmt.Printf("baseline:    %s  (nodes=%d)\n", *baselinePath, b.Nodes)
	fmt.Printf("distributed: %s  (nodes=%d, strategy=%s)\n\n", *distPath, d.Nodes, d.Strategy)

	speedup := 0.0
	if b.ThroughputRPS > 0 {
		speedup = d.ThroughputRPS / b.ThroughputRPS
	}
	fmt.Println("throughput:")
	fmt.Printf("  baseline:    %.1f req/s\n", b.ThroughputRPS)
	fmt.Printf("  distributed: %.1f req/s  (%.2fx baseline)\n\n", d.ThroughputRPS, speedup)

	fmt.Println("latency p95 (ms):")
	fmt.Printf("  baseline:    %.2f\n", b.Latency.P95Ms)
	fmt.Printf("  distributed: %.2f  (delta %+.2f ms)\n\n", d.Latency.P95Ms, d.Latency.P95Ms-b.Latency.P95Ms)

	if b.FailedRequests > 0 || d.FailedRequests > 0 {
		fmt.Printf("failed requests (excluded from classification below):\n")
		fmt.Printf("  baseline:    %d / %d\n", b.FailedRequests, b.TotalRequests)
		fmt.Printf("  distributed: %d / %d\n\n", d.FailedRequests, d.TotalRequests)
		if b.FailedRequests != d.FailedRequests {
			fmt.Println("  note: unequal failure counts mean the two runs classified a different")
			fmt.Println("  effective sample set — treat the classification comparison below as")
			fmt.Println("  indicative, not a clean apples-to-apples equivalence check.")
		}
	}

	fmt.Println("classification equivalence (same detector+threshold on the same inputs should match exactly):")
	bc, dc := b.Classification.Counts, d.Classification.Counts
	fmt.Printf("  baseline:    TP=%d FP=%d TN=%d FN=%d  (n=%d)\n", bc.TP, bc.FP, bc.TN, bc.FN, b.Classification.Samples)
	fmt.Printf("  distributed: TP=%d FP=%d TN=%d FN=%d  (n=%d)\n", dc.TP, dc.FP, dc.TN, dc.FN, d.Classification.Samples)
	if bc.TP != dc.TP || bc.FP != dc.FP || bc.TN != dc.TN || bc.FN != dc.FN {
		fmt.Println("  MISMATCH — distributed decisions differ from baseline on the same inputs.")
		if b.FailedRequests > 0 || d.FailedRequests > 0 {
			fmt.Println("  Likely explained by the failed-request counts above, not a real accuracy")
			fmt.Println("  difference — re-run with a longer --http-timeout-seconds and check again.")
		} else {
			fmt.Println("  Expected if datasets/thresholds differ between runs; otherwise this points")
			fmt.Println("  at a bug in aggregation/dispatch rather than a performance trade-off.")
		}
	} else {
		fmt.Println("  match — distributing the workload did not change any decision.")
	}

	if len(d.NodeBreakdown) > 0 {
		fmt.Printf("\nload fairness (Jain's index, 1.0 = perfectly even):\n")
		fmt.Printf("  by requests:  %.4f\n", d.FairnessByRequests)
		fmt.Printf("  by busy time: %.4f\n", d.FairnessByBusyTime)
		ideal := 1.0 / float64(d.Nodes)
		if math.Abs(d.FairnessByRequests-1.0) > 0.05 {
			fmt.Printf("  note: noticeably below 1.0 (floor for %d nodes is %.3f) — load is unevenly spread.\n", d.Nodes, ideal)
		}
	}

	return nil
}
