package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/pkg/apis"
)

type SecurityScanStore interface {
	CreateFromResult(ctx context.Context, result apis.SecurityResult, promptHash string, promptLength int, promptText *string, userID *uuid.UUID, workerID *string) (*store.SecurityScan, error)
	RecordScanMetric(ctx context.Context, decision string, latencyMs float64) error
}

type InferenceStore interface {
	Create(ctx context.Context, p store.CreateInferenceParams) (*store.InferenceRecord, error)
	AddEvent(ctx context.Context, inferenceID uuid.UUID, eventType string, payload map[string]interface{}) error
}

type WorkerStore interface {
	ListResources(ctx context.Context) ([]apis.WorkerResource, error)
}
