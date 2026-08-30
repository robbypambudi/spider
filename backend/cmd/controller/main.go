package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spider/spider/dao"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/reconciler"
)

type workerStoreAdapter struct {
	repo *store.WorkerRepo
}

func (a *workerStoreAdapter) ListWorkers(ctx context.Context) ([]reconciler.WorkerRow, error) {
	rows, err := a.repo.ListWorkers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]reconciler.WorkerRow, len(rows))
	for i, w := range rows {
		out[i] = reconciler.WorkerRow{WorkerID: w.WorkerID, Status: w.Status}
	}
	return out, nil
}

func (a *workerStoreAdapter) LastHeartbeatAt(ctx context.Context, workerID string) (*time.Time, error) {
	return a.repo.LastHeartbeatAt(ctx, workerID)
}

func (a *workerStoreAdapter) MarkOffline(ctx context.Context, workerID string) error {
	return a.repo.MarkOffline(ctx, workerID)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	settings, err := config.Load()
	if err != nil {
		os.Exit(1)
	}
	ctx := context.Background()
	db, err := dao.Connect(ctx, settings)
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()
	_ = db.Migrate(settings)

	workers := store.NewWorkerRepo(db.Pool)
	adapter := &workerStoreAdapter{repo: workers}
	wr := &reconciler.WorkerReconciler{
		Store:          adapter,
		OfflineTimeout: time.Duration(settings.WorkerOfflineTimeout) * time.Second,
	}
	snr := &reconciler.ServingNodeReconciler{}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if marked, err := wr.Reconcile(ctx); err != nil {
				slog.Error("worker_reconcile", "error", err)
			} else if len(marked) > 0 {
				slog.Info("workers_marked_offline", "count", len(marked), "workers", marked)
			}
			if err := snr.Reconcile(ctx); err != nil {
				slog.Error("serving_reconcile", "error", err)
			}
		}
	}
}
