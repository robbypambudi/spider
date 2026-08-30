// Command spider-bench simulates SPIDER's security detection pipeline
// running on a single node versus a simulated cluster of N nodes, so the
// two can be compared on classification (FP/FN/TPR/FPR), throughput,
// latency, and load fairness. See README.md for what is and isn't
// actually simulated.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "gendata":
		err = runGenData(os.Args[2:])
	case "bench":
		err = runBench(os.Args[2:])
	case "compare":
		err = runCompare(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`spider-bench — Prompt-Shield baseline vs distributed evaluation (spider-internal/labs methodology)

Usage:
  go run . gendata --out datasets/load-smoke.jsonl --n 2000 [--seed 42]
  go run . bench    --dataset /path/to/PromptShield/test.json --out results/name.json [--nodes 1]
  go run . compare  --baseline results/a.json --distributed results/b.json

Prerequisite: prompt-shield sidecar at http://localhost:8081
Defaults: detector=prompt-shield, chunker=token, chunk-size=256, overlap=0

Run "go run . <command> -h" for flags. See README.md.`)
}
