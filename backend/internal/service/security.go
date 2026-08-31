package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"github.com/google/uuid"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/telemetry"
	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/security/detectors"
	"github.com/spider/spider/pkg/security/evaluation"
	"github.com/spider/spider/pkg/security/pipeline"
)

func PromptHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

type SecurityService struct {
	Pipeline *pipeline.Pipeline
	Policy   *PolicyService
	Scans    SecurityScanStore
	Settings *config.Settings
	Metrics  *telemetry.Metrics
}

func (s *SecurityService) Inspect(ctx context.Context, text string, opts InspectOptions) (apis.SecurityResult, *store.SecurityScan, error) {
	if s.Policy != nil {
		_ = s.Policy.ApplyActivePolicy(ctx)
	}
	request := apis.SecurityRequest{
		RequestID: opts.RequestID,
		Text:      text,
		Source:    opts.Source,
		Model:     opts.Model,
		Metadata:  opts.Metadata,
	}
	if request.RequestID == "" {
		request.RequestID = uuid.New().String()
	}
	if opts.UserID != nil {
		uid := opts.UserID.String()
		request.UserID = &uid
	}

	result := s.Pipeline.Inspect(ctx, request)
	digest := PromptHash(text)

	logAttrs := []any{
		"request_id", result.RequestID,
		"decision", result.Decision,
		"score", result.Score,
		"detector", result.Metadata["detector"],
		"threshold", result.Metadata["threshold"],
		"latency_ms", result.TotalLatencyMs,
		"prompt_hash", digest,
		"prompt_length", len(text),
	}
	if opts.Model != nil {
		logAttrs = append(logAttrs, "model", *opts.Model)
	}
	if s.Settings.LogPromptContent {
		logAttrs = append(logAttrs, "prompt", text)
	}
	slog.Info("security_scan", logAttrs...)

	detectorName := "unknown"
	if v, ok := result.Metadata["detector"].(string); ok {
		detectorName = v
	}
	detectorLatency := result.TotalLatencyMs
	if len(result.DetectorResults) > 0 {
		detectorLatency = result.DetectorResults[0].LatencyMs
	}
	s.Metrics.ObserveScan(string(result.Decision), result.TotalLatencyMs, result.ChunksScanned, detectorName, result.Score, detectorLatency)
	_ = s.Scans.RecordScanMetric(ctx, string(result.Decision), result.TotalLatencyMs)

	var scan *store.SecurityScan
	if opts.Persist {
		shouldPersist := true
		if s.Settings != nil && s.Settings.AuditStoreMode != "all" {
			// In blocked_only mode (default), only persist threat incidents (BLOCK, REVIEW, ERROR)
			shouldPersist = result.Decision != apis.DecisionAllow
		}

		if shouldPersist {
			var promptText *string
			if s.Settings != nil && s.Settings.PersistPromptContent {
				promptText = &text
			}
			stored, err := s.Scans.CreateFromResult(ctx, result, digest, len(text), promptText, opts.UserID, opts.WorkerID)
			if err != nil {
				return result, nil, err
			}
			scan = stored
			slog.Info("security_scan_persisted", "scan_id", scan.ID.String(), "request_id", result.RequestID, "decision", result.Decision)
		}
	}
	return result, scan, nil
}

type InspectOptions struct {
	RequestID string
	Source    *string
	UserID    *uuid.UUID
	Model     *string
	Metadata  map[string]interface{}
	Persist   bool
	WorkerID  *string
}

func (s *SecurityService) ListDetectors() []map[string]string {
	var implemented []map[string]string
	for _, name := range detectors.Names() {
		entry := map[string]string{"name": name, "status": "implemented"}
		if name == "rule-based" {
			entry["warning"] = "rule-based is development/testing only"
		} else if name == "prompt-shield" {
			entry["warning"] = "loads models from huggingface.co/collections/robbypambudi/prompt-shield"
		} else {
			entry["warning"] = ""
		}
		implemented = append(implemented, entry)
	}
		prepared := []map[string]string{
		{"name": "flan-t5", "status": "prepared", "warning": "use prompt-shield detector with HF sidecar"},
		{"name": "transformer", "status": "prepared", "warning": "not implemented"},
		{"name": "remote", "status": "prepared", "warning": "not implemented"},
		{"name": "ensemble", "status": "prepared", "warning": "not implemented"},
	}
	return append(implemented, prepared...)
}

func (s *SecurityService) Evaluate(ctx context.Context, samples []apis.EvaluateSample, threshold *float64, targetFPRs []float64) apis.EvaluationReport {
	var scores []float64
	var labels []bool
	for _, sample := range samples {
		result, _, _ := s.Inspect(ctx, sample.Text, InspectOptions{Source: strPtr("evaluate"), Persist: false})
		scores = append(scores, result.Score)
		labels = append(labels, sample.IsInjection)
	}
	usedThreshold := s.Pipeline.Policy.Threshold()
	if threshold != nil {
		usedThreshold = *threshold
	}
	if len(targetFPRs) == 0 {
		targetFPRs = []float64{0.0005, 0.001, 0.005, 0.01}
	}
	return evaluation.EvaluateScores(labels, scores, usedThreshold, targetFPRs)
}

func strPtr(s string) *string { return &s }
