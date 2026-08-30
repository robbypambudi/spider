package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/spider/spider/pkg/apis"
)

type LLMProvider interface {
	Name() string
	Infer(ctx context.Context, request apis.InferenceRequest) (apis.InferenceResponse, error)
}

type MockLLMProvider struct {
	mu    sync.Mutex
	Calls []apis.InferenceRequest
}

func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{Calls: make([]apis.InferenceRequest, 0)}
}

func (m *MockLLMProvider) Name() string { return "mock" }

func (m *MockLLMProvider) Infer(ctx context.Context, request apis.InferenceRequest) (apis.InferenceResponse, error) {
	_ = ctx
	m.mu.Lock()
	m.Calls = append(m.Calls, request)
	m.mu.Unlock()

	preview := request.Prompt
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}
	output := fmt.Sprintf("[mock-llm:%s] %s", request.Model, preview)
	finish := "stop"
	tokens := len(request.Prompt)/4 + 32
	return apis.InferenceResponse{
		RequestID:    uuid.New().String(),
		Model:        request.Model,
		Output:       &output,
		FinishReason: &finish,
		Usage: map[string]int{
			"prompt_tokens":     len(request.Prompt) / 4,
			"completion_tokens": 32,
			"total_tokens":      tokens,
		},
		Metadata: map[string]interface{}{"provider": "mock"},
	}, nil
}

func (m *MockLLMProvider) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Calls)
}

var Registry = map[string]func() LLMProvider{
	"mock": func() LLMProvider { return NewMockLLMProvider() },
}
