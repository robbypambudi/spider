package reconciler

import (
	"context"
	"time"

	"github.com/spider/spider/pkg/runtime"
)

type WorkerRow struct {
	WorkerID string
	Status   string
}

type WorkerStore interface {
	ListWorkers(ctx context.Context) ([]WorkerRow, error)
	LastHeartbeatAt(ctx context.Context, workerID string) (*time.Time, error)
	MarkOffline(ctx context.Context, workerID string) error
}

type WorkerReconciler struct {
	Store            WorkerStore
	OfflineTimeout   time.Duration
	Now              func() time.Time
}

func (r *WorkerReconciler) Reconcile(ctx context.Context) ([]string, error) {
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	workers, err := r.Store.ListWorkers(ctx)
	if err != nil {
		return nil, err
	}
	var marked []string
	for _, w := range workers {
		if w.Status == runtime.WorkerStatusOffline || w.Status == runtime.WorkerStatusError {
			continue
		}
		last, err := r.Store.LastHeartbeatAt(ctx, w.WorkerID)
		if err != nil {
			return marked, err
		}
		if last == nil || now().Sub(*last) > r.OfflineTimeout {
			if err := r.Store.MarkOffline(ctx, w.WorkerID); err != nil {
				return marked, err
			}
			marked = append(marked, w.WorkerID)
		}
	}
	return marked, nil
}

type ServingNodeReconciler struct{}

func (r *ServingNodeReconciler) Reconcile(ctx context.Context) error {
	_ = ctx
	return nil
}
