package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLabConfigRejectsRuleBased(t *testing.T) {
	require.Error(t, validateLabConfig("rule-based", "token"))
	require.NoError(t, validateLabConfig("prompt-shield", "token"))
	require.Error(t, validateLabConfig("prompt-shield", "fixed"))
}

func TestExtractPromptMatchesLab(t *testing.T) {
	require.Equal(t, "inst\nuser", extractPrompt("inst", "user"))
	require.Equal(t, "inst only", extractPrompt("inst only", ""))
	require.Equal(t, "inst", extractPrompt("inst", "   "))
}

func TestLoadBenchmarkArrayFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bench.json")
	err := os.WriteFile(path, []byte(`[
		{"instruction":"hello","input":"world","flag":0},
		{"instruction":"attack","input":"","flag":1}
	]`), 0o644)
	require.NoError(t, err)

	samples, err := loadDataset(path)
	require.NoError(t, err)
	require.Len(t, samples, 2)
	require.Equal(t, "hello\nworld", samples[0].Text)
	require.False(t, samples[0].IsInjection)
	require.Equal(t, "attack", samples[1].Text)
	require.True(t, samples[1].IsInjection)
}
