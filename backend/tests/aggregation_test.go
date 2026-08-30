package spider_test

import (
	"testing"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/aggregation"
	"github.com/stretchr/testify/require"
)

func TestMaxScoreAggregator(t *testing.T) {
	a := &aggregation.MaxScoreAggregator{}
	res := a.Aggregate([]apis.DetectionResult{
		{Score: 0.1}, {Score: 0.9}, {Score: 0.3},
	})
	require.Equal(t, 0.9, res.Score)
	require.True(t, res.IsInjection == false) // is_injection from any - need IsInjection set
}

func TestMaxScoreAggregatorEmpty(t *testing.T) {
	a := &aggregation.MaxScoreAggregator{}
	res := a.Aggregate(nil)
	require.Equal(t, 0.0, res.Score)
	require.Equal(t, 0, res.ChunksScanned)
}
