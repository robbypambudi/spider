package aggregation

import "github.com/spider/spider/pkg/apis"

type Aggregator interface {
	Name() string
	Aggregate(results []apis.DetectionResult) apis.AggregatedDetectionResult
}

type MaxScoreAggregator struct{}

func (a *MaxScoreAggregator) Name() string { return "max-score" }

func (a *MaxScoreAggregator) Aggregate(results []apis.DetectionResult) apis.AggregatedDetectionResult {
	if len(results) == 0 {
		return apis.AggregatedDetectionResult{
			Score:         0,
			IsInjection:   false,
			ChunksScanned: 0,
		}
	}
	maxScore := 0.0
	isInjection := false
	for _, r := range results {
		if r.Score > maxScore {
			maxScore = r.Score
		}
		if r.IsInjection {
			isInjection = true
		}
	}
	return apis.AggregatedDetectionResult{
		Score:           maxScore,
		IsInjection:     isInjection,
		ChunksScanned:   len(results),
		DetectorResults: results,
	}
}
