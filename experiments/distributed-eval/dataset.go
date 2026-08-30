package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Sample is one labeled row: text plus ground truth (is this an injection attempt?).
type Sample struct {
	Text        string `json:"text"`
	IsInjection bool   `json:"is_injection"`
}

// promptShieldRow matches the {"prompt": "...", "label": 0|1} format used by
// the real Prompt-Shield fine-tuning dataset (spider-internal/labs/PromptShield/
// {train,test,validation}.json — a single JSON array, label 1 = injection).
type promptShieldRow struct {
	Prompt string `json:"prompt"`
	Label  int    `json:"label"`
}

// loadDataset accepts two formats, auto-detected from the first non-blank byte:
//
//   - JSONL, one {"text": "...", "is_injection": true|false} object per line
//     (this tool's own format — see `gendata`).
//   - a single JSON array of {"prompt": "...", "label": 0|1} objects (the
//     real Prompt-Shield dataset format). label 1 maps to is_injection=true.
func loadDataset(path string) ([]Sample, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}

	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return loadPromptShieldArray(trimmed)
	}
	return loadJSONL(data)
}

func loadPromptShieldArray(data []byte) ([]Sample, error) {
	var rows []promptShieldRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("parse Prompt-Shield JSON array: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("dataset has no samples")
	}
	samples := make([]Sample, len(rows))
	for i, r := range rows {
		samples[i] = Sample{Text: r.Prompt, IsInjection: r.Label == 1}
	}
	return samples, nil
}

func loadJSONL(data []byte) ([]Sample, error) {
	var samples []Sample
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s Sample
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			return nil, fmt.Errorf("parse line: %w", err)
		}
		samples = append(samples, s)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read dataset: %w", err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("dataset has no samples")
	}
	return samples, nil
}
