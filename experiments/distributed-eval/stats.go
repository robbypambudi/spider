package main

import "sort"

// LatencyStats summarizes a set of per-request latencies (milliseconds).
type LatencyStats struct {
	Count  int     `json:"count"`
	MeanMs float64 `json:"mean_ms"`
	P50Ms  float64 `json:"p50_ms"`
	P95Ms  float64 `json:"p95_ms"`
	P99Ms  float64 `json:"p99_ms"`
	MaxMs  float64 `json:"max_ms"`
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	idx := p / 100 * float64(len(sorted)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := idx - float64(lo)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}

func computeLatencyStats(samples []float64) LatencyStats {
	if len(samples) == 0 {
		return LatencyStats{}
	}
	s := append([]float64(nil), samples...)
	sort.Float64s(s)
	sum, max := 0.0, s[0]
	for _, v := range s {
		sum += v
		if v > max {
			max = v
		}
	}
	return LatencyStats{
		Count:  len(s),
		MeanMs: sum / float64(len(s)),
		P50Ms:  percentile(s, 50),
		P95Ms:  percentile(s, 95),
		P99Ms:  percentile(s, 99),
		MaxMs:  max,
	}
}

// jainsFairnessIndex computes Jain's Fairness Index over a set of per-node
// load values (e.g. request counts or busy time). Returns 1.0 for perfectly
// even distribution, down toward 1/n for maximally skewed distribution.
// An all-zero input (no load anywhere) is reported as perfectly fair.
func jainsFairnessIndex(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
	}
	if sumSq == 0 {
		return 1
	}
	n := float64(len(values))
	return (sum * sum) / (n * sumSq)
}
