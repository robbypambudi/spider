package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spider/spider/pkg/apis"
)

type SecurityScan struct {
	ID               uuid.UUID
	RequestID        uuid.UUID
	UserID           *uuid.UUID
	Decision         string
	Score            float64
	Threshold        *float64
	Detector         string
	DetectorVersion  string
	Policy           string
	ChunksScanned    int
	ChunkingStrategy string
	LatencyMs        float64
	PromptHash       string
	PromptLength     int
	PromptText       *string
	ModelTarget      *string
	WorkerID         *string
	Source           *string
	MetadataJSON     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SecurityRepo struct {
	pool *pgxpool.Pool
}

func NewSecurityRepo(pool *pgxpool.Pool) *SecurityRepo {
	return &SecurityRepo{pool: pool}
}

func (r *SecurityRepo) CreateFromResult(ctx context.Context, result apis.SecurityResult, promptHash string, promptLength int, promptText *string, userID *uuid.UUID, workerID *string) (*SecurityScan, error) {
	meta := result.Metadata
	if meta == nil {
		meta = map[string]interface{}{}
	}
	metaJSON, _ := json.Marshal(meta)
	now := time.Now().UTC()
	id := uuid.New()

	var threshold *float64
	if v, ok := meta["threshold"].(float64); ok {
		threshold = &v
	}
	detector := strMeta(meta, "detector", "unknown")
	detectorVersion := strMeta(meta, "detector_version", "unknown")
	chunker := strMeta(meta, "chunker", "fixed")
	var modelTarget, source *string
	if v, ok := meta["model"].(string); ok {
		modelTarget = &v
	}
	if v, ok := meta["source"].(string); ok {
		source = &v
	}

	reqID, _ := uuid.Parse(result.RequestID)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO security_scans (id, request_id, user_id, decision, score, threshold, detector,
			detector_version, policy, chunks_scanned, chunking_strategy, latency_ms, prompt_hash,
			prompt_length, prompt_text, model_target, worker_id, source, metadata_json, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		id, reqID, userID, string(result.Decision), result.Score, threshold, detector,
		detectorVersion, result.Policy, result.ChunksScanned, chunker, result.TotalLatencyMs,
		promptHash, promptLength, promptText, modelTarget, workerID, source, string(metaJSON), now, now)
	if err != nil {
		return nil, err
	}

	for index, detection := range result.DetectorResults {
		dmeta, _ := json.Marshal(detection.Metadata)
		_, err = r.pool.Exec(ctx, `
			INSERT INTO security_chunk_results (id, scan_id, chunk_index, detector, score, is_injection, latency_ms, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			uuid.New(), id, index, detection.Detector, detection.Score, detection.IsInjection,
			detection.LatencyMs, now, now)
		if err != nil {
			return nil, err
		}
		version := strMeta(detection.Metadata, "version", "unknown")
		_, err = r.pool.Exec(ctx, `
			INSERT INTO detector_executions (id, scan_id, detector, detector_version, threshold, score,
				is_injection, latency_ms, metadata_json, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			uuid.New(), id, detection.Detector, version, detection.Threshold, detection.Score,
			detection.IsInjection, detection.LatencyMs, string(dmeta), now, now)
		if err != nil {
			return nil, err
		}
	}

	return &SecurityScan{
		ID: id, RequestID: reqID, UserID: userID, Decision: string(result.Decision),
		Score: result.Score, Threshold: threshold, Detector: detector,
		DetectorVersion: detectorVersion, Policy: result.Policy,
		ChunksScanned: result.ChunksScanned, ChunkingStrategy: chunker,
		LatencyMs: result.TotalLatencyMs, PromptHash: promptHash, PromptLength: promptLength,
		PromptText: promptText, ModelTarget: modelTarget, WorkerID: workerID, Source: source,
		MetadataJSON: string(metaJSON), CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (r *SecurityRepo) Get(ctx context.Context, scanID uuid.UUID) (*SecurityScan, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, request_id, user_id, decision, score, threshold, detector, detector_version,
			policy, chunks_scanned, chunking_strategy, latency_ms, prompt_hash, prompt_length,
			prompt_text, model_target, worker_id, source, metadata_json, created_at, updated_at
		FROM security_scans WHERE id = $1`, scanID)
	return scanSecurityRow(row)
}

func (r *SecurityRepo) ListScans(ctx context.Context, limit, offset int) ([]SecurityScan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id, user_id, decision, score, threshold, detector, detector_version,
			policy, chunks_scanned, chunking_strategy, latency_ms, prompt_hash, prompt_length,
			prompt_text, model_target, worker_id, source, metadata_json, created_at, updated_at
		FROM security_scans ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SecurityScan
	for rows.Next() {
		s, err := scanSecurityRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

type ChunkResult struct {
	ChunkIndex  int
	Detector    string
	Score       float64
	IsInjection bool
	LatencyMs   float64
}

type DetectorExecution struct {
	Detector        string
	DetectorVersion string
	Threshold       *float64
	Score           float64
	IsInjection     bool
	LatencyMs       float64
	MetadataJSON    string
}

func (r *SecurityRepo) ChunkResults(ctx context.Context, scanID uuid.UUID) ([]ChunkResult, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT chunk_index, detector, score, is_injection, latency_ms
		FROM security_chunk_results WHERE scan_id = $1 ORDER BY chunk_index`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChunkResult
	for rows.Next() {
		var c ChunkResult
		if err := rows.Scan(&c.ChunkIndex, &c.Detector, &c.Score, &c.IsInjection, &c.LatencyMs); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *SecurityRepo) DetectorExecutions(ctx context.Context, scanID uuid.UUID) ([]DetectorExecution, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT detector, detector_version, threshold, score, is_injection, latency_ms, metadata_json
		FROM detector_executions WHERE scan_id = $1`, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DetectorExecution
	for rows.Next() {
		var d DetectorExecution
		if err := rows.Scan(&d.Detector, &d.DetectorVersion, &d.Threshold, &d.Score, &d.IsInjection, &d.LatencyMs, &d.MetadataJSON); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *SecurityRepo) RecordScanMetric(ctx context.Context, decision string, latencyMs float64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO scan_metrics_counters (decision, count, total_latency_ms)
		VALUES ($1, 1, $2)
		ON CONFLICT (decision) DO UPDATE
		SET count = scan_metrics_counters.count + 1,
		    total_latency_ms = scan_metrics_counters.total_latency_ms + EXCLUDED.total_latency_ms
	`, decision, latencyMs)
	return err
}

func (r *SecurityRepo) SummaryCounts(ctx context.Context) (map[string]interface{}, error) {
	// Try aggregate counters table first for zero-loss telemetry
	rows, err := r.pool.Query(ctx, `SELECT decision, count, total_latency_ms FROM scan_metrics_counters`)
	if err == nil {
		defer rows.Close()
		var total, allowed, blocked, review int64
		var totalLatency float64
		hasData := false
		for rows.Next() {
			var dec string
			var cnt int64
			var lat float64
			if err := rows.Scan(&dec, &cnt, &lat); err == nil {
				hasData = true
				total += cnt
				totalLatency += lat
				switch dec {
				case "ALLOW":
					allowed += cnt
				case "BLOCK":
					blocked += cnt
				case "REVIEW":
					review += cnt
				}
			}
		}
		if hasData {
			avgLatency := 0.0
			if total > 0 {
				avgLatency = totalLatency / float64(total)
			}
			return map[string]interface{}{
				"total_scans":              int(total),
				"allowed":                  int(allowed),
				"blocked":                  int(blocked),
				"review":                   int(review),
				"avg_detection_latency_ms": avgLatency,
			}, nil
		}
	}

	// Fallback to querying security_scans table if counters table is not yet populated
	var total, allowed, blocked, review int
	var avgLatency float64
	err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM security_scans`).Scan(&total)
	if err != nil {
		return nil, err
	}
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM security_scans WHERE decision='ALLOW'`).Scan(&allowed)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM security_scans WHERE decision='BLOCK'`).Scan(&blocked)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM security_scans WHERE decision='REVIEW'`).Scan(&review)
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(AVG(latency_ms), 0) FROM security_scans`).Scan(&avgLatency)
	return map[string]interface{}{
		"total_scans": total, "allowed": allowed, "blocked": blocked, "review": review,
		"avg_detection_latency_ms": avgLatency,
	}, nil
}

func strMeta(m map[string]interface{}, key, fallback string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return fallback
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanSecurityRow(row scannable) (*SecurityScan, error) {
	var s SecurityScan
	err := row.Scan(&s.ID, &s.RequestID, &s.UserID, &s.Decision, &s.Score, &s.Threshold,
		&s.Detector, &s.DetectorVersion, &s.Policy, &s.ChunksScanned, &s.ChunkingStrategy,
		&s.LatencyMs, &s.PromptHash, &s.PromptLength, &s.PromptText, &s.ModelTarget,
		&s.WorkerID, &s.Source, &s.MetadataJSON, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanSecurityRows(rows scannable) (*SecurityScan, error) {
	return scanSecurityRow(rows)
}
