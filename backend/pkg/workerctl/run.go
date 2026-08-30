package workerctl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/protocol"
	"github.com/spider/spider/pkg/runtime"
)

// RunOptions configures a worker's registration + heartbeat loop.
type RunOptions struct {
	Settings *config.Settings
	// WorkerID overrides identity resolution (env var / .spider-worker-id file).
	WorkerID string
	Site     *string
	Logger   *slog.Logger
	// Models overrides the advertised model list; defaults to a single
	// entry for Settings.DefaultModel when nil.
	Models []apis.LoadedModel
	// Metadata is merged into registration/heartbeat metadata alongside
	// the always-present "platform" entry.
	Metadata map[string]interface{}
}

// Run registers the worker with the SPIDER API and heartbeats until ctx is
// cancelled. It is the shared implementation behind the standalone `worker`
// binary and `spider worker join` in the CLI.
func Run(ctx context.Context, opts RunOptions) error {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	identity, err := LoadIdentity(opts.WorkerID, opts.Site, "0.1.0")
	if err != nil {
		return fmt.Errorf("identity: %w", err)
	}

	resources := DiscoverResources()
	gpus := make([]apis.GPUResource, 0, len(resources.GPUs))
	for _, g := range resources.GPUs {
		gpus = append(gpus, apis.GPUResource{
			Index: g.Index, Vendor: g.Vendor, Name: g.Name,
			MemoryTotalMB: g.MemoryTotalMB, MemoryUsedMB: g.MemoryUsedMB, Utilization: g.Utilization,
		})
	}

	client := &http.Client{Timeout: 15 * time.Second}
	base := opts.Settings.APIBaseURL

	models := opts.Models
	if models == nil {
		models = []apis.LoadedModel{{Name: opts.Settings.DefaultModel, Status: runtime.ModelStatusReady}}
	}
	meta := map[string]interface{}{"platform": PlatformSummary()}
	for k, v := range opts.Metadata {
		meta[k] = v
	}

	reg := apis.WorkerResource{
		WorkerID: identity.WorkerID, Hostname: identity.Hostname, Site: identity.Site,
		Version: identity.Version, Status: runtime.WorkerStatusOnline,
		Resources: apis.WorkerResources{
			CPUTotal: resources.CPUTotal, MemoryTotalMB: resources.MemoryTotalMB, GPUs: gpus,
		},
		Models:   models,
		Metadata: meta,
	}

	if err := postJSON(ctx, client, base+protocol.RegisterPath, opts.Settings.WorkerToken, reg); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	logger.Info("worker_registered", "worker_id", identity.WorkerID)

	ticker := time.NewTicker(time.Duration(opts.Settings.WorkerHeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker_stopped", "worker_id", identity.WorkerID)
			return nil
		case <-ticker.C:
			hb := apis.WorkerHeartbeat{
				WorkerID: identity.WorkerID, Status: runtime.WorkerStatusOnline,
				Resources: &reg.Resources, Models: reg.Models, RunningRequests: 0,
				Metadata: map[string]interface{}{"heartbeat": time.Now().UTC().Format(time.RFC3339)},
			}
			path := fmt.Sprintf("%s/api/v1/workers/%s/heartbeat", base, identity.WorkerID)
			if err := postJSON(ctx, client, path, opts.Settings.WorkerToken, hb); err != nil {
				logger.Error("heartbeat", "error", err)
			}
		}
	}
}

func postJSON(ctx context.Context, client *http.Client, url, token string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(protocol.WorkerTokenHeader, token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
