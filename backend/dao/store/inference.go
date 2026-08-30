package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InferenceRecord struct {
	ID                   uuid.UUID
	RequestID            uuid.UUID
	UserID               *uuid.UUID
	ScanID               *uuid.UUID
	Model                string
	Status               string
	Decision             string
	WorkerID             *string
	EndToEndLatencyMs    float64
	SecurityOverheadMs   float64
	InferenceLatencyMs   *float64
	OutputPreview        *string
	MetadataJSON         string
	CreatedAt            time.Time
}

type InferenceRepo struct {
	pool *pgxpool.Pool
}

func NewInferenceRepo(pool *pgxpool.Pool) *InferenceRepo {
	return &InferenceRepo{pool: pool}
}

type CreateInferenceParams struct {
	RequestID           uuid.UUID
	Model               string
	Status              string
	Decision            string
	ScanID              *uuid.UUID
	UserID              *uuid.UUID
	WorkerID            *string
	EndToEndLatencyMs   float64
	SecurityOverheadMs  float64
	InferenceLatencyMs  *float64
	OutputPreview       *string
	Metadata            map[string]interface{}
}

func (r *InferenceRepo) Create(ctx context.Context, p CreateInferenceParams) (*InferenceRecord, error) {
	now := time.Now().UTC()
	id := uuid.New()
	meta := p.Metadata
	if meta == nil {
		meta = map[string]interface{}{}
	}
	metaJSON, _ := json.Marshal(meta)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inference_requests (id, request_id, user_id, scan_id, model, status, decision,
			worker_id, end_to_end_latency_ms, security_overhead_ms, inference_latency_ms,
			output_preview, metadata_json, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		id, p.RequestID, p.UserID, p.ScanID, p.Model, p.Status, p.Decision,
		p.WorkerID, p.EndToEndLatencyMs, p.SecurityOverheadMs, p.InferenceLatencyMs,
		p.OutputPreview, string(metaJSON), now, now)
	if err != nil {
		return nil, err
	}
	return &InferenceRecord{
		ID: id, RequestID: p.RequestID, UserID: p.UserID, ScanID: p.ScanID,
		Model: p.Model, Status: p.Status, Decision: p.Decision, WorkerID: p.WorkerID,
		EndToEndLatencyMs: p.EndToEndLatencyMs, SecurityOverheadMs: p.SecurityOverheadMs,
		InferenceLatencyMs: p.InferenceLatencyMs, OutputPreview: p.OutputPreview,
		MetadataJSON: string(metaJSON), CreatedAt: now,
	}, nil
}

func (r *InferenceRepo) AddEvent(ctx context.Context, inferenceID uuid.UUID, eventType string, payload map[string]interface{}) error {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	data, _ := json.Marshal(payload)
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inference_events (id, inference_id, event_type, payload_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), inferenceID, eventType, string(data), now, now)
	return err
}

func (r *InferenceRepo) ListRecent(ctx context.Context, limit int) ([]InferenceRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id, user_id, scan_id, model, status, decision, worker_id,
			end_to_end_latency_ms, security_overhead_ms, inference_latency_ms, output_preview,
			metadata_json, created_at
		FROM inference_requests ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InferenceRecord
	for rows.Next() {
		var rec InferenceRecord
		if err := rows.Scan(&rec.ID, &rec.RequestID, &rec.UserID, &rec.ScanID, &rec.Model,
			&rec.Status, &rec.Decision, &rec.WorkerID, &rec.EndToEndLatencyMs,
			&rec.SecurityOverheadMs, &rec.InferenceLatencyMs, &rec.OutputPreview,
			&rec.MetadataJSON, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *InferenceRepo) AvgSecurityOverhead(ctx context.Context) (float64, error) {
	var avg float64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(AVG(security_overhead_ms), 0) FROM inference_requests`).Scan(&avg)
	return avg, err
}

func (r *InferenceRepo) P95DetectionLatency(ctx context.Context) (float64, error) {
	var p95 float64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms), 0)
		FROM security_scans`).Scan(&p95)
	return p95, err
}
