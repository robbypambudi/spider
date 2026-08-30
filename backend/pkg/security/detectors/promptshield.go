package detectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spider/spider/pkg/apis"
)

// PromptShieldDetector calls a local inference sidecar that loads models from
// https://huggingface.co/collections/robbypambudi/prompt-shield
type PromptShieldDetector struct {
	Endpoint   string
	Model      string
	HTTPClient *http.Client
}

type detectRequest struct {
	Text  string `json:"text"`
	Model string `json:"model,omitempty"`
}

type detectResponse struct {
	Score       float64 `json:"score"`
	IsInjection bool    `json:"is_injection"`
	LatencyMs   float64 `json:"latency_ms"`
	Model       string  `json:"model"`
	Error       string  `json:"error,omitempty"`
}

func NewPromptShieldDetector(endpoint, model string) *PromptShieldDetector {
	return &PromptShieldDetector{
		Endpoint: endpoint,
		Model:    model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (d *PromptShieldDetector) Name() string { return "prompt-shield" }

func (d *PromptShieldDetector) Version() string { return d.Model }

func (d *PromptShieldDetector) Detect(ctx context.Context, text string) (apis.DetectionResult, error) {
	start := time.Now()
	body, _ := json.Marshal(detectRequest{Text: text, Model: d.Model})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.Endpoint+"/detect", bytes.NewReader(body))
	if err != nil {
		return apis.DetectionResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return apis.DetectionResult{}, fmt.Errorf("prompt-shield unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return apis.DetectionResult{}, fmt.Errorf("prompt-shield status %d: %s", resp.StatusCode, string(raw))
	}

	var result detectResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return apis.DetectionResult{}, err
	}
	if result.Error != "" {
		return apis.DetectionResult{}, fmt.Errorf("prompt-shield: %s", result.Error)
	}

	latency := result.LatencyMs
	if latency == 0 {
		latency = float64(time.Since(start).Microseconds()) / 1000.0
	}

	return apis.DetectionResult{
		Detector:    d.Name(),
		Score:       result.Score,
		IsInjection: result.IsInjection,
		LatencyMs:   latency,
		Metadata: map[string]interface{}{
			"version": d.Model,
			"kind":    "prompt-shield",
			"source":  "huggingface",
			"repo":    d.Model,
		},
	}, nil
}
