package main

import (
	"fmt"
	"strings"

	"github.com/spider/spider/pkg/config"
)

// Defaults mirror spider-internal/labs/eval_chunked.py (wave-1 token chunking).
const (
	defaultDetector             = "prompt-shield"
	defaultChunker              = "token"
	defaultChunkSize            = 256
	defaultChunkOverlap         = 0
	defaultPromptShieldEndpoint = "http://localhost:8081"
	defaultPromptShieldModel    = config.PromptShieldSmall
	defaultTargetFPR            = "0.0005,0.001,0.005,0.01"
)

func validateLabConfig(detector, chunker string) error {
	if detector != "prompt-shield" {
		return fmt.Errorf(
			"detector %q is not supported in this harness; use %q (matches spider-internal/labs Prompt-Shield eval)",
			detector, defaultDetector,
		)
	}
	if chunker != "token" {
		return fmt.Errorf(
			"chunker %q is not supported; use %q with chunk-size=256 overlap=0 (matches spider-internal/labs eval_chunked.py)",
			chunker, defaultChunker,
		)
	}
	return nil
}

// extractPrompt matches BenchmarkDataset.extract_prompt in spider-internal/labs.
func extractPrompt(instruction, input string) string {
	instruction = strings.TrimSpace(instruction)
	userInput := strings.TrimSpace(input)
	if userInput != "" {
		return instruction + "\n" + userInput
	}
	return instruction
}
