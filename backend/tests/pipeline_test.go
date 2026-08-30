package spider_test

import (
	"context"
	"testing"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/aggregation"
	"github.com/spider/spider/pkg/security/chunking"
	"github.com/spider/spider/pkg/security/enforcement"
	"github.com/spider/spider/pkg/security/pipeline"
	"github.com/spider/spider/pkg/security/policies"
	"github.com/spider/spider/pkg/security/preprocessing"
	"github.com/stretchr/testify/require"
)

type panicDetector struct{}

func (p *panicDetector) Name() string    { return "panic" }
func (p *panicDetector) Version() string { return "test" }
func (p *panicDetector) Detect(_ context.Context, _ string) (apis.DetectionResult, error) {
	panic("detector failure")
}

func TestPipelinePanicReturnsErrorDecision(t *testing.T) {
	pipe := &pipeline.Pipeline{
		Preprocessor: &preprocessing.DefaultPreprocessor{},
		Chunker:      chunking.NewFixedSizeChunker(2048, 128),
		Detector:     &panicDetector{},
		Aggregator:   &aggregation.MaxScoreAggregator{},
		Policy:       policies.NewThresholdPolicy(0.5, "block"),
		Enforcer:     enforcement.NewEnforcer("closed"),
	}

	result := pipe.Inspect(context.Background(), apis.SecurityRequest{
		RequestID: "panic-test",
		Text:      "hello",
	})
	require.Equal(t, apis.DecisionError, result.Decision)
	require.Contains(t, result.Metadata["error"], "pipeline panic")
	require.Equal(t, enforcement.ActionReject, pipe.Enforcer.Resolve(result.Decision))
}
