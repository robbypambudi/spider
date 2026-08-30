package service

import (
	"context"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/pkg/config"
)

type PolicyService struct {
	Policies *store.PolicyRepo
	Settings *config.Settings
}

func (s *PolicyService) ListPolicies(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.Policies.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []map[string]interface{}{{
			"name": s.Settings.DefaultSecurityPolicy, "kind": "threshold",
			"threshold": s.Settings.DefaultThreshold, "action_on_detection": "block",
			"status": "implemented",
		}}, nil
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		out = append(out, map[string]interface{}{
			"name": p.Name, "kind": p.Kind, "threshold": p.Threshold,
			"action_on_detection": p.ActionOnDetection, "status": "implemented",
		})
	}
	prepared := []map[string]interface{}{
		{"name": "adaptive", "kind": "adaptive", "status": "prepared"},
		{"name": "ensemble", "kind": "ensemble", "status": "prepared"},
	}
	return append(out, prepared...), nil
}

type ServingService struct {
	Workers *store.WorkerRepo
}

func (s *ServingService) ListNodes(ctx context.Context) ([]map[string]interface{}, error) {
	return s.Workers.ListServingNodes(ctx)
}

func (s *ServingService) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	return s.Workers.ListModels(ctx)
}

type MetricsService struct {
	Scans    *store.SecurityRepo
	Workers  *store.WorkerRepo
	Infer    *store.InferenceRepo
}

func (s *MetricsService) DashboardSummary(ctx context.Context) (map[string]interface{}, error) {
	counts, err := s.Scans.SummaryCounts(ctx)
	if err != nil {
		return nil, err
	}
	total := counts["total_scans"].(int)
	blocked := counts["blocked"].(int)
	detectionRate := 0.0
	if total > 0 {
		detectionRate = float64(blocked) / float64(total)
	}
	avgOverhead, _ := s.Infer.AvgSecurityOverhead(ctx)
	p95, _ := s.Infer.P95DetectionLatency(ctx)
	activeNodes, _ := s.Workers.CountOnlineNodes(ctx)
	totalGPUs, _ := s.Workers.CountGPUs(ctx)
	workersTotal, _ := s.Workers.CountWorkers(ctx)

	return map[string]interface{}{
		"total_scans":                counts["total_scans"],
		"allowed":                    counts["allowed"],
		"blocked":                    counts["blocked"],
		"review":                     counts["review"],
		"detection_rate":             detectionRate,
		"avg_detection_latency_ms":   counts["avg_detection_latency_ms"],
		"p95_detection_latency_ms":   p95,
		"avg_security_overhead_ms":   avgOverhead,
		"active_serving_nodes":       activeNodes,
		"total_gpus":                 totalGPUs,
		"workers_total":              workersTotal,
	}, nil
}
