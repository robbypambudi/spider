package spider_test

import (
	"testing"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/pipeline"
	"github.com/spider/spider/pkg/security/policies"
	"github.com/stretchr/testify/require"
)

func TestThresholdPolicyBlocksAtOrAbove(t *testing.T) {
	p := policies.NewThresholdPolicy(0.5, "block")
	decision := p.Evaluate(apis.AggregatedDetectionResult{Score: 0.5})
	require.Equal(t, apis.DecisionBlock, decision)
}

func TestThresholdAllowsBelow(t *testing.T) {
	p := policies.NewThresholdPolicy(0.5, "block")
	decision := p.Evaluate(apis.AggregatedDetectionResult{Score: 0.49})
	require.Equal(t, apis.DecisionAllow, decision)
}

func TestThresholdPolicyReview(t *testing.T) {
	p := policies.NewThresholdPolicy(0.5, "review")
	decision := p.Evaluate(apis.AggregatedDetectionResult{Score: 0.9})
	require.Equal(t, apis.DecisionReview, decision)
}

func TestParsePolicyConfig(t *testing.T) {
	raw := store.DefaultPolicyConfigJSON("token", 256, 64)
	chunker, size, overlap := store.ParsePolicyConfig(raw)
	require.Equal(t, "token", chunker)
	require.Equal(t, 256, size)
	require.Equal(t, 64, overlap)
}

func TestParsePolicyConfigDefaults(t *testing.T) {
	chunker, size, overlap := store.ParsePolicyConfig("")
	require.Equal(t, "token", chunker)
	require.Equal(t, 256, size)
	require.Equal(t, 0, overlap)
}

func TestPipelineNewChunkerNames(t *testing.T) {
	token := pipeline.NewChunker("token", "http://127.0.0.1:8081", 256, 0)
	require.Equal(t, "token", token.Name())
	fixed := pipeline.NewChunker("fixed", "", 2048, 128)
	require.Equal(t, "fixed", fixed.Name())
}
