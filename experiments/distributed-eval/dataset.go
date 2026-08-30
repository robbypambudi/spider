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

// promptShieldRow matches {"prompt": "...", "label": 0|1} in spider-internal/labs/PromptShield/*.json
type promptShieldRow struct {
	Prompt string `json:"prompt"`
	Label  int    `json:"label"`
}

// benchmarkRow matches camera-ready eval JSON (instruction/input/flag) used in
// spider-internal/labs/PromptShieldCode and eval_chunked.py.
type benchmarkRow struct {
	Instruction string `json:"instruction"`
	Input       string `json:"input"`
	Flag        int    `json:"flag"`
}

// loadDataset accepts formats auto-detected from the first JSON object:
//
//   - Camera-ready benchmark array: {"instruction","input","flag",...} — text via extractPrompt (lab)
//   - Prompt-Shield fine-tune array: {"prompt","label"} — label 1 = injection
//   - JSONL: {"text","is_injection"} per line (load/smoke only, not for paper metrics)
func loadDataset(path string) ([]Sample, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}

	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return loadJSONArray(trimmed)
	}
	return loadJSONL(data)
}

func loadJSONArray(data []byte) ([]Sample, error) {
	var peek []json.RawMessage
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, fmt.Errorf("parse JSON array: %w", err)
	}
	if len(peek) == 0 {
		return nil, fmt.Errorf("dataset has no samples")
	}

	var meta map[string]json.RawMessage
	if err := json.Unmarshal(peek[0], &meta); err != nil {
		return nil, fmt.Errorf("parse first row: %w", err)
	}
	if _, ok := meta["instruction"]; ok {
		return loadBenchmarkArray(data)
	}
	return loadPromptShieldArray(data)
}

func loadBenchmarkArray(data []byte) ([]Sample, error) {
	var rows []benchmarkRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("parse camera-ready benchmark JSON: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("dataset has no samples")
	}
	samples := make([]Sample, len(rows))
	for i, r := range rows {
		samples[i] = Sample{
			Text:        extractPrompt(r.Instruction, r.Input),
			IsInjection: r.Flag == 1,
		}
	}
	return samples, nil
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
