package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/protocol"
	"github.com/spider/spider/pkg/runtime"
	"github.com/spider/spider/pkg/workerctl"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	settings, err := config.Load()
	if err != nil {
		os.Exit(1)
	}

	identity, err := workerctl.LoadIdentity(nil, "0.1.0")
	if err != nil {
		slog.Error("identity", "error", err)
		os.Exit(1)
	}

	resources := workerctl.DiscoverResources()
	gpus := make([]apis.GPUResource, 0, len(resources.GPUs))
	for _, g := range resources.GPUs {
		gpus = append(gpus, apis.GPUResource{
			Index: g.Index, Vendor: g.Vendor, Name: g.Name,
			MemoryTotalMB: g.MemoryTotalMB, MemoryUsedMB: g.MemoryUsedMB, Utilization: g.Utilization,
		})
	}

	client := &http.Client{Timeout: 15 * time.Second}
	base := settings.APIBaseURL

	reg := apis.WorkerResource{
		WorkerID: identity.WorkerID, Hostname: identity.Hostname, Site: identity.Site,
		Version: identity.Version, Status: runtime.WorkerStatusOnline,
		Resources: apis.WorkerResources{
			CPUTotal: resources.CPUTotal, MemoryTotalMB: resources.MemoryTotalMB, GPUs: gpus,
		},
		Models: []apis.LoadedModel{{Name: settings.DefaultModel, Status: runtime.ModelStatusReady}},
		Metadata: map[string]interface{}{"platform": workerctl.PlatformSummary()},
	}

	if err := postJSON(client, base+protocol.RegisterPath, settings.WorkerToken, reg, nil); err != nil {
		slog.Error("register", "error", err)
		os.Exit(1)
	}
	slog.Info("worker_registered", "worker_id", identity.WorkerID)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(time.Duration(settings.WorkerHeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			hb := apis.WorkerHeartbeat{
				WorkerID: identity.WorkerID, Status: runtime.WorkerStatusOnline,
				Resources: &reg.Resources, Models: reg.Models, RunningRequests: 0,
				Metadata: map[string]interface{}{"heartbeat": time.Now().UTC().Format(time.RFC3339)},
			}
			path := fmt.Sprintf("%s/api/v1/workers/%s/heartbeat", base, identity.WorkerID)
			if err := postJSON(client, path, settings.WorkerToken, hb, nil); err != nil {
				slog.Error("heartbeat", "error", err)
			}
		}
	}
}

func postJSON(client *http.Client, url, token string, body, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(data))
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
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
