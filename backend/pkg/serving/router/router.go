package router

import (
	"context"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/scheduler"
	"github.com/spider/spider/pkg/serving/providers"
)

type ServingRouter struct {
	Provider  providers.LLMProvider
	Scheduler scheduler.Scheduler
}

func New(provider providers.LLMProvider, sched scheduler.Scheduler) *ServingRouter {
	if sched == nil {
		sched = &scheduler.LeastLoadedScheduler{}
	}
	return &ServingRouter{Provider: provider, Scheduler: sched}
}

func (r *ServingRouter) Route(ctx context.Context, request apis.InferenceRequest, workers []apis.WorkerResource) (apis.InferenceResponse, *apis.WorkerResource, error) {
	var selected *apis.WorkerResource
	if len(workers) > 0 {
		selected = r.Scheduler.SelectWorker(ctx, request.Model, nil, workers)
	}
	response, err := r.Provider.Infer(ctx, request)
	if err != nil {
		return apis.InferenceResponse{}, selected, err
	}
	return response, selected, nil
}
