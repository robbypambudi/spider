package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (r *PolicyRepo) EnsureDefault(ctx context.Context, name string, threshold float64) error {
	_, err := r.GetByName(ctx, name)
	if err == nil {
		return nil
	}
	if err != pgx.ErrNoRows {
		return err
	}
	now := time.Now().UTC()
	cfg, _ := json.Marshal(map[string]interface{}{"threshold": threshold})
	_, err = r.pool.Exec(ctx, `
		INSERT INTO policies (id, name, kind, threshold, action_on_detection, config_json, is_default, created_at, updated_at)
		VALUES ($1, $2, 'threshold', $3, 'block', $4, true, $5, $6)`,
		uuid.New(), name, threshold, string(cfg), now, now)
	return err
}
