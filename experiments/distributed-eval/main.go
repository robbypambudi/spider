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
	fmt.Println(`spider-bench — baseline vs distributed evaluation harness for SPIDER's security pipeline

Usage:
  go run . gendata --out datasets/set.jsonl --n 2000 [--seed 42] [--injection-ratio 0.3]
  go run . bench    --dataset datasets/set.jsonl --out results/name.json [--nodes 1] [flags...]
  go run . compare  --baseline results/a.json --distributed results/b.json

Run "go run . <command> -h" for the full flag list of that command.
See README.md for a walkthrough and what the "distributed" simulation does and does not model.`)
}
