package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/spidererrors"
	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/security/pipeline"
)

type PolicyService struct {
	Policies *store.PolicyRepo
	Settings *config.Settings
	Pipeline *pipeline.Pipeline
}

func (s *PolicyService) ListPolicies(ctx context.Context) ([]apis.PolicyView, error) {
	rows, err := s.Policies.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []apis.PolicyView{{
			Name:              s.Settings.DefaultSecurityPolicy,
			Kind:              "threshold",
			Threshold:         s.Settings.DefaultThreshold,
			ActionOnDetection: "block",
			Chunker:           s.Settings.Chunker,
			ChunkSize:         s.Settings.ChunkSize,
			ChunkOverlap:      s.Settings.ChunkOverlap,
			IsDefault:         true,
			Status:            "implemented",
		}}, nil
	}
	out := make([]apis.PolicyView, 0, len(rows))
	for _, p := range rows {
		out = append(out, policyToView(p))
	}
	return out, nil
}

func (s *PolicyService) GetPolicy(ctx context.Context, id uuid.UUID) (apis.PolicyView, error) {
	p, err := s.Policies.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrPolicyNotFound) {
			return apis.PolicyView{}, spidererrors.NotFound("Policy not found")
		}
		return apis.PolicyView{}, err
	}
	return policyToView(*p), nil
}

func (s *PolicyService) CreatePolicy(ctx context.Context, req apis.CreatePolicyRequest) (apis.PolicyView, error) {
	if req.Name == "" {
		return apis.PolicyView{}, spidererrors.Validation("name is required")
	}
	if err := validatePolicyInput(req.Threshold, req.ActionOnDetection, req.Chunker, req.ChunkSize, req.ChunkOverlap); err != nil {
		return apis.PolicyView{}, err
	}
	cfg := store.DefaultPolicyConfigJSON(req.Chunker, req.ChunkSize, req.ChunkOverlap)
	p := store.Policy{
		Name:              req.Name,
		Kind:              "threshold",
		Threshold:         req.Threshold,
		ActionOnDetection: req.ActionOnDetection,
		ConfigJSON:        cfg,
		IsDefault:         req.IsDefault,
	}
	if p.ActionOnDetection == "" {
		p.ActionOnDetection = "block"
	}
	if err := p.ValidateChunkOverlap(); err != nil {
		return apis.PolicyView{}, spidererrors.Validation(err.Error())
	}
	created, err := s.Policies.Create(ctx, p)
	if err != nil {
		return apis.PolicyView{}, err
	}
	if req.IsDefault {
		if err := s.Policies.SetDefault(ctx, created.ID); err != nil {
			return apis.PolicyView{}, err
		}
		created, _ = s.Policies.GetByID(ctx, created.ID)
	}
	_ = s.ApplyActivePolicy(ctx)
	return policyToView(*created), nil
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, id uuid.UUID, req apis.UpdatePolicyRequest) (apis.PolicyView, error) {
	existing, err := s.Policies.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrPolicyNotFound) {
			return apis.PolicyView{}, spidererrors.NotFound("Policy not found")
		}
		return apis.PolicyView{}, err
	}
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}
	threshold := existing.Threshold
	if req.Threshold != nil {
		threshold = *req.Threshold
	}
	action := existing.ActionOnDetection
	if req.ActionOnDetection != nil {
		action = *req.ActionOnDetection
	}
	chunker, chunkSize, chunkOverlap := store.ParsePolicyConfig(existing.ConfigJSON)
	if req.Chunker != nil {
		chunker = *req.Chunker
	}
	if req.ChunkSize != nil {
		chunkSize = *req.ChunkSize
	}
	if req.ChunkOverlap != nil {
		chunkOverlap = *req.ChunkOverlap
	}
	if err := validatePolicyInput(threshold, action, chunker, chunkSize, chunkOverlap); err != nil {
		return apis.PolicyView{}, err
	}
	isDefault := existing.IsDefault
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}
	updated := store.Policy{
		ID:                id,
		Name:              name,
		Kind:              existing.Kind,
		Threshold:         threshold,
		ActionOnDetection: action,
		ConfigJSON:        store.DefaultPolicyConfigJSON(chunker, chunkSize, chunkOverlap),
		IsDefault:         isDefault,
	}
	if err := updated.ValidateChunkOverlap(); err != nil {
		return apis.PolicyView{}, spidererrors.Validation(err.Error())
	}
	saved, err := s.Policies.Update(ctx, updated)
	if err != nil {
		return apis.PolicyView{}, err
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := s.Policies.SetDefault(ctx, id); err != nil {
			return apis.PolicyView{}, err
		}
		saved, _ = s.Policies.GetByID(ctx, id)
	}
	_ = s.ApplyActivePolicy(ctx)
	return policyToView(*saved), nil
}

func (s *PolicyService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	err := s.Policies.Delete(ctx, id)
	if errors.Is(err, store.ErrPolicyNotFound) {
		return spidererrors.NotFound("Policy not found")
	}
	if errors.Is(err, store.ErrLastDefaultPolicy) {
		return spidererrors.Validation("Cannot delete the only default policy")
	}
	if err != nil {
		return err
	}
	return s.ApplyActivePolicy(ctx)
}

func (s *PolicyService) SetDefaultPolicy(ctx context.Context, id uuid.UUID) (apis.PolicyView, error) {
	if err := s.Policies.SetDefault(ctx, id); err != nil {
		if errors.Is(err, store.ErrPolicyNotFound) {
			return apis.PolicyView{}, spidererrors.NotFound("Policy not found")
		}
		return apis.PolicyView{}, err
	}
	if err := s.ApplyActivePolicy(ctx); err != nil {
		return apis.PolicyView{}, err
	}
	p, err := s.Policies.GetByID(ctx, id)
	if err != nil {
		return apis.PolicyView{}, err
	}
	return policyToView(*p), nil
}

func (s *PolicyService) ApplyActivePolicy(ctx context.Context) error {
	if s.Pipeline == nil {
		return nil
	}
	p, err := s.Policies.GetDefault(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.Pipeline.SetPolicy(s.Settings.DefaultThreshold, "block")
			s.Pipeline.SetChunker(pipeline.NewChunker(
				s.Settings.Chunker,
				s.Settings.PromptShieldEndpoint,
				s.Settings.ChunkSize,
				s.Settings.ChunkOverlap,
			))
			return nil
		}
		return err
	}
	chunker, chunkSize, chunkOverlap := store.ParsePolicyConfig(p.ConfigJSON)
	s.Pipeline.SetPolicy(p.Threshold, p.ActionOnDetection)
	s.Pipeline.SetChunker(pipeline.NewChunker(
		chunker,
		s.Settings.PromptShieldEndpoint,
		chunkSize,
		chunkOverlap,
	))
	_ = ctx
	return nil
}

func policyToView(p store.Policy) apis.PolicyView {
	chunker, chunkSize, chunkOverlap := store.ParsePolicyConfig(p.ConfigJSON)
	return apis.PolicyView{
		ID:                p.ID.String(),
		Name:              p.Name,
		Kind:              p.Kind,
		Threshold:         p.Threshold,
		ActionOnDetection: p.ActionOnDetection,
		Chunker:           chunker,
		ChunkSize:         chunkSize,
		ChunkOverlap:      chunkOverlap,
		IsDefault:         p.IsDefault,
		Status:            "implemented",
	}
}

func validatePolicyInput(threshold float64, action, chunker string, chunkSize, chunkOverlap int) error {
	if threshold < 0 || threshold > 1 {
		return spidererrors.Validation("threshold must be between 0.0 and 1.0")
	}
	if action != "" && action != "block" && action != "review" {
		return spidererrors.Validation("action_on_detection must be block or review")
	}
	if chunker != "" && chunker != "token" && chunker != "fixed" {
		return spidererrors.Validation("chunker must be token or fixed")
	}
	if chunkSize < 0 || chunkOverlap < 0 {
		return spidererrors.Validation("chunk_size and chunk_overlap must be non-negative")
	}
	if chunkSize > 0 && chunkOverlap >= chunkSize {
		return spidererrors.Validation("chunk_overlap must be smaller than chunk_size")
	}
	return nil
}

type MetricsService struct {
	Scans   *store.SecurityRepo
	Workers *store.WorkerRepo
	Infer   *store.InferenceRepo
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
		"total_scans":              counts["total_scans"],
		"allowed":                  counts["allowed"],
		"blocked":                  counts["blocked"],
		"review":                   counts["review"],
		"detection_rate":           detectionRate,
		"avg_detection_latency_ms": counts["avg_detection_latency_ms"],
		"p95_detection_latency_ms": p95,
		"avg_security_overhead_ms": avgOverhead,
		"active_serving_nodes":     activeNodes,
		"total_gpus":               totalGPUs,
		"workers_total":            workersTotal,
	}, nil
}

func (s *PolicyService) PreparedPolicyKinds() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "adaptive", "kind": "adaptive", "status": "prepared"},
		{"name": "ensemble", "kind": "ensemble", "status": "prepared"},
	}
}