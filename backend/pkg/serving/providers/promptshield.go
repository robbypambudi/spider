package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spider/spider/pkg/apis"
)

type PromptShieldProvider struct {
	Endpoint   string
	Model      string
	HTTPClient *http.Client
}

func NewPromptShieldProvider(endpoint, model string) *PromptShieldProvider {
	return &PromptShieldProvider{
		Endpoint: endpoint,
		Model:    model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *PromptShieldProvider) Name() string { return "prompt-shield" }

func (p *PromptShieldProvider) SetModel(model string) {
	p.Model = model
}

func (p *PromptShieldProvider) Infer(ctx context.Context, request apis.InferenceRequest) (apis.InferenceResponse, error) {
	model := request.Model
	if model == "" {
		model = p.Model
	}
	body, _ := json.Marshal(map[string]string{"text": request.Prompt, "model": model})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint+"/detect", bytes.NewReader(body))
	if err != nil {
		return apis.InferenceResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return apis.InferenceResponse{}, fmt.Errorf("prompt-shield unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return apis.InferenceResponse{}, fmt.Errorf("prompt-shield status %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Score       float64 `json:"score"`
		IsInjection bool    `json:"is_injection"`
		LatencyMs   float64 `json:"latency_ms"`
		Model       string  `json:"model"`
		Error       string  `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return apis.InferenceResponse{}, err
	}
	if result.Error != "" {
		return apis.InferenceResponse{}, fmt.Errorf("prompt-shield: %s", result.Error)
	}

	label := "benign"
	if result.IsInjection {
		label = "injection"
	}
	output := fmt.Sprintf(
		"Prompt-Shield classification [%s]\nscore=%.4f label=%s latency_ms=%.1f",
		result.Model, result.Score, label, result.LatencyMs,
	)
	finish := "stop"
	return apis.InferenceResponse{
		RequestID:    uuid.New().String(),
		Model:        result.Model,
		Output:       &output,
		FinishReason: &finish,
		Usage:        map[string]int{"total_tokens": len(request.Prompt) / 4},
		Metadata: map[string]interface{}{
			"provider":       "prompt-shield",
			"score":          result.Score,
			"is_injection":   result.IsInjection,
			"latency_ms":     result.LatencyMs,
			"collection_url": "https://huggingface.co/collections/robbypambudi/prompt-shield",
		},
	}, nil
}
