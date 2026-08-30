package spider_test

import (
	"testing"

	"github.com/spider/spider/pkg/apis"
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
