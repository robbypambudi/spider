package workerctl

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Identity struct {
	WorkerID string
	Hostname string
	Site     *string
	Version  string
}

func LoadIdentity(site *string, version string) (*Identity, error) {
	if version == "" {
		version = "0.1.0"
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	workerID := os.Getenv("SPIDER_WORKER_ID")
	if workerID == "" {
		workerID = loadOrCreateWorkerIDFile()
	}
	if workerID == "" {
		workerID = fmt.Sprintf("%s-%s", hostname, randomSuffix(4))
	}
	return &Identity{
		WorkerID: workerID,
		Hostname: hostname,
		Site:     site,
		Version:  version,
	}, nil
}

func loadOrCreateWorkerIDFile() string {
	path := filepath.Join(".", ".spider-worker-id")
	if data, err := os.ReadFile(path); err == nil {
		return strings.TrimSpace(string(data))
	}
	id := fmt.Sprintf("%s-%s", mustHostname(), randomSuffix(4))
	_ = os.WriteFile(path, []byte(id), 0600)
	return id
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "worker"
	}
	return h
}

func randomSuffix(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b)[:n*2]
}

func PlatformSummary() map[string]string {
	return map[string]string{
		"system":       runtime.GOOS,
		"release":      runtime.Version(),
		"architecture": runtime.GOARCH,
		"go":           runtime.Version(),
	}
}

func DiscoverResources() WorkerResources {
	cpu := runtime.NumCPU()
	memMB := 8192
	return WorkerResources{
		CPUTotal:      cpu,
		MemoryTotalMB: memMB,
		GPUs:          []GPUResource{},
	}
}

type GPUResource struct {
	Index         int    `json:"index"`
	Vendor        string `json:"vendor"`
	Name          string `json:"name"`
	MemoryTotalMB int    `json:"memory_total_mb"`
	MemoryUsedMB  int    `json:"memory_used_mb"`
	Utilization   int    `json:"utilization"`
}

type WorkerResources struct {
	CPUTotal      int           `json:"cpu_total"`
	MemoryTotalMB int           `json:"memory_total_mb"`
	GPUs          []GPUResource `json:"gpus"`
}

func (w WorkerResources) ToAPI() []struct {
	Index         int    `json:"index"`
	Vendor        string `json:"vendor"`
	Name          string `json:"name"`
	MemoryTotalMB int    `json:"memory_total_mb"`
	MemoryUsedMB  int    `json:"memory_used_mb"`
	Utilization   int    `json:"utilization"`
} {
	return nil
}
