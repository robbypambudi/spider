package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/runtime"
)

type WorkerRow struct {
	ID               uuid.UUID
	WorkerID         string
	Hostname         string
	Site             *string
	Version          string
	Status           string
	CPUTotal         int
	MemoryTotalMB    int
	RunningRequests  int
	LastHeartbeatAt  *time.Time
	ModelsJSON       string
	MetadataJSON     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type WorkerRepo struct {
	pool *pgxpool.Pool
}

func NewWorkerRepo(pool *pgxpool.Pool) *WorkerRepo {
	return &WorkerRepo{pool: pool}
}

func (r *WorkerRepo) UpsertRegistration(ctx context.Context, resource apis.WorkerResource) (*WorkerRow, error) {
	now := time.Now().UTC()
	modelsJSON, _ := json.Marshal(resource.Models)
	metaJSON, _ := json.Marshal(resource.Metadata)

	existing, err := r.GetByWorkerID(ctx, resource.WorkerID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	var worker WorkerRow
	if existing == nil {
		id := uuid.New()
		_, err = r.pool.Exec(ctx, `
			INSERT INTO workers (id, worker_id, hostname, site, version, status, cpu_total, memory_total_mb,
				running_requests, last_heartbeat_at, models_json, metadata_json, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			id, resource.WorkerID, resource.Hostname, resource.Site, resource.Version,
			runtime.WorkerStatusOnline, resource.Resources.CPUTotal, resource.Resources.MemoryTotalMB,
			resource.RunningRequests, now, string(modelsJSON), string(metaJSON), now, now)
		if err != nil {
			return nil, err
		}
		_, err = r.pool.Exec(ctx, `
			INSERT INTO serving_nodes (id, worker_pk, worker_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), id, resource.WorkerID, runtime.WorkerStatusOnline, now, now)
		if err != nil {
			return nil, err
		}
		worker = WorkerRow{ID: id, WorkerID: resource.WorkerID}
	} else {
		_, err = r.pool.Exec(ctx, `
			UPDATE workers SET hostname=$2, site=$3, version=$4, status=$5, cpu_total=$6, memory_total_mb=$7,
				running_requests=$8, last_heartbeat_at=$9, models_json=$10, metadata_json=$11, updated_at=$12
			WHERE worker_id=$1`,
			resource.WorkerID, resource.Hostname, resource.Site, resource.Version,
			runtime.WorkerStatusOnline, resource.Resources.CPUTotal, resource.Resources.MemoryTotalMB,
			resource.RunningRequests, now, string(modelsJSON), string(metaJSON), now)
		if err != nil {
			return nil, err
		}
		worker = *existing
	}

	if err := r.replaceGPUs(ctx, worker.ID, resource.WorkerID, resource.Resources.GPUs); err != nil {
		return nil, err
	}
	if err := r.replaceModels(ctx, resource.WorkerID, resource.Models); err != nil {
		return nil, err
	}
	return r.GetByWorkerID(ctx, resource.WorkerID)
}

func (r *WorkerRepo) RecordHeartbeat(ctx context.Context, hb apis.WorkerHeartbeat) (*WorkerRow, error) {
	worker, err := r.GetByWorkerID(ctx, hb.WorkerID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	modelsJSON, _ := json.Marshal(hb.Models)
	metaJSON, _ := json.Marshal(hb.Metadata)

	cpu := worker.CPUTotal
	mem := worker.MemoryTotalMB
	if hb.Resources != nil {
		cpu = hb.Resources.CPUTotal
		mem = hb.Resources.MemoryTotalMB
		if err := r.replaceGPUs(ctx, worker.ID, hb.WorkerID, hb.Resources.GPUs); err != nil {
			return nil, err
		}
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE workers SET status=$2, running_requests=$3, last_heartbeat_at=$4, cpu_total=$5,
			memory_total_mb=$6, models_json=$7, updated_at=$8 WHERE worker_id=$1`,
		hb.WorkerID, hb.Status, hb.RunningRequests, now, cpu, mem, string(modelsJSON), now)
	if err != nil {
		return nil, err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO worker_heartbeats (id, worker_pk, worker_id, status, payload_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), worker.ID, hb.WorkerID, hb.Status, string(metaJSON), now)
	if err != nil {
		return nil, err
	}

	_, _ = r.pool.Exec(ctx, `UPDATE serving_nodes SET status=$2, updated_at=$3 WHERE worker_id=$1`,
		hb.WorkerID, hb.Status, now)

	if err := r.replaceModels(ctx, hb.WorkerID, hb.Models); err != nil {
		return nil, err
	}
	return r.GetByWorkerID(ctx, hb.WorkerID)
}

func (r *WorkerRepo) GetByWorkerID(ctx context.Context, workerID string) (*WorkerRow, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, worker_id, hostname, site, version, status, cpu_total, memory_total_mb,
			running_requests, last_heartbeat_at, models_json, metadata_json, created_at, updated_at
		FROM workers WHERE worker_id = $1`, workerID)
	var w WorkerRow
	if err := row.Scan(&w.ID, &w.WorkerID, &w.Hostname, &w.Site, &w.Version, &w.Status,
		&w.CPUTotal, &w.MemoryTotalMB, &w.RunningRequests, &w.LastHeartbeatAt,
		&w.ModelsJSON, &w.MetadataJSON, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WorkerRepo) ListWorkers(ctx context.Context) ([]WorkerRow, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, worker_id, hostname, site, version, status, cpu_total,
		memory_total_mb, running_requests, last_heartbeat_at, models_json, metadata_json, created_at, updated_at
		FROM workers ORDER BY worker_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkers(rows)
}

func (r *WorkerRepo) MarkOffline(ctx context.Context, workerID string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `UPDATE workers SET status=$2, updated_at=$3 WHERE worker_id=$1`,
		workerID, runtime.WorkerStatusOffline, now)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE serving_nodes SET status=$2, updated_at=$3 WHERE worker_id=$1`,
		workerID, runtime.WorkerStatusOffline, now)
	return err
}

func (r *WorkerRepo) LastHeartbeatAt(ctx context.Context, workerID string) (*time.Time, error) {
	w, err := r.GetByWorkerID(ctx, workerID)
	if err != nil {
		return nil, err
	}
	return w.LastHeartbeatAt, nil
}

func (r *WorkerRepo) DeleteWorker(ctx context.Context, workerID string) error {
	_, _ = r.pool.Exec(ctx, `DELETE FROM serving_models WHERE worker_id = $1`, workerID)
	_, _ = r.pool.Exec(ctx, `DELETE FROM serving_nodes WHERE worker_id = $1`, workerID)
	res, err := r.pool.Exec(ctx, `DELETE FROM workers WHERE worker_id = $1`, workerID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *WorkerRepo) PruneOfflineWorkers(ctx context.Context) (int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT worker_id FROM workers WHERE status = $1`, runtime.WorkerStatusOffline)
	if err != nil {
		return 0, err
	}
	var offlineIDs []string
	for rows.Next() {
		var wid string
		if err := rows.Scan(&wid); err == nil {
			offlineIDs = append(offlineIDs, wid)
		}
	}
	rows.Close()

	for _, wid := range offlineIDs {
		_, _ = r.pool.Exec(ctx, `DELETE FROM serving_models WHERE worker_id = $1`, wid)
		_, _ = r.pool.Exec(ctx, `DELETE FROM serving_nodes WHERE worker_id = $1`, wid)
	}

	res, err := r.pool.Exec(ctx, `DELETE FROM workers WHERE status = $1`, runtime.WorkerStatusOffline)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *WorkerRepo) UpdateWorker(ctx context.Context, workerID string, req apis.UpdateWorkerRequest) (*WorkerRow, error) {
	w, err := r.GetByWorkerID(ctx, workerID)
	if err != nil {
		return nil, err
	}
	hostname := w.Hostname
	if req.Hostname != nil && *req.Hostname != "" {
		hostname = *req.Hostname
	}
	site := w.Site
	if req.Site != nil {
		site = req.Site
	}
	metaJSON := w.MetadataJSON
	if req.Metadata != nil {
		b, _ := json.Marshal(req.Metadata)
		metaJSON = string(b)
	}
	now := time.Now().UTC()
	_, err = r.pool.Exec(ctx, `UPDATE workers SET hostname=$2, site=$3, metadata_json=$4, updated_at=$5 WHERE worker_id=$1`,
		workerID, hostname, site, metaJSON, now)
	if err != nil {
		return nil, err
	}
	return r.GetByWorkerID(ctx, workerID)
}

func (r *WorkerRepo) ListServingNodes(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, worker_id, status FROM serving_nodes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var workerID, status string
		if err := rows.Scan(&id, &workerID, &status); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id": id.String(), "worker_id": workerID, "status": status,
		})
	}
	return out, rows.Err()
}

func (r *WorkerRepo) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, `SELECT worker_id, name, status FROM serving_models`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var workerID, name, status string
		if err := rows.Scan(&workerID, &name, &status); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"worker_id": workerID, "name": name, "status": status,
		})
	}
	return out, rows.Err()
}

func (r *WorkerRepo) GPUsFor(ctx context.Context, workerID string) ([]apis.GPUResource, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT gpu_index, vendor, name, memory_total_mb, memory_used_mb, utilization
		FROM worker_gpus WHERE worker_id = $1 ORDER BY gpu_index`, workerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var gpus []apis.GPUResource
	for rows.Next() {
		var g apis.GPUResource
		if err := rows.Scan(&g.Index, &g.Vendor, &g.Name, &g.MemoryTotalMB, &g.MemoryUsedMB, &g.Utilization); err != nil {
			return nil, err
		}
		gpus = append(gpus, g)
	}
	return gpus, rows.Err()
}

func (r *WorkerRepo) AsResource(ctx context.Context, w *WorkerRow) (apis.WorkerResource, error) {
	gpus, err := r.GPUsFor(ctx, w.WorkerID)
	if err != nil {
		return apis.WorkerResource{}, err
	}
	var models []apis.LoadedModel
	_ = json.Unmarshal([]byte(w.ModelsJSON), &models)
	var meta map[string]interface{}
	_ = json.Unmarshal([]byte(w.MetadataJSON), &meta)
	return apis.WorkerResource{
		WorkerID: w.WorkerID, Hostname: w.Hostname, Site: w.Site, Version: w.Version,
		Status: w.Status,
		Resources: apis.WorkerResources{CPUTotal: w.CPUTotal, MemoryTotalMB: w.MemoryTotalMB, GPUs: normalizeGPUs(gpus)},
		Models: models, RunningRequests: w.RunningRequests, Metadata: meta,
	}, nil
}

func normalizeGPUs(gpus []apis.GPUResource) []apis.GPUResource {
	if gpus == nil {
		return []apis.GPUResource{}
	}
	return gpus
}

func (r *WorkerRepo) ListResources(ctx context.Context) ([]apis.WorkerResource, error) {
	workers, err := r.ListWorkers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]apis.WorkerResource, 0, len(workers))
	for i := range workers {
		res, err := r.AsResource(ctx, &workers[i])
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, nil
}

func (r *WorkerRepo) replaceGPUs(ctx context.Context, workerPK uuid.UUID, workerID string, gpus []apis.GPUResource) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM worker_gpus WHERE worker_id = $1`, workerID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, gpu := range gpus {
		_, err = r.pool.Exec(ctx, `
			INSERT INTO worker_gpus (id, worker_pk, worker_id, gpu_index, vendor, name,
				memory_total_mb, memory_used_mb, utilization, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			uuid.New(), workerPK, workerID, gpu.Index, gpu.Vendor, gpu.Name,
			gpu.MemoryTotalMB, gpu.MemoryUsedMB, gpu.Utilization, now, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *WorkerRepo) replaceModels(ctx context.Context, workerID string, models []apis.LoadedModel) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM serving_models WHERE worker_id = $1`, workerID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, m := range models {
		status := m.Status
		if status == "" {
			status = runtime.ModelStatusReady
		}
		_, err = r.pool.Exec(ctx, `
			INSERT INTO serving_models (id, worker_id, name, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), workerID, m.Name, status, now, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *WorkerRepo) CountWorkers(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM workers`).Scan(&n)
	return n, err
}

func (r *WorkerRepo) CountGPUs(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM worker_gpus`).Scan(&n)
	return n, err
}

func (r *WorkerRepo) CountOnlineNodes(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM serving_nodes WHERE status IN ('ONLINE','BUSY')`).Scan(&n)
	return n, err
}

func scanWorkers(rows pgx.Rows) ([]WorkerRow, error) {
	var out []WorkerRow
	for rows.Next() {
		var w WorkerRow
		if err := rows.Scan(&w.ID, &w.WorkerID, &w.Hostname, &w.Site, &w.Version, &w.Status,
			&w.CPUTotal, &w.MemoryTotalMB, &w.RunningRequests, &w.LastHeartbeatAt,
			&w.ModelsJSON, &w.MetadataJSON, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// Reconciler adapter
func (r *WorkerRepo) ListWorkerRows(ctx context.Context) ([]struct {
	WorkerID string
	Status   string
}, error) {
	workers, err := r.ListWorkers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]struct {
		WorkerID string
		Status   string
	}, len(workers))
	for i, w := range workers {
		out[i].WorkerID = w.WorkerID
		out[i].Status = w.Status
	}
	return out, nil
}
