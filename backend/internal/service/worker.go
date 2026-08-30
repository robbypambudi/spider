package service

import (
	"context"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/spidererrors"
	"github.com/spider/spider/pkg/apis"
)

type WorkerService struct {
	Workers *store.WorkerRepo
	Token   string
}

func (s *WorkerService) AuthenticateToken(token string) error {
	if token != s.Token {
		return spidererrors.WorkerAuth("Invalid worker token")
	}
	return nil
}

func (s *WorkerService) Register(ctx context.Context, resource apis.WorkerResource) (apis.WorkerResource, error) {
	_, err := s.Workers.UpsertRegistration(ctx, resource)
	if err != nil {
		return apis.WorkerResource{}, err
	}
	w, err := s.Workers.GetByWorkerID(ctx, resource.WorkerID)
	if err != nil {
		return apis.WorkerResource{}, err
	}
	return s.Workers.AsResource(ctx, w)
}

func (s *WorkerService) Heartbeat(ctx context.Context, hb apis.WorkerHeartbeat) (apis.WorkerResource, error) {
	w, err := s.Workers.RecordHeartbeat(ctx, hb)
	if err != nil || w == nil {
		return apis.WorkerResource{}, spidererrors.NotFound("Worker not found")
	}
	return s.Workers.AsResource(ctx, w)
}

func (s *WorkerService) ListWorkers(ctx context.Context) ([]apis.WorkerResource, error) {
	return s.Workers.ListResources(ctx)
}

func (s *WorkerService) Inspect(ctx context.Context, workerID string) (apis.WorkerResource, error) {
	w, err := s.Workers.GetByWorkerID(ctx, workerID)
	if err != nil {
		return apis.WorkerResource{}, spidererrors.NotFound("Worker not found")
	}
	return s.Workers.AsResource(ctx, w)
}
