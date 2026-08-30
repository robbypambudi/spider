package scheduler

import (
	"context"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/runtime"
)

type Scheduler interface {
	Name() string
	SelectWorker(ctx context.Context, model string, site *string, workers []apis.WorkerResource) *apis.WorkerResource
}

type LeastLoadedScheduler struct{}

func (s *LeastLoadedScheduler) Name() string { return "least-loaded" }

func (s *LeastLoadedScheduler) SelectWorker(ctx context.Context, model string, site *string, workers []apis.WorkerResource) *apis.WorkerResource {
	_ = ctx
	_ = model
	var best *apis.WorkerResource
	bestKey := []int{1, 1<<30, 1<<30, 1<<30}

	for i := range workers {
		w := &workers[i]
		if w.Status != runtime.WorkerStatusOnline && w.Status != runtime.WorkerStatusBusy {
			continue
		}
		if site != nil && w.Site != nil && *site != *w.Site {
			continue
		}
		locality := 0
		if site != nil && w.Site != nil && *site != *w.Site {
			locality = 1
		}
		gpuUtil := 0
		vramUsed := 0
		for _, gpu := range w.Resources.GPUs {
			if gpu.Utilization > gpuUtil {
				gpuUtil = gpu.Utilization
			}
			if gpu.MemoryUsedMB > vramUsed {
				vramUsed = gpu.MemoryUsedMB
			}
		}
		key := []int{locality, w.RunningRequests, gpuUtil, vramUsed}
		if best == nil || lessKey(key, bestKey) {
			best = w
			bestKey = key
		}
	}
	return best
}

func lessKey(a, b []int) bool {
	for i := range a {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}
