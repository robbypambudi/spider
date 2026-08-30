package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPolicyNotFound = errors.New("policy not found")
var ErrLastDefaultPolicy = errors.New("cannot delete the only default policy")

type Policy struct {
	ID                uuid.UUID
	Name              string
	Kind              string
	Threshold         float64
	ActionOnDetection string
	ConfigJSON        string
	IsDefault         bool
}

type PolicyRepo struct {
	pool *pgxpool.Pool
}

func NewPolicyRepo(pool *pgxpool.Pool) *PolicyRepo {
	return &PolicyRepo{pool: pool}
}

func (r *PolicyRepo) ListPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, kind, threshold, action_on_detection, config_json, is_default
		FROM policies ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Name, &p.Kind, &p.Threshold, &p.ActionOnDetection, &p.ConfigJSON, &p.IsDefault); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PolicyRepo) GetByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, kind, threshold, action_on_detection, config_json, is_default
		FROM policies WHERE id = $1`, id)
	var p Policy
	if err := row.Scan(&p.ID, &p.Name, &p.Kind, &p.Threshold, &p.ActionOnDetection, &p.ConfigJSON, &p.IsDefault); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PolicyRepo) GetByName(ctx context.Context, name string) (*Policy, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, kind, threshold, action_on_detection, config_json, is_default
		FROM policies WHERE name = $1`, name)
	var p Policy
	if err := row.Scan(&p.ID, &p.Name, &p.Kind, &p.Threshold, &p.ActionOnDetection, &p.ConfigJSON, &p.IsDefault); err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &p, nil
}

func (r *PolicyRepo) GetDefault(ctx context.Context) (*Policy, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, kind, threshold, action_on_detection, config_json, is_default
		FROM policies WHERE is_default = true
		ORDER BY updated_at DESC
		LIMIT 1`)
	var p Policy
	if err := row.Scan(&p.ID, &p.Name, &p.Kind, &p.Threshold, &p.ActionOnDetection, &p.ConfigJSON, &p.IsDefault); err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &p, nil
}

func (r *PolicyRepo) Create(ctx context.Context, p Policy) (*Policy, error) {
	now := time.Now().UTC()
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Kind == "" {
		p.Kind = "threshold"
	}
	if p.ActionOnDetection == "" {
		p.ActionOnDetection = "block"
	}
	if p.ConfigJSON == "" {
		p.ConfigJSON = "{}"
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO policies (id, name, kind, threshold, action_on_detection, config_json, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, p.Kind, p.Threshold, p.ActionOnDetection, p.ConfigJSON, p.IsDefault, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.ID)
}

func (r *PolicyRepo) Update(ctx context.Context, p Policy) (*Policy, error) {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE policies
		SET name = $2, kind = $3, threshold = $4, action_on_detection = $5,
		    config_json = $6, is_default = $7, updated_at = $8
		WHERE id = $1`,
		p.ID, p.Name, p.Kind, p.Threshold, p.ActionOnDetection, p.ConfigJSON, p.IsDefault, now)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrPolicyNotFound
	}
	return r.GetByID(ctx, p.ID)
}

func (r *PolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p.IsDefault {
		var count int
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policies`).Scan(&count); err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastDefaultPolicy
		}
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM policies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPolicyNotFound
	}
	if p.IsDefault {
		var fallback uuid.UUID
		if err := r.pool.QueryRow(ctx, `SELECT id FROM policies ORDER BY updated_at DESC LIMIT 1`).Scan(&fallback); err == nil {
			_ = r.SetDefault(ctx, fallback)
		}
	}
	return nil
}

func (r *PolicyRepo) SetDefault(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `UPDATE policies SET is_default = false WHERE is_default = true`)
	if err != nil {
		return err
	}
	_ = tag

	now := time.Now().UTC()
	tag, err = tx.Exec(ctx, `
		UPDATE policies SET is_default = true, updated_at = $2 WHERE id = $1`, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPolicyNotFound
	}
	return tx.Commit(ctx)
}

func (r *PolicyRepo) EnsureDefault(ctx context.Context, name string, threshold float64, chunker string, chunkSize, chunkOverlap int) error {
	_, err := r.GetByName(ctx, name)
	if err == nil {
		return nil
	}
	if err != pgx.ErrNoRows {
		return err
	}
	if chunker == "" {
		chunker = "token"
	}
	if chunkSize <= 0 {
		chunkSize = 256
	}
	cfg, _ := json.Marshal(map[string]interface{}{
		"chunker":       chunker,
		"chunk_size":    chunkSize,
		"chunk_overlap": chunkOverlap,
	})
	now := time.Now().UTC()
	_, err = r.pool.Exec(ctx, `
		INSERT INTO policies (id, name, kind, threshold, action_on_detection, config_json, is_default, created_at, updated_at)
		VALUES ($1, $2, 'threshold', $3, 'block', $4, true, $5, $6)`,
		uuid.New(), name, threshold, string(cfg), now, now)
	return err
}

func (r *PolicyRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policies`).Scan(&count)
	return count, err
}

func DefaultPolicyConfigJSON(chunker string, chunkSize, chunkOverlap int) string {
	if chunker == "" {
		chunker = "token"
	}
	if chunkSize <= 0 {
		chunkSize = 256
	}
	cfg, _ := json.Marshal(map[string]interface{}{
		"chunker":       chunker,
		"chunk_size":    chunkSize,
		"chunk_overlap": chunkOverlap,
	})
	return string(cfg)
}

func ParsePolicyConfig(raw string) (chunker string, chunkSize, chunkOverlap int) {
	chunker = "token"
	chunkSize = 256
	chunkOverlap = 0
	if raw == "" {
		return
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return
	}
	if v, ok := cfg["chunker"].(string); ok && v != "" {
		chunker = v
	}
	if v, ok := cfg["chunk_size"].(float64); ok {
		chunkSize = int(v)
	}
	if v, ok := cfg["chunk_overlap"].(float64); ok {
		chunkOverlap = int(v)
	}
	return
}

func (p Policy) ValidateChunkOverlap() error {
	_, size, overlap := ParsePolicyConfig(p.ConfigJSON)
	if overlap >= size {
		return fmt.Errorf("chunk_overlap must be smaller than chunk_size")
	}
	return nil
}
