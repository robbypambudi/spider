package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/telemetry"
	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/runtime"
	"github.com/spider/spider/pkg/security/enforcement"
	"github.com/spider/spider/pkg/serving/router"
)

type InferenceService struct {
	Security   *SecurityService
	Enforcer   *enforcement.Enforcer
	Router     *router.ServingRouter
	Inferences InferenceStore
	Workers    WorkerStore
	Metrics    *telemetry.Metrics
}

func (s *InferenceService) Infer(ctx context.Context, request apis.InferenceRequest, userID *uuid.UUID) (apis.ProtectedInferenceResponse, error) {
	requestID := uuid.New()
	started := time.Now()

	result, scan, err := s.Security.Inspect(ctx, request.Prompt, InspectOptions{
		RequestID: requestID.String(),
		Source:    strPtr("inference"),
		UserID:    userID,
		Model:     &request.Model,
		Persist:   true,
	})
	if err != nil {
		return apis.ProtectedInferenceResponse{}, err
	}

	securityOverhead := result.TotalLatencyMs
	action := s.Enforcer.Resolve(result.Decision)

	if action != enforcement.ActionForward {
		status := runtime.InferenceStatusBlocked
		if action == enforcement.ActionHold {
			status = runtime.InferenceStatusReview
		}
		e2e := float64(time.Since(started).Microseconds()) / 1000.0
		var scanID *uuid.UUID
		if scan != nil {
			scanID = &scan.ID
		}
		rec, err := s.Inferences.Create(ctx, store.CreateInferenceParams{
			RequestID: requestID, Model: request.Model, Status: status,
			Decision: string(result.Decision), ScanID: scanID, UserID: userID,
			EndToEndLatencyMs: e2e, SecurityOverheadMs: securityOverhead,
		})
		if err != nil {
			return apis.ProtectedInferenceResponse{}, err
		}
		eventType := "blocked"
		if status == runtime.InferenceStatusReview {
			eventType = "held"
		}
		_ = s.Inferences.AddEvent(ctx, rec.ID, eventType, map[string]interface{}{"action": string(action)})
		s.Metrics.ObserveInference(status, e2e, securityOverhead)
		slog.Info("inference_not_forwarded", "request_id", requestID.String(), "decision", result.Decision, "action", action)
		return apis.ProtectedInferenceResponse{
			RequestID: requestID.String(), Status: status, Decision: result.Decision,
			Model: request.Model, Security: result, SecurityOverheadMs: securityOverhead,
			EndToEndLatencyMs: e2e,
		}, nil
	}

	inventory, err := s.Workers.ListResources(ctx)
	if err != nil {
		return apis.ProtectedInferenceResponse{}, err
	}
	llmStarted := time.Now()
	llmResponse, selected, err := s.Router.Route(ctx, request, inventory)
	if err != nil {
		return apis.ProtectedInferenceResponse{}, err
	}
	inferenceLatency := float64(time.Since(llmStarted).Microseconds()) / 1000.0
	e2e := float64(time.Since(started).Microseconds()) / 1000.0

	var workerID *string
	if selected != nil {
		workerID = &selected.WorkerID
	}
	var scanID *uuid.UUID
	if scan != nil {
		scanID = &scan.ID
	}
	var preview *string
	if llmResponse.Output != nil {
		p := *llmResponse.Output
		if len(p) > 512 {
			p = p[:512]
		}
		preview = &p
	}

	rec, err := s.Inferences.Create(ctx, store.CreateInferenceParams{
		RequestID: requestID, Model: request.Model, Status: runtime.InferenceStatusCompleted,
		Decision: string(result.Decision), ScanID: scanID, UserID: userID, WorkerID: workerID,
		EndToEndLatencyMs: e2e, SecurityOverheadMs: securityOverhead,
		InferenceLatencyMs: &inferenceLatency, OutputPreview: preview,
	})
	if err != nil {
		return apis.ProtectedInferenceResponse{}, err
	}
	_ = s.Inferences.AddEvent(ctx, rec.ID, "completed", map[string]interface{}{"provider": "mock"})
	s.Metrics.ObserveInference(runtime.InferenceStatusCompleted, e2e, securityOverhead)
	slog.Info("inference_completed", "request_id", requestID.String(), "decision", result.Decision, "worker_id", workerID, "latency_ms", e2e)

	return apis.ProtectedInferenceResponse{
		RequestID: requestID.String(), Status: runtime.InferenceStatusCompleted,
		Decision: result.Decision, Model: request.Model, Output: llmResponse.Output,
		Security: result, SecurityOverheadMs: securityOverhead,
		InferenceLatencyMs: &inferenceLatency, EndToEndLatencyMs: e2e, WorkerID: workerID,
	}, nil
}
